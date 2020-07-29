package cluster

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/segmentio/ksuid"
	"github.com/thoas/go-funk"

	bootstraputil "k8s.io/cluster-bootstrap/token/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/provider/phases/kubeadm"
	"github.com/gostship/kunkka/pkg/provider/phases/kubemisc"
	"github.com/gostship/kunkka/pkg/provider/phases/system"
	"github.com/gostship/kunkka/pkg/provider/preflight"
	"github.com/gostship/kunkka/pkg/util/apiclient"
	"github.com/gostship/kunkka/pkg/util/hosts"

	"bytes"

	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/addons/cni"
	"github.com/gostship/kunkka/pkg/provider/addons/flannel"
	"github.com/gostship/kunkka/pkg/provider/addons/metricsserver"
	"github.com/gostship/kunkka/pkg/provider/phases/certs"

	"sync"

	"github.com/gostship/kunkka/pkg/provider/phases/component"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/pkiutil"
	"github.com/gostship/kunkka/pkg/util/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (p *Provider) EnsureCopyFiles(ctx context.Context, c *common.Cluster) error {
	for _, file := range c.Spec.Features.Files {
		for _, machine := range c.Spec.Machines {
			machineSSH, err := machine.SSH()
			if err != nil {
				return err
			}

			err = system.CopyFile(machineSSH, &file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) EnsurePreflight(ctx context.Context, c *common.Cluster) error {
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		klog.Infof("start check node: %s ... ", machine.IP)
		err = preflight.RunMasterChecks(machineSSH, c)
		if err != nil {
			klog.Errorf("node:%s check err: %+v", machine.IP, err)
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureClusterComplete(ctx context.Context, c *common.Cluster) error {
	funcs := []func(cluster *common.Cluster) error{
		completeK8sVersion,
		completeNetworking,
		completeDNS,
		completeAddresses,
		completeCredential,
	}
	for _, f := range funcs {
		if err := f(c); err != nil {
			return err
		}
	}

	c.Cluster.Status.Version = c.Spec.Version
	return nil
}

func completeK8sVersion(cluster *common.Cluster) error {
	cluster.Cluster.Status.Version = cluster.Spec.Version
	return nil
}

func completeNetworking(cluster *common.Cluster) error {
	var (
		serviceCIDR      string
		nodeCIDRMaskSize int32
		err              error
	)

	if cluster.Spec.ServiceCIDR != nil {
		serviceCIDR = *cluster.Spec.ServiceCIDR
		nodeCIDRMaskSize, err = k8sutil.GetNodeCIDRMaskSize(cluster.Spec.ClusterCIDR, *cluster.Spec.Properties.MaxNodePodNum)
		if err != nil {
			return errors.Wrap(err, "GetNodeCIDRMaskSize error")
		}
	} else {
		serviceCIDR, nodeCIDRMaskSize, err = k8sutil.GetServiceCIDRAndNodeCIDRMaskSize(cluster.Spec.ClusterCIDR, *cluster.Spec.Properties.MaxClusterServiceNum, *cluster.Spec.Properties.MaxNodePodNum)
		if err != nil {
			return errors.Wrap(err, "GetServiceCIDRAndNodeCIDRMaskSize error")
		}
	}
	cluster.Cluster.Status.ServiceCIDR = serviceCIDR
	cluster.Cluster.Status.NodeCIDRMaskSize = nodeCIDRMaskSize

	return nil
}

func completeDNS(cluster *common.Cluster) error {
	ip, err := k8sutil.GetIndexedIP(cluster.Cluster.Status.ServiceCIDR, constants.DNSIPIndex)
	if err != nil {
		return errors.Wrap(err, "get DNS IP error")
	}
	cluster.Cluster.Status.DNSIP = ip.String()

	return nil
}

func completeAddresses(cluster *common.Cluster) error {
	for _, m := range cluster.Spec.Machines {
		cluster.AddAddress(devopsv1.AddressReal, m.IP, 6443)
	}

	if cluster.Spec.Features.HA != nil {
		if cluster.Spec.Features.HA.DKEHA != nil {
			cluster.AddAddress(devopsv1.AddressAdvertise, cluster.Spec.Features.HA.DKEHA.VIP, 6443)
		}
		if cluster.Spec.Features.HA.ThirdPartyHA != nil {
			cluster.AddAddress(devopsv1.AddressAdvertise, cluster.Spec.Features.HA.ThirdPartyHA.VIP, cluster.Spec.Features.HA.ThirdPartyHA.VPort)
		}
	}

	return nil
}

func completeCredential(cluster *common.Cluster) error {
	token := ksuid.New().String()
	cluster.ClusterCredential.Token = &token

	bootstrapToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return err
	}
	cluster.ClusterCredential.BootstrapToken = &bootstrapToken

	certBytes := make([]byte, 32)
	if _, err := rand.Read(certBytes); err != nil {
		return err
	}
	certificateKey := hex.EncodeToString(certBytes)
	cluster.ClusterCredential.CertificateKey = &certificateKey

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx context.Context, c *common.Cluster) error {
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		err = kubemisc.Install(machineSSH, c)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitKubeletStartPhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg),
		fmt.Sprintf("kubelet-start --node-name=%s", c.Spec.Machines[0].IP))
}

func (p *Provider) EnsureCerts(ctx context.Context, c *common.Cluster) error {
	cfg := kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg)
	err := kubeadm.InitCerts(cfg, c, false)
	if err != nil {
		return err
	}

	for _, machine := range c.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		for pathFile, va := range c.ClusterCredential.CertsBinaryData {
			klog.Infof("node: %s start write BinaryData [%s] ...", sh.HostIP(), pathFile)
			err = sh.WriteFile(bytes.NewReader(va), pathFile)
			if err != nil {
				klog.Errorf("write [%s] err: %v", pathFile, err)
				return err
			}
		}
	}

	return nil
}

