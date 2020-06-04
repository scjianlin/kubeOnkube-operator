package docker

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/gostship/kunkka/pkg/util/supervisor"
	"github.com/gostship/kunkka/pkg/util/template"
)

type Option struct {
	InsecureRegistries string
	RegistryDomain     string
	Options            string
	IsGPU              bool
	ExtraArgs          map[string]string
}

const (
	dockerDaemonFile = "/etc/docker/daemon.json"
)

func Install(s ssh.Interface, option *Option) error {
	// dstFile, err := res.Docker.CopyToNodeWithDefault(s)
	// if err != nil {
	// 	return err
	// }

	// cmd := "tar xvaf %s -C %s --strip-components=1"
	// _, stderr, exit, err := s.Execf(cmd, dstFile, constants.DstBinDir)
	// if err != nil {
	// 	return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", cmd, exit, stderr, err)
	// }

	var args []string
	for k, v := range option.ExtraArgs {
		args = append(args, fmt.Sprintf(`--%s="%s"`, k, v))
	}
	err := s.WriteFile(strings.NewReader(fmt.Sprintf("DOCKER_EXTRA_ARGS=%s", strings.Join(args, " "))), "/etc/sysconfig/docker")
	if err != nil {
		return err
	}

	data, err := template.ParseFile(path.Join(constants.ConfDir, "docker/daemon.json"), option)
	if err != nil {
		return err
	}
	err = s.WriteFile(bytes.NewReader(data), dockerDaemonFile)
	if err != nil {
		return errors.Wrapf(err, "write %s error", dockerDaemonFile)
	}

	data, err = template.ParseFile(path.Join(constants.ConfDir, "docker/docker.service"), option)
	if err != nil {
		return err
	}
	ss := &supervisor.SystemdSupervisor{Name: "docker", SSH: s}
	err = ss.Deploy(bytes.NewReader(data))
	if err != nil {
		return err
	}

	err = ss.Start()
	if err != nil {
		return err
	}
	return nil
}
