package k8smanager

import (
	"fmt"
	"github.com/gostship/kunkka/pkg/provider/monitoring/prometheus"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// Add ...
func (m *ClusterManager) AddMonitor(name string, monit *prometheus.Prometheus) error {
	if _, err := m.GetMonitor(name); err == nil {
		return fmt.Errorf("monitor name: %s is already add to manager", name)
	}
	m.Lock()
	defer m.Unlock()
	m.monitor[name] = monit
	return nil
}

// Get ...
func (m *ClusterManager) GetMonitor(name string) (*prometheus.Prometheus, error) {
	m.Lock()
	defer m.Unlock()

	if name == "" || name == "all" {
		return nil, fmt.Errorf("single query not support: %s ", name)
	}
	var findMonitor *prometheus.Prometheus
	if prom, ok := m.monitor[name]; ok {
		findMonitor = prom
	}
	if findMonitor == nil {
		return nil, fmt.Errorf("monitor: %s not found", name)
	}
	return findMonitor, nil
}

// update..
func (m *ClusterManager) UpdateMonitor(name string, monitor *prometheus.Prometheus) error {
	if _, err := m.GetMonitor(name); err == nil {
		//	update
		m.monitor[name] = monitor
		klog.Infof("the monitor update %s has been updated.", name)
		return nil
	}
	klog.Error("monitor %s,not found.", name)
	return errors.New("monitor not found.")
}

// Delete ...
func (m *ClusterManager) DeleteMonitir(name string) error {
	if name == "" {
		return nil
	}
	m.Lock()
	defer m.Unlock()
	if len(m.monitor) == 0 {
		klog.Errorf("monitor list is empty, nothing to delete")
		return nil
	}
	delete(m.monitor, name)
	klog.Infof("the monitor %s has been deleted.", name)
	return nil
}
