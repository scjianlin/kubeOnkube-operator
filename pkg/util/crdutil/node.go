package crdutil

import (
	"bytes"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/template"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

var nodeTemplate = `
apiVersion: devops.gostship.io/v1
kind: Machine
metadata:
  labels:
    name: 10.248.224.210
    clusterName: idc-test
  name: 10.248.224.210
  namespace: idc-test
spec:
  clusterName: idc-test
  type: Baremetal
  machine:
    ip: 10.248.224.210
    port: 22
    username: root
    password: hNKKTFCAOp6r58A
  feature:
    hooks:
      installType: kubeadm
`

func BuildNodeCrd(node *model.ClusterNode) ([]runtime.Object, error) {
	data, err := template.ParseString(nodeTemplate, node)
	if err != nil {
		return nil, err
	}

	objs, err := k8sutil.LoadObjs(bytes.NewReader(data))
	if err != nil {
		klog.Errorf("bremetal load objs err: %v", err)
		return nil, err
	}
	return objs, nil
}
