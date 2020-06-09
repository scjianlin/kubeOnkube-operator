package system

import (
	"bytes"

	"strconv"
	"strings"

	"github.com/gostship/kunkka/pkg/constants"
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
	ExtraArgs          map[string]string
}

func Install(s ssh.Interface, option *Option) error {
	initData, err := template.ParseString(initShellTemplate, option)
	if err != nil {
		return err
	}

	klog.Infof("write init.sh to node: %s", option.HostIP)
	err = s.WriteFile(bytes.NewReader(initData), constants.SystemInitFile)
	if err != nil {
		return err
	}

	execf, stderr, exit, err := s.Execf("chmod a+x %s && %s", constants.SystemInitFile, constants.SystemInitFile)
	if err != nil {
		klog.Errorf("%q %q %q", execf, stderr, exit)
		return err
	}

	klog.Infof("init node: %s system success, info execf:\n %s \n info stderr: \n %s", option.HostIP, execf, stderr)
	result, err := s.CombinedOutput("uname -r")
	if err != nil {
		klog.Errorf("err: %q", err)
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
