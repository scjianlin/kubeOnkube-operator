package joinNode

import (
	"fmt"
	"os"

	"strings"

	kubeadmv1beta2 "github.com/gostship/kunkka/pkg/apis/kubeadm/v1beta2"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/provider/certs"
	"github.com/gostship/kunkka/pkg/provider/config"
	"github.com/gostship/kunkka/pkg/provider/phases/kubeadm"
	"github.com/gostship/kunkka/pkg/util/pkiutil"
	"github.com/gostship/kunkka/pkg/util/ssh"
	"github.com/gostship/kunkka/pkg/util/template"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

func ApplyPodManifest(hostIP string, c *common.Cluster, cfg *config.Config, pathName string, podManifest string, fileMaps map[string]string) error {
	opt := &kubeadm.Option{
		HostIP:           hostIP,
		Images:           cfg.KubeAllImageFullName(constants.KubernetesAllImageName, c.Cluster.Spec.Version),
		EtcdPeerCluster:  kubeadm.BuildMasterEtcdPeerCluster(c),
		TokenClusterName: c.Cluster.Name,
	}

	serialized, err := template.ParseString(podManifest, opt)
	if err != nil {
		return err
	}

	fileMaps[pathName] = string(serialized)
	return nil
}

func buildKubeletKubeconfig(hostIP string, c *common.Cluster, apiserver string, fileMaps map[string]string) error {
	cfgMaps, err := certs.CreateKubeConfigFiles(c.ClusterCredential.CAKey, c.ClusterCredential.CACert,
		apiserver, hostIP, c.Cluster.Name, pkiutil.KubeletKubeConfigFileName)
	if err != nil {
		klog.Errorf("create node: %s kubelet kubeconfg err: %+v", hostIP, err)
		return err
	}

	var kubeletConf []byte
	for _, v := range cfgMaps {
		data, err := certs.BuildKubeConfigByte(v)
		if err != nil {
			klog.Errorf("covert node: %s kubelet kubeconfg err: %+v", hostIP, err)
			return err
		}

		kubeletConf = data
		break
	}

	if kubeletConf == nil {
		return fmt.Errorf("node: %s can't build kubeletConf", hostIP)
	}

	fileMaps[constants.KubeletKubeConfigFileName] = string(kubeletConf)
	return nil
}

func JoinMasterNode(hostIP string, c *common.Cluster, cfg *config.Config, isMaster bool, fileMaps map[string]string) error {
	if !isMaster {
		fileMaps[constants.CACertName] = string(c.ClusterCredential.CACert)
		return nil
	}

	for pathName, va := range c.ClusterCredential.CertsBinaryData {
		fileMaps[pathName] = string(va)
	}

	for pathName, va := range c.ClusterCredential.KubeData {
		fileMaps[pathName] = va
	}

	for pathName, va := range c.ClusterCredential.ManifestsData {
		ApplyPodManifest(hostIP, c, cfg, pathName, va, fileMaps)
	}

	return nil
}

func JoinNodePhase(s ssh.Interface, cfg *config.Config, c *common.Cluster, apiserver string, isMaster bool) error {
	hostIP := s.HostIP()
	fileMaps := make(map[string]string)
	err := JoinMasterNode(hostIP, c, cfg, isMaster, fileMaps)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed build misc file", hostIP)
	}

	err = buildKubeletKubeconfig(hostIP, c, apiserver, fileMaps)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed build kubelet file", hostIP)
	}

	nodeOpt := &kubeadmv1beta2.NodeRegistrationOptions{
		Name: hostIP,
	}
	flagsEnv := BuildKubeletDynamicEnvFile(cfg.Registry.Prefix, nodeOpt)
	fileMaps[constants.KubeletEnvFileName] = flagsEnv

	kubeletCfg := kubeadm.GetFullKubeletConfiguration(c)
	cfgYaml, err := KubeletMarshal(kubeletCfg)
	if err != nil {
		return errors.Wrapf(err, "node: %s failed marshal kubelet file", hostIP)
	}

	fileMaps[constants.KubeletConfigurationFileName] = string(cfgYaml)
	fileMaps[constants.KubeletServiceRunConfig] = kubeletEnvironmentTemplate

	for pathName, va := range fileMaps {
		klog.V(4).Infof("node: %s start write [%s] ...", hostIP, pathName)
		err = s.WriteFile(strings.NewReader(va), pathName)
		if err != nil {
			return errors.Wrapf(err, "node: %s failed to write for %s ", hostIP, pathName)
		}
	}

	klog.Infof("node: %s restart kubelet ... ", hostIP)
	cmd := fmt.Sprintf("systemctl enable kubelet && systemctl daemon-reload && systemctl restart kubelet")
	exit, err := s.ExecStream(cmd, os.Stdout, os.Stderr)
	if err != nil {
		klog.Errorf("%q %+v", exit, err)
		return err
	}
	return nil
}
