package apimanager

import (
	"context"
	"github.com/gostship/kunkka/pkg/apimanager/healthcheck"
	"github.com/gostship/kunkka/pkg/apimanager/router"
	apiv1 "github.com/gostship/kunkka/pkg/apimanager/v1"
	"github.com/gostship/kunkka/pkg/controllers/apictl"
	"github.com/gostship/kunkka/pkg/controllers/k8smanager"
	"github.com/gostship/kunkka/pkg/gmanager"
	"github.com/gostship/kunkka/pkg/provider/monitoring/prometheus"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	rt.AddRoutes("kapi", v1.Routes())
	apiMgr.Router = rt

	// run api ctrl
	apictl.Add(mgr, &gmanager.GManager{ClusterManager: k8sMgr})

	// set meta monitor
	m, _ := prometheus.NewPrometheus(&prometheus.Options{Endpoint: "http://10.248.225.17:32032/"})
	k8sMgr.AddMonitor("host", m)

	apiMgr.Cluster = k8sMgr
	v1.Cluster = k8sMgr

	err = preStart(k8sMgr)
	if err != nil {
		klog.Error("cluster: host client preStart error:%s", err)
		return nil, err
	}

	return apiMgr, nil
}

func GetClusterLs() map[string]string {
	return map[string]string{
		"ClusterOwner": "kunkka-api",
	}
}

func preStart(cli *k8smanager.ClusterManager) error {
	if err := cli.GetFieldIndexer().IndexField(context.TODO(), &corev1.Pod{}, "spec.nodeName", func(object runtime.Object) []string {
		pod := object.(*corev1.Pod)
		return []string{pod.Status.HostIP}
	}); err != nil {
		klog.Warning("cluster: host add field index pod status.hostIP, err: %#v", err)
		return errors.New("cluster add field index pod spec.nodeName failed")
	} else {
		klog.Warning("########### cluster: host add field index pod status.hostIP, successfully ##################")
	}

	if err := cli.GetFieldIndexer().IndexField(context.TODO(), &corev1.Event{}, "source.host", func(object runtime.Object) []string {
		event := object.(*corev1.Event)
		return []string{event.Source.Host}
	}); err != nil {
		klog.Warning("cluster: host add field index pod source.host, err: %#v", err)
		return errors.New("cluster add field index pod involvedObject.name failed")
	} else {
		klog.Warning("########### cluster: host add field index pod source.host, successfully ##################")
	}

	if err := cli.GetFieldIndexer().IndexField(context.TODO(), &corev1.Event{}, "involvedObject.kind", func(object runtime.Object) []string {
		event := object.(*corev1.Event)
		return []string{event.InvolvedObject.Kind}
	}); err != nil {
		klog.Warning("cluster: host add field index pod involvedObject.kind, err: %#v", err)
		return errors.New("cluster add field index pod involvedObject.kind failed")
	} else {
		klog.Warning("########### cluster: host add field index pod involvedObject.kind, successfully ##################")
	}

	return nil
}
