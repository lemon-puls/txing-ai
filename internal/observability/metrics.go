package observability

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"txing-ai/internal/global"
)

var (
	mu             sync.RWMutex
	activeMeter    Meter                  = nopMeter{}
	activeGatherer prometheus.Gatherer    // prom 实现时非 nil
	activeStore    *Store
)

// ---- 指标定义（Setup 时注册；初始为 nop，未启用时零开销）----

var (
	metricHTTPRequests Counter
	metricHTTPDuration Histogram
	metricLLMRequests  Counter
	metricLLMErrors    Counter
	metricLLMDuration  Histogram
	metricLLMTTFT      Histogram
	metricLLMTokens    Counter
	metricLLMActive    Gauge
	metricChatMessages Counter
	metricWSActive     Gauge
	metricWFNodes      Counter
	metricWFNodeDur    Histogram
)

// 时延分桶（秒）
var (
	httpDurationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	llmDurationBuckets  = append(append([]float64{}, httpDurationBuckets...), 30, 60, 120)
	llmTTFTBuckets      = []float64{.05, .1, .25, .5, 1, 2, 5, 10, 30}
	wfNodeDurBuckets    = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60}
)

// Setup 构建指标实现并注册全部指标（app.New 启动时调用一次）。
// 无论配置 enabled 与否都构建 prom 实现（开销可忽略），
// 使热改配置 enabled=true 时无需重启即完整生效
func Setup(cfg *global.ObservabilityConfig) {
	m := newPromMeter()

	mu.Lock()
	activeMeter = m
	activeGatherer = m.reg
	activeStore = NewStore(slotIntervalConfig(), retentionConfig(), maxSeriesConfig())
	mu.Unlock()

	// external /metrics 端点开关（startup 决策，改配置需重启）
	promMu.Lock()
	activeProm = m
	externalOn = false
	externalTok = ""
	if cfg != nil && cfg.External != nil {
		externalOn = cfg.External.Enabled
		externalTok = cfg.External.BearerToken
	}
	promMu.Unlock()

	metricHTTPRequests = m.NewCounter("txing_http_requests_total",
		"Total HTTP requests (route is the gin route template)", []string{"route", "method", "code"})
	metricHTTPDuration = m.NewHistogram("txing_http_request_duration_seconds",
		"HTTP request latency in seconds", httpDurationBuckets, []string{"route"})

	metricLLMRequests = m.NewCounter("txing_llm_requests_total",
		"Total LLM streaming requests", []string{"channel", "model", "source"})
	metricLLMErrors = m.NewCounter("txing_llm_errors_total",
		"Total LLM errors by kind (no_channel/no_mapping/canceled/upstream)", []string{"channel", "model", "kind"})
	metricLLMDuration = m.NewHistogram("txing_llm_request_duration_seconds",
		"LLM streaming request total duration in seconds", llmDurationBuckets, []string{"channel", "model"})
	metricLLMTTFT = m.NewHistogram("txing_llm_ttft_seconds",
		"LLM time-to-first-token in seconds", llmTTFTBuckets, []string{"channel", "model"})
	metricLLMTokens = m.NewCounter("txing_llm_tokens_total",
		"LLM token usage (best-effort, depends on provider)", []string{"channel", "model", "direction"})
	metricLLMActive = m.NewGauge("txing_llm_streams_active",
		"In-flight LLM streaming requests", []string{"channel", "model"})
	metricChatMessages = m.NewCounter("txing_chat_messages_total",
		"User chat messages sent (after rate limit check)", []string{"model"})

	metricWSActive = m.NewGauge("txing_ws_connections_active",
		"Active websocket connections", nil)

	metricWFNodes = m.NewCounter("txing_workflow_node_total",
		"Workflow node executions by type and terminal status", []string{"node_type", "status"})
	metricWFNodeDur = m.NewHistogram("txing_workflow_node_duration_seconds",
		"Workflow node execution duration in seconds", wfNodeDurBuckets, []string{"node_type"})

	// 基础设施包直接用 zap 全局实例：项目 log 包装在 InitLogger 前为 nil，
	// 本包可能在日志系统初始化前/单测环境运行，zap.L() 永不为 nil
	zap.L().Info("observability setup done",
		zap.Bool("enabled", Enabled()),
		zap.Bool("external", ExternalEnabled()))
}

