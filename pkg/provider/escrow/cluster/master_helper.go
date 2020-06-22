package cluster

import (
	"context"

	"fmt"
	"strings"

	"sort"

	"github.com/gostship/kunkka/pkg/controllers/common"
	"github.com/gostship/kunkka/pkg/util/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KubeApiServer         = "kube-apiserver"
	KubeKubeScheduler     = "kube-scheduler"
	KubeControllerManager = "kube-controller-manager"
	KubeApiServerCerts    = "kube-apiserver-certs"
	KubeApiServerConfig   = "kube-apiserver-config"
	KubeApiServerAudit    = "kube-apiserver-audit"
)

var KubeApiServerLabels = map[string]string{
	"component": KubeApiServer,
}

var KubeKubeSchedulerLabels = map[string]string{
	"component": KubeKubeScheduler,
}

var KubeControllerManagerLabels = map[string]string{
	"component": KubeControllerManager,
}

var Labels = map[string]string{
	"createBy": "controller",
}

type Reconciler struct {
	Obj     *common.Cluster
	dynamic dynamic.Interface
	*Provider
}

// GetHPAReplicaCountOrDefault get desired replica count from HPA if exists, returns the given default otherwise
func GetHPAReplicaCountOrDefault(client client.Client, name types.NamespacedName, defaultReplicaCount int32) int32 {
	var hpa autoscalev2beta1.HorizontalPodAutoscaler
	err := client.Get(context.Background(), name, &hpa)
	if err != nil {
		return defaultReplicaCount
	}

	if hpa.Spec.MinReplicas != nil && hpa.Status.DesiredReplicas < *hpa.Spec.MinReplicas {
		return *hpa.Spec.MinReplicas
	}

	return hpa.Status.DesiredReplicas
}

func (r *Reconciler) apiServerCertSecret() runtime.Object {
	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Immutable:  nil,
		Data:       nil,
		StringData: nil,
		Type:       "",
	}

	return secret
}

