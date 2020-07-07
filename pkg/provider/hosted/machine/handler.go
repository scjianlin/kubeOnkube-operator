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

	"bytes"

	"os"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	kubeadmv1beta2 "github.com/gostship/kunkka/pkg/apis/kubeadm/v1beta2"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/certs"
	"github.com/gostship/kunkka/pkg/provider/phases/k8sComponent"
	"github.com/gostship/kunkka/pkg/provider/phases/kubeconfig"
	"github.com/gostship/kunkka/pkg/provider/phases/system"
	"github.com/gostship/kunkka/pkg/provider/preflight"
	"github.com/gostship/kunkka/pkg/util/apiclient"
	"github.com/gostship/kunkka/pkg/util/hosts"
	"github.com/gostship/kunkka/pkg/util/pkiutil"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

func (p *Provider) EnsureCopyFiles(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	for _, file := range cluster.Spec.Features.Files {
		err = machineSSH.CopyFile(file.Src, file.Dst)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
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

func (p *Provider) EnsureClean(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
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

func (p *Provider) EnsurePreflight(ctx context.Context, machine *devopsv1.Machine, cluster *common.Cluster) error {
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
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		err = k8sComponent.Install(machineSSH, c)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	machineSSH, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.Features.HA.ThirdPartyHA.VIP, int(c.Cluster.Spec.Features.HA.ThirdPartyHA.VPort))

	option := &kubeconfig.Option{
		MasterEndpoint: apiserver,
		ClusterName:    c.Cluster.Name,
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
	ip := machine.Spec.Machine.IP
	nodeOpt := &kubeadmv1beta2.NodeRegistrationOptions{
		Name: ip,
	}
	flagsEnv := BuildKubeletDynamicEnvFile(p.Cfg.Registry.Prefix, nodeOpt)
	sh, err := machine.Spec.SSH()
	if err != nil {
		return err
	}

	err = sh.WriteFile(strings.NewReader(flagsEnv), constants.KubeletEnvFileName)
	if err != nil {
		return err
	}

	kubeletCfg := p.getKubeletConfiguration(c)
	cfgYaml, err := KubeletMarshal(kubeletCfg)
	if err != nil {
		return err
	}

	err = sh.WriteFile(bytes.NewReader(cfgYaml), constants.KubeletConfigurationFileName)
	if err != nil {
		return err
	}

	klog.Infof("node: %s start waite ca: %s", ip, constants.CACertName)
	err = sh.WriteFile(bytes.NewReader(c.ClusterCredential.CACert), constants.CACertName)
	if err != nil {
		return err
	}

	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.Features.HA.ThirdPartyHA.VIP, int(c.Cluster.Spec.Features.HA.ThirdPartyHA.VPort))
	cfgMaps, err := certs.CreateKubeConfigFiles(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, ip, c.Cluster.Name, pkiutil.KubeletKubeConfigFileName)
	if err != nil {
		klog.Errorf("create node: %s kubelet kubeconfg err: %+v", ip, err)
		return err
	}

	klog.Infof("[%s/%s] start build node: %s kubelet kubeconfig ...", c.Cluster.Namespace, c.Cluster.Name, ip)
	for _, v := range cfgMaps {
		kubeletConf, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			klog.Errorf("covert node: %s kubelet kubeconfg err: %+v", ip, err)
			return err
		}
		err = sh.WriteFile(bytes.NewReader(kubeletConf), constants.KubeletKubeConfigFileName)
		if err != nil {
			return err
		}

		klog.Infof("node: %s write kubelet kubeconfg success", ip)
		break
	}

	err = sh.WriteFile(strings.NewReader(kubeletEnvironmentTemplate), constants.KubeletServiceRunConfig)
	if err != nil {
		return err
	}

	klog.Infof("node: %s start kubelet ... ", ip)
	cmd := fmt.Sprintf("systemctl enable kubelet && systemctl daemon-reload && systemctl restart kubelet")
	exit, err := sh.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("%q %+v", exit, err)
		return err
	}
	return nil
}

func (p *Provider) EnsureMarkNode(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}

	err = apiclient.MarkNode(ctx, clusterCtx.KubeCli, machine.Spec.Machine.IP, machine.Spec.Machine.Labels, machine.Spec.Machine.Taints)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) EnsureNodeReady(ctx context.Context, machine *devopsv1.Machine, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		node, err := clusterCtx.KubeCli.CoreV1().Nodes().Get(ctx, machine.Spec.Machine.IP, metav1.GetOptions{})
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