func (p *Provider) EnsureKubeMiscPhase(ctx context.Context, c *common.Cluster) error {
	sh, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	apiserver := certs.BuildApiserverEndpoint(c.Spec.Machines[0].IP, 6443)
	kubeMaps := make(map[string]string)
	err = kubemisc.ApplyKubeletKubeconfig(c, apiserver, sh.HostIP(), kubeMaps)
	if err != nil {
		return err
	}

	err = kubemisc.ApplyMasterMisc(c, apiserver)
	if err != nil {
		return err
	}

	for k, v := range c.ClusterCredential.KubeData {
		kubeMaps[k] = v
	}

	for pathName, va := range kubeMaps {
		klog.V(4).Infof("node: %s start write misc [%s] ...", sh.HostIP(), pathName)
		err = sh.WriteFile(strings.NewReader(va), pathName)
		if err != nil {
			klog.Errorf("write kubeconfg: %s err: %+v", pathName, err)
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureKubeadmInitControlPlanePhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg), "control-plane all")
}

func (p *Provider) EnsureKubeadmInitEtcdPhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg), "etcd local")
}

func (p *Provider) EnsureKubeadmInitUploadConfigPhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg), "upload-config all ")
}

func (p *Provider) EnsureKubeadmInitUploadCertsPhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg), "upload-certs --upload-certs")
}

func (p *Provider) EnsureKubeadmInitBootstrapTokenPhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg), "bootstrap-token")
}

func (p *Provider) EnsureKubeadmInitAddonPhase(ctx context.Context, c *common.Cluster) error {
	machineSSH, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}
	return kubeadm.Init(machineSSH, kubeadm.GetKubeadmConfigByMaster0(c, p.Cfg), "addon all")
}

func (p *Provider) EnsureJoinControlePlane(ctx context.Context, c *common.Cluster) error {
	for _, machine := range c.Spec.Machines[1:] {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		clientset, err := c.ClientsetForBootstrap()
		if err != nil {
			klog.Errorf("ClientsetForBootstrap err: %v", clientset)
			return err
		}

		_, err = clientset.CoreV1().Nodes().Get(context.TODO(), sh.HostIP(), metav1.GetOptions{})
		if err == nil {
			return nil
		}

		// apiserver := certs.BuildApiserverEndpoint(c.Spec.Machines[0].IP, 6443)
		// err = joinNode.JoinNodePhase(sh, p.Cfg, c, apiserver, true)
		// if err != nil {
		// 	return errors.Wrapf(err, "node: %s JoinNodePhase", sh.HostIP())
		// }

		err = kubeadm.JoinControlPlane(sh, c)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}

		if p.Cfg.CustomeImages {
			err = kubeadm.ApplyCustomMaster(sh, c, p.Cfg)
			if err != nil {
				return errors.Wrap(err, machine.IP)
			}
		}
	}

	return nil
}

