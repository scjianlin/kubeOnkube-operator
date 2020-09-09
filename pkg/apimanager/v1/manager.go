package v1

import (
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/provider/monitoring/prometheus"
)

type Manager struct {
	Cluster *k8smanager.ClusterManager
	Monitor *prometheus.Prometheus
}
