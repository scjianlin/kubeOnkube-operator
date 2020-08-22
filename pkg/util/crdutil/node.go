package crdutil

import (
	"bytes"
	"fmt"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	v1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
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
    name: {{ index .Node.AddressList 0 }}
    clusterName: {{ .Node.ClusterName }}
  name: {{ index .Node.AddressList 0 }}
  namespace: {{ .Node.ClusterName }}
spec:
  clusterName: {{ .Node.ClusterName }}
  type: Baremetal
  machine:
    ip: {{ index .Node.AddressList 0 }}
    port: 22
    username: {{ .Node.UserName }}
    password: {{ .Node.Password }}
    hostCni:
      id: {{ .Cni.ID }}
      subnet: {{ .Cni.Subnet }}
      useState: {{ .Cni.UseState }}
      rangeStart: {{ .Cni.RangeStart }}
      rangeEnd: {{ .Cni.RangeEnd }}
      defaultRoute: {{ .Cni.DefaultRoute }}
      gw: {{ .Cni.GW }}
      useState: 1
  dockerExtraArgs:
    registry-mirrors: https://4xr1qpsp.mirror.aliyuncs.com
    version: {{ .Node.DockerVersion }}
  feature:
    hooks:
      installType: kubeadm
`

func BuildNodeCrd(node *model.ClusterNode, cni *v1.ClusterCni) ([]runtime.Object, error) {

	type Option struct {
		Node *model.ClusterNode
		Cni  *v1.ClusterCni
	}
	opt := &Option{
		node,
		cni,
	}
	fmt.Println("opt==", opt)
	data, err := template.ParseString(nodeTemplate, opt)
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
