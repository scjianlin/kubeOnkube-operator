package apimanager

import (
	"context"
	"encoding/json"
	"errors"
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
	ctx := context.Background()
	cli := m.Cluster.GetClient()
	listMap := []*model.Rack{}

	cluster, err := resp.Bind(newCluster)
	if err != nil {
		klog.Error("add cluster bind params error.", err)
		resp.RespError("add cluster faild params.")
		return
	}

	cms := &corev1.ConfigMap{}

	err = cli.Get(ctx, types.NamespacedName{
		Name:      ConfigMapName,
		Namespace: ConfigMapName,
	}, cms)

	if err != nil {
		resp.RespError("not found rack cfg.")
		return
	}

	// 获取confiMap的数据
	data, ok := cms.Data["List"]
	if !ok {
		klog.Info("no ConfigMap list!")
		resp.RespError("no configMap list!")
		return
	}
	// 将yaml转换为json
	yamlToRack, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		klog.Errorf("yamlToJson error", err)
		resp.RespError("yaml to struct error!")
		return
	}
	// 转换为结构体
	err = json.Unmarshal(yamlToRack, &listMap)
	if err != nil {
		klog.Errorf("Unmarshal json err", err)
		resp.RespError("Unmarshal list json error.")
		return
	}
	//cniOpt := &model.CniOption{}
	cniOptList := []*model.CniOption{}
	// 处理集群配置
	for _, rack := range listMap {
		cniOpt := &model.CniOption{}
		if metautil.StringofContains(rack.RackTag, cluster.(*model.AddCluster).ClusterRack) {
			cniOpt.Racks = rack.RackCidr
			for _, machine := range rack.HostAddr {
				if metautil.StringofContains(machine.IPADDR, cluster.(*model.AddCluster).ClusterIP) {
					cniOpt.Machine = machine.IPADDR
				}
			}
			for _, pod := range rack.PodCidr {
				if metautil.StringofContains(pod.ID, cluster.(*model.AddCluster).PodPool) {
					cniOpt.Cni = pod
				}
			}
			cniOptList = append(cniOptList, cniOpt)
		}
	}

	cls, err := crdutil.BuildBremetalCrd(cluster.(*model.AddCluster), cniOptList)
	if err != nil {
		klog.Error("Build Object Bremetal err.", err)
		resp.RespError("Build Object Bremetal err.")
		return
	}

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
			break
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

	resultList := []*model.RuntimeCondition{}

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

	for _, condit := range metautil.ProviderClusterSteps(clusterType) {
		resultList = append(resultList, metautil.ClusterConditionOfContains(cluster.Status.Conditions, condit))
	}
	resp.RespSuccess(true, "success", resultList, len(resultList))
}

// get node getNodeCondition
func (m *APIManager) getNodeCondition(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clusterName := c.Query("clusterName")
	ipAddr := c.Query("ipAddr")

	cli := m.Cluster.GetClient()
	ctx := context.Background()

	resultList := []*model.RuntimeCondition{}

	cluster := &devopsv1.Machine{}

	err := cli.Get(ctx, types.NamespacedName{
		Namespace: clusterName,
		Name:      ipAddr,
	}, cluster)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errors.New("machine is not found.")
		}
		klog.Error(err)
		resp.RespError(err.Error())
		return
	}

	for _, condit := range metautil.ProviderClusterSteps("Machine") {
		resultList = append(resultList, metautil.MachineConditionOfContains(cluster.Status.Conditions, condit))
	}
	resp.RespSuccess(true, "success", resultList, len(resultList))
}

func (m *APIManager) GetClusterCounts(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clusterName := c.Query("clusterName")

	cli := m.getClient(clusterName)

	ctx := context.Background()
	nodes := &corev1.NodeList{}

	err := cli.List(ctx, nodes)
	if err != nil {
		resp.RespError("get cluster count error")
		return
	}
	clusterCount := map[string]int{"masterCount": 0, "masterWorkerCount": 0}

	for _, cls := range nodes.Items {
		if _, ok := cls.Labels["node-role.kubernetes.io/master"]; ok {
			clusterCount["masterCount"] += 1
		} else {
			clusterCount["masterWorkerCount"] += 1
		}

	}
	resp.RespSuccess(true, "success", clusterCount, len(clusterCount))
}
