package provider

import (
	baremetalcluster "github.com/gostship/kunkka/pkg/provider/baremetal/cluster"
	baremetalmachine "github.com/gostship/kunkka/pkg/provider/baremetal/machine"
	"github.com/gostship/kunkka/pkg/provider/cluster"
	clusterprovider "github.com/gostship/kunkka/pkg/provider/cluster"
	"github.com/gostship/kunkka/pkg/provider/config"
	hostedcluster "github.com/gostship/kunkka/pkg/provider/hosted/cluster"
	hostedmachine "github.com/gostship/kunkka/pkg/provider/hosted/machine"
	"github.com/gostship/kunkka/pkg/provider/machine"
	machineprovider "github.com/gostship/kunkka/pkg/provider/machine"
)

type ProviderManager struct {
	*cluster.CpManager
	*machine.MpManager
	Cfg *config.Config
}

var AddToCpManagerFuncs []func(*clusterprovider.CpManager, *config.Config) error
var AddToMpManagerFuncs []func(*machineprovider.MpManager, *config.Config) error

func NewProvider() (*ProviderManager, error) {
	AddToCpManagerFuncs = append(AddToCpManagerFuncs, baremetalcluster.Add)
	AddToCpManagerFuncs = append(AddToCpManagerFuncs, hostedcluster.Add)

	AddToMpManagerFuncs = append(AddToMpManagerFuncs, baremetalmachine.Add)
	AddToMpManagerFuncs = append(AddToMpManagerFuncs, hostedmachine.Add)

	cfg, _ := config.NewDefaultConfig()
	mgr := &ProviderManager{
		CpManager: cluster.New(),
		MpManager: machine.New(),
		Cfg:       cfg,
	}

	for _, f := range AddToCpManagerFuncs {
		if err := f(mgr.CpManager, mgr.Cfg); err != nil {
			return nil, err
		}
	}

	for _, f := range AddToMpManagerFuncs {
		if err := f(mgr.MpManager, mgr.Cfg); err != nil {
			return nil, err
		}
	}

	return mgr, nil
}
