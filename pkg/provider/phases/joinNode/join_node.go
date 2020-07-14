package joinNode

import (
	"context"
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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

type Option struct {
	HostIP string
}

func ApplyPodManifest(hostIP string, pathName string, podManifest string, fileMaps map[string]string) error {
	option := &Option{
		HostIP: hostIP,
	}

	serialized, err := template.ParseString(podManifest, option)
	if err != nil {
		return err
	}

	// obj, err := k8sutil.UnmarshalFromYaml([]byte(podManifest), corev1.SchemeGroupVersion)
	// if err != nil {
	// 	return fmt.Errorf("node: %s unmarshal failed err: %v", s.HostIP(), err)
	// }
	//
	// var pod *corev1.Pod
	// switch obj.(type) {
	// case *corev1.Pod:
	// 	pod = obj.(*corev1.Pod)
	// default:
	// 	return fmt.Errorf("unknown type")
	// }
	//
	// switch pod.Spec.Containers[0].Name {
	// case "etcd":
	// case "kube-apiserver":
	// case "kube-controller-manager":
	// case "kube-scheduler":
	// }
	//
	// serialized, err := k8sutil.MarshalToYaml(pod, corev1.SchemeGroupVersion)
	// if err != nil {
	// 	return errors.Wrapf(err, "node: %s failed to marshal manifest for %s to YAML", s.HostIP(), pathName)
	// }

	// klog.Infof("node: %s start write [%s]: %s", s.HostIP(), pathName, string(serialized))
	// err = s.WriteFile(bytes.NewReader(serialized), pathName)
	// if err != nil {
	// 	return errors.Wrapf(err, "node: %s failed to write manifest for %s ", s.HostIP(), pathName)
	// }

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

func JoinMasterNode(hostIP string, c *common.Cluster, isMaster bool, fileMaps map[string]string) error {
	if !isMaster {
		fileMaps[constants.CACertName] = string(c.ClusterCredential.CACert)
		return nil
	}

	certsCfgMap := &corev1.ConfigMap{}
	err := c.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Cluster.Namespace, Name: constants.KubeApiServerCerts}, certsCfgMap)
	if err != nil {
		return errors.Wrapf(err, "failed get KubeApiServerCerts err: %v", err)
	}

	for pathName, va := range certsCfgMap.BinaryData {
		fileMaps[pathName] = string(va)
	}

	manifestsCfgMap := &corev1.ConfigMap{}
	err = c.Client.Get(context.TODO(), types.NamespacedName{Namespace: c.Cluster.Namespace, Name: constants.KubeMasterManifests}, manifestsCfgMap)
	if err != nil {
		return errors.Wrapf(err, "failed get KubeMasterManifests err: %v", err)
	}

	for pathName, va := range manifestsCfgMap.Data {
		ApplyPodManifest(hostIP, pathName, va, fileMaps)
	}

	return nil
}

func JoinNodePhase(s ssh.Interface, cfg *config.Config, c *common.Cluster, apiserver string, isMaster bool) error {
	hostIP := s.HostIP()
	fileMaps := make(map[string]string)
	err := JoinMasterNode(hostIP, c, isMaster, fileMaps)
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
		klog.V(4).Infof("node: %s start write [%s]: %s", hostIP, pathName, va)
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
