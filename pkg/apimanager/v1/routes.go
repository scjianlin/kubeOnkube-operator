package v1

import "github.com/gostship/kunkka/pkg/apimanager/router"

// Routes ...
func (m *Manager) Routes() []*router.Route {
	var routes []*router.Route
	apiRoutes := []*router.Route{
		{
			Method:  "GET",
			Path:    "/oauth/authorize",
			Handler: m.AuthorizeHandler,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/users",
			Handler: m.getClusterUser,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/globalroles",
			Handler: m.getGlobalRole,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/configs/oauth",
			Handler: m.getAuthConfig,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/users/:username",
			Handler: m.getUserDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/workspaces",
			Handler: m.getWorkSpace,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/configs/configz",
			Handler: m.getClusterConfig,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getRackCidr",
			Handler: m.GetRackMap,
		},
		{
			Method:  "POST",
			Path:    "/apis/cluster/addRackCidr",
			Handler: m.AddRackCidr,
		},
		{
			Method:  "POST",
			Path:    "/apis/cluster/updateRackCidr",
			Handler: m.UptConfigMap,
		},
		{
			Method:  "DELETE",
			Path:    "/apis/cluster/delRackCidr",
			Handler: m.DelConfigMap,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getPodCidr",
			Handler: m.GetPodCidr,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getClusterVersion",
			Handler: m.GetClusterVersion,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getMetaList",
			Handler: m.getClusterList,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getMemberList",
			Handler: m.getClusterList,
		},
		{
			Method:  "POST",
			Path:    "/apis/cluster/addCluster",
			Handler: m.AddCluster,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getClusterDetail",
			Handler: m.GetClusterDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getMasterRack",
			Handler: m.getMasterRack,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getClusterCondition",
			Handler: m.GetClusterCondition,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getNodeCondition",
			Handler: m.getNodeCondition,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getMemberMeta",
			Handler: m.GetMemberMetaData,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getClusterCounts",
			Handler: m.GetClusterCounts,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getClusterRole",
			Handler: m.GetClusterRole,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getNodeCount",
			Handler: m.GetNodeCount,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Test",
			Handler: m.TestGet,
		},
		{
			Method:  "POST",
			Path:    "/apis/cluster/addClusterNode",
			Handler: m.addClusterNode,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/getNoreadyNode",
			Handler: m.getNoreadyNode,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/nodes",
			Handler: m.getNodeMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/cluster",
			Handler: m.getClusterMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/namespaces/:namespace",
			Handler: m.getClusterNsMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/namespaces/:namespace/pods",
			Handler: m.getClusterNsPodsMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/namespaces/:namespace/workloads/:kind/:workload/pods",
			Handler: m.getClusterNsPodsMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/namespaces/:namespace/pods/:pod",
			Handler: m.getClusterNsPodsMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/namespaces",
			Handler: m.getClusterNsMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/components/:component",
			Handler: m.getApiserverMonitor,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/nodes/:node",
			Handler: m.getNodeDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/components",
			Handler: m.getClusterComponents,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/components/:component",
			Handler: m.getComponentsDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/services/:service",
			Handler: m.getServiceDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/Monitoring/:name/nodes/:node/pods",
			Handler: m.getNodePods,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/pods",
			Handler: m.getNodePodDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/services",
			Handler: m.getServiceList,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/ingresses",
			Handler: m.getIngressesDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/secrets",
			Handler: m.getSecretsDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/configmaps",
			Handler: m.getConfigmapsDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/persistentvolumeclaims",
			Handler: m.getPvcDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/storageclasses",
			Handler: m.getStorageClassesDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/deployments",
			Handler: m.getDeploymentList,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/events",
			Handler: m.getNsPodEvents,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/resource/klusters/:name/namespaces",
			Handler: m.getClusterAllNameSpace,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace",
			Handler: m.getClusterNameSpace,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/pods",
			Handler: m.getClusterNameSpacePods,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/statefulsets/:workload",
			Handler: m.getStatefulsetsWorkLoad,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/deployments/:workload",
			Handler: m.getDeploymentDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/daemonsets/:workload",
			Handler: m.getDaemonsetsDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/statefulsets",
			Handler: m.getStatefulsetsDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/daemonsets",
			Handler: m.getDaemonsetsList,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/jobs",
			Handler: m.getJobsDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/cronjobs",
			Handler: m.getCronJobDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/events",
			Handler: m.getNodeEvents,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/componenthealth",
			Handler: m.getComponentHealth,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/users/:user/kubectl",
			Handler: m.getKubectlPod,
		},
		{
			Method:  "GET",
			Path:    "/apis/clusters/:name/namespaces/:namespace/pods/:pod",
			Handler: m.getTerminalSession,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/users/:user/kubeconfig",
			Handler: m.getKubeConfig,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/pods/:pod",
			Handler: m.getPodDetail,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/pods/:pod/log",
			Handler: m.getPodLogs,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/replicasets",
			Handler: m.getDeploymentReplicaset,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/controllerrevisions",
			Handler: m.getControllerRevisions,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/endpoints/:service",
			Handler: m.getServicePods,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/deployments",
			Handler: m.getServiceDep,
		},
		{
			Method:  "GET",
			Path:    "/apis/cluster/klusters/:name/namespaces/:namespace/statefulsets",
			Handler: m.getServiceDae,
		},
		//
		//{
		//	Method:  "GET",
		//	Path:    "/apis/cluster/watch/klusters/:name/namespaces/:namespaces",
		//	Handler: m.getWatchNameSpace,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/exec",
		//	Handler: m.ExecOnceWithHTTP,
		//	Desc:    ExecOnceWithHTTPDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/services/:appName",
		//	Handler: m.GetServices,
		//	Desc:    GetServicesDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/service/:svcName",
		//	Handler: m.GetServiceInfo,
		//	Desc:    GetServiceInfoDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/deployments/:appName",
		//	Handler: m.GetDeployments,
		//	Desc:    GetDeploymentsDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/deployment/:deployName",
		//	Handler: m.GetDeploymentInfo,
		//	Desc:    GetDeploymentInfoDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/deployments/stat",
		//	Handler: m.GetDeploymentsStat,
		//	Desc:    GetDeploymentsStatDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/pods/:podName/event",
		//	Handler: m.GetPodEvent,
		//	Desc:    GetPodEventDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/events/warning",
		//	Handler: m.GetWarningEvents,
		//	Desc:    GetWarningEventsDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/pod/logfiles",
		//	Handler: m.GetFiles,
		//	Desc:    GetFilesDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/pods/:podName/logs",
		//	Handler: m.HandleLogs,
		//	Desc:    HandleLogsDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/pods/:podName/logs/file",
		//	Handler: m.HandleFileLogs,
		//	Desc:    HandleFileLogsDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/offlineWorkloadDeploy",
		//	Handler: m.HandleOfflineWorkloadDeploy,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/offlinePodAppList/all",
		//	Handler: m.GetAllOfflineApp,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/namespace/:namespace/appname/:appname/offlinepodlist",
		//	Handler: m.GetOfflinePods,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/offlineWorkloadPod/terminal",
		//	Handler: m.GetOfflineLogTerminal,
		//},
	}

	routes = append(routes, apiRoutes...)
	return routes
}
