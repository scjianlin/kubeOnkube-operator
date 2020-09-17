package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func (m *Manager) getServicePods(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	ctx := context.Background()
	clsName := c.Param("name")
	nsName := c.Param("namespace")
	svcName := c.Param("service")

	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get client error.")
		resp.RespError("get client error.")
		return
	}

	svc := &corev1.Endpoints{}

	err = cli.Get(ctx, types.NamespacedName{
		Name:      svcName,
		Namespace: nsName,
	}, svc)
	if err != nil {
		klog.Error("get Endpoints error")
		resp.RespError("get Endpoints error")
		return
	}
	resp.RespJson(svc)
}

// 获取service 下关联的deployment
func (m *Manager) getServiceDep(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	nsName := c.Param("namespace")
	depName := c.Query("name")

	lab := c.Query("labelSelector")
	label := labels.Set{}
	if lab != "" {
		labs := strings.Split(lab, ",")
		for _, v := range labs {
			setlab := strings.Split(v, "=")
			label[setlab[0]] = setlab[1]
		}
	}

	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get client error.")
		resp.RespError("get client error.")
		return
	}

	ctx := context.Background()
	dep := &v1.DeploymentList{}
	err = cli.List(ctx, dep, &client.ListOptions{
		Namespace:     nsName,
		LabelSelector: label.AsSelector(),
	})

	if err != nil {
		klog.Error("get service deployment error.", err)
		resp.RespError("get service deployment error.")
		return
	}

	de := []v1.Deployment{}
	for _, v := range dep.Items {
		if v.Name == depName {
			de = append(de, v)
		}
	}

	resp.RespJson(de)
}

// 获取service 下关联的daemonset
func (m *Manager) getServiceDae(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	nsName := c.Param("namespace")
	depName := c.Query("name")

	lab := c.Query("labelSelector")
	label := labels.Set{}
	if lab != "" {
		labs := strings.Split(lab, ",")
		for _, v := range labs {
			setlab := strings.Split(v, "=")
			label[setlab[0]] = setlab[1]
		}
	}

	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get client error.")
		resp.RespError("get client error.")
		return
	}

	ctx := context.Background()
	dep := &v1.DaemonSetList{}
	err = cli.List(ctx, dep, &client.ListOptions{
		Namespace:     nsName,
		LabelSelector: label.AsSelector(),
	})

	if err != nil {
		klog.Error("get service DaemonSetList error.", err)
		resp.RespError("get service DaemonSetList error.")
		return
	}

	de := []v1.DaemonSet{}
	for _, v := range dep.Items {
		if v.Name == depName {
			de = append(de, v)
		}
	}

	resp.RespJson(de)
}
