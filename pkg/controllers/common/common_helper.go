package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultTimeout = 30 * time.Second
	defaultQPS     = 100
	defaultBurst   = 200
)

type Cluster struct {
	*devopsv1.Cluster
	ClusterCredential *devopsv1.ClusterCredential
}

func GetCluster(ctx context.Context, cli client.Client, cluster *devopsv1.Cluster) (*Cluster, error) {
	result := new(Cluster)
	result.Cluster = cluster

	clusterCredential := &devopsv1.ClusterCredential{}
	err := cli.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, clusterCredential)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(3).Infof("cluster: %s not find credential, start create ...", cluster.Name)
			credential := &devopsv1.ClusterCredential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cluster.Name,
					Namespace: cluster.Namespace,
				},
				CredentialInfo: devopsv1.CredentialInfo{
					TenantID:    cluster.Spec.TenantID,
					ClusterName: cluster.Name,
				},
			}
			err := cli.Create(ctx, credential)
			if err != nil && !apierrors.IsAlreadyExists(err) {
				return nil, err
			}

			result.ClusterCredential = credential
			return result, nil
		} else {
			klog.Errorf("cluster: %s faild to get credential, err: %v", cluster.Name, err)
			return nil, err
		}
	}

	result.ClusterCredential = clusterCredential

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

func (c *Cluster) IPs() []string {
	ips := []string{}
	for _, m := range c.Spec.Machines {
		ips = append(ips, m.IP)
	}
	return ips
}
