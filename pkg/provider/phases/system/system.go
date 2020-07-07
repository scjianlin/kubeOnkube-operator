package system

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/gostship/kunkka/pkg/util/template"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

type Option struct {
	InsecureRegistries string
	RegistryDomain     string
	Options            string
	K8sVersion         string
	DockerVersion      string
	Cgroupdriver       string
	HostIP             string
	KernelRepo         string
	ResolvConf         string
	CentosVersion      string
	ExtraArgs          map[string]string
}

func Install(s ssh.Interface, c *common.Cluster) error {
	dockerVersion := "19.03.9"
	if v, ok := c.Spec.DockerExtraArgs["version"]; ok {
		dockerVersion = v
	}
	option := &Option{
		K8sVersion:    c.Spec.Version,
		DockerVersion: dockerVersion,
		Cgroupdriver:  "systemd", // cgroupfs or systemd
		ExtraArgs:     c.Spec.KubeletExtraArgs,
		HostIP:        s.HostIP(),
		KernelRepo:    "yum-mirrors.dmall.com",
	}

	initData, err := template.ParseString(initShellTemplate, option)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(initData), constants.SystemInitFile)
	if err != nil {
		return err
	}

	klog.Infof("node: %s start exec init system ... ", option.HostIP)
	cmd := fmt.Sprintf("chmod a+x %s && %s", constants.SystemInitFile, constants.SystemInitFile)
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("%q %+v", exit, err)
		return errors.Wrapf(err, "node: %s exec init", option.HostIP)
	}

	klog.Infof("node: %s exec init system success", option.HostIP)
	result, err := s.CombinedOutput("uname -r")
	if err != nil {
		klog.Errorf("err: %+v", err)
		return err
	}
	versionStr := strings.TrimSpace(string(result))
	versions := strings.Split(strings.TrimSpace(string(result)), ".")
	if len(versions) < 2 {
		return errors.Errorf("parse version error:%s", versionStr)
	}
	kernelVersion, err := strconv.Atoi(versions[0])
	if err != nil {
		return errors.Wrapf(err, "parse kernelVersion")
	}

	if kernelVersion >= 4 {
		return nil
	}

	klog.Infof("node: %s now kernel: %s,  start reboot ... ", option.HostIP, string(result))
	_, _ = s.CombinedOutput("reboot")
	return nil
}
