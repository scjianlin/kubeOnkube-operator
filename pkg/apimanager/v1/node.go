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
	corev1 "k8s.io/api/core/v1"
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
	err = cli.List(ctx, pod, &client.ListOptions{FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName})})
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
