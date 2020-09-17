package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func (m *Manager) getClusterAllNameSpace(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	nsName := c.Param("namespace")
	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get cluster client error")
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	cms := &corev1.NamespaceList{}
	resultCms := &corev1.Namespace{}

	err = cli.List(ctx, cms)
	if err != nil {
		klog.Error("get cluster namespace error")
		resp.RespError("get cluster namespace error")
		return
	}
	if nsName != "" {
		for _, v := range cms.Items {
			if v.Name == nsName {
				resultCms = &v
			}
		}
		resp.RespJson(resultCms)
	}
	resp.RespJson(cms)
}

func (m *Manager) getClusterNameSpace(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	nsName := c.Param("namespace")
	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get cluster client error")
		resp.RespError("get cluster client error")
		return
	}
	ctx := context.Background()
	cms := &corev1.Namespace{}

	err = cli.Get(ctx, types.NamespacedName{
		Name: nsName,
	}, cms)
	if err != nil {
		klog.Error("get cluster namespace error")
		resp.RespError("get cluster namespace error")
		return
	}
	resp.RespJson(cms)
}

func (m *Manager) getClusterNameSpacePods(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	nsName := c.Param("namespace")
	labelSelector := c.DefaultQuery("labelSelector", "")

	var label []string
	seLabel := labels.Set{}
	if labelSelector != "" {
		label = strings.Split(labelSelector, ",")
		for _, v := range label {
			la := strings.Split(v, "=")
			seLabel[la[0]] = la[1]
		}
	}

	ctx := context.Background()
	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get clienet error:%s", err)
		resp.RespError("get client error")
		return
	}

	pods := &corev1.PodList{}
	if labelSelector == "" {
		err = cli.List(ctx, pods, &client.ListOptions{Namespace: nsName})
	} else {
		err = cli.List(ctx, pods, &client.ListOptions{Namespace: nsName, LabelSelector: seLabel.AsSelector()})
	}
	if err != nil {
		klog.Error("get node pods error")
		resp.RespError("get node pods error")
		return
	}

	resp.RespSuccess(true, "OK", pods.Items, len(pods.Items))
}
