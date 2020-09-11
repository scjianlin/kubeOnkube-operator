package v1

import (
	"errors"
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/provider/monitoring/prometheus"
	"k8s.io/klog"
	"sync"
)

type Manager struct {
	Cluster *k8smanager.ClusterManager
	Monitor map[string]*prometheus.Prometheus
	sync.RWMutex
}

func (m *Manager) getMonitClient(name string) (*prometheus.Prometheus, error) {
	m.Lock()
	defer m.Unlock()
	if cli, ok := m.Monitor[name]; ok {
		return cli, nil
	}
	return nil, errors.New("not found client")
}

func (m *Manager) addMonitClient(name string, endpoint string) (*prometheus.Prometheus, error) {
	if cli, _ := m.getMonitClient(name); cli == nil {
		cli := m.newMonitor(endpoint)
		m.Monitor[name] = cli
		return cli, nil
	}
	return m.Monitor[name], nil
}

func (m *Manager) newMonitor(endpoint string) *prometheus.Prometheus {
	opt := &prometheus.Options{Endpoint: endpoint}
	client, err := prometheus.NewPrometheus(opt)
	if err != nil {
		klog.Error("get prometheus client error")
		return nil
	}
	return &client
}

func (m *Manager) updateMonitor(name string, endpoint string) (*prometheus.Prometheus, error) {
	m.Lock()
	m.Unlock()
	if _, err := m.getMonitClient(name); err == nil {
		m.Monitor[name] = m.newMonitor(endpoint)
	} else {
		return nil, errors.New("not found monit client")
	}
	return m.Monitor[name], nil
}
