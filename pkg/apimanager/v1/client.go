package v1

import (
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
