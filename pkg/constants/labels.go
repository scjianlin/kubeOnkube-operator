package constants

const (
	ComponentNameCrd = "crds"
	CreatedByLabel   = "k8s.io/created-by"
	CreatedBy        = "operator"

	KubeApiServer         = "kube-apiserver"
	KubeKubeScheduler     = "kube-scheduler"
	KubeControllerManager = "kube-controller-manager"
	KubeApiServerCerts    = "kube-apiserver-certs"
	KubeApiServerConfig   = "kube-apiserver-config"
	KubeApiServerAudit    = "kube-apiserver-audit"
	KubeMasterManifests   = "kube-master-manifests"
)

const (
	ClusterAnnotationAction = "k8s.io/action"
	ClusterPhaseRestore     = "k8s.io/phaseRestore"
	ClusterApiSvcType       = "k8s.io/apiSvcType"
	ClusterApiSvcVip        = "k8s.io/apiSvcVip"
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

var CtrlLabels = map[string]string{
	"createBy": "controller",
}

func GetAnnotationKey(annotation map[string]string, key string) string {
	if annotation == nil {
		return ""
	}

	return annotation[key]
}
