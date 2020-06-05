package system

import (
	"bytes"

	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/gostship/kunkka/pkg/util/template"
)

type Option struct {
	InsecureRegistries string
	RegistryDomain     string
	Options            string
	K8sVersion         string
	DockerVersion      string
	Cgroupdriver       string
	HostName           string
	HostIP             string
	ExtraArgs          map[string]string
}

func Install(s ssh.Interface, option *Option) error {

	// var args []string
	// for k, v := range option.ExtraArgs {
	// 	args = append(args, fmt.Sprintf(`--%s="%s"`, k, v))
	// }
	// err := s.WriteFile(strings.NewReader(fmt.Sprintf("KUBELET_EXTRA_ARGS=%s", strings.Join(args, " "))), "/etc/sysconfig/kubelet")
	// if err != nil {
	// 	return err
	// }

	initData, err := template.ParseString(initShellTemplate, option)
	if err != nil {
		return err
	}

	err = s.WriteFile(bytes.NewReader(initData), "/opt/k8s/init.sh")
	if err != nil {
		return err
	}
	return nil
}
