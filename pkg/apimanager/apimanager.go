package apimanager

import (
	//"github.com/gostship/kunkka/pkg/apimanager/config"
	"github.com/gostship/kunkka/pkg/apimanager/healthcheck"
	"github.com/gostship/kunkka/pkg/apimanager/router"
	"github.com/gostship/kunkka/pkg/controllers/apictl"
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/gmanager"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)

// Option ...
type Option struct {
	Threadiness        int
	GoroutineThreshold int
	IsMeta             bool
	ResyncPeriod       time.Duration
	Features           []string

	// use expose /metrics, /read, /live, /pprof, /api.
	HTTPAddr       string
	GinLogEnabled  bool
	GinLogSkipPath []string
	PprofEnabled   bool
}

// APIManager ...
type APIManager struct {
	Opt           *Option
	Cluster       *k8smanager.ClusterManager
	Router        *router.Router
	HealthHandler healthcheck.Handler
}

// DefaultOption ...
func DefaultOption() *Option {
	return &Option{
		HTTPAddr:           ":8888",
		IsMeta:             true,
		GoroutineThreshold: 1000,
		GinLogSkipPath:     []string{"/ready", "/live"},
		GinLogEnabled:      true,
		PprofEnabled:       true,
	}
}

// NewAPIManager ...
func NewAPIManager(mgr manager.Manager, cli k8smanager.MasterClient, opt *Option, componentName string) (*APIManager, error) {
	healthHandler := healthcheck.GetHealthHandler()
	healthHandler.AddLivenessCheck("goroutine_threshold",
		healthcheck.GoroutineCountCheck(opt.GoroutineThreshold))

	apiMgr := &APIManager{
		Opt:           opt,
		HealthHandler: healthHandler,
	}

	klog.Info("start init kunkka api manager... ")
	k8sMgr, err := k8smanager.NewManager(cli)
	if err != nil {
		klog.Fatalf("unable to new k8s manager err: %v", err)
	}

	routerOptions := &router.Options{
		GinLogEnabled:    opt.GinLogEnabled,
		GinLogSkipPath:   opt.GinLogSkipPath,
		MetricsEnabled:   true,
		PprofEnabled:     opt.PprofEnabled,
		Addr:             opt.HTTPAddr,
		MetricsPath:      "metrics",
		MetricsSubsystem: componentName,
	}
	rt := router.NewRouter(routerOptions)

	rt.AddRoutes("kapi", apiMgr.Routes())
	apiMgr.Router = rt

	// run api ctrl
	apictl.Add(mgr, &gmanager.GManager{ClusterManager: k8sMgr})

	apiMgr.Cluster = k8sMgr

	return apiMgr, nil
}

// get cluster client
func (m *APIManager) getClient(cliName string) (client.Client, error) {
	var cli client.Client
	if cliName == MetaClusterName {
		cli = m.Cluster.GetClient()
	} else {
		cls, err := m.Cluster.Get(cliName)
		if err != nil {
			return nil, nil
		}
		cli = cls.Client
	}
	return cli, nil
}

// Routes ...
func (m *APIManager) Routes() []*router.Route {
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

func GetClusterLs() map[string]string {
	return map[string]string{
		"ClusterOwner": "kunkka-api",
	}
}
