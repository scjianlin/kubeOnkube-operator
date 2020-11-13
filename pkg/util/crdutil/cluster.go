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
  clusterCIDR: {{ (index $.Cfg 0).ClusterCIDR }}
  serviceCIDR: {{ (index $.Cfg 0).Cni.DefaultRoute }}
  dnsDomain: cluster.local
  publicAlternativeNames:
    - {{ .Cls.ClusterName }}.k8s.example.com
  features:
    ipvs: true
    internalLB: true
    enableMasterSchedule: true
    ha:
      thirdParty:
        vip: {{ .Cls.ClusterName }}.k8s.example.com
        vport: 6443
    hooks:
      cniInstall: dke-cni
      postInstall: addnode
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

var hostedTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  labels:
    name: {{ .Cls.ClusterName }}
  name: {{ .Cls.ClusterName }}
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  namespace: {{ .Cls.ClusterName }}
  name: etcd
  labels:
    app: etcd
spec:
  serviceName: etcd
  replicas: 3
  selector:
    matchLabels:
      app: etcd
  template:
    metadata:
      name: etcd
      labels:
        app: etcd
    spec:
      containers:
        - name: etcd
          image: symcn.tencentcloudcr.com/symcn/kubernetes:{{ .Cls.ClusterVersion }}
          ports:
            - containerPort: 2379
              name: client
            - containerPort: 2380
              name: peer
          env:
            - name: INITIAL_CLUSTER_SIZE
              value: "3"
            - name: CLUSTER_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          resources:
            requests:
              cpu: 100m
              memory: 512Mi
          volumeMounts:
            - name: datadir
              mountPath: /var/run/etcd
          command:
            - /bin/sh
            - -c
            - |
              PEERS="etcd-0=http://etcd-0.etcd:2380,etcd-1=http://etcd-1.etcd:2380,etcd-2=http://etcd-2.etcd:2380"
              exec etcd --name ${HOSTNAME} \
                --listen-peer-urls http://0.0.0.0:2380 \
                --listen-client-urls http://0.0.0.0:2379 \
                --advertise-client-urls http://${HOSTNAME}.etcd:2379 \
                --initial-advertise-peer-urls http://${HOSTNAME}:2380 \
                --initial-cluster-token etcd-cluster-1 \
                --initial-cluster ${PEERS} \
                --initial-cluster-state new \
                --data-dir /var/run/etcd/default.etcd
  volumeClaimTemplates:
    - metadata:
        name: datadir
      spec:
        storageClassName: local-storage
        accessModes:
          - "ReadWriteOnce"
        resources:
          requests:
            storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  namespace: {{ .Cls.ClusterName }}
  name: etcd
  labels:
    app: etcd
spec:
  ports:
    - port: 2380
      name: etcd-server
    - port: 2379
      name: etcd-client
  clusterIP: None
  selector:
    app: etcd
  publishNotReadyAddresses: true
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  namespace: {{ .Cls.ClusterName }}
  name: etcd-pdb
  labels:
    pdb: etcd
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: etcd
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  labels:
    createdBy: controller
  name: local-storage
provisioner: kubernetes.io/no-provisioner
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
---
{{  range $index, $elem := .Cfg }} 
apiVersion: v1
kind: PersistentVolume
metadata:
  name: etcd{{ $index }}-{{ $.Cls.ClusterName }}-lpv
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 10Gi
  local:
    path: /web/{{ $.Cls.ClusterName }}
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - {{ $elem.Machine }}
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  volumeMode: Filesystem
---
{{ end }}
apiVersion: devops.gostship.io/v1
kind: Cluster
metadata:
  name: {{ .Cls.ClusterName }}
  namespace: {{ .Cls.ClusterName }}
  annotations:
    kunkka.io/description: {{ .Cls.Description }}
    k8s.io/apiSvcVip: "10.248.225.11"
    k8s.io/action: EnsureKubeMaster,EnsureExtKubeconfig,EnsureAddons,EnsureCni
  labels:
    cluster-role.kunkka.io/cluster-role: "member"
    cluster.kunkka.io/group: {{ .Cls.ClusterGroup }}
spec:
  pause: false
  tenantID: kunkka
  displayName: demo
  type: {{ .Cls.ClusterType }}
  version: {{ .Cls.ClusterVersion }}
  networkType: eth0
  clusterCIDR: {{ (index $.Cfg 1).ClusterCIDR }}
  serviceCIDR: {{ (index $.Cfg 1).ServiceCIDR }}
  dnsDomain: cluster.local
  publicAlternativeNames:
    - {{ .Cls.ClusterName }}.k8s.example.com
    - kube-apiserver
  features:
    ipvs: true
    internalLB: true
    enableMasterSchedule: true
    ha:
      thirdParty:
        vip: "10.248.225.11"
        vport: 30443
    files:
      - src: "/k8s/bin/k9s"
        dst: "/usr/local/bin/k9s"
    hooks:
      cniInstall: flannel
      postInstall: addnode
  properties:
    maxNodePodNum: 64
  apiServerExtraArgs:
    audit-log-maxage: "30"
    audit-log-maxbackup: "3"
    audit-log-maxsize: "100"
    audit-log-truncate-enabled: "true"
    audit-policy-file: "/etc/kubernetes/audit-policy.yaml"
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
	var data []byte
	var err error
	if cluster.ClusterType == "Baremetal" {
		data, err = template.ParseString(baremetalTemplate, opt)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = template.ParseString(hostedTemplate, opt)
		if err != nil {
			return nil, err
		}
	}

	objs, err := k8sutil.LoadObjs(bytes.NewReader(data))
	if err != nil {
		klog.Errorf("bremetal load objs err: %v", err)
		return nil, err
	}
	return objs, nil
}
