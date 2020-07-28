package constants

import (
	"github.com/thoas/go-funk"
	"k8s.io/klog"
)

var (
	OSs              = []string{"linux"}
	K8sVersions      = []string{"v1.16.13", "v1.18.5"}
	K8sVersionsWithV = funk.Map(K8sVersions, func(s string) string {
		return "v" + s
	}).([]string)
	K8sVersionConstraint = ">= 1.10"
	DockerVersions       = []string{"18.09.9"}
)

func IsK8sSupport(version string) bool {
	for _, v := range K8sVersions {
		if v == version {
			return true
		}
	}

	klog.Errorf("k8s version only support: %#v", K8sVersions)
	return false
}