func (p *Provider) EnsureComponent(ctx context.Context, c *common.Cluster) error {
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		err = component.Install(machineSSH, c)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureSystem(ctx context.Context, c *common.Cluster) error {
	wg := sync.WaitGroup{}
	quitErrors := make(chan error)
	wgDone := make(chan struct{})
	for _, mach := range c.Spec.Machines {
		sh, err := mach.SSH()
		if err != nil {
			return err
		}

		wg.Add(1)

		go func(s ssh.Interface) {
			defer wg.Done()
			err = system.Install(s, c)
			if err != nil {
				quitErrors <- errors.Wrap(err, mach.IP)
			}
		}(sh)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received through the channel
	select {
	case <-wgDone:
		break
	case err := <-quitErrors:
		close(quitErrors)
		klog.Errorf("err: %+v", err)
		return err
	}

	klog.Infof("clster: %s ensureSystem all host executed successfully", c.Cluster.Name)
	return nil
}

func (p *Provider) EnsureKubeadmInitWaitControlPlanePhase(ctx context.Context, c *common.Cluster) error {
	sh, err := c.Spec.Machines[0].SSH()
	if err != nil {
		return err
	}

	err = kubeadm.ApplyCustomMaster(sh, c, p.Cfg)
	if err != nil {
		klog.Errorf("ApplyCustomImagesMaster err: %v", err)
		return err
	}

	start := time.Now()
	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		healthStatus := 0
		clientset, err := c.ClientsetForBootstrap()
		if err != nil {
			log.Warn(err.Error())
			return false, nil
		}

		res := clientset.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx)
		res.StatusCode(&healthStatus)
		if healthStatus != http.StatusOK {
			klog.Errorf("Discovery healthz err: %+v", res.Error())
			return false, nil
		}

		log.Infof("All control plane components are healthy after %f seconds\n", time.Since(start).Seconds())
		return true, nil
	})
}

func (p *Provider) EnsureMarkControlPlane(ctx context.Context, c *common.Cluster) error {
	clientset, err := c.ClientsetForBootstrap()
	if err != nil {
		return err
	}

	for _, machine := range c.Spec.Machines {
		if machine.Labels == nil {
			machine.Labels = make(map[string]string)
		}

		machine.Labels[constants.LabelNodeRoleMaster] = ""
		if !c.Spec.Features.EnableMasterSchedule {
			taint := corev1.Taint{
				Key:    constants.LabelNodeRoleMaster,
				Effect: corev1.TaintEffectNoSchedule,
			}
			if !funk.Contains(machine.Taints, taint) {
				machine.Taints = append(machine.Taints, taint)
			}
		}
		err := apiclient.MarkNode(ctx, clientset, machine.IP, machine.Labels, machine.Taints)
		if err != nil {
			return errors.Wrap(err, machine.IP)
		}
	}

	return nil
}

func (p *Provider) EnsureRegistryHosts(ctx context.Context, c *common.Cluster) error {
	if !p.Cfg.NeedSetHosts() {
		return nil
	}

	domains := []string{
		p.Cfg.Registry.Domain,
		c.Spec.TenantID + "." + p.Cfg.Registry.Domain,
	}
	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		for _, one := range domains {
			remoteHosts := &hosts.RemoteHosts{Host: one, SSH: machineSSH}
			err := remoteHosts.Set(p.Cfg.Registry.IP)
			if err != nil {
				return errors.Wrap(err, machine.IP)
			}
		}
	}

	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx context.Context, c *common.Cluster) error {
	if c.Spec.Features.Hooks == nil {
		return nil
	}

	hook := c.Spec.Features.Hooks[devopsv1.HookPreInstall]
	if hook == "" {
		return nil
	}
	cmd := strings.Split(hook, " ")[0]

	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		machineSSH.Execf("chmod +x %s", cmd)
		_, stderr, exit, err := machineSSH.Exec(hook)
		if err != nil || exit != 0 {
			return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
		}
	}
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx context.Context, c *common.Cluster) error {
	if c.Spec.Features.Hooks == nil {
		return nil
	}

	hook := c.Spec.Features.Hooks[devopsv1.HookPostInstall]
	if hook == "" {
		return nil
	}
	cmd := strings.Split(hook, " ")[0]

	for _, machine := range c.Spec.Machines {
		machineSSH, err := machine.SSH()
		if err != nil {
			return err
		}

		machineSSH.Execf("chmod +x %s", cmd)
		_, stderr, exit, err := machineSSH.Exec(hook)
		if err != nil || exit != 0 {
			return fmt.Errorf("exec %q failed:exit %d:stderr %s:error %s", hook, exit, stderr, err)
		}
	}
	return nil
}

