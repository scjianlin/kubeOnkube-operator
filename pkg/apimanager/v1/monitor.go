package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/gostship/kunkka/pkg/apimanager/model/monit"
	"github.com/gostship/kunkka/pkg/provider/monitoring"
	"github.com/gostship/kunkka/pkg/util/responseutil"
	"regexp"
)

func (m *Manager) getNodeMonitor(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	params := parseRequestParams(c)
	opt, err := makeQueryOptions(m, params, monitoring.LevelNode)
	if err != nil {
		resp.RespError("make node query options error")
		return
	}
	m.handleNameMetricsQuery(c, opt)
}

func (m *Manager) getClusterMonitor(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}
	params := parseRequestParams(c)
	opt, err := makeQueryOptions(m, params, monitoring.LevelCluster)
	if err != nil {
		resp.RespError("make cluster options error")
		return
	}
	m.handleNameMetricsQuery(c, opt)
}

func (m *Manager) getApiserverMonitor(c *gin.Context) {
	resp := responseutil.Gin{Ctx: c}

	params := parseRequestParams(c)
	opt, err := makeQueryOptions(m, params, monitoring.LevelComponent)
	if err != nil {
		resp.RespError("make comonent options error")
		return
	}
	m.handleNameMetricsQuery(c, opt)
}

func (m *Manager) handleNameMetricsQuery(c *gin.Context, q queryOptions) {
	resp := responseutil.Gin{Ctx: c}
	var res monit.Metrics
	var metrics []string
	for _, metric := range q.namedMetrics {
		ok, _ := regexp.MatchString(q.metricFilter, metric)
		if ok {
			metrics = append(metrics, metric)
		}
	}
	if len(metrics) == 0 {
		resp.RespSuccess(true, "OK", res, 0)
		return
	}

	if q.isRangeQuery() {
		res.Results = m.Monitor.GetNamedMetricsOverTime(metrics, q.start, q.end, q.step, q.option)
	} else {
		res.Results = m.Monitor.GetNamedMetrics(metrics, q.time, q.option)
		if q.shouldSort() {
			res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
		}
	}
	resp.RespSuccess(true, "OK", res, len(res.Results))
}
