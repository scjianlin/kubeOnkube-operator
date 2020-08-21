package apimanager

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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (m *APIManager) addClusterNode(c *gin.Context) {
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
