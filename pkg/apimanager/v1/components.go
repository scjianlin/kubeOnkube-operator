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

	ns := &corev1.NamespaceList{}
	svc := &corev1.ServiceList{}
	pod := &corev1.PodList{}
	err = cli.List(ctx, ns)
	if err != nil {
		resp.RespError("get all ns error")
		return
	}

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

func isAllContainersReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.ContainerStatuses {
		if !c.Ready {
			return false
		}
	}
	return true
}
