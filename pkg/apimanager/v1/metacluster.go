package v1

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/crdutil"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/metautil"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
)

var (
	VersionMap = "version"
)

// get cluster Available version
func (m *Manager) GetClusterVersion(c *gin.Context) {
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
func (m *Manager) getClusterList(c *gin.Context) {
	lable := c.Query("labelSelector")

	resp := responseutil.Gin{Ctx: c}

	cli := m.Cluster.GetClient()
	ctx := context.Background()

	clusters := &devopsv1.ClusterList{}
	clusterList := []*devopsv1.Cluster{}

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

	for i := 0; i < len(clusters.Items); i++ {
		clusters.Items[i].Status.NodeCount = m.getCount(clusters.Items[i].Name)
		if lable == "meta" && clusters.Items[i].Labels["cluster-role.kunkka.io/cluster-role"] == lable {
			clusterList = append(clusterList, &clusters.Items[i])
		}
		if lable == "member" && clusters.Items[i].Labels["cluster-role.kunkka.io/cluster-role"] == lable {
			clusterList = append(clusterList, &clusters.Items[i])
		}
	}
	resp.RespSuccess(true, "success", clusterList, len(clusterList))
}

// add member cluster
func (m *Manager) AddCluster(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	newCluster := &model.AddCluster{}
	listRack := []*model.Rack{}

	cli := m.Cluster.GetClient()

	cluster, err := resp.Bind(newCluster)
	if err != nil {
		klog.Error("add cluster bind params error.", err)
		resp.RespError("add cluster faild params.")
		return
	}
	if cluster.(*model.AddCluster).ClusterType == "Baremetal" {
		for _, host := range cluster.(*model.AddCluster).ClusterIP {
			listRack = append(listRack, m.getHostRack(host, c, cluster.(*model.AddCluster).ClusterType))
		}
	} else {
		for _, rack := range cluster.(*model.AddCluster).ClusterRack {
			listRack = append(listRack, m.getHostRack(rack, c, cluster.(*model.AddCluster).ClusterType))
		}
	}

	if cluster.(*model.AddCluster).ClusterType == "Baremetal" && len(listRack) != len(cluster.(*model.AddCluster).ClusterIP) {
		klog.Error("address list lerge rack list!")
		resp.RespError("address list lerge rack list!")
		return
	}

	// 处理集群配置
	cniOptList := []*model.CniOption{}
	for i, rack := range listRack {
		cniOpt := &model.CniOption{}

		if metautil.StringofContains(rack.RackTag, cluster.(*model.AddCluster).ClusterRack) {
			cniOpt.Racks = rack.RackCidr
			cniOpt.ClusterCIDR = rack.ProviderCidr //获取clusterCIDR
			cniOpt.ServiceCIDR = rack.ServiceRoute //获取serviceCIDR
			if cluster.(*model.AddCluster).ClusterType == "Baremetal" {
				for _, machine := range rack.HostAddr {
					if (cluster.(*model.AddCluster).ClusterIP)[i] == machine.IPADDR {
						cniOpt.Machine = machine.IPADDR
					}
				}
				for _, pod := range rack.PodCidr {
					if cluster.(*model.AddCluster).PodPool[i] == pod.ID {
						cniOpt.Cni = pod
					}
				}
			} else {
				// 托管集群获取meta
				for _, machine := range rack.HostAddr {
					if machine.IsMeta == 1 {
						cniOpt.Machine = machine.IPADDR
					}
				}
			}
			cniOptList = append(cniOptList, cniOpt)
			//err = m.UptRackStatePhase(cniOpt.Racks, cniOpt.Machine, cniOpt.Cni.ID, 1) //更新机器/CNI状态为使用状态。
			//klog.Info("update rack state: %s", err)
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
func (m *Manager) GetClusterDetail(c *gin.Context) {
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
func (m *Manager) GetClusterCondition(c *gin.Context) {
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
func (m *Manager) getNodeCondition(c *gin.Context) {
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

func (m *Manager) GetClusterCounts(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clusterName := c.Query("clusterName")

	cli, err := m.getClient(clusterName)
	if err != nil {
		resp.RespError("get cluster client error.")
		return
	}

	// 获取集群版本
	cls, _ := m.getClient("host")

	ctx := context.Background()
	nodes := &corev1.NodeList{}

	cluster := &devopsv1.Cluster{}
	cls.Get(ctx, types.NamespacedName{
		Namespace: clusterName,
		Name:      clusterName,
	}, cluster)

	err = cli.List(ctx, nodes)
	if err != nil {
		resp.RespError("get cluster count error")
		return
	}
	clusterCount := map[string]string{"masterCount": "", "masterWorkerCount": "", "masterVersion": ""}

	masterCount, workCount := 0, 0
	for _, cls := range nodes.Items {
		if _, ok := cls.Labels["node-role.kubernetes.io/master"]; ok {
			masterCount += 1
		} else {
			workCount += 1
		}
	}
	clusterCount["masterWorkerCount"] = strconv.Itoa(workCount)
	clusterCount["masterCount"] = strconv.Itoa(masterCount)
	clusterCount["masterVersion"] = cluster.Spec.Version
	resp.RespSuccess(true, "success", clusterCount, len(clusterCount))
}

func (m *Manager) getCount(name string) int {
	cli, err := m.getClient(name)
	if cli == nil || err != nil { // 如果是创建新的集群,此时client还未缓存,直接返回默认值。
		return 0
	}
	nodes := &corev1.NodeList{}
	ctx := context.Background()
	err = cli.List(ctx, nodes)
	if err != nil {
		return 0
	}
	return len(nodes.Items)
}
