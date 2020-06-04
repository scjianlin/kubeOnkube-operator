package kubeadm

import (
	"fmt"
	"testing"

	kubeadmv1beta2 "github.com/gostship/kunkka/pkg/apis/kubeadm/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfig_Marshal(t *testing.T) {
	c := &Config{
		InitConfiguration: &kubeadmv1beta2.InitConfiguration{
			TypeMeta:       metav1.TypeMeta{},
			CertificateKey: "a",
		},
	}
	data, err := c.Marshal()
	fmt.Println(string(data), err)
}
