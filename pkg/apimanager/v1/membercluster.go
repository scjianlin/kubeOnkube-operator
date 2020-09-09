package v1

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/util/metautil"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

var (
	MetaClusterName = "host"
)

// get cluster metadata
func (m *Manager) GetMemberMetaData(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clusterName := c.Query("clusterName")

	cli, err := m.getClient(clusterName)
	if err != nil {
		resp.RespError("get cluster client error.")
		return
	}

	ctx := context.Background()

	resultList := &corev1.NamespaceList{}

	err = cli.List(ctx, resultList)

	if err != nil {
		if apierrors.IsNotFound(err) {
			err = errors.New("cluster namespace is not found.")
		}
		klog.Error(err)
		resp.RespError(err.Error())
		return
	}

	resp.RespSuccess(true, "success", resultList.Items, len(resultList.Items))
}

// get cluster role
func (m *Manager) GetClusterRole(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	cls, err := metautil.BuildClusterRole()
	if err != nil {
		resp.RespError(err.Error())
		return
	}
	resp.RespSuccess(true, "success", cls, len(cls))
}

// get cluster all nodes
func (m *Manager) GetNodeCount(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Query("clusterName")

	cli, err := m.getClient(clsName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}

	nodeList := &corev1.NodeList{}
	ctx := context.Background()
	err = cli.List(ctx, nodeList)
	if err != nil {
		resp.RespError("get node list error")
		return
	}
	resp.RespSuccess(true, "success", nodeList.Items, len(nodeList.Items))
}

func (m *Manager) TestGet(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	cliName := c.Query("clusterName")

	cli, err := m.getClient(cliName)
	if err != nil {
		resp.RespError("get cluster client error")
		return
	}

	role := &rbacv1.ClusterRoleList{}
	ctx := context.Background()

	err = cli.List(ctx, role)
	if err != nil {
		resp.RespError("get cluster role error")
		return
	}

	resp.RespSuccess(true, "success", role.Items, len(role.Items))
}
