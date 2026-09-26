package observability

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// promMeter 基于 prometheus/client_golang 的默认实现。
// 使用独立 Registry（非全局 DefaultRegisterer），附带 Go runtime/process 采集器
type promMeter struct {
	reg *prometheus.Registry
}

func newPromMeter() *promMeter {
	reg := prometheus.NewRegistry()
	// Go runtime / process 指标白送：go_goroutines、go_memstats_*、process_* 等
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return &promMeter{reg: reg}
}

type promCounter struct {
	vec *prometheus.CounterVec
}

func (c promCounter) Inc(labels ...string)  { c.vec.WithLabelValues(labels...).Inc() }
func (c promCounter) Add(delta float64, labels ...string) {
	c.vec.WithLabelValues(labels...).Add(delta)
}

type promGauge struct {
	vec *prometheus.GaugeVec
}

func (g promGauge) Set(v float64, labels ...string) { g.vec.WithLabelValues(labels...).Set(v) }
func (g promGauge) Inc(labels ...string)            { g.vec.WithLabelValues(labels...).Inc() }
func (g promGauge) Dec(labels ...string)            { g.vec.WithLabelValues(labels...).Dec() }

type promHistogram struct {
	vec *prometheus.HistogramVec
}

func (h promHistogram) Observe(v float64, labels ...string) {
	h.vec.WithLabelValues(labels...).Observe(v)
}

func (m *promMeter) NewCounter(name, help string, labelNames []string) Counter {
	vec := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: help}, labelNames)
	m.reg.MustRegister(vec)
	return promCounter{vec: vec}
}

func (m *promMeter) NewGauge(name, help string, labelNames []string) Gauge {
	vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, labelNames)
	m.reg.MustRegister(vec)
	return promGauge{vec: vec}
}

func (m *promMeter) NewHistogram(name, help string, buckets []float64, labelNames []string) Histogram {
	vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labelNames)
	m.reg.MustRegister(vec)
	return promHistogram{vec: vec}
}

// ---- /metrics 端点（external 模式，startup 决策注册）----

var (
	promMu       sync.RWMutex
	activeProm   *promMeter
	externalOn   bool
	externalTok  string
)

// PromHTTPHandler 返回 /metrics 抓取 handler；
// 未启用 external 模式时返回 nil（调用方不注册路由）
func PromHTTPHandler() http.Handler {
	promMu.RLock()
	defer promMu.RUnlock()
	if activeProm == nil || !externalOn {
		return nil
	}
	return promhttp.HandlerFor(activeProm.reg, promhttp.HandlerOpts{})
}

// CheckMetricsToken 校验 /metrics 请求的 Bearer token；
// 未配置 token 时放行；配置后 constant-time 比较
func CheckMetricsToken(authHeader string) bool {
	promMu.RLock()
	tok := externalTok
	promMu.RUnlock()
	if tok == "" {
		return true
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(authHeader, prefix)), []byte(tok)) == 1
}
