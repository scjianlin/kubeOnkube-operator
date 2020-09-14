package v1

import (
	"context"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	devopsv1 "github.com/gostship/kunkka/pkg/apis/devops/v1"
	"github.com/gostship/kunkka/pkg/util/crdutil"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	v13 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *Manager) addClusterNode(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	cli := m.Cluster.GetClient()
	ctx := context.Background()
	nodeParm := &model.ClusterNode{}
	listMap := []*model.Rack{}

	cni := &devopsv1.ClusterCni{}

	node, err := resp.Bind(nodeParm)
	if err != nil {
		klog.Error("bind http params error: ", err)
		resp.RespError("bind http params error")
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

	for _, rack := range listMap {
		if rack.RackTag == node.(*model.ClusterNode).NodeRack[0] {
			for _, pod := range rack.PodCidr {
				cni = pod //找到node节点的podcidr
				break
			}
		}
	}

	nodeObj, err := crdutil.BuildNodeCrd(node.(*model.ClusterNode), cni)
	if err != nil {
		klog.Error("build node crd cfg error: ", err)
		resp.RespError("build node crd cfg error")
		return
	}

	logger := ctrl.Log.WithValues("cluster", node.(*model.ClusterNode).ClusterName)
	logger.Info("create node reconcile ...")
	for _, obj := range nodeObj {
		err := k8sutil.Reconcile(logger, cli, obj, k8sutil.DesiredStatePresent)
		if err != nil {
			resp.RespError("create node reconcile error")
			return
		}
	}
	resp.RespSuccess(true, "success", "OK", 0)
}

// get Noready machine
func (m *Manager) getNoreadyNode(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	cli := m.Cluster.GetClient()
	ctx := context.Background()
	clusterName := c.Query("clusterName")
	resultList := []*devopsv1.Machine{}
	machine := &devopsv1.MachineList{}

	err := cli.List(ctx, machine)
	if err != nil {
		klog.Error("get list no ready machine err:", err)
		resp.RespError("get list no ready machine err.")
		return
	}

	for _, ma := range machine.Items {
		if ma.Status.Phase != devopsv1.MachineRunning && ma.Spec.ClusterName == clusterName { // 未就绪的节点
			resultList = append(resultList, &ma)
		}
	}
	resp.RespSuccess(true, "success", resultList, len(resultList))
}

// 获取node节点CRD
func (m *Manager) getNodeDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clusterName := c.Param("name")
	nodeName := c.Param("node")

	cli, err := m.getClient(clusterName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	nodes := &corev1.NodeList{}
	result := &corev1.Node{}
	err = cli.List(ctx, nodes)

	if err != nil {
		resp.RespError("get node kind error!")
		return
	}
	for _, node := range nodes.Items {
		if node.Name == nodeName {
			result = &node
			break
		}
	}
	resp.RespJson(result)
}

// 获取node节点pod list
func (m *Manager) getNodePodDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	nodeName := c.Query("nodeName")

	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get clienet error:%s", err)
		resp.RespError("get client error")
		return
	}

	pod := &corev1.PodList{}
	ctx := context.Background()
	if nodeName != "" {
		err = cli.List(ctx, pod, &client.ListOptions{FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName})})
	} else {
		err = cli.List(ctx, pod)
	}
	if err != nil {
		klog.Error("get node pods error")
		resp.RespError("get node pods error")
		return
	}
	resp.RespSuccess(true, "OK", pod.Items, len(pod.Items))
}

// 获取node节点event
func (m *Manager) getNodeEvents(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	sourceHost := c.Query("source.host")
	kindNmae := c.Query("involvedObject.kind")

	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get client error.")
		return
	}

	evList := &corev1.EventList{}
	result := []*corev1.Event{}
	ctx := context.Background()

	err = cli.List(ctx, evList, &client.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"involvedObject.kind": kindNmae}),
	})

	if err != nil {
		klog.Error("get envent list error", err)
		resp.RespError("get event list error")
		return
	}

	for _, ev := range evList.Items {
		if ev.Source.Host == sourceHost {
			result = append(result, &ev)
		}
	}
	resp.RespJson(evList)
}

