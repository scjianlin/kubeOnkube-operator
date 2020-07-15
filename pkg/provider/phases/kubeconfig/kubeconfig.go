package kubeconfig

import (
	"bytes"

	"fmt"

	"path/filepath"

	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/certs"
	"github.com/gostship/kunkka/pkg/util/ssh"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/klog"
)

const (
	additPolicy = `
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
`
)

type Option struct {
	MasterEndpoint string
	ClusterName    string
	CACert         []byte
	Token          string
}

func GetBindPort(obj *devopsv1.Cluster) int {
	bindPort := 6443
	if obj.Spec.Features.HA != nil && obj.Spec.Features.HA.ThirdPartyHA != nil {
		bindPort = int(obj.Spec.Features.HA.ThirdPartyHA.VPort)
	}

	return bindPort
}

func install(s ssh.Interface, option *Option) error {
	config := CreateWithToken(option.MasterEndpoint, option.ClusterName, "kubernetes-admin", option.CACert, option.Token)
	data, err := runtime.Encode(clientcmdlatest.Codec, config)
	if err != nil {
		return err
	}
	err = s.WriteFile(bytes.NewReader(data), "/root/.kube/config") // fixme ssh not support $HOME or ~
	if err != nil {
		return err
	}

	return nil
}

// Install creates all the requested kubeconfig files.
func Install(s ssh.Interface, c *common.Cluster) error {
	option := &Option{
		MasterEndpoint: "https://127.0.0.1:6443",
		ClusterName:    c.Name,
		CACert:         c.ClusterCredential.CACert,
		Token:          *c.ClusterCredential.Token,
	}

	return install(s, option)
}

func InstallNode(s ssh.Interface, option *Option) error {
	return install(s, option)
}

func ApplyKubeletKubeconfig(c *common.Cluster, apiserver string, kubeletNodeAddr string, kubeMaps map[string]string) error {
	if c.ClusterCredential.CACert == nil {
		return fmt.Errorf("ca is nil")
	}

	cfgMaps, err := certs.CreateKubeletKubeConfigFile(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, kubeletNodeAddr, c.Cluster.Name)
	if err != nil {
		klog.Errorf("create kubeconfg err: %+v", err)
		return err
	}

	for noPathFile, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		key := filepath.Join(constants.KubernetesDir, noPathFile)
		kubeMaps[key] = string(by)
	}

	return nil
}

func ApplyMasterKubeconfig(c *common.Cluster, apiserver string) error {
	if c.ClusterCredential.CACert == nil {
		return fmt.Errorf("ca is nil")
	}

	cfgMaps, err := certs.CreateMasterKubeConfigFile(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, c.Cluster.Name)
	if err != nil {
		klog.Errorf("create kubeconfg err: %+v", err)
		return err
	}

	if c.ClusterCredential.KubeData == nil {
		c.ClusterCredential.KubeData = make(map[string]string)
	}

	klog.Infof("[%s/%s] start build kubeconfig ...", c.Cluster.Namespace, c.Cluster.Name)
	for noPathFile, v := range cfgMaps {
		by, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			return err
		}

		key := filepath.Join(constants.KubernetesDir, noPathFile)
		c.ClusterCredential.KubeData[key] = string(by)
	}

	key := filepath.Join(constants.KubernetesDir, "audit-policy.yaml")
	c.ClusterCredential.KubeData[key] = additPolicy

	return nil
}
