package machine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/phases/k8sComponent"
	"github.com/gostship/kunkka/pkg/provider/phases/kubeadm"
	"github.com/gostship/kunkka/pkg/provider/phases/kubeconfig"
	"github.com/gostship/kunkka/pkg/provider/phases/system"
	"github.com/gostship/kunkka/pkg/provider/preflight"
	"github.com/gostship/kunkka/pkg/util/apiclient"
	"github.com/gostship/kunkka/pkg/util/hosts"
	"github.com/pkg/errors"
)

func (p *Provider) EnsureCopyFiles(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	for _, file := range c.Spec.Features.Files {
		err = system.CopyFile(machineSSH, &file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
	hook := cluster.Spec.Features.Hooks[devopsv1.HookPreInstall]
	if hook == "" {
		return nil
	}

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	cmd := strings.Split(hook, " ")[0]

	machineSSH.Execf("chmod +x %s", cmd)
	_, stderr, exit, err := machineSSH.Exec(hook)
	if err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
	}
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
	hook := cluster.Spec.Features.Hooks[devopsv1.HookPostInstall]
	if hook == "" {
		return nil
	}

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	cmd := strings.Split(hook, " ")[0]

	machineSSH.Execf("chmod +x %s", cmd)
	_, stderr, exit, err := machineSSH.Exec(hook)
	if err != nil || exit != 0 {
		return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
	}
	return nil
}

func (p *Provider) EnsureClean(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	_, err = machineSSH.CombinedOutput(fmt.Sprintf("rm -rf %s", constants.KubernetesDir))
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsurePreflight(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = preflight.RunNodeChecks(machineSSH)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsureRegistryHosts(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
	if !p.Cfg.NeedSetHosts() {
		return nil
	}

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	domains := []string{
		p.Cfg.Registry.Domain,
		machine.Spec.TenantID + "." + p.Cfg.Registry.Domain,
	}
	for _, one := range domains {
		remoteHosts := hosts.RemoteHosts{Host: one, SSH: machineSSH}
		err := remoteHosts.Set(p.Cfg.Registry.IP)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureSystem(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = system.Install(sh, c)
	if err != nil {
		return errors.Wrap(err, sh.HostIP())
	}

	return nil
}

func (p *Provider) EnsureK8sComponent(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = k8sComponent.Install(sh, c)
	if err != nil {
		return errors.Wrap(err, sh.HostIP())
	}

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	masterEndpoint, err := GetMasterEndpoint(c.Cluster.Status.Addresses)
	if err != nil {
		return err
	}

	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	option := &kubeconfig.Option{
		MasterEndpoint: masterEndpoint,
		ClusterName:    c.Name,
		CACert:         c.ClusterCredential.CACert,
		Token:          *c.ClusterCredential.Token,
	}
	err = kubeconfig.Install(machineSSH, option)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) EnsureJoinNode(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	host, err := c.Host()
	if err != nil {
		return err
	}
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	option := &kubeadm.JoinNodeOption{
		NodeName:             machine.Spec.Machine.IP,
		BootstrapToken:       *c.ClusterCredential.BootstrapToken,
		ControlPlaneEndpoint: host,
	}
	err = kubeadm.JoinNode(machineSSH, option)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) EnsureMarkNode(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	clientset, err := c.Clientset()
	if err != nil {
		return err
	}

	err = apiclient.MarkNode(ctx, clientset, machine.Spec.Machine.IP, machine.Spec.Machine.Labels, machine.Spec.Machine.Taints)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) EnsureNodeReady(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	clientset, err := c.Clientset()
	if err != nil {
		return err
	}

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		node, err := clientset.CoreV1().Nodes().Get(ctx, machine.Spec.Machine.IP, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, one := range node.Status.Conditions {
			if one.Type == corev1.NodeReady && one.Status == corev1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	})
}

func GetMasterEndpoint(addresses []devopsv1.ClusterAddress) (string, error) {
	var advertise, internal []*devopsv1.ClusterAddress
	for _, one := range addresses {
		if one.Type == devopsv1.AddressAdvertise {
			advertise = append(advertise, &one)
		}
		if one.Type == devopsv1.AddressReal {
			internal = append(internal, &one)
		}
	}

	var address *devopsv1.ClusterAddress
	if advertise != nil {
		address = advertise[rand.Intn(len(advertise))]
	} else {
		if internal != nil {
			address = internal[rand.Intn(len(internal))]
		}
	}
	if address == nil {
		return "", errors.New("no advertise or internal address for the cluster")
	}

	return fmt.Sprintf("https://%s:%d", address.Host, address.Port), nil
}