func (p *Provider) EnsureApplyEtcd(ctx context.Context, c *common.Cluster) error {
	etcdPeerEndpoints := []string{}
	etcdClusterEndpoints := []string{}
	for _, machine := range c.Spec.Machines {
		etcdPeerEndpoints = append(etcdPeerEndpoints, fmt.Sprintf("%s=https://%s:2380", machine.IP, machine.IP))
		etcdClusterEndpoints = append(etcdClusterEndpoints, fmt.Sprintf("https://%s:2379", machine.IP))
	}

	for _, machine := range c.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		etcdByte, err := sh.ReadFile(constants.EtcdPodManifestFile)
		if err != nil {
			return fmt.Errorf("node: %s ReadFile: %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		etcdObj, err := k8sutil.UnmarshalFromYaml(etcdByte, corev1.SchemeGroupVersion)
		if err != nil {
			return fmt.Errorf("node: %s marshalling %s failed error: %v", machine.IP, constants.EtcdPodManifestFile, err)
		}

		if etcdPod, ok := etcdObj.(*corev1.Pod); ok {
			isFindState := false
			isFindLogger := false
			klog.Infof("etcd pod name: %s, cmd: %s", etcdPod.Name, etcdPod.Spec.Containers[0].Command)
			for i, arg := range etcdPod.Spec.Containers[0].Command {
				if strings.HasPrefix(arg, "--initial-cluster=") {
					etcdPod.Spec.Containers[0].Command[i] = fmt.Sprintf("--initial-cluster=%s", strings.Join(etcdPeerEndpoints, ","))
				}
				if strings.HasPrefix(arg, "--initial-cluster-state=") {
					isFindState = true
				}

				if strings.HasPrefix(arg, "--logger=") {
					isFindLogger = true
				}
			}

			if !isFindState {
				etcdPod.Spec.Containers[0].Command = append(etcdPod.Spec.Containers[0].Command, "--initial-cluster-state=existing")
			}

			if !isFindLogger {
				etcdPod.Spec.Containers[0].Command = append(etcdPod.Spec.Containers[0].Command, "--logger=zap")
			}
			serialized, err := k8sutil.MarshalToYaml(etcdPod, corev1.SchemeGroupVersion)
			if err != nil {
				return errors.Wrapf(err, "failed to marshal manifest for %q to YAML", etcdPod.Name)
			}

			sh.WriteFile(bytes.NewReader(serialized), constants.EtcdPodManifestFile)
		}

		apiServerByte, err := sh.ReadFile(constants.KubeAPIServerPodManifestFile)
		if err != nil {
			return fmt.Errorf("node: %s ReadFile: %s failed error: %v", machine.IP, constants.KubeAPIServerPodManifestFile, err)
		}

		apiServerObj, err := k8sutil.UnmarshalFromYaml(apiServerByte, corev1.SchemeGroupVersion)
		if err != nil {
			return fmt.Errorf("node: %s marshalling %s failed error: %v", machine.IP, constants.KubeAPIServerPodManifestFile, err)
		}

		var ok bool
		var apiServerPod *corev1.Pod
		if apiServerPod, ok = apiServerObj.(*corev1.Pod); !ok {
			continue
		}

		klog.Infof("apiServer pod name: %s, cmd: %s", apiServerPod.Name, apiServerPod.Spec.Containers[0].Command)
		for i, arg := range apiServerPod.Spec.Containers[0].Command {
			if !strings.HasPrefix(arg, "--etcd-servers=") {
				continue
			}

			apiServerPod.Spec.Containers[0].Command[i] = fmt.Sprintf("--etcd-servers=%s", strings.Join(etcdClusterEndpoints, ","))
			break
		}

		serialized, err := k8sutil.MarshalToYaml(apiServerPod, corev1.SchemeGroupVersion)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal manifest for %q to YAML", apiServerPod.Name)
		}

		sh.WriteFile(bytes.NewReader(serialized), constants.KubeAPIServerPodManifestFile)
	}

	return nil
}

func (p *Provider) EnsureApplyControlPlane(ctx context.Context, c *common.Cluster) error {
	for _, machine := range c.Spec.Machines[1:] {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}
		err = kubemisc.CovertMasterKubeConfig(sh, c)
		if err != nil {
			return err
		}

		_, _, _, err = sh.Execf("systemctl enable kubelet && systemctl restart kubelet")
		if err != nil {
			return err
		}
		// err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-apiserver"))
		// if err != nil {
		// 	return err
		// }
		// err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-controller-manager"))
		// if err != nil {
		// 	return err
		// }
		// err = kubeadm.RestartContainerByFilter(sh, kubeadm.DockerFilterForControlPlane("kube-scheduler"))
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}

