package workload

import (
	v2 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var Kind = map[string]runtime.Object{
	"deployment":   &v2.Deployment{},
	"statefulsets": &v2.StatefulSet{},
	"services":     &corev1.Service{},
	"pods":         &corev1.PodList{},
	"pod":          &corev1.Pod{},
	"eventList":    &corev1.EventList{},
}