func (r *Reconciler) apiServerDeployment() runtime.Object {
	containers := []corev1.Container{}
	vms := []corev1.VolumeMount{
		{
			Name:      KubeApiServerCerts,
			MountPath: "/etc/kubernetes/pki/",
			ReadOnly:  true,
		},
		{
			Name:      KubeApiServerConfig,
			MountPath: "/etc/kubernetes/",
		},
		{
			Name:      KubeApiServerAudit,
			MountPath: "/var/log/kubernetes",
		},
	}
	hostPathType := corev1.HostPathDirectoryOrCreate
	volumes := []corev1.Volume{
		{
			Name: KubeApiServerCerts,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: KubeApiServerCerts,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: KubeApiServerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: KubeApiServerConfig,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: KubeApiServerAudit,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/web/%s/kube-apiserver/audit", r.Obj.Cluster.Name),
					Type: &hostPathType,
				},
			},
		},
	}

	cmds := []string{
		"kube-apiserver",
		"--advertise-address=$(INSTANCE_IP)",
		"--authorization-mode=Node,RBAC",
		"--client-ca-file=/etc/kubernetes/pki/ca.crt",
		"--enable-admission-plugins=NodeRestriction",
		"--enable-bootstrap-token-auth=true",
		"--kubelet-client-certificate=/etc/kubernetes/pki/apiserver-kubelet-client.crt",
		"--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key",
		"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
		"--proxy-client-cert-file=/etc/kubernetes/pki/front-proxy-client.crt",
		"--proxy-client-key-file=/etc/kubernetes/pki/front-proxy-client.key",
		"--requestheader-allowed-names=front-proxy-client",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--requestheader-extra-headers-prefix=X-Remote-Extra-",
		"--requestheader-group-headers=X-Remote-Group",
		"--requestheader-username-headers=X-Remote-User",
		"--secure-port=6443",
		"--service-account-key-file=/etc/kubernetes/pki/sa.pub",
		"--tls-cert-file=/etc/kubernetes/pki/apiserver.crt",
		"--tls-private-key-file=/etc/kubernetes/pki/apiserver.key",
		"--token-auth-file=/etc/kubernetes/pki/known_tokens.csv",
	}

	if r.Obj.Cluster.Spec.APIServerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Obj.Cluster.Spec.APIServerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	svcCidr := "10.96.0.0/16"
	if r.Obj.Cluster.Spec.ServiceCIDR != nil {
		svcCidr = *r.Obj.Cluster.Spec.ServiceCIDR
		cmds = append(cmds, fmt.Sprintf("--service-cluster-ip-range=%s", svcCidr))
	}

	if r.Obj.Cluster.Spec.Etcd != nil && r.Obj.Cluster.Spec.Etcd.External != nil {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", strings.Join(r.Obj.Cluster.Spec.Etcd.External.Endpoints, ",")))
		// tode check
		if strings.Contains(r.Obj.Cluster.Spec.Etcd.External.Endpoints[0], "https") {
			cmds = append(cmds, fmt.Sprintf("--etcd-cafile=%s", r.Obj.Cluster.Spec.Etcd.External.CAFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-certfile=%s", r.Obj.Cluster.Spec.Etcd.External.CertFile))
			cmds = append(cmds, fmt.Sprintf("--etcd-keyfile=%s", r.Obj.Cluster.Spec.Etcd.External.KeyFile))
		}
	} else {
		cmds = append(cmds, fmt.Sprintf("--etcd-servers=%s", "http://etcd-0.etcd:2379,http://etcd-1.etcd:2379,http://etcd-2.etcd:2379"))
	}

	c := corev1.Container{
		Name:            KubeApiServer,
		Image:           r.Provider.Cfg.Registry.ImageFullName(KubeApiServer, r.Obj.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "https",
				ContainerPort: 6443,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.FromInt(6443),
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Obj),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts:             vms,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers = append(containers, c)

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(KubeApiServer, KubeApiServerLabels, r.Obj.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: KubeApiServerLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: KubeApiServerLabels,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Obj.Cluster.Namespace, KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) apiServerSvc() runtime.Object {
	svc := &corev1.Service{
		ObjectMeta: k8sutil.ObjectMeta(KubeApiServer, KubeApiServerLabels, r.Obj.Cluster),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       6443,
					TargetPort: intstr.FromString("https"),
				},
			},
			Selector: KubeApiServerLabels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}

	svc.Annotations["contour.heptio.com/upstream-protocol.tls"] = "443,https"

	return svc
}

func (r *Reconciler) controllerManagerDeployment() runtime.Object {
	containers := []corev1.Container{}
	vms := []corev1.VolumeMount{
		{
			Name:      KubeApiServerCerts,
			MountPath: "/etc/kubernetes/pki/",
			ReadOnly:  true,
		},
		{
			Name:      KubeApiServerConfig,
			MountPath: "/etc/kubernetes/",
		},
	}
	volumes := []corev1.Volume{
		{
			Name: KubeApiServerCerts,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: KubeApiServerCerts,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
		{
			Name: KubeApiServerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: KubeApiServerConfig,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
	}

	cmds := []string{
		"kube-controller-manager",
		"--authentication-kubeconfig=/etc/kubernetes/controller-manager.conf",
		"--authorization-kubeconfig=/etc/kubernetes/controller-manager.conf",
		"--client-ca-file=/etc/kubernetes/pki/ca.crt",
		"--cluster-signing-cert-file=/etc/kubernetes/pki/ca.crt",
		"--cluster-signing-key-file=/etc/kubernetes/pki/ca.key",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--kubeconfig=/etc/kubernetes/controller-manager.conf",
		"--leader-elect=true",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--root-ca-file=/etc/kubernetes/pki/ca.crt",
		"--service-account-private-key-file=/etc/kubernetes/pki/sa.key",
		"--requestheader-client-ca-file=/etc/kubernetes/pki/front-proxy-ca.crt",
		"--use-service-account-credentials=true",
	}

	if r.Obj.Cluster.Spec.ControllerManagerExtraArgs != nil {
		extraArgs := []string{}
		for k, v := range r.Obj.Cluster.Spec.ControllerManagerExtraArgs {
			extraArgs = append(extraArgs, fmt.Sprintf("--%s=%s", k, v))
		}
		sort.Strings(extraArgs)
		cmds = append(cmds, extraArgs...)
	}

	if r.Obj.Cluster.Status.NodeCIDRMaskSize > 0 {
		cmds = append(cmds, "--allocate-node-cidrs=true")
		cmds = append(cmds, fmt.Sprintf("--cluster-cidr=%s", r.Obj.Cluster.Spec.ClusterCIDR))
		cmds = append(cmds, fmt.Sprintf("--cluster-name=%s", r.Obj.Cluster.Name))
		cmds = append(cmds, fmt.Sprintf("--node-cidr-mask-size=%d", r.Obj.Cluster.Status.NodeCIDRMaskSize))
	}

	c := corev1.Container{
		Name:            KubeControllerManager,
		Image:           r.Provider.Cfg.Registry.ImageFullName(KubeControllerManager, r.Obj.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 10252,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "https-healthz",
				ContainerPort: 10257,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.FromInt(10257),
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Obj),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts:             vms,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers = append(containers, c)

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(KubeControllerManager, KubeControllerManagerLabels, r.Obj.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: KubeControllerManagerLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: KubeControllerManagerLabels,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Obj.Cluster.Namespace, KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}

func (r *Reconciler) schedulerDeployment() runtime.Object {
	containers := []corev1.Container{}
	vms := []corev1.VolumeMount{
		{
			Name:      KubeApiServerConfig,
			MountPath: "/etc/kubernetes/",
		},
	}

	volumes := []corev1.Volume{

		{
			Name: KubeApiServerConfig,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: KubeApiServerConfig,
					},
					DefaultMode: k8sutil.IntPointer(420),
				},
			},
		},
	}

	cmds := []string{
		"kube-scheduler",
		"--authentication-kubeconfig=/etc/kubernetes/scheduler.conf",
		"--authorization-kubeconfig=/etc/kubernetes/scheduler.conf",
		"--bind-address=0.0.0.0",
		"--kubeconfig=/etc/kubernetes/scheduler.conf",
		"--leader-elect=true",
	}

	c := corev1.Container{
		Name:            KubeKubeScheduler,
		Image:           r.Provider.Cfg.Registry.ImageFullName(KubeKubeScheduler, r.Obj.Cluster.Spec.Version),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         cmds,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 10251,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "https-healthz",
				ContainerPort: 10259,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/healthz",
					Port:   intstr.FromInt(10259),
					Scheme: corev1.URISchemeHTTPS,
				},
			},

			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			TimeoutSeconds:      15,
			FailureThreshold:    8,
			SuccessThreshold:    1,
		},
		Env: common.ComponentEnv(r.Obj),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("0.1"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},

		VolumeMounts:             vms,
		TerminationMessagePath:   corev1.TerminationMessagePathDefault,
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}

	containers = append(containers, c)

	deployment := &appsv1.Deployment{
		ObjectMeta: k8sutil.ObjectMeta(KubeKubeScheduler, KubeKubeSchedulerLabels, r.Obj.Cluster),
		Spec: appsv1.DeploymentSpec{
			Replicas: k8sutil.IntPointer(3),
			Strategy: common.DefaultRollingUpdateStrategy(),
			Selector: &metav1.LabelSelector{
				MatchLabels: KubeKubeSchedulerLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: KubeKubeSchedulerLabels,
				},
				Spec: corev1.PodSpec{
					Containers:  containers,
					Volumes:     volumes,
					Affinity:    common.ComponentAffinity(r.Obj.Cluster.Namespace, KubeApiServerLabels),
					Tolerations: common.ComponentTolerations(),
				},
			},
		},
	}

	return deployment
}
