package k8smanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/gostship/kunkka/pkg/provider/monitoring/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sort"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	logger = logf.Log.WithName("controller")
)

// key config
const (
	ClustersAll = "all"
)

// MasterClient ...
type MasterClient struct {
	KubeCli kubernetes.Interface
	manager.Manager
}

// ClusterManager ...
type ClusterManager struct {
	MasterClient
	clusters []*Cluster
	monitor  map[string]*prometheus.Prometheus
	Started  bool
	sync.RWMutex
}

// NewManager ...
func NewManager(cli MasterClient) (*ClusterManager, error) {
	cMgr := &ClusterManager{
		MasterClient: cli,
		clusters:     make([]*Cluster, 0, 4),
		monitor:      map[string]*prometheus.Prometheus{},
	}

	cMgr.Started = true
	return cMgr, nil
}

// GetAll get all cluster
func (m *ClusterManager) GetAll(name ...string) []*Cluster {
	m.RLock()
	defer m.RUnlock()

	isAll := true
	var ObserveName string
	if len(name) > 0 {
		if name[0] != ClustersAll {
			isAll = false
		}
	}

	list := make([]*Cluster, 0, 4)
	for _, c := range m.clusters {
		if c.Status == ClusterOffline {
			continue
		}

		if isAll {
			list = append(list, c)
		} else {
			if ObserveName != "" && ObserveName == c.Name {
				list = append(list, c)
				break
			}
		}
	}

	return list
}

// Add ...
func (m *ClusterManager) Add(cluster *Cluster) error {
	if _, err := m.Get(cluster.Name); err == nil {
		return fmt.Errorf("cluster name: %s is already add to manager", cluster.Name)
	}

	m.Lock()
	defer m.Unlock()
	m.clusters = append(m.clusters, cluster)
	sort.Slice(m.clusters, func(i int, j int) bool {
		return m.clusters[i].Name > m.clusters[j].Name
	})

	return nil
}

// update..
func (m *ClusterManager) Update(cluster *Cluster) error {
	if cls, err := m.Get(cluster.Name); err == nil {
		//	update
		index, _ := m.GetClusterIndex(cls.Name)
		m.clusters[index] = cls
		klog.Infof("the cluster update %s has been updated.", cls.Name)
		return nil
	}
	klog.Error("cluster %s,not found.", cluster.Name)
	return errors.New("cluster not found.")
}

// GetClusterIndex ...
func (m *ClusterManager) GetClusterIndex(name string) (int, bool) {
	for i, r := range m.clusters {
		if r.Name == name {
			return i, true
		}
	}
	return 0, false
}

// Delete ...
func (m *ClusterManager) Delete(name string) error {
	if name == "" {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	if len(m.clusters) == 0 {
		klog.Errorf("clusters list is empty, nothing to delete")
		return nil
	}

	index, ok := m.GetClusterIndex(name)
	if !ok {
		klog.Warningf("cluster:%s  is not found in the registries list, nothing to delete", name)
		return nil
	}

	m.clusters[index].Stop()
	clusters := m.clusters
	clusters = append(clusters[:index], clusters[index+1:]...)
	m.clusters = clusters
	klog.Infof("the cluster %s has been deleted.", name)
	return nil
}

// Get ...
func (m *ClusterManager) Get(name string) (*Cluster, error) {
	m.Lock()
	defer m.Unlock()

	if name == "" || name == "all" {
		return nil, fmt.Errorf("single query not support: %s ", name)
	}

	var findCluster *Cluster
	for _, c := range m.clusters {
		if name == c.Name {
			findCluster = c
			break
		}
	}
	if findCluster == nil {
		return nil, fmt.Errorf("cluster: %s not found", name)
	}

	if findCluster.Status == ClusterOffline {
		return nil, fmt.Errorf("cluster: %s found, but offline", name)
	}

	return findCluster, nil
}

func (m *ClusterManager) cluterCheck() {
	klog.V(5).Info("cluster health check.")
	for _, c := range m.clusters {
		if !c.healthCheck() {
			klog.Warningf("cluster: %s healthCheck fail", c.Name)
		}
	}
}

func (m *ClusterManager) AddNewClusters(name string, kubeconfig string) (*Cluster, error) {
	if c, _ := m.Get(name); c != nil {
		return c, nil
	}

	nc, err := NewCluster(name, []byte(kubeconfig), logger)
	if err != nil {
		klog.Errorf("cluster: %s new client err: %v", name, err)
		return nil, err
	}

	err = m.preStart(nc)
	if err != nil {
		klog.Error("cluster client preStart error:%s", nc.Name)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	nc.StartCache(ctx.Done())
	err = m.Add(nc)
	if err != nil {
		klog.Errorf("cluster: %s add err: %+v", name, err)
		return nil, err
	}
	return nc, nil
}

// Start timer check cluster health
func (m *ClusterManager) Start(stopCh <-chan struct{}) error {
	klog.V(4).Info("multi cluster manager start check loop ... ")
	wait.Until(m.cluterCheck, time.Minute, stopCh)

	klog.V(4).Info("multi cluster manager stoped ... ")
	m.Stop()
	return nil
}

func (m *ClusterManager) Stop() {
	m.Lock()
	defer m.Unlock()

	for _, cluster := range m.clusters {
		cluster.Stop()
	}
}

// 增加对象索引
func (r *ClusterManager) preStart(cls *Cluster) error {
	if err := cls.Mgr.GetFieldIndexer().IndexField(context.TODO(), &corev1.Pod{}, "spec.nodeName", func(object runtime.Object) []string {
		pod := object.(*corev1.Pod)
		return []string{pod.Status.HostIP}
	}); err != nil {
		klog.Warning("cluster: %#v add field index pod status.hostIP, err: %#v", cls.Name, err)
		return errors.New("cluster add field index pod spec.nodeName failed")
	} else {
		klog.Warning("########### cluster: %#v add field index pod status.hostIP, successfully ##################", cls.Name)
	}

	if err := cls.Mgr.GetFieldIndexer().IndexField(context.TODO(), &corev1.Event{}, "source.host", func(object runtime.Object) []string {
		event := object.(*corev1.Event)
		return []string{event.Source.Host}
	}); err != nil {
		klog.Warning("cluster: %#v add field index pod source.host, err: %#v", cls.Name, err)
		return errors.New("cluster add field index pod source.host failed")
	} else {
		klog.Warning("########### cluster: %#v  add field index pod source.host, successfully ##################", cls.Name)
	}

	if err := cls.Mgr.GetFieldIndexer().IndexField(context.TODO(), &corev1.Event{}, "involvedObject.kind", func(object runtime.Object) []string {
		event := object.(*corev1.Event)
		return []string{event.InvolvedObject.Kind}
	}); err != nil {
		klog.Warning("cluster: %#v add field index pod involvedObject.kind, err: %#v", cls.Name, err)
		return errors.New("cluster add field index pod involvedObject.kind failed")
	} else {
		klog.Warning("########### cluster: %#v  add field index pod involvedObject.kind, successfully ##################", cls.Name)
	}
	return nil
}
