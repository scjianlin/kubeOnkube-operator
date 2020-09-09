package v1

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model/monit"
	"github.com/gostship/kunkka/pkg/provider/monitoring"
	"github.com/gostship/kunkka/pkg/provider/monitoring/prometheus"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"strconv"
	"time"
)

const (
	DefaultStep   = 10 * time.Minute
	DefaultFilter = ".*"
	DefaultOrder  = "desc"
	DefaultPage   = 1
	DefaultLimit  = 5

	ComponentEtcd      = "etcd"
	ComponentAPIServer = "apiserver"
	ComponentScheduler = "scheduler"

	ErrNoHit           = "'end' must be after the namespace creation time."
	ErrParamConflict   = "'time' and the combination of 'start' and 'end' are mutually exclusive."
	ErrInvalidStartEnd = "'start' must be before 'end'."
	ErrInvalidPage     = "Invalid parameter 'page'."
	ErrInvalidLimit    = "Invalid parameter 'limit'."
)

type reqParams struct {
	time             string
	start            string
	end              string
	step             string
	target           string
	order            string
	page             string
	limit            string
	metricFilter     string
	resourceFilter   string
	nodeName         string
	workspaceName    string
	namespaceName    string
	workloadKind     string
	workloadName     string
	podName          string
	containerName    string
	pvcName          string
	storageClassName string
	componentType    string
	expression       string
	metric           string
}

type queryOptions struct {
	metricFilter string
	namedMetrics []string

	start time.Time
	end   time.Time
	time  time.Time
	step  time.Duration

	target     string
	identifier string
	order      string
	page       int
	limit      int

	option monitoring.QueryOption
}

type Context struct {
	responseutil.Gin
}

func (q queryOptions) isRangeQuery() bool {
	return q.time.IsZero()
}

func (q queryOptions) shouldSort() bool {
	return q.target != "" && q.identifier != ""
}

func parseRequestParams(g *gin.Context) reqParams {
	var r reqParams
	r.time = g.DefaultQuery("time", "")
	r.start = g.DefaultQuery("start", "")
	r.end = g.DefaultQuery("end", "")
	r.step = g.DefaultQuery("step", "")
	r.target = g.DefaultQuery("sort_metric", "")
	r.order = g.DefaultQuery("sort_type", "")
	r.page = g.DefaultQuery("page", "")
	r.limit = g.DefaultQuery("limit", "")
	r.metricFilter = g.DefaultQuery("metrics_filter", "")
	r.resourceFilter = g.DefaultQuery("resources_filter", "")
	r.nodeName = g.Param("node")
	r.workspaceName = g.Param("workspace")
	r.namespaceName = g.Param("namespace")
	r.workloadKind = g.Param("kind")
	r.workloadName = g.Param("workload")
	r.podName = g.Param("pod")
	r.containerName = g.Param("container")
	r.pvcName = g.Param("pvc")
	r.storageClassName = g.Param("storageclass")
	r.componentType = g.Param("component")
	r.expression = g.DefaultQuery("expr", "")
	r.metric = g.DefaultQuery("metric", "")
	return r
}

