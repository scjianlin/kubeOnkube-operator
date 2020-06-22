package cluster

import (
	"fmt"

	kubeadmv1beta2 "github.com/gostship/kunkka/pkg/apis/kubeadm/v1beta2"
	kubeletv1beta1 "github.com/gostship/kunkka/pkg/apis/kubelet/config/v1beta1"
	kubeproxyv1alpha1 "github.com/gostship/kunkka/pkg/apis/kubeproxy/config/v1alpha1"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/baremetal/phases/kubeadm"
	"github.com/gostship/kunkka/pkg/util/json"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func (p *Provider) getKubeadmConfig(c *common.Cluster) *kubeadm.Config {
	config := new(kubeadm.Config)
	config.InitConfiguration = p.getInitConfiguration(c)
	config.ClusterConfiguration = p.getClusterConfiguration(c)
	config.KubeProxyConfiguration = p.getKubeProxyConfiguration(c)
	config.KubeletConfiguration = p.getKubeletConfiguration(c)

	return config
}

func (p *Provider) getInitConfiguration(c *common.Cluster) *kubeadmv1beta2.InitConfiguration {
	token, _ := kubeadmv1beta2.NewBootstrapTokenString(*c.ClusterCredential.BootstrapToken)

	return &kubeadmv1beta2.InitConfiguration{
		BootstrapTokens: []kubeadmv1beta2.BootstrapToken{
			{
				Token:       token,
				Description: "dke kubeadm bootstrap token",
				TTL:         &metav1.Duration{Duration: 0},
			},
		},
		CertificateKey: *c.ClusterCredential.CertificateKey,
	}
}

func (p *Provider) getClusterConfiguration(c *common.Cluster) *kubeadmv1beta2.ClusterConfiguration {
	controlPlaneEndpoint := fmt.Sprintf("%s:6443", KubeApiServer)

	config := &kubeadmv1beta2.ClusterConfiguration{
		// CertificatesDir: constants.CertificatesDir,
		Networking: kubeadmv1beta2.Networking{
			DNSDomain:     c.Spec.DNSDomain,
			ServiceSubnet: c.Cluster.Status.ServiceCIDR,
		},
		KubernetesVersion:    c.Spec.Version,
		ControlPlaneEndpoint: controlPlaneEndpoint,
		APIServer: kubeadmv1beta2.APIServer{
			CertSANs: k8sutil.GetAPIServerCertSANs(c.Cluster),
		},
		DNS: kubeadmv1beta2.DNS{
			Type: kubeadmv1beta2.CoreDNS,
		},
		ImageRepository: p.Cfg.Registry.Prefix,
		ClusterName:     c.Name,
	}

	utilruntime.Must(json.Merge(&config.Etcd, &c.Spec.Etcd))

	return config
}

func (p *Provider) getKubeProxyConfiguration(c *common.Cluster) *kubeproxyv1alpha1.KubeProxyConfiguration {
	kubeProxyMode := "iptables"
	if c.Spec.Features.IPVS != nil && *c.Spec.Features.IPVS {
		kubeProxyMode = "ipvs"
	}

	return &kubeproxyv1alpha1.KubeProxyConfiguration{
		Mode: kubeproxyv1alpha1.ProxyMode(kubeProxyMode),
	}
}

func (p *Provider) getKubeletConfiguration(c *common.Cluster) *kubeletv1beta1.KubeletConfiguration {
	return &kubeletv1beta1.KubeletConfiguration{
		KubeReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
		SystemReserved: map[string]string{
			"cpu":    "100m",
			"memory": "500Mi",
		},
	}
}
