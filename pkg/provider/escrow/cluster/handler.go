package cluster

import (
	"context"

	"fmt"
	"net/http"

	"crypto/rand"
	"encoding/hex"

	"crypto/x509"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	kubeadmv1beta2 "github.com/gostship/kunkka/pkg/apis/kubeadm/v1beta2"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/addons/coredns"
	"github.com/gostship/kunkka/pkg/provider/addons/flannel"
	"github.com/gostship/kunkka/pkg/provider/addons/kubeproxy"
	"github.com/gostship/kunkka/pkg/provider/addons/metricsserver"
	"github.com/gostship/kunkka/pkg/provider/certs"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/pkiutil"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	tokenMap := make(map[string]string)

	tokenData := fmt.Sprintf(tokenFileTemplate, *c.ClusterCredential.Token)
	tokenMap["known_tokens.csv"] = tokenData

	warp := &kubeadmv1beta2.WarpperConfiguration{
		InitConfiguration:    p.getInitConfiguration(c),
		ClusterConfiguration: p.getClusterConfiguration(c),
	}

	var lastCACert *certs.CaAll
	cfgMaps := make(map[string][]byte)
	for _, cert := range certs.GetCertsWithoutEtcd() {
		if cert.CAName == "" {
			ret, err := certs.CreateCACertAndKeyFiles(cert, warp, cfgMaps)
			if err != nil {
				return err
			}
			lastCACert = ret
		} else {
			if lastCACert == nil {
				return fmt.Errorf("not hold CertificateAuthority by create cert: %s", cert.Name)
			}
			err := certs.CreateCertAndKeyFilesWithCA(cert, lastCACert, warp, cfgMaps)
			if err != nil {
				return errors.Wrapf(err, "create cert: %s", cert.Name)
			}
		}
	}

	err := certs.CreateServiceAccountKeyAndPublicKeyFiles(warp.ClusterConfiguration.CertificatesDir, x509.RSA, cfgMaps)
	if err != nil {
		return errors.Wrapf(err, "create sa public key")
	}

	if len(cfgMaps) == 0 {
		return fmt.Errorf("no cert build")
	}

	k8sconfigmap := &corev1.ConfigMap{
		ObjectMeta: k8sutil.ObjectMeta(KubeApiServerCerts, Labels, c.Cluster),
		Data:       tokenMap,
		BinaryData: make(map[string][]byte),
	}

	for fileName, v := range cfgMaps {
		k8sconfigmap.BinaryData[fileName] = v
		if fileName == "ca.crt" {
			c.ClusterCredential.CACert = v
		}
		if fileName == "ca.key" {
			c.ClusterCredential.CAKey = v
		}
	}

	err = c.Client.Create(ctx, k8sconfigmap)
	if err != nil {
		return errors.Wrapf(err, "create pki config err: %v", err)
	}

	return nil
}

func (p *Provider) EnsureKubeconfig(ctx context.Context, c *common.Cluster) error {
	if c.ClusterCredential.CACert == nil {
		certsMap := &corev1.ConfigMap{}
		err := c.Client.Get(ctx, types.NamespacedName{Namespace: c.Cluster.Namespace, Name: KubeApiServerCerts}, certsMap)
		if err != nil {
			return errors.Wrapf(err, "get certs configmap err: %v", err)
		}
		c.ClusterCredential.CACert = certsMap.BinaryData["ca.crt"]
		c.ClusterCredential.CAKey = certsMap.BinaryData["ca.key"]
	}

	bindPort := 6443
	if c.Cluster.Spec.Features.HA != nil && c.Cluster.Spec.Features.HA.ThirdPartyHA != nil {
		bindPort = int(c.Cluster.Spec.Features.HA.ThirdPartyHA.VPort)
	}

	cfgMaps, err := certs.CreateMasterKubeConfigFile(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		certs.BuildApiserverEndpoint(KubeApiServer, bindPort), "", c.Cluster.Name)
	if err != nil {
		klog.Errorf("create kubeconfg err: %+v", err)
		return err
	}

	k8sconfigmap := &corev1.ConfigMap{
		ObjectMeta: k8sutil.ObjectMeta(KubeApiServerConfig, Labels, c.Cluster),
		Data:       make(map[string]string),
	}

	klog.Infof("[%s/%s] start build kubeconfig ...", c.Cluster.Namespace, c.Cluster.Name)
	for noPathFile, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}
		k8sconfigmap.Data[noPathFile] = string(by)
	}

	k8sconfigmap.Data["audit-policy.yaml"] = additPolicy

	err = c.Client.Create(ctx, k8sconfigmap)
	if err != nil {
		return errors.Wrapf(err, "create kubeconfig err: %v", err)
	}
	return nil
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
			return errors.Wrapf(err, "create kubeconfig err: %v", err)
		}
	}

	return nil
}

func (p *Provider) EnsureTemp(ctx context.Context, c *common.Cluster) error {
	cfgMap := &corev1.ConfigMap{}
	err := c.Client.Get(ctx, types.NamespacedName{Namespace: c.Cluster.Namespace, Name: KubeApiServerConfig}, cfgMap)
	if err != nil {
		return errors.Wrapf(err, "get certs cfgMap err: %v", err)
	}

	if _, ok := cfgMap.Data[pkiutil.ExternalAdminKubeConfigFileName]; ok {
		return nil
	}

	apiserver := certs.BuildApiserverEndpoint(c.Cluster.Spec.Features.HA.ThirdPartyHA.VIP, int(c.Cluster.Spec.Features.HA.ThirdPartyHA.VPort))
	cfgMaps, err := certs.CreateKubeConfigFiles(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, "", c.Cluster.Name, pkiutil.AdminKubeConfigFileName)
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
		cfgMap.Data[pkiutil.ExternalAdminKubeConfigFileName] = string(by)
	}

	logger := ctrl.Log.WithValues("cluster", c.Name)
	err = k8sutil.Reconcile(logger, c.Client, cfgMap, k8sutil.DesiredStatePresent)
	if err != nil {
		return errors.Wrapf(err, "create k8sSecret err: %v", err)
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
		return errors.Wrapf(err, "build kube-proxy err: %v", err)
	}

	logger := ctrl.Log.WithValues("cluster", c.Name)
	for _, obj := range kubeproxyObjs {
		err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
		}
	}

	corednsObjs, err := coredns.BuildCoreDNSAddon(p.Cfg, c)
	if err != nil {
		return errors.Wrapf(err, "build kube-proxy err: %v", err)
	}
	for _, obj := range corednsObjs {
		err = k8sutil.Reconcile(logger, clusterCtx.Client, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			return errors.Wrapf(err, "Reconcile  err: %v", err)
		}
	}
	return nil
}

func (p *Provider) EnsureFlannel(ctx context.Context, c *common.Cluster) error {
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

	return nil
}

func (p *Provider) EnsureMetricsServer(ctx context.Context, c *common.Cluster) error {
	clusterCtx, err := c.ClusterManager.Get(c.Name)
	if err != nil {
		return nil
	}
	objs, err := metricsserver.BuildMetricsServerAddon(c)
	if err != nil {
		return errors.Wrapf(err, "build flannel err: %v", err)
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
