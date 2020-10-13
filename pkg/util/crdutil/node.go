package crdutil

import (
	"bytes"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	v1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/template"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

var nodeTemplate = `
{{ range $index, $element := .Node.AddressList }}
apiVersion: devops.gostship.io/v1
kind: Machine
metadata:
  labels:
    name: {{ $element }}
    clusterName: {{ $.Node.ClusterName }}
  name: {{ $element }}
  namespace: {{ $.Node.ClusterName }}
spec:
  clusterName: {{ $.Node.ClusterName }}
  type: Baremetal
  machine:
    ip: {{ $element }}
    port: 22
    username: {{ $.Node.UserName }}
    password: {{ $.Node.Password }}
    hostCni:
      id: {{  (index $.Cni $index).ID }}
      subnet: {{  (index $.Cni $index).Subnet }}
      useState: {{ (index $.Cni $index).UseState }}
      rangeStart: {{ (index $.Cni $index).RangeStart }}
      rangeEnd: {{ (index $.Cni $index).RangeEnd }}
      defaultRoute: {{ (index $.Cni $index).DefaultRoute }}
      gw: {{ (index $.Cni $index).GW }}
      useState: 1
  dockerExtraArgs:
    registry-mirrors: https://4xr1qpsp.mirror.aliyuncs.com
    version: {{ $.Node.DockerVersion }}
  feature:
    hooks:
      installType: kubeadm
---
{{ end }}
`

func BuildNodeCrd(node *model.ClusterNode, cni []*v1.ClusterCni) ([]runtime.Object, error) {

	type Option struct {
		Node *model.ClusterNode
		Cni  []*v1.ClusterCni
	}
	opt := &Option{
		node,
		cni,
	}
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
