package apimanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/crdutil"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/metautil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	//"fmt"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	//"go/ast"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	//"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	VersionMap = "version"
)

// get cluster Available version
func (m *APIManager) GetClusterVersion(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	cli := m.Cluster.GetClient()
	ctx := context.Background()
	cms := []model.ClusterVersion{}

	cmList := &corev1.ConfigMap{}
	err := cli.Get(ctx, types.NamespacedName{
		Namespace: ConfigMapName,
		Name:      VersionMap,
	}, cmList)

	if err != nil {
		klog.Error("Get ConfigMap error %v: ", err)
		resp.RespError("can't found clusterVersion configMap, please create!")
		return
	}

	data := cmList.Data["List"]

	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yamlToJson error")
		return
	}

	rerr := json.Unmarshal(yamlToRack, &cms)
	if rerr != nil {
		klog.Errorf("failed to Unmarshal err: %v", rerr)
		resp.RespError("failed to Unmarshal error.")
		return
	}

	resp.RespSuccess(true, "success", cms, len(cms))
}

// get list of cluster
func (m *APIManager) GetClusterList(c *gin.Context) {
	lable := c.Query("labelSelector")
	name := c.DefaultQuery("name", "all")

	resp := responseutil.Gin{Ctx: c}

	cli := m.Cluster.GetClient()
	ctx := context.Background()

	clusters := &devopsv1.ClusterList{}
	clusterList := []*devopsv1.Cluster{}
	memberList := []*devopsv1.Cluster{}

	err := cli.List(ctx, clusters)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errors.New("cluster is not found.")
		}
		klog.Error(err)
		resp.RespError(err.Error())
		return
	}
	// append meta cluster
	metaObj, err := metautil.BuildMetaObj()
	if err != nil {
		klog.Error("build meta cluster error")
		resp.RespError("build meta cluster error!")
		return
	}
	clusters.Items = append(clusters.Items, *metaObj)

	tag := false
	if name != "all" {
		tag = true
	}
	for i := 0; i < len(clusters.Items); i++ {
		clusters.Items[i].Status.NodeCount = len(clusters.Items[i].Spec.Machines)
		if clusters.Items[i].Labels["cluster-role.kunkka.io/cluster-role"] == "meta" {
			if tag && clusters.Items[i].Name == name {
				clusterList = append(clusterList, &clusters.Items[i])
			} else {
				clusterList = append(clusterList, &clusters.Items[i])
			}
		} else {
			if tag && clusters.Items[i].Name == name {
				memberList = append(clusterList, &clusters.Items[i])
			} else {
				memberList = append(clusterList, &clusters.Items[i])
			}
		}
	}

	if lable == "meta" {
		resp.RespSuccess(true, "success", clusterList, len(clusterList))
	} else {
		resp.RespSuccess(true, "success", memberList, len(memberList))
	}
}

// add member cluster
func (m *APIManager) AddCluster(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	newCluster := &model.AddCluster{}
	cluster, err := resp.Bind(newCluster)
	if err != nil {
		klog.Error("add cluster bind params error.", err)
		resp.RespError("add cluster faild params.")
		return
	}

	cls, err := crdutil.BuildBremetalCrd(cluster.(*model.AddCluster))
	if err != nil {
		klog.Error("Build Object Bremetal err.", err)
		resp.RespError("Build Object Bremetal err.")
		return
	}

	cli := m.Cluster.GetClient()

	logger := ctrl.Log.WithValues("cluster", cluster.(*model.AddCluster).ClusterName)
	logger.Info("create cluster reconcile ...")
	for _, obj := range cls {
		err := k8sutil.Reconcile(logger, cli, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			resp.RespError("create cluster reconcile error")
			return
		}
	}
	resp.RespSuccess(true, "success", "OK", 0)
}

// get meta cluster detail
func (m *APIManager) GetClusterDetail(c *gin.Context) {
	name := c.Query("name")

	resp := responseutil.Gin{Ctx: c}

	cli := m.Cluster.GetClient()
	ctx := context.Background()

	clusters := &devopsv1.ClusterList{}
	clusterDetail := &devopsv1.Cluster{}

	err := cli.List(ctx, clusters)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errors.New("cluster is not found.")
		}
		klog.Error(err)
		resp.RespError(err.Error())
		return
	}
	// append meta cluster
	metaObj, err := metautil.BuildMetaObj()
	if err != nil {
		klog.Error("build meta cluster error")
		resp.RespError("build meta cluster error!")
		return
	}
	clusters.Items = append(clusters.Items, *metaObj)

	for _, cls := range clusters.Items {
		if cls.Name == name {
			clusterDetail = &cls
		}
	}
	resp.RespSuccess(true, "success", clusterDetail, 1)
}

// get cluster conditions
func (m *APIManager) GetClusterCondition(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clusterName := c.Query("clusterName")
	clusterType := c.Query("clusterType")

	cli := m.Cluster.GetClient()
	ctx := context.Background()

	resultList := []*model.ClusterCondition{}

	cluster := &devopsv1.Cluster{}

	err := cli.Get(ctx, types.NamespacedName{
		Namespace: clusterName,
		Name:      clusterName,
	}, cluster)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errors.New("cluster is not found.")
		}
		klog.Error(err)
		resp.RespError(err.Error())
		return
	}

	for _, condit := range providerSteps(clusterType) {
		fmt.Println("condtion==>", condit)
		resultList = append(resultList, metautil.ConditionOfContains(cluster.Status.Conditions, condit))
	}
	fmt.Println("resut-->", resultList)
	resp.RespSuccess(true, "success", resultList, len(resultList))
}

func providerSteps(cType string) []*model.ClusterCondition {
	condition := map[string][]*model.ClusterCondition{
		"Baremetal": {
			&model.ClusterCondition{
				Type: "EnsureSystem",
				Name: "初始化操作系统",
			},
			&model.ClusterCondition{
				Type: "EnsureCerts",
				Name: "生成集群证书",
			},
			&model.ClusterCondition{
				Type: "EnsureKubeadmInitEtcdPhase",
				Name: "初始化ETCD集群",
			},
			&model.ClusterCondition{
				Type: "EnsureJoinControlePlane",
				Name: "安装集群组件",
			},
			&model.ClusterCondition{
				Type: "EnsureApplyControlPlane",
				Name: "初始化集群",
			},
		},
		"Hosted": {
			&model.ClusterCondition{
				Type: "EnsureSystem",
				Name: "初始化操作系统",
			},
			&model.ClusterCondition{
				Type: "EnsureCerts",
				Name: "生成集群证书",
			},
			&model.ClusterCondition{
				Type: "EnsureKubeadmInitEtcdPhase",
				Name: "初始化ETCD集群",
			},
			&model.ClusterCondition{
				Type: "EnsureJoinControlePlane",
				Name: "安装集群组件",
			},
			&model.ClusterCondition{
				Type: "EnsureApplyControlPlane",
				Name: "初始化集群",
			},
		},
	}
	return condition[cType]
}