func makeQueryOptions(mgr *Manager, r reqParams, lvl monitoring.Level) (q queryOptions, err error) {
	if r.resourceFilter == "" {
		r.resourceFilter = DefaultFilter
	}

	q.metricFilter = r.metricFilter
	if r.metricFilter == "" {
		q.metricFilter = DefaultFilter
	}

	switch lvl {
	case monitoring.LevelCluster:
		q.option = monitoring.ClusterOption{}
		q.namedMetrics = monit.ClusterMetrics
	case monitoring.LevelNode:
		q.identifier = monit.IdentifierNode
		q.namedMetrics = monit.NodeMetrics
		q.option = monitoring.NodeOption{
			ResourceFilter: r.resourceFilter,
			NodeName:       r.nodeName,
		}
	case monitoring.LevelWorkspace:
		q.identifier = monit.IdentifierWorkspace
		q.namedMetrics = monit.WorkspaceMetrics
		q.option = monitoring.WorkspaceOption{
			ResourceFilter: r.resourceFilter,
			WorkspaceName:  r.workspaceName,
		}
	case monitoring.LevelNamespace:
		q.identifier = monit.IdentifierNamespace
		q.namedMetrics = monit.NamespaceMetrics
		q.option = monitoring.NamespaceOption{
			ResourceFilter: r.resourceFilter,
			WorkspaceName:  r.workspaceName,
			NamespaceName:  r.namespaceName,
		}
	case monitoring.LevelWorkload:
		q.identifier = monit.IdentifierWorkload
		q.namedMetrics = monit.WorkloadMetrics
		q.option = monitoring.WorkloadOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			WorkloadKind:   r.workloadKind,
		}
	case monitoring.LevelPod:
		q.identifier = monit.IdentifierPod
		q.namedMetrics = monit.PodMetrics
		q.option = monitoring.PodOption{
			ResourceFilter: r.resourceFilter,
			NodeName:       r.nodeName,
			NamespaceName:  r.namespaceName,
			WorkloadKind:   r.workloadKind,
			WorkloadName:   r.workloadName,
			PodName:        r.podName,
		}
	case monitoring.LevelContainer:
		q.identifier = monit.IdentifierContainer
		q.namedMetrics = monit.ContainerMetrics
		q.option = monitoring.ContainerOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			PodName:        r.podName,
			ContainerName:  r.containerName,
		}
	case monitoring.LevelPVC:
		q.identifier = monit.IdentifierPVC
		q.namedMetrics = monit.PVCMetrics
		q.option = monitoring.PVCOption{
			ResourceFilter:            r.resourceFilter,
			NamespaceName:             r.namespaceName,
			StorageClassName:          r.storageClassName,
			PersistentVolumeClaimName: r.pvcName,
		}
	case monitoring.LevelComponent:
		q.option = monitoring.ComponentOption{}
		switch r.componentType {
		case ComponentEtcd:
			q.namedMetrics = monit.EtcdMetrics
		case ComponentAPIServer:
			q.namedMetrics = monit.APIServerMetrics
		case ComponentScheduler:
			q.namedMetrics = monit.SchedulerMetrics
		}
	}

	// Parse time params
	if r.start != "" && r.end != "" {
		startInt, err := strconv.ParseInt(r.start, 10, 64)
		if err != nil {
			return q, err
		}
		q.start = time.Unix(startInt, 0)

		endInt, err := strconv.ParseInt(r.end, 10, 64)
		if err != nil {
			return q, err
		}
		q.end = time.Unix(endInt, 0)

		if r.step == "" {
			q.step = DefaultStep
		} else {
			q.step, err = time.ParseDuration(r.step)
			if err != nil {
				return q, err
			}
		}

		if q.start.After(q.end) {
			return q, errors.New(ErrInvalidStartEnd)
		}
	} else if r.start == "" && r.end == "" {
		if r.time == "" {
			q.time = time.Now()
		} else {
			timeInt, err := strconv.ParseInt(r.time, 10, 64)
			if err != nil {
				return q, err
			}
			q.time = time.Unix(timeInt, 0)
		}
	} else {
		return q, errors.Errorf(ErrParamConflict)
	}

	// Ensure query start time to be after the namespace creation time
	if r.namespaceName != "" {
		ns, err := mgr.Cluster.KubeCli.CoreV1().Namespaces().Get(context.Background(), r.namespaceName, v1.GetOptions{})
		if err != nil {
			return q, err
		}
		cts := ns.CreationTimestamp.Time

		// Query should happen no earlier than namespace's creation time.
		// For range query, check and mutate `start`. For instant query, check and mutate `time`.
		// In range query, if `start` and `end` are both before namespace's creation time, it causes no hit.
		if !q.isRangeQuery() {
			if q.time.Before(cts) {
				q.time = cts
			}
		} else {
			if q.start.Before(cts) {
				q.start = cts
			}
			if q.end.Before(cts) {
				return q, errors.New(ErrNoHit)
			}
		}

	}

	// Parse sorting and paging params
	if r.target != "" {
		q.target = r.target
		q.page = DefaultPage
		q.limit = DefaultLimit
		q.order = r.order
		if r.order != monit.OrderAscending {
			q.order = DefaultOrder
		}
		if r.page != "" {
			q.page, err = strconv.Atoi(r.page)
			if err != nil || q.page <= 0 {
				return q, errors.New(ErrInvalidPage)
			}
		}
		if r.limit != "" {
			q.limit, err = strconv.Atoi(r.limit)
			if err != nil || q.limit <= 0 {
				return q, errors.New(ErrInvalidLimit)
			}
		}
	}

	return q, nil
}

func NewMonitor() *prometheus.Prometheus {
	opt := &prometheus.Options{Endpoint: "http://10.248.224.210:32002/"}
	client, err := prometheus.NewPrometheus(opt)
	if err != nil {
		klog.Error("get prometheus client error")
		return nil
	}
	return &client
}
