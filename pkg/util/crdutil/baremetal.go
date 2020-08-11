package crdutil

import (
	"bytes"
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
    name: {{ .ClusterName }}
  name: {{ .ClusterName }}
  annotations:
    k8s.io/action: EnsureCni,EnsureExtKubeconfig,EnsureMetricsServer
---
apiVersion: devops.gostship.io/v1
kind: Cluster
metadata:
  name: {{ .ClusterName }}
  namespace: {{ .ClusterName }}
  labels:
    cluster-role.kunkka.io/cluster-role: "member"
spec:
  pause: false
  tenantID: kunkka
  displayName: {{ .ClusterName }}
  type: {{ .ClusterType }}
  version: {{ .ClusterVersion }}
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
  properties:
    maxNodePodNum: 128
  machines:
    {{ range $elem := .ClusterIP }}
    - ip: {{ $elem }}
      port: 22
      username: {{  $.UserName }}
      password: {{ $.Password }}
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
    version: {{ .DockerVersion }}
`

func BuildBremetalCrd(cluster *model.AddCluster) ([]runtime.Object, error) {
	data, err := template.ParseString(baremetalTemplate, cluster)
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
