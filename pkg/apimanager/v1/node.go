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
	"github.com/gostship/kunkka/pkg/util/metautil"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	"github.com/gostship/kunkka/pkg/util/workload"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	v13 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// 添加集群节点, 根据前端提交的机器和cni配置,组装Machine CRD然后提交到集群。
func (m *Manager) addClusterNode(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	cli := m.Cluster.GetClient()
	ctx := context.Background()
	nodeParm := &model.ClusterNode{}
	listMap := []*model.Rack{}

	//cni := &devopsv1.ClusterCni{}

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

	cniList := []*devopsv1.ClusterCni{}

	listRack := []*model.Rack{}

	for _, host := range node.(*model.ClusterNode).AddressList {
		listRack = append(listRack, m.getHostRack(host, c, "Baremetal"))
	}

	for i, rack := range listRack {
		if metautil.StringofContains(rack.RackTag, node.(*model.ClusterNode).NodeRack) {
			for _, pod := range rack.PodCidr {
				if pod.ID == node.(*model.ClusterNode).PodPool[i] {
					cniList = append(cniList, pod)
				}
			}
		}
	}

	nodeObj, err := crdutil.BuildNodeCrd(node.(*model.ClusterNode), cniList)
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
	resp.RespSuccess(true, "OK", result, len(result))
}

// 获取pod事件
func (m *Manager) getNsPodEvents(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	nsName := c.Param("namespace")

	kindNmae := c.Query("involvedObject.kind")
	objName := c.Query("involvedObject.name")

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
		klog.Error("get pod envent list error", err)
		resp.RespError("get pod event list error")
		return
	}

	for _, ev := range evList.Items {
		if ev.InvolvedObject.Namespace == nsName && ev.InvolvedObject.Name == objName {
			result = append(result, &ev)
		}
	}
	resp.RespSuccess(true, "OK", result, len(result))
}

// 获取集群的deployment信息
func (m *Manager) getDeploymentList(c *gin.Context) {
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

// 获取集群的deployment信息
func (m *Manager) getDeploymentDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	work := c.Param("workload")
	nsName := c.Param("namespace")

	k, ok := workload.Kind["deployment"]
	if !ok {
		klog.Error("get kind error")
		resp.RespError("get kind error")
		return
	}
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()

	err = cli.Get(ctx, types.NamespacedName{
		Namespace: nsName,
		Name:      work,
	}, k)
	if err != nil {
		klog.Error("get cluster workload error.")
		resp.RespError("get cluster workload error")
		return
	}
	resp.RespJson(k)
}

// 获取集群的deployment信息
func (m *Manager) getStatefulsetsWorkLoad(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	work := c.Param("workload")
	nsName := c.Param("namespace")

	k, ok := workload.Kind["statefulsets"]
	if !ok {
		klog.Error("get kind error")
		resp.RespError("get kind error")
		return
	}
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()

	err = cli.Get(ctx, types.NamespacedName{
		Namespace: nsName,
		Name:      work,
	}, k)
	if err != nil {
		klog.Error("get cluster workload error.")
		resp.RespError("get cluster workload error")
		return
	}
	resp.RespJson(k)
}

// 获取集群的deployment信息
func (m *Manager) getDeploymentReplicaset(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	nsName := c.Param("namespace")
	SelectLabel := c.DefaultQuery("labelSelector", "")
	lab := labels.Set{}
	if SelectLabel != "" {
		label := strings.Split(SelectLabel, ",")
		for _, v := range label {
			setlab := strings.Split(v, "=")
			lab[setlab[0]] = setlab[1]
		}
	}

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	repList := &v1.ReplicaSetList{}

	err = cli.List(ctx, repList, &client.ListOptions{
		Namespace:     nsName,
		LabelSelector: lab.AsSelector(),
	})
	if err != nil {
		klog.Error("get cluster ReplicaSetList error.")
		resp.RespError("get cluster ReplicaSetList error")
		return
	}

	resp.RespSuccess(true, "OK", repList.Items, len(repList.Items))
}

// 获取集群的ControllerRevisions 信息
func (m *Manager) getControllerRevisions(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	nsName := c.Param("namespace")
	SelectLabel := c.DefaultQuery("labelSelector", "")
	lab := labels.Set{}
	if SelectLabel != "" {
		label := strings.Split(SelectLabel, ",")
		for _, v := range label {
			setlab := strings.Split(v, "=")
			lab[setlab[0]] = setlab[1]
		}
	}
	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	repList := &v1.ControllerRevisionList{}

	err = cli.List(ctx, repList, &client.ListOptions{
		Namespace:     nsName,
		LabelSelector: lab.AsSelector(),
	})
	if err != nil {
		klog.Error("get cluster ReplicaSetList error.")
		resp.RespError("get cluster ReplicaSetList error")
		return
	}

	resp.RespSuccess(true, "OK", repList.Items, len(repList.Items))
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
func (m *Manager) getDaemonsetsList(c *gin.Context) {
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

// 获取DaemonsetsDetail 信息
func (m *Manager) getDaemonsetsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	nsName := c.Param("namespace")
	appName := c.Param("workload")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	Daemon := &v1.DaemonSet{}

	err = cli.Get(ctx, types.NamespacedName{
		Namespace: nsName,
		Name:      appName,
	}, Daemon)
	if err != nil {
		klog.Error("get cluster DaemonSet error.")
		resp.RespError("get cluster DaemonSet error")
		return
	}
	resp.RespJson(Daemon)
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
func (m *Manager) getServiceList(c *gin.Context) {
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
func (m *Manager) getServiceDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	ctx := context.Background()
	clusName := c.Param("name")
	nsName := c.Param("namespace")
	svcName := c.Param("service")
	cli, err := m.getClient(clusName)
	if err != nil {
		resp.RespError("get client error")
		return
	}
	svc := &corev1.Service{}
	err = cli.Get(ctx, types.NamespacedName{
		Namespace: nsName,
		Name:      svcName,
	}, svc)
	if err != nil {
		klog.Error("get service error:", err)
		resp.RespError("get service error")
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