// ---- 配置 live-read（viper 热重载即时生效，镜像 wiki.AskEnabled 范式）----

// cfgOverride 测试注入配置（生产恒 nil）；SetConfigForTest 返回还原函数
var cfgOverride *global.ObservabilityConfig

// SetConfigForTest 单测注入配置，绕过 viper live-read（测试环境无配置文件）
func SetConfigForTest(c *global.ObservabilityConfig) (restore func()) {
	old := cfgOverride
	cfgOverride = c
	return func() { cfgOverride = old }
}

func obsConfig() (c *global.ObservabilityConfig) {
	if cfgOverride != nil {
		return cfgOverride
	}
	defer func() {
		// 单测环境下 LoadConfig 可能因无配置文件 panic：按未启用处理
		if r := recover(); r != nil {
			c = nil
		}
	}()
	cfg := global.LoadConfig()
	if cfg == nil {
		return nil
	}
	return cfg.ObservabilityConfig
}

// Enabled 总开关（热生效）
func Enabled() bool {
	c := obsConfig()
	return c != nil && c.Enabled
}

// ExternalEnabled 是否暴露 /metrics（startup 决策）
func ExternalEnabled() bool {
	c := obsConfig()
	return c != nil && c.External != nil && c.External.Enabled
}

func sampleIntervalConfig() time.Duration {
	if c := obsConfig(); c != nil && c.SampleInterval > 0 {
		return c.SampleInterval
	}
	return 15 * time.Second
}

func slotIntervalConfig() time.Duration {
	if c := obsConfig(); c != nil && c.SlotInterval > 0 {
		return c.SlotInterval
	}
	return time.Minute
}

func retentionConfig() time.Duration {
	if c := obsConfig(); c != nil && c.Retention > 0 {
		return c.Retention
	}
	return 24 * time.Hour
}

func maxSeriesConfig() int {
	if c := obsConfig(); c != nil && c.MaxSeries > 0 {
		return c.MaxSeries
	}
	return 2000
}

func gathererSnapshot() prometheus.Gatherer {
	mu.RLock()
	defer mu.RUnlock()
	return activeGatherer
}

// ---- 业务 facade：单行调用，未启用时 no-op ----

// HTTPRequest HTTP 请求计数（route 为 gin 路由模板）
func HTTPRequest(route, method string, code int) {
	if !Enabled() {
		return
	}
	mu.RLock()
	c := metricHTTPRequests
	mu.RUnlock()
	c.Inc(route, method, strconv.Itoa(code))
}

// HTTPDuration HTTP 请求时延（秒）
func HTTPDuration(route string, seconds float64) {
	if !Enabled() {
		return
	}
	mu.RLock()
	h := metricHTTPDuration
	mu.RUnlock()
	h.Observe(seconds, route)
}

// ChatMessage 用户聊天消息计数（通过限流检查后调用）
func ChatMessage(model string) {
	if !Enabled() {
		return
	}
	mu.RLock()
	c := metricChatMessages
	mu.RUnlock()
	c.Inc(model)
}

// WSConnect WebSocket 连接建立
func WSConnect() {
	if !Enabled() {
		return
	}
	mu.RLock()
	g := metricWSActive
	mu.RUnlock()
	g.Inc()
}

// WSDisconnect WebSocket 连接断开
func WSDisconnect() {
	if !Enabled() {
		return
	}
	mu.RLock()
	g := metricWSActive
	mu.RUnlock()
	g.Dec()
}

