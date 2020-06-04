package provider

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cluster struct {
	*devopsv1.Cluster
	ClusterCredential *devopsv1.ClusterCredential
}

const (
	defaultTimeout = 30 * time.Second
	defaultQPS     = 100
	defaultBurst   = 200
)

func GetClusterByName(ctx context.Context, cli client.Client, ns, name string) (*Cluster, error) {
	res := &Cluster{}
	cluster := &devopsv1.Cluster{}
	err := cli.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, cluster)
	if err != nil {
		return nil, err
	}

	res.Cluster = cluster
	if cluster.Spec.ClusterCredentialRef != nil {
		clusterCredential := &devopsv1.ClusterCredential{}
		err := cli.Get(ctx, types.NamespacedName{Name: cluster.Spec.ClusterCredentialRef.Name, Namespace: ns}, clusterCredential)
		if err != nil {
			return nil, fmt.Errorf("get cluster's credential error: %w", err)
		}
		res.ClusterCredential = clusterCredential
	}

	return res, nil
}

func GetCluster(ctx context.Context, cli client.Client, cluster *devopsv1.Cluster) (*Cluster, error) {
	result := new(Cluster)
	result.Cluster = cluster
	if cluster.Spec.ClusterCredentialRef != nil {
		clusterCredential := &devopsv1.ClusterCredential{}
		err := cli.Get(ctx, types.NamespacedName{Name: cluster.Spec.ClusterCredentialRef.Name, Namespace: cluster.Namespace}, clusterCredential)
		if err != nil {
			return nil, fmt.Errorf("get cluster's credential error: %w", err)
		}
		result.ClusterCredential = clusterCredential
	}

	return result, nil
}

func Clientset(cluster *devopsv1.Cluster, credential *devopsv1.ClusterCredential) (kubernetes.Interface, error) {
	return (&Cluster{Cluster: cluster, ClusterCredential: credential}).Clientset()
}

func (c *Cluster) Clientset() (kubernetes.Interface, error) {
	config, err := c.RESTConfig(&rest.Config{})
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func (c *Cluster) ClientsetForBootstrap() (kubernetes.Interface, error) {
	config, err := c.RESTConfigForBootstrap(&rest.Config{})
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func (c *Cluster) RESTConfigForBootstrap(config *rest.Config) (*rest.Config, error) {
	host, err := c.HostForBootstrap()
	if err != nil {
		return nil, err
	}
	config.Host = host

	return c.RESTConfig(config)
}
func (c *Cluster) RESTConfig(config *rest.Config) (*rest.Config, error) {
	if config.Host == "" {
		host, err := c.Host()
		if err != nil {
			return nil, err
		}
		config.Host = host
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.QPS == 0 {
		config.QPS = defaultQPS
	}
	if config.Burst == 0 {
		config.Burst = defaultBurst
	}

	if c.ClusterCredential.CACert != nil {
		config.TLSClientConfig.CAData = c.ClusterCredential.CACert
	} else {
		config.TLSClientConfig.Insecure = true
	}
	if c.ClusterCredential.ClientCert != nil && c.ClusterCredential.ClientKey != nil {
		config.TLSClientConfig.CertData = c.ClusterCredential.ClientCert
		config.TLSClientConfig.KeyData = c.ClusterCredential.ClientKey
	}

	if c.ClusterCredential.Token != nil {
		config.BearerToken = *c.ClusterCredential.Token
	}

	return config, nil
}

func (c *Cluster) Host() (string, error) {
	addrs := make(map[devopsv1.AddressType][]devopsv1.ClusterAddress)
	for _, one := range c.Status.Addresses {
		addrs[one.Type] = append(addrs[one.Type], one)
	}

	var address *devopsv1.ClusterAddress
	if len(addrs[devopsv1.AddressInternal]) != 0 {
		address = &addrs[devopsv1.AddressInternal][rand.Intn(len(addrs[devopsv1.AddressInternal]))]
	} else if len(addrs[devopsv1.AddressAdvertise]) != 0 {
		address = &addrs[devopsv1.AddressAdvertise][rand.Intn(len(addrs[devopsv1.AddressAdvertise]))]
	} else {
		if len(addrs[devopsv1.AddressReal]) != 0 {
			address = &addrs[devopsv1.AddressReal][rand.Intn(len(addrs[devopsv1.AddressReal]))]
		}
	}

	if address == nil {
		return "", errors.New("can't find valid address")
	}

	return fmt.Sprintf("%s:%d", address.Host, address.Port), nil
}

func (c *Cluster) HostForBootstrap() (string, error) {
	for _, one := range c.Status.Addresses {
		if one.Type == devopsv1.AddressReal {
			return fmt.Sprintf("%s:%d", one.Host, one.Port), nil
		}
	}

	return "", errors.New("can't find bootstrap address")
}
