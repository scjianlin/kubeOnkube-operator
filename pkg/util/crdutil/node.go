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
{{ range $index, $element := .Cni }}
apiVersion: devops.gostship.io/v1
kind: Machine
metadata:
  labels:
    name: {{ $element.Machine }}
    clusterName: {{ $.Node.ClusterName }}
  name: {{ $element.Machine }}
  namespace: {{ $.Node.ClusterName }}
spec:
  clusterName: {{ $.Node.ClusterName }}
  type: Baremetal
  machine:
    ip: {{ $element.Machine }}
    port: 22
    username: {{ $.Node.UserName }}
    password: {{ $.Node.Password }}
    hostCni:
      id: {{  $element.Cni.ID }}
      subnet: {{  $element.Cni.Subnet }}
      useState: {{ $element.Cni.UseState }}
      rangeStart: {{ $element.Cni.RangeStart }}
      rangeEnd: {{ $element.Cni.RangeEnd }}
      defaultRoute: {{ $element.Cni.DefaultRoute }}
      gw: {{ $element.Cni.GW }}
      rackTag: {{ $element.Cni.RackTag }}
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

func BuildNodeCrd(node *model.ClusterNode, cni []*model.CniOption) ([]runtime.Object, error) {

	type Option struct {
		Node *model.ClusterNode
		Cni  []*model.CniOption
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
