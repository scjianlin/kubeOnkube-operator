package apimanager

import (
	//"github.com/gostship/kunkka/pkg/apimanager/config"
	"github.com/gostship/kunkka/pkg/apimanager/healthcheck"
	"github.com/gostship/kunkka/pkg/apimanager/router"
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"k8s.io/klog"
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
	//Oauth         *jwt.OauthhMgr
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
func NewAPIManager(cli k8smanager.MasterClient, opt *Option, componentName string) (*APIManager, error) {
	healthHandler := healthcheck.GetHealthHandler()
	healthHandler.AddLivenessCheck("goroutine_threshold",
		healthcheck.GoroutineCountCheck(opt.GoroutineThreshold))

	//New Oauth
	//authOption := authentication.NewAuthenticateOptions()
	//jwtOption := jwt.NewJwtTokenIssuer(authOption)
	//authMgr := &jwt.OauthhMgr{
	//	Options: authOption,
	//	Jwt:     jwtOption,
	//}

	apiMgr := &APIManager{
		Opt:           opt,
		HealthHandler: healthHandler,
		//Oauth:         authMgr,
	}

	klog.Info("start init kunkka api manager ... ")
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
	//rt.AddRoutes("index", rt.DefaultRoutes())
	//rt.AddRoutes("health", healthHandler.Routes())
	rt.AddRoutes("kapi", apiMgr.Routes())
	apiMgr.Router = rt

	apiMgr.Cluster = k8sMgr
	return apiMgr, nil
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

		//{
		//	Method:  "POST",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/app/:appName/restart",
		//	Handler: m.DeletePodByGroup,
		//	Desc:    DeletePodByGroupDesc,
		//},
		//{
		//	Method:  "DELETE",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/pod/:podName",
		//	Handler: m.DeletePodByName,
		//	Desc:    DeletePodByNameDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/namespace/:namespace/pod/:podName",
		//	Handler: m.GetPodByName,
		//	Desc:    GetPodByNameDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/endpoints/:appName",
		//	Handler: m.GetEndpoints,
		//	Desc:    GetEndpointsDesc,
		//},
		//{
		//	Method:  "GET",
		//	Path:    "/kapi/cluster/:name/terminal",
		//	Handler: m.GetTerminal,
		//	Desc:    GetTerminalDesc,
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

func GetClusterLs() map[string]string {
	return map[string]string{
		"ClusterOwner": "kunkka-api",
	}
}
