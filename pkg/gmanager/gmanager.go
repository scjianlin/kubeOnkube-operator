package gmanager

import (
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/provider"
)

type GManager struct {
	*provider.ProviderManager
	*k8smanager.ClusterManager
}
