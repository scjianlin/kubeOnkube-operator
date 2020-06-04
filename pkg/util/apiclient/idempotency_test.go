package apiclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
)

func TestCreateOrUpdateNamespace(t *testing.T) {
	cfg, err := clientcmd.BuildConfigFromFlags("", "/Users/chenglong/.kube/config")
	assert.Nil(t, err)
	client := kubernetes.NewForConfigOrDie(cfg)
	CreateOrUpdateNamespace(context.Background(), client, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a",
		},
	})
}
