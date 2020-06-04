package constants

import "github.com/thoas/go-funk"

var (
	Archs            = []string{"amd64", "arm64"}
	Arm64            = "arm64"
	Arm64Variants    = []string{"v8", "unknown"}
	OSs              = []string{"linux"}
	K8sVersions      = []string{"1.14.10", "1.16.6"}
	K8sVersionsWithV = funk.Map(K8sVersions, func(s string) string {
		return "v" + s
	}).([]string)
	K8sVersionConstraint           = ">= 1.10"
	DockerVersions                 = []string{"18.09.9"}
	CNIPluginsVersions             = []string{"v0.8.5"}
	KubeadmVersions                = []string{"v1.15.1"}
	NvidiaDriverVersions           = []string{"440.31"}
	NvidiaContainerRuntimeVersions = []string{"3.1.4"}
)
