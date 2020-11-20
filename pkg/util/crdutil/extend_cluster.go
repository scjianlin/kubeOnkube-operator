package crdutil

import (
	"context"
	"errors"
	"fmt"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/controllers/apictl"
	"github.com/gostship/kunkka/pkg/util/template"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var includeTemplate = `
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
  displayName: host
  type: {{ .Cls.ClusterType }}
  version: {{ .Cls.ClusterVersion }}
  machines:
    {{ range $elem := .Cls.ClusterIP }}
    - ip: {{ $elem }}
      port: 22
      username: "root"
      password: "123123"
    {{ end }}
status:
 conditions:
 - lastProbeTime: "2020-11-16T12:22:14Z"
   reason: "Ready"
   status: "True"
   type: "Ready"
   lastTransitionTime: "2020-11-16T12:22:14Z"
   message: "Cluster is available now"
 version: {{ .Cls.ClusterVersion }}
 phase: "Running"
 nodeCount: 3
`

var (
	ConfigMapName = "extend-cluster"
)

func BuildExtendCrd(cluster *model.AddCluster, cli client.Client) error {
	type option struct {
		Cls *model.AddCluster
	}

	opt := &option{
		Cls: cluster,
	}

	data, err := template.ParseString(includeTemplate, opt)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cm := &corev1.ConfigMap{}
	cmName := fmt.Sprintf("extend-%s", cluster.ClusterName)

	// 获取configMap
	err = cli.Get(ctx, types.NamespacedName{Namespace: ConfigMapName, Name: cmName}, cm)
	if err != nil {
		if apierrors.IsNotFound(err) {
			extendCm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: ConfigMapName,
				},
				Data: map[string]string{"List": "", "Cfg": ""},
			}

			// 创建 comfilMap
			err := cli.Create(ctx, extendCm)
			if err != nil {
				klog.Errorf("failed to create rack configMaps, %s", err)
				return errors.New("failed to create rack configMaps.")
			}
			cm = extendCm
		}
	}

	klog.Info("create ConfigMap Name: ", cmName)

	_, err = apictl.Reconciler.AddNewClusters(cluster.ClusterName, cluster.CustomConfig)
	if err != nil {
		klog.Errorf("Add extend cluster errors,")
		return errors.New("Add extend cluster error.")
	}

	// 写入configMap
	cm.Data["List"] = string(data)
	cm.Data["Cfg"] = cluster.CustomConfig

	// 更新configMap
	uerr := cli.Update(ctx, cm)
	if uerr != nil {
		klog.Errorf("failed to update rack configMap.")
		return errors.New("failed to update  rack configMap.")
	}

	return nil
}
