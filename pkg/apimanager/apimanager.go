package apimanager

import (
	"github.com/gostship/kunkka/pkg/apimanager/healthcheck"
	"github.com/gostship/kunkka/pkg/apimanager/router"
	apiv1 "github.com/gostship/kunkka/pkg/apimanager/v1"
	"github.com/gostship/kunkka/pkg/controllers/apictl"
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/gmanager"
	"k8s.io/klog"
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

	v1 := apiv1.Manager{}

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
	monit := apiv1.NewMonitor()

	rt.AddRoutes("kapi", v1.Routes())
	apiMgr.Router = rt

	// run api ctrl
	apictl.Add(mgr, &gmanager.GManager{ClusterManager: k8sMgr})

	apiMgr.Cluster = k8sMgr
	v1.Cluster = k8sMgr
	v1.Monitor = monit
	return apiMgr, nil
}

func GetClusterLs() map[string]string {
	return map[string]string{
		"ClusterOwner": "kunkka-api",
	}
}