// 获取集群的deployment信息
func (m *Manager) getDeploymentDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	dep := &v1.DeploymentList{}

	err = cli.List(ctx, dep)
	if err != nil {
		klog.Error("get cluster deployment error.")
		resp.RespError("get cluster deployment error")
		return
	}
	resp.RespJson(dep)
}

// 获取集群的Statefulsets信息
func (m *Manager) getStatefulsetsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	dep := &v1.StatefulSetList{}

	err = cli.List(ctx, dep)
	if err != nil {
		klog.Error("get cluster StatefulSetList error.")
		resp.RespError("get cluster StatefulSetList error")
		return
	}
	resp.RespJson(dep)
}

// 获取集群的Daemonsets信息
func (m *Manager) getDaemonsetsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	dep := &v1.DaemonSetList{}

	err = cli.List(ctx, dep)
	if err != nil {
		klog.Error("get cluster DaemonSetList error.")
		resp.RespError("get cluster DaemonSetList error")
		return
	}
	resp.RespJson(dep)
}

// 获取集群的Job信息
func (m *Manager) getJobsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	job := &v12.JobList{}

	err = cli.List(ctx, job)
	if err != nil {
		klog.Error("get cluster JobList error.")
		resp.RespError("get cluster JobList error")
		return
	}
	resp.RespJson(job)
}

// 获取集群的CronJob信息
func (m *Manager) getCronJobDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	job := &v1beta1.CronJobList{}

	err = cli.List(ctx, job)
	if err != nil {
		klog.Error("get cluster CronJob error.")
		resp.RespError("get cluster CronJob error")
		return
	}
	resp.RespJson(job)
}

// 获取集群Svc信息
func (m *Manager) getServiceDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	svc := &corev1.ServiceList{}

	err = cli.List(ctx, svc)
	if err != nil {
		klog.Error("get cluster ServiceList error.")
		resp.RespError("get cluster ServiceList error")
		return
	}
	resp.RespJson(svc)
}

// 获取集群Svc信息
func (m *Manager) getIngressesDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	ing := &v1beta12.IngressList{}

	err = cli.List(ctx, ing)
	if err != nil {
		klog.Error("get cluster IngressList error.")
		resp.RespError("get cluster IngressList error")
		return
	}
	resp.RespJson(ing)
}

// 获取secrets列表
func (m *Manager) getSecretsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	sec := &corev1.SecretList{}

	err = cli.List(ctx, sec)
	if err != nil {
		klog.Error("get cluster SecretList error.")
		resp.RespError("get cluster SecretList error")
		return
	}
	resp.RespJson(sec)
}

// 获取secrets列表
func (m *Manager) getConfigmapsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	sec := &corev1.ConfigMapList{}

	err = cli.List(ctx, sec)
	if err != nil {
		klog.Error("get cluster ConfigMapList error.")
		resp.RespError("get cluster ConfigMapList error")
		return
	}
	resp.RespJson(sec)
}

// 获取自定义PVC类型
func (m *Manager) getPvcDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}

	ctx := context.Background()
	cus := &corev1.PersistentVolumeClaimList{}

	err = cli.List(ctx, cus)
	if err != nil {
		klog.Error("get cluster PersistentVolumeClaimList error:", err)
		resp.RespError("get cluster PersistentVolumeClaimList error")
		return
	}
	resp.RespJson(cus)
}

// 获取自定义PVC类型
func (m *Manager) getStorageClassesDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}

	ctx := context.Background()
	cus := &v13.StorageClassList{}

	err = cli.List(ctx, cus)
	if err != nil {
		klog.Error("get cluster StorageClassList error:", err)
		resp.RespError("get cluster StorageClassList error")
		return
	}
	resp.RespJson(cus)
}