// WorkflowNode 工作流节点终态计数（status: completed/failed，running 不记）
func WorkflowNode(nodeType, status string, durationSeconds float64) {
	if !Enabled() {
		return
	}
	mu.RLock()
	c := metricWFNodes
	h := metricWFNodeDur
	mu.RUnlock()
	c.Inc(nodeType, status)
	h.Observe(durationSeconds, nodeType)
}

// LLMFail 渠道选择失败（无可用渠道/映射），不计入 requests_total（未发起 LLM 调用）
func LLMFail(model, kind string) {
	if !Enabled() {
		return
	}
	mu.RLock()
	c := metricLLMErrors
	mu.RUnlock()
	c.Inc("none", model, kind)
}

// llmTrace 一次 LLM 流式调用的观测上下文；nil-safe（未启用时 LLM 返回 nil）
type llmTrace struct {
	channel string
	model   string
	start   time.Time
	first   bool
	ended   bool
}

// LLM 开始一次 LLM 调用观测（channel/model 为映射后真实渠道与模型）
func LLM(channel, model string) *llmTrace {
	if !Enabled() {
		return nil
	}
	mu.RLock()
	g := metricLLMActive
	mu.RUnlock()
	g.Inc(channel, model)
	return &llmTrace{channel: channel, model: model, start: time.Now(), first: true}
}

// WrapHook 包装流式回调：首个内容 chunk 记 TTFT；
// 拦截 usage-only chunk 记 token 后不再向下游传递（避免空调量广播）
func (t *llmTrace) WrapHook(hook global.Hook) global.Hook {
	if t == nil {
		return hook
	}
	return func(chunk *global.Chunk) error {
		if chunk != nil {
			if t.first && (chunk.Content != "" || chunk.ReasoningContent != "") {
				t.first = false
				mu.RLock()
				h := metricLLMTTFT
				mu.RUnlock()
				h.Observe(time.Since(t.start).Seconds(), t.channel, t.model)
			}
			if u := chunk.Usage; u != nil {
				mu.RLock()
				c := metricLLMTokens
				mu.RUnlock()
				c.Add(float64(u.PromptTokens), t.channel, t.model, "prompt")
				c.Add(float64(u.CompletionTokens), t.channel, t.model, "completion")
				if chunk.Content == "" && chunk.ReasoningContent == "" && u.TotalTokens > 0 {
					return nil // usage-only chunk：吞掉
				}
			}
		}
		return hook(chunk)
	}
}

// End 调用结束：记时长、错误分类、活跃流减一（幂等）
func (t *llmTrace) End(err error) {
	if t == nil || t.ended {
		return
	}
	t.ended = true
	secs := time.Since(t.start).Seconds()

	mu.RLock()
	req := metricLLMRequests
	errs := metricLLMErrors
	dur := metricLLMDuration
	g := metricLLMActive
	mu.RUnlock()

	g.Dec(t.channel, t.model)
	req.Inc(t.channel, t.model, "chat")
	if err != nil {
		errs.Inc(t.channel, t.model, classifyErr(err))
		return
	}
	dur.Observe(secs, t.channel, t.model)
}

// classifyErr 错误分类为有界枚举（绝不放原始 error 字符串，防基数爆炸）
func classifyErr(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	default:
		return "upstream"
	}
}

// ---- 查询透出（controller 使用）----

// Query 时序查询（activeStore 未初始化时返回 nil）
func Query(opt QueryOptions) []SeriesOut {
	mu.RLock()
	st := activeStore
	mu.RUnlock()
	return st.Query(opt)
}

// Summarize 窗口汇总
func Summarize(metric, groupBy string, from, to time.Time) []AggSummary {
	mu.RLock()
	st := activeStore
	mu.RUnlock()
	return st.Summarize(metric, groupBy, from, to)
}

// SinceStart 进程内数据起点（unix 秒），0 表示未初始化
func SinceStart() int64 {
	mu.RLock()
	st := activeStore
	mu.RUnlock()
	if st == nil {
		return 0
	}
	return st.StartedAt()
}