func (p *Provider) EnsureExtKubeconfig(ctx context.Context, c *common.Cluster) error {
	if c.ClusterCredential.ExtData == nil {
		c.ClusterCredential.ExtData = make(map[string]string)
	}

	apiserver := certs.BuildExternalApiserverEndpoint(c)
	klog.Infof("external apiserver url: %s", apiserver)
	cfgMaps, err := certs.CreateApiserverKubeConfigFile(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, c.Cluster.Name)
	if err != nil {
		klog.Errorf("build apiserver kubeconfg err: %+v", err)
		return err
	}
	klog.Infof("[%s/%s] start convert apiserver kubeconfig ...", c.Cluster.Namespace, c.Cluster.Name)
	for _, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		externalKubeconfig := string(by)
		klog.Infof("cluster: %s externalKubeconfig: \n%s", c.Cluster.Name, externalKubeconfig)
		c.ClusterCredential.ExtData[pkiutil.ExternalAdminKubeConfigFileName] = externalKubeconfig
	}

	return nil
}

func (p *Provider) EnsureMetricsServer(ctx context.Context, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}
	objs, err := metricsserver.BuildMetricsServerAddon(c)
	if err != nil {
		return errors.Wrapf(err, "build metrics-server err: %v", err)
	}

	logger := ctrl.Log.WithValues("cluster", c.Name, "component", "metrics-server")
	logger.Info("start reconcile ...")
	for _, obj := range objs {
		err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
		}
	}

	return nil
}

func (p *Provider) EnsureEth(ctx context.Context, c *common.Cluster) error {
	var cniType string
	var ok bool

	if cniType, ok = c.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; !ok {
		return nil
	}

	if cniType != "dke-cni" {
		return nil
	}

	for _, machine := range c.Spec.Machines {
		sh, err := machine.SSH()
		if err != nil {
			return err
		}

		err = cni.ApplyEth(sh, c)
		if err != nil {
			klog.Errorf("node: %s apply eth err: %v", sh.HostIP(), err)
			return err
		}
	}

	return nil
}

func (p *Provider) EnsureCni(ctx context.Context, c *common.Cluster) error {
	var cniType string
	var ok bool

	if cniType, ok = c.Cluster.Spec.Features.Hooks[devopsv1.HookCniInstall]; !ok {
		return nil
	}

	switch cniType {
	case "dke-cni":
		for _, machine := range c.Spec.Machines {
			sh, err := machine.SSH()
			if err != nil {
				return err
			}

			err = cni.ApplyCniCfg(sh, c)
			if err != nil {
				klog.Errorf("node: %s apply cni cfg err: %v", sh.HostIP(), err)
				return err
			}
		}
	case "flannel":
		clusterCtx, err := c.ClusterManager.Get(c.Name)
		if err != nil {
			return nil
		}
		objs, err := flannel.BuildFlannelAddon(p.Cfg, c)
		if err != nil {
			return errors.Wrapf(err, "build flannel err: %v", err)
		}

		logger := ctrl.Log.WithValues("cluster", c.Name, "component", "flannel")
		logger.Info("start reconcile ...")
		for _, obj := range objs {
			err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStatePresent)
			if err != nil {
				return errors.Wrapf(err, "Reconcile  err: %v", err)
			}
		}
	default:
		return fmt.Errorf("unknown cni type: %s", cniType)
	}

	return nil
}

func (p *Provider) EnsureMasterNode(ctx context.Context, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}

	node := &corev1.Node{}
	var noReadNode *devopsv1.ClusterMachine
	for _, machine := range c.Spec.Machines {
		err := clusterCtx.Client.Get(ctx, types.NamespacedName{Name: machine.IP}, node)
		if err != nil {
			klog.Warningf("failed get cluster: %s node: %s", c.Cluster.Name, machine.IP)
			return errors.Wrapf(err, "failed get cluster: %s node: %s", c.Cluster.Name, machine.IP)
		}

		isNoReady := false
		for j := range node.Status.Conditions {
			if node.Status.Conditions[j].Type != corev1.NodeReady {
				continue
			}

			if node.Status.Conditions[j].Status != corev1.ConditionTrue {
				isNoReady = true
			}
			break
		}

		if isNoReady {
			noReadNode = machine
			break
		}
	}

	if noReadNode == nil {
		return nil
	}

	klog.Infof("start reconcile node: %s", noReadNode.IP)
	sh, err := noReadNode.SSH()
	if err != nil {
		return err
	}

	for _, file := range c.Spec.Features.Files {
		err = system.CopyFile(sh, &file)
		if err != nil {
			return err
		}
	}

	phases := []func(s ssh.Interface, c *common.Cluster) error{
		system.Install,
		component.Install,
		preflight.RunMasterChecks,
		kubemisc.Install,
		kubeadm.JoinControlPlane,
	}

	for _, phase := range phases {
		err = phase(sh, c)
		if err != nil {
			return err
		}
	}

	return nil
}
