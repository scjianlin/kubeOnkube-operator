package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model"
	"github.com/gostship/kunkka/pkg/constants"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *Manager) getClusterComponents(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	ctx := context.Background()
	clusName := c.Param("name")
	cli, err := m.getClient(clusName)
	if err != nil {
		resp.RespError("get client error")
		return
	}

	svc := &corev1.ServiceList{}
	pod := &corev1.PodList{}

	components := make([]model.ComponentStatus, 0)
	for _, ns := range constants.SystemNamespaces { //range namespace
		listOptions := &client.ListOptions{Namespace: ns}
		err = cli.List(ctx, svc, listOptions)
		if err != nil {
			klog.Error(err)
			continue
		}

		for _, service := range svc.Items {

			// skip services without a selector
			if len(service.Spec.Selector) == 0 {
				continue
			}

			component := model.ComponentStatus{
				Name:            service.Name,
				Namespace:       service.Namespace,
				SelfLink:        service.SelfLink,
				Label:           service.Spec.Selector,
				StartedAt:       service.CreationTimestamp.Time,
				HealthyBackends: 0,
				TotalBackends:   0,
			}
			listOptions.LabelSelector = labels.SelectorFromValidatedSet(service.Spec.Selector)

			err = cli.List(ctx, pod, listOptions)
			if err != nil {
				klog.Errorln(err)
				continue
			}

			for _, pod := range pod.Items {
				component.TotalBackends++
				if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(&pod) {
					component.HealthyBackends++
				}
			}

			components = append(components, component)
		}
	}
	resp.RespJson(components)
}

func (m *Manager) getComponentsDetail(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	ctx := context.Background()
	clusName := c.Param("name")
	componets := c.Param("component")
	cli, err := m.getClient(clusName)
	if err != nil {
		resp.RespError("get client error")
		return
	}

	svc := &corev1.ServiceList{}
	pod := &corev1.PodList{}

	components := &model.ComponentStatus{}
	for _, ns := range constants.SystemNamespaces { //range namespace
		listOptions := &client.ListOptions{Namespace: ns}
		err = cli.List(ctx, svc, listOptions)
		if err != nil {
			klog.Error(err)
			continue
		}

		for _, service := range svc.Items {
			if service.Name == componets {

				// skip services without a selector
				if len(service.Spec.Selector) == 0 {
					continue
				}

				component := model.ComponentStatus{
					Name:            service.Name,
					Namespace:       service.Namespace,
					SelfLink:        service.SelfLink,
					Label:           service.Spec.Selector,
					StartedAt:       service.CreationTimestamp.Time,
					HealthyBackends: 0,
					TotalBackends:   0,
				}
				listOptions.LabelSelector = labels.SelectorFromValidatedSet(service.Spec.Selector)

				err = cli.List(ctx, pod, listOptions)
				if err != nil {
					klog.Errorln(err)
					continue
				}

				for _, pod := range pod.Items {
					component.TotalBackends++
					if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(&pod) {
						component.HealthyBackends++
					}
				}
				components = &component
			}
		}
	}
	resp.RespJson(components)

}

func isAllContainersReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.ContainerStatuses {
		if !c.Ready {
			return false
		}
	}
	return true
}

func (m *Manager) getComponentHealth(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	status := model.HealthStatus{}

	clsName := c.Param("name")
	cli, err := m.getClient(clsName)
	if err != nil {
		klog.Error("get cluster error.")
		return
	}

	components, err := m.GetAllComponentsStatus(cli)
	if err != nil {
		klog.Error("get all componentsStatus error", err)
		resp.RespError("get all componentsStatus error")
		return
	}
	status.KubeSphereComponents = components

	// get node status
	ctx := context.Background()
	nodes := &corev1.NodeList{}
	err = cli.List(ctx, nodes)
	if err != nil {
		klog.Error("get node list error:", err)
		resp.RespError("get node list error.")
		return
	}

	totalNodes := 0
	healthyNodes := 0
	for _, nodes := range nodes.Items {
		totalNodes++
		for _, condition := range nodes.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				healthyNodes++
			}
		}
	}
	nodeStatus := model.NodeStatus{TotalNodes: totalNodes, HealthyNodes: healthyNodes}

	status.NodeStatus = nodeStatus

	resp.RespJson(status)
}

func (m *Manager) GetAllComponentsStatus(cli client.Client) ([]model.ComponentStatus, error) {

	components := make([]model.ComponentStatus, 0)
	ctx := context.Background()
	svc := &corev1.ServiceList{}
	pods := &corev1.PodList{}

	var err error
	for _, ns := range constants.SystemNamespaces {
		listOptions := &client.ListOptions{Namespace: ns}
		err := cli.List(ctx, svc, listOptions)
		if err != nil {
			klog.Error(err)
			continue
		}

		for _, service := range svc.Items {

			// skip services without a selector
			if len(service.Spec.Selector) == 0 {
				continue
			}

			component := model.ComponentStatus{
				Name:            service.Name,
				Namespace:       service.Namespace,
				SelfLink:        service.SelfLink,
				Label:           service.Spec.Selector,
				StartedAt:       service.CreationTimestamp.Time,
				HealthyBackends: 0,
				TotalBackends:   0,
			}

			listOptions.LabelSelector = labels.SelectorFromValidatedSet(service.Spec.Selector)

			err = cli.List(ctx, pods, listOptions)
			if err != nil {
				klog.Errorln(err)
				continue
			}

			for _, pod := range pods.Items {
				component.TotalBackends++
				if pod.Status.Phase == corev1.PodRunning && isAllContainersReady(&pod) {
					component.HealthyBackends++
				}
			}
			components = append(components, component)
		}
	}

	return components, err
}
