package kubeadm

import (
	"bytes"
	"reflect"

	"github.com/gostship/kunkka/pkg/apis"
	kubeadmv1beta2 "github.com/gostship/kunkka/pkg/apis/kubeadm/v1beta2"
	kubeletv1beta1 "github.com/gostship/kunkka/pkg/apis/kubelet/config/v1beta1"
	kubeproxyv1alpha1 "github.com/gostship/kunkka/pkg/apis/kubeproxy/config/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Config struct {
	InitConfiguration      *kubeadmv1beta2.InitConfiguration
	ClusterConfiguration   *kubeadmv1beta2.ClusterConfiguration
	JoinConfiguration      *kubeadmv1beta2.JoinConfiguration
	KubeletConfiguration   *kubeletv1beta1.KubeletConfiguration
	KubeProxyConfiguration *kubeproxyv1alpha1.KubeProxyConfiguration
}

func (c *Config) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)
	v := reflect.ValueOf(*c)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			continue
		}
		obj, ok := v.Field(i).Interface().(runtime.Object)
		if !ok {
			panic("no runtime.Object")
		}
		gvks, _, err := apis.GetScheme().ObjectKinds(obj)
		if err != nil {
			return nil, err
		}

		yamlData, err := apis.MarshalToYAML(obj, gvks[0].GroupVersion())
		if err != nil {
			return nil, err
		}
		buf.WriteString("---\n")
		buf.Write(yamlData)
	}

	return buf.Bytes(), nil
}
