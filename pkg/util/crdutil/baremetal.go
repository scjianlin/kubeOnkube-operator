package crdutil

import (
	"bytes"
	"fmt"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/template"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

var baremetalTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    name: {{ .Cls.ClusterName }}
  name: {{ .Cls.ClusterName }}
  annotations:
    k8s.io/action: EnsureCni,EnsureExtKubeconfig,EnsureMetricsServer
---
apiVersion: devops.gostship.io/v1
kind: Cluster
metadata:
  name: {{ .Cls.ClusterName }}
  namespace: {{ .Cls.ClusterName }}
  annotations:
    kunkka.io/description: {{ .Cls.Description }}
  labels:
    cluster-role.kunkka.io/cluster-role: "member"
    cluster.kunkka.io/group: {{ .Cls.ClusterGroup }}
spec:
  pause: false
  tenantID: kunkka
  displayName: {{ .Cls.ClusterName }}
  type: {{ .Cls.ClusterType }}
  version: {{ .Cls.ClusterVersion }}
  networkType: eth0
  clusterCIDR: 10.27.184.0/21
  serviceCIDR: 172.27.248.0/22
  dnsDomain: cluster.local
  publicAlternativeNames:
    - idct1-cluster.dke.k8s.io
  features:
    ipvs: true
    internalLB: true
    enableMasterSchedule: true
    hooks:
      cniInstall: dke-cni
  properties:
    maxNodePodNum: 128
  machines:
    {{ range $elem := .Cfg }}
    - ip: {{ $elem.Machine }}
      port: 22
      username: {{  $.Cls.UserName }}
      password: {{ $.Cls.Password }}
      hostCni:
        id: {{ $elem.Cni.ID }}
        subnet: {{ $elem.Cni.Subnet }}
        rangeStart: {{ $elem.Cni.RangeStart }}
        rangeEnd: {{ $elem.Cni.RangeEnd }}
        defaultRoute: {{ $elem.Cni.DefaultRoute }}
        gw: {{ $elem.Cni.GW }}
        useState: 1       
    {{ end }}
  apiServerExtraArgs:
    audit-log-maxage: "30"
    audit-log-maxbackup: "3"
    audit-log-maxsize: "100"
    audit-log-truncate-enabled: "true"
    audit-log-path: "/var/log/kubernetes/k8s-audit.log"
  controllerManagerExtraArgs:
    "bind-address": "0.0.0.0"
  schedulerExtraArgs:
    "bind-address": "0.0.0.0"
  dockerExtraArgs:
    registry-mirrors: https://4xr1qpsp.mirror.aliyuncs.com
    version: {{ .Cls.DockerVersion }}
`

func BuildBremetalCrd(cluster *model.AddCluster, cni []*model.CniOption) ([]runtime.Object, error) {

	type option struct {
		Cls *model.AddCluster
		Cfg []*model.CniOption
	}

	opt := &option{
		Cls: cluster,
		Cfg: cni,
	}

	data, err := template.ParseString(baremetalTemplate, opt)
	if err != nil {
		return nil, err
	}

	fmt.Println("str-crd==>", string(data))

	objs, err := k8sutil.LoadObjs(bytes.NewReader(data))
	if err != nil {
		klog.Errorf("bremetal load objs err: %v", err)
		return nil, err
	}
	return objs, nil
}
