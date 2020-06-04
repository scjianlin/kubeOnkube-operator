package cluster

import (
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/server/mux"

	clusterprovider "github.com/gostship/kunkka/pkg/provider/cluster"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/provider"
	"github.com/gostship/kunkka/pkg/provider/baremetal/config"
	"github.com/gostship/kunkka/pkg/provider/baremetal/validation"
	"github.com/gostship/kunkka/pkg/util/containerregistry"
	"github.com/gostship/kunkka/pkg/util/pointer"
	"k8s.io/klog"
)

func init() {
	p, err := NewProvider()
	if err != nil {
		klog.Errorf("init cluster provider error: %s", err)
		return
	}
	clusterprovider.Register(p.Name(), p)
}

type Provider struct {
	*clusterprovider.DelegateProvider

	config *config.Config
}

var _ clusterprovider.Provider = &Provider{}

func NewProvider() (*Provider, error) {
	p := new(Provider)

	p.DelegateProvider = &clusterprovider.DelegateProvider{
		ProviderName: "Baremetal",
		CreateHandlers: []clusterprovider.Handler{
			p.EnsureCopyFiles,
			// p.EnsurePreInstallHook,
			//
			// p.EnsureRegistryHosts,
			p.EnsureKernelModule,
			p.EnsureSysctl,
			p.EnsureDisableSwap,

			p.EnsurePreflight, // wait basic setting done

			p.EnsureClusterComplete,

			// p.EnsureNvidiaDriver,
			// p.EnsureNvidiaContainerRuntime,
			// p.EnsureDocker,
			p.EnsureKubelet,
			// p.EnsureCNIPlugins,
			// p.EnsureKubeadm,
			//
			// p.EnsurePrepareForControlplane,

			p.EnsureKubeadmInitKubeletStartPhase,
			p.EnsureKubeadmInitCertsPhase,
			p.EnsureStoreCredential,
			p.EnsureKubeconfig,
			p.EnsureKubeadmInitKubeConfigPhase,
			p.EnsureKubeadmInitControlPlanePhase,
			p.EnsureKubeadmInitEtcdPhase,
			p.EnsureKubeadmInitWaitControlPlanePhase,
			p.EnsureKubeadmInitUploadConfigPhase,
			p.EnsureKubeadmInitUploadCertsPhase,
			p.EnsureKubeadmInitBootstrapTokenPhase,
			p.EnsureKubeadmInitAddonPhase,

			p.EnsureJoinControlePlane,
			p.EnsurePatchAnnotation, // wait rest master ready
			p.EnsureMarkControlPlane,

			p.EnsureCleanup,

			// p.EnsurePostInstallHook,
		},
		UpdateHandlers: []clusterprovider.Handler{
			p.EnsureRenewCerts,
			p.EnsureAPIServerCert,
			p.EnsureStoreCredential,
		},
	}

	cfg, err := config.New(constants.ConfigFile)
	if err != nil {
		return nil, err
	}
	p.config = cfg

	containerregistry.Init(cfg.Registry.Domain, cfg.Registry.Namespace)

	return p, nil
}

func (p *Provider) RegisterHandler(mux *mux.PathRecorderMux) {
	prefix := "/provider/" + strings.ToLower(p.Name())

	mux.HandleFunc(path.Join(prefix, "ping"), p.ping)
}

func (p *Provider) Validate(cluster *provider.Cluster) field.ErrorList {
	return validation.ValidateCluster(cluster)
}

func (p *Provider) PreCreate(cluster *provider.Cluster) error {
	if cluster.Spec.Version == "" {
		cluster.Spec.Version = constants.K8sVersions[0]
	}
	if cluster.Spec.ClusterCIDR == "" {
		cluster.Spec.ClusterCIDR = "10.244.0.0/16"
	}
	if cluster.Spec.NetworkDevice == "" {
		cluster.Spec.NetworkDevice = "eth0"
	}

	if cluster.Spec.Features.IPVS == nil {
		cluster.Spec.Features.IPVS = pointer.ToBool(true)
	}

	if cluster.Spec.Properties.MaxClusterServiceNum == nil && cluster.Spec.ServiceCIDR == nil {
		cluster.Spec.Properties.MaxClusterServiceNum = pointer.ToInt32(256)
	}
	if cluster.Spec.Properties.MaxNodePodNum == nil {
		cluster.Spec.Properties.MaxNodePodNum = pointer.ToInt32(256)
	}
	if cluster.Spec.Features.SkipConditions == nil {
		cluster.Spec.Features.SkipConditions = p.config.Feature.SkipConditions
	}

	if cluster.Spec.Etcd == nil {
		cluster.Spec.Etcd = &devopsv1.Etcd{Local: &devopsv1.LocalEtcd{}}
	}

	return nil
}
