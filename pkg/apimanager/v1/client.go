package v1

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// get cluster client
func (m *Manager) getClient(cliName string) (client.Client, error) {
	var cli client.Client
	if cliName == MetaClusterName {
		cli = m.Cluster.GetClient()
	} else {
		cls, err := m.Cluster.Get(cliName)
		if err != nil {
			return nil, nil
		}
		cli = cls.Client
	}
	return cli, nil
}

// 获取集群client interface
func (m *Manager) getClientInterface(cliName string) (kubernetes.Interface, error) {
	var cli kubernetes.Interface
	if cliName == MetaClusterName {
		cli = m.Cluster.KubeCli
	} else {
		cls, err := m.Cluster.Get(cliName)
		if err != nil {
			return nil, nil
		}
		cli = cls.KubeCli
	}
	return cli, nil
}

// 获取集群client restconfig
func (m *Manager) getClientRestCfg(cliName string) (rest.Config, error) {
	var cfg rest.Config
	if cliName == MetaClusterName {
		cfg = *m.Cluster.GetConfig()
	} else {
		cls, err := m.Cluster.Get(cliName)
		if err != nil {
			return cfg, nil
		}
		cfg = *cls.RestConfig
	}
	return cfg, nil
}

// 获取集群kubeconfgi
func (m *Manager) getConfig(clsName string) ([]byte, error) {
	cls, err := m.Cluster.Get(clsName)
	if err != nil {
		return nil, err
	}
	return cls.RawKubeconfig, nil
}
