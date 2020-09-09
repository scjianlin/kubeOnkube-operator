package v1

import "github.com/gostship/kunkka/pkg/apimanager/router"

// Routes ...
func (m *Manager) Routes() []*router.Route {
	var routes []*router.Route
	apiRoutes := []*router.Route{
		//{
		//	Method:  "GET",
		//	Path:    "/oauth/authorize",
		//	Handler: m.Oauth.AuthorizeHandler,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapis/config.kubesphere.io/v1alpha2/configs/configz",
		//	Handler: m.Oauth.GetConfigMap,
		//},
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
		//{
		//	Method:  "GET",
		//	Path:    "/apis/cluster/Test",
		//	Handler: m.TestGet,
		//},
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
			Path:    "/apis/cluster/Monitoring/:name/components/:component",
			Handler: m.getApiserverMonitor,
		},
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
