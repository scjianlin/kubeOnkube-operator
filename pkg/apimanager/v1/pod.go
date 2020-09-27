package v1

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	websocket2 "github.com/gostship/kunkka/pkg/util/websocket"
	"io/ioutil"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

const (
	namespace        = "kunkka-api"
	deployNameFormat = "kubectl-%s"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (m *Manager) getKubectlPod(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	user := c.Param("user")

	ctx := context.Background()
	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get client error:", err)
		resp.RespError("get client error.")
		return
	}

	dep := &v1.Deployment{}
	err = cli.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      fmt.Sprintf(deployNameFormat, user),
	}, dep)

	if err != nil {
		klog.Error("get deployment error:", err)
		resp.RespError("get deployment error")
		return
	}
	selector := dep.Spec.Selector.MatchLabels
	labeSelector := labels.Set(selector).AsSelector()
	pods := &corev1.PodList{}
	err = cli.List(ctx, pods, &client.ListOptions{
		LabelSelector: labeSelector,
	})
	if err != nil {
		klog.Error("get Pod error:", err)
		resp.RespError("get Pod error")
		return
	}
	var kubectlPodList corev1.Pod
	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				kubectlPodList = pod
				break
			}
		}
	}
	info := model.PodInfo{Namespace: kubectlPodList.Namespace, Pod: kubectlPodList.Name, Container: kubectlPodList.Status.ContainerStatuses[0].Name}
	resp.RespJson(info)
}

func (m *Manager) getTerminalSession(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	nsName := c.Param("namespace")
	podName := c.Param("pod")
	clsName := c.Param("name")

	containerName := c.Query("container")
	shell := c.Query("shell")

	cli, err := m.getClientInterface(clsName)
	if err != nil {
		klog.Error("get client interface error,", err)
		resp.RespError("get client interface error")
		return
	}
	cfg, err := m.getClientRestCfg(clsName)
	if err != nil {
		klog.Error("get client restconfig error,", err)
		resp.RespError("get client restconfig error")
		return
	}

	handle := websocket2.NewTerminaler(cli, &cfg)

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Error("update websocker error", err)
		resp.RespError("update websocket error.")
		return
	}
	//defer ws.Close()
	handle.HandleSession(shell, nsName, podName, containerName, ws)
}

func (m *Manager) getPodDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	clsName := c.Param("name")
	nsName := c.Param("namespace")
	podName := c.Param("pod")

	ctx := context.Background()
	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get client error,", err)
		resp.RespError("get client error.")
		return
	}
	pods := &corev1.Pod{}

	err = cli.Get(ctx, types.NamespacedName{
		Namespace: nsName,
		Name:      podName,
	}, pods)
	if err != nil {
		resp.RespError("get pod error")
		return
	}

	resp.RespJson(pods)
}

func (m *Manager) getPodLogs(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	clsName := c.Param("name")
	nsName := c.Param("namespace")
	podName := c.Param("pod")

	tailLines, _ := strconv.ParseInt(c.DefaultQuery("tail", "1000"), 10, 64)
	follow, _ := strconv.ParseBool(c.DefaultQuery("follow", "false"))
	previous, _ := strconv.ParseBool(c.DefaultQuery("previous", "false"))
	timestamps, _ := strconv.ParseBool(c.DefaultQuery("timestamps", "false"))
	contName := c.DefaultQuery("container", "")

	logOptions := &corev1.PodLogOptions{
		Follow:     follow,
		Previous:   previous,
		Timestamps: timestamps,
		TailLines:  &tailLines,
		Container:  contName,
	}

	clsInterface, err := m.getClientInterface(clsName)
	if err != nil {
		klog.Error("get cluster interface error", err)
		resp.RespError("get cluster interface error")
		return
	}
	req, err := clsInterface.CoreV1().RESTClient().Get().
		Namespace(nsName).
		Name(podName).
		Resource("pods").
		SubResource("log").
		VersionedParams(logOptions, scheme.ParameterCodec).
		Stream(context.TODO())
	if err != nil {
		klog.Error("get pod logs error: ", err)
		resp.RespError("get pod logs error")
		return
	}
	defer req.Close()

	result, err := ioutil.ReadAll(req)
	if err != nil {
		klog.Error("get pod log io read error:", err)
		resp.RespError("get pod log io read error")
		return
	}
	resp.RespJson(string(result))

}
