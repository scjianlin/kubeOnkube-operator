package clean

import (
	"fmt"
	"os"

	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

func CleanNode(s ssh.Interface) error {
	cmd := "kubeadm reset -f && rm -rf /var/lib/etcd /var/lib/kubelet /var/lib/dockershim /var/run/kubernetes /var/lib/cni /etc/kubernetes /etc/cni /root/.kube && ipvsadm --clear"
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("cmd: %s exit: %q err: %+v", cmd, exit, err)
		return errors.Wrapf(err, "node: %s exec: \n%s", s.HostIP(), cmd)
	}
	return nil
}

func DleNode(s ssh.Interface, name string) error {
	cmd := fmt.Sprintf("kubectl delete node %s", name) //kubectl delete node 10.248.224.171
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Error("cmd: %s exit: %q err: %+v", cmd, exit, err)
		return errors.Wrapf(err, "node delete exec: \n%s", s.HostIP(), cmd)
	}
	return nil
}
