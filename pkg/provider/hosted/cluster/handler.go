package cluster

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"strings"
	"time"

	"crypto/rand"
	"encoding/hex"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/addons/cni"
	"github.com/gostship/kunkka/pkg/provider/addons/coredns"
	"github.com/gostship/kunkka/pkg/provider/addons/flannel"
	"github.com/gostship/kunkka/pkg/provider/addons/kubeproxy"
	"github.com/gostship/kunkka/pkg/provider/addons/metricsserver"
	"github.com/gostship/kunkka/pkg/provider/phases/certs"
	"github.com/gostship/kunkka/pkg/provider/phases/kubeadm"
	"github.com/gostship/kunkka/pkg/provider/phases/kubemisc"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/pkiutil"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"k8s.io/apimachinery/pkg/runtime"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	tokenFileTemplate = `%s,admin,admin,system:masters
`
	additPolicy = `
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
`
)

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

func (p *Provider) ping(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprint(resp, "pong")
}

func (p *Provider) EnsureCopyFiles(ctx context.Context, c *common.Cluster) error {
	return nil
}

func (p *Provider) EnsurePreInstallHook(ctx context.Context, c *common.Cluster) error {
	return nil
}

func (p *Provider) EnsurePostInstallHook(ctx context.Context, c *common.Cluster) error {
	return nil
}

func (p *Provider) EnsureClusterReady(ctx context.Context, c *common.Cluster) error {

	client, err := c.Clientset()
	if err != nil {
		klog.Error("client set error")
		return err
	}

	var defaultRetry = wait.Backoff{
		Steps:    20,
		Duration: 6 * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}
	werr := wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		body, berr := client.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.TODO()).Raw()
		if berr != nil {
			klog.Error("Failed to do cluster health check for cluster: %s", c.Name)
			return false, nil
		}
		if !strings.EqualFold(string(body), "ok") {
			klog.Error("cluster to do health error")
			return false, nil
		}

		// append cluster client to cache
		if extKubeconfig, ok := c.ClusterCredential.ExtData[pkiutil.ExternalAdminKubeConfigFileName]; ok {
			_, err = c.ClusterManager.AddNewClusters(c.Name, extKubeconfig)
			if err != nil {
				klog.Error("add cluster client to cache error.")
				return false, nil
			}
			klog.Info("add cluster cache success!")
			return true, nil
		} else {
			return false, nil
		}
	})
	if werr != nil {
		return werr
	}
	return nil
}

func (p *Provider) EnsureClusterComplete(ctx context.Context, cluster *common.Cluster) error {
	funcs := []func(cluster *common.Cluster) error{
		completeK8sVersion,
		completeNetworking,
		completeDNS,
		completeAddresses,
		completeCredential,
	}
	for _, f := range funcs {
		if err := f(cluster); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) EnsureCerts(ctx context.Context, c *common.Cluster) error {
	apiserver := certs.BuildApiserverEndpoint(constants.KubeApiServer, 6443)
	err := kubeadm.InitCerts(kubeadm.GetKubeadmConfig(c, p.Cfg, apiserver), c, true)
	if err != nil {
		return err
	}

	return ApplyCertsConfigmap(c.Client, c, c.ClusterCredential.CertsBinaryData)
}

func (p *Provider) EnsureKubeMisc(ctx context.Context, c *common.Cluster) error {
	apiserver := certs.BuildApiserverEndpoint(constants.KubeApiServer, kubemisc.GetBindPort(c.Cluster))
	err := kubemisc.ApplyMasterMisc(c, apiserver)
	if err != nil {
		return err
	}
	return ApplyKubeMiscConfigmap(c.Client, c, c.ClusterCredential.KubeData)
}

func (p *Provider) EnsureEtcd(ctx context.Context, c *common.Cluster) error {
	return nil
}

func (p *Provider) EnsureKubeMaster(ctx context.Context, c *common.Cluster) error {
	r := &Reconciler{
		Obj:      c,
		Provider: p,
	}

	var fs []func() runtime.Object
	fs = append(fs, r.apiServerDeployment)
	fs = append(fs, r.apiServerSvc)
	fs = append(fs, r.controllerManagerDeployment)
	fs = append(fs, r.schedulerDeployment)

	logger := ctrl.Log.WithValues("cluster", c.Name)
	for _, f := range fs {
		obj := f()
		err := k8sutil.Reconcile(logger, c.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "apply object err: %v", err)
		}
	}

	return nil
}

func (p *Provider) EnsureExtKubeconfig(ctx context.Context, c *common.Cluster) error {
	if c.ClusterCredential.ExtData == nil {
		c.ClusterCredential.ExtData = make(map[string]string)
	}

	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.PublicAlternativeNames[0], kubemisc.GetBindPort(c.Cluster))
	klog.Infof("external apiserver url: %s", apiserver)
	cfgMaps, err := certs.CreateApiserverKubeConfigFile(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, c.Cluster.Name)
	if err != nil {
		klog.Errorf("create kubeconfg err: %+v", err)
		return err
	}
	klog.Infof("[%s/%s] start build kubeconfig ...", c.Cluster.Namespace, c.Cluster.Name)
	for _, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}
		c.ClusterCredential.ExtData[pkiutil.ExternalAdminKubeConfigFileName] = string(by)
	}

	return nil
}

func (p *Provider) EnsureAddons(ctx context.Context, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}
	kubeproxyObjs, err := kubeproxy.BuildKubeproxyAddon(p.Cfg, c)
	if err != nil {
		return errors.Wrapf(err, "build kube-proxy err: %+v", err)
	}

	logger := ctrl.Log.WithValues("cluster", c.Name)
	logger.Info("start apply kube-proxy")
	for _, obj := range kubeproxyObjs {
		err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
		}
	}

	logger.Info("start apply coredns")
	corednsObjs, err := coredns.BuildCoreDNSAddon(p.Cfg, c)
	if err != nil {
		return errors.Wrapf(err, "build coredns err: %+v", err)
	}
	for _, obj := range corednsObjs {
		err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
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

			err = cni.ApplyClusterCni(sh, c, machine)
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

	//return nil
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
		err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStateAbsent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
		}
	}

	return nil
}
