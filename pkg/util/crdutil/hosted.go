package crdutil

var HostedTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    name: {{ .clusterName }}
  name: {{ .clusterName }}
  annotations:
    k8s.io/action: EnsureCni,EnsureExtKubeconfig,EnsureMetricsServer
---
apiVersion: devops.gostship.io/v1
kind: Cluster
metadata:
  name: {{ .clusterName }}
  namespace: {{ .clusterName }}
spec:
  pause: false
  tenantID: kunkka
  displayName: {{ .clusterName }}
  type: {{ .clusterType }}
  version: {{ .clusterVersion }}
  networkType: eth0
  clusterCIDR: 10.27.184.0/21
  serviceCIDR: 172.27.248.0/22
  dnsDomain: cluster.local
  publicAlternativeNames:
    - idct1-cluster.dke.k8s.io
  features:
    ipvs: true
    internalLB: truecat /in
    enableMasterSchedule: true
  properties:
    maxNodePodNum: 128
  machines:
	{{ range $i, $v := .clusterIp }}
    - ip: v
      port: 22
      username: root
      password: "123456"
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
    version: {{ .dockerVersion }} 
`

//func BuildHostedCrd(cfg *config.Config, c *common.Cluster) ([]runtime.Object, error) {
//
//}
