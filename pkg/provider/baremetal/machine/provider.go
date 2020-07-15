package machine

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gostship/kunkka/pkg/provider/baremetal/validation"
	machineprovider "github.com/gostship/kunkka/pkg/provider/machine"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/provider/config"
	"k8s.io/klog"
)

func Add(mgr *machineprovider.MpManager, cfg *config.Config) error {
	p, err := NewProvider(mgr, cfg)
	if err != nil {
		klog.Errorf("init cluster provider error: %s", err)
		return err
	}
	mgr.Register(p.Name(), p)
	return nil
}

type Provider struct {
	*machineprovider.DelegateProvider
	Mgr *machineprovider.MpManager
	Cfg *config.Config
}

func NewProvider(mgr *machineprovider.MpManager, cfg *config.Config) (*Provider, error) {
	p := &Provider{
		Mgr: mgr,
		Cfg: cfg,
	}

	p.DelegateProvider = &machineprovider.DelegateProvider{
		ProviderName: "Baremetal",
		CreateHandlers: []machineprovider.Handler{
			p.EnsureCopyFiles,
			p.EnsurePreInstallHook,
			p.EnsureClean,
			p.EnsureRegistryHosts,

			p.EnsureEth,
			p.EnsureSystem,
			p.EnsureK8sComponent,
			p.EnsurePreflight, // wait basic setting done

			p.EnsureJoinNode,
			p.EnsureKubeconfig,
			p.EnsureMarkNode,
			p.EnsureCni,
			p.EnsureNodeReady,

			p.EnsurePostInstallHook,
		},
		UpdateHandlers: []machineprovider.Handler{
			p.EnsureCni,
			p.EnsurePostInstallHook,
			p.EnsureRegistryHosts,
		},
	}

	return p, nil
}

var _ machineprovider.Provider = &Provider{}

func (p *Provider) Validate(machine *devopsv1.Machine) field.ErrorList {
	return validation.ValidateMachine(machine)
}
