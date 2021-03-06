package apiclient

import (
	"context"
	"errors"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetClientset return clientset
func GetClientset(masterEndpoint string, token string, caCert []byte) (*kubernetes.Clientset, error) {
	restConfig := &rest.Config{
		Host:        masterEndpoint,
		BearerToken: token,
		Timeout:     5 * time.Second,
	}
	if caCert != nil {
		restConfig.TLSClientConfig = rest.TLSClientConfig{
			CAData: caCert,
		}
	} else {
		restConfig.TLSClientConfig = rest.TLSClientConfig{
			Insecure: true,
		}
	}

	return kubernetes.NewForConfig(restConfig)
}

// CheckAPIHealthz check healthz
func CheckAPIHealthz(ctx context.Context, client rest.Interface) bool {
	healthStatus := 0
	client.Get().AbsPath("/healthz").Do(ctx).StatusCode(&healthStatus)
	return healthStatus == http.StatusOK
}

// CheckDeployment check Deployment current replicas is equal to desired and all pods are running
func CheckDeployment(ctx context.Context, client kubernetes.Interface, namespace string, name string) (bool, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if *deployment.Spec.Replicas != deployment.Status.Replicas {
		return false, errors.New("deployment.Spec.Replicas != deployment.Status.Replicas")
	}

	labelSelector := metav1.FormatLabelSelector(deployment.Spec.Selector)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		if !IsPodReady(&pod) {
			return false, nil
		}
	}

	return true, nil
}

// CheckStatefulSet check StatefulSet current replicas is equal to desired and all pods are running
func CheckStatefulSet(ctx context.Context, client kubernetes.Interface, namespace string, name string) (bool, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if *statefulSet.Spec.Replicas != statefulSet.Status.Replicas {
		return false, errors.New("statefulSet.Spec.Replicas != statefulSet.Status.Replicas")
	}

	labelSelector := metav1.FormatLabelSelector(statefulSet.Spec.Selector)
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		if !IsPodReady(&pod) {
			return false, nil
		}
	}

	return true, nil
}

// CheckDaemonset check daemonset current replicas is equal to desired and all pods are running
func CheckDaemonset(ctx context.Context, client kubernetes.Interface, namespace string, name string) (bool, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if daemonSet.Status.NumberReady == 0 {
		return false, err
	}
	if daemonSet.Status.DesiredNumberScheduled != daemonSet.Status.NumberReady {
		return false, errors.New("daemonSet.Status.DesiredNumberScheduled != daemonSet.Status.NumberReady")
	}

	return true, nil
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *corev1.Pod) bool {
	return isPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func isPodReadyConditionTrue(status corev1.PodStatus) bool {
	condition := getPodReadyCondition(status)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func getPodReadyCondition(status corev1.PodStatus) *corev1.PodCondition {
	_, condition := getPodCondition(&status, corev1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func getPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return getPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func getPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
