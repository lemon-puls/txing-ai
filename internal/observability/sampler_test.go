package observability

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"txing-ai/internal/global"
)

func enabledConfig() global.ObservabilityConfig {
	return global.ObservabilityConfig{Enabled: true}
}

func contentChunk(content string) *global.Chunk {
	return &global.Chunk{Content: content}
}

func usageChunk(prompt, completion, total int64) *global.Chunk {
	return &global.Chunk{Usage: &global.UsageInfo{
		PromptTokens: prompt, CompletionTokens: completion, TotalTokens: total,
	}}
}

func cfgPtr(c global.ObservabilityConfig) *global.ObservabilityConfig {
	return &c
}

// TestSamplerEndToEnd 用真实 registry 验证 Gather → delta → store 全链路
func TestSamplerEndToEnd(t *testing.T) {
	restore := SetConfigForTest(cfgPtr(enabledConfig()))
	defer restore()

	reg := prometheus.NewRegistry()
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "s_total", Help: "t"}, []string{"k"})
	reg.MustRegister(counter)
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "s_gauge", Help: "t"}, nil)
	reg.MustRegister(gauge)

	st := NewStore(time.Minute, time.Hour, 100)
	now := time.Now().Unix()

	counter.WithLabelValues("v1").Add(3)
	gauge.WithLabelValues().Set(7)
	sampleOnce(reg, st) // 基线

	counter.WithLabelValues("v1").Add(4)
	gauge.WithLabelValues().Set(9)
	sampleOnce(reg, st)

	sums := st.Summarize("s_total", "", time.Unix(now-60, 0), time.Unix(now+60, 0))
	if len(sums) != 1 || sums[0].Sum != 7 { // 3（首见全量）+ 4（增量）
		t.Fatalf("expect counter total 7, got %+v", sums)
	}
	out := st.Query(QueryOptions{Metric: "s_gauge", From: time.Unix(now-60, 0), To: time.Unix(now+60, 0), Mode: QueryLast})
	if len(out) != 1 || out[0].Points[0].Value != 9 {
		t.Fatalf("expect gauge 9, got %+v", out)
	}
}

// TestLLMTraceLifecycle 验证 LLM facade 记账（经 prom 实现落 registry）
func TestLLMTraceLifecycle(t *testing.T) {
	cfg := enabledConfig()
	restoreCfg := SetConfigForTest(&cfg)
	defer restoreCfg()
	Setup(&cfg)
	defer func() {
		mu.Lock()
		activeMeter = nopMeter{}
		activeGatherer = nil
		activeStore = nil
		mu.Unlock()
	}()

	// 渠道选择失败
	LLMFail("m1", "no_channel")

	// 成功调用 + TTFT + usage chunk 拦截
	trace := LLM("ch1", "m1")
	var downstream int
	hook := trace.WrapHook(func(chunk *global.Chunk) error {
		downstream++
		return nil
	})
	_ = hook(contentChunk("hello"))
	_ = hook(usageChunk(10, 20, 30)) // usage-only：应被吞掉
	trace.End(nil)

	_ = hook(contentChunk("after-end")) // End 后仍透传

	if downstream != 2 {
		t.Fatalf("expect 2 downstream chunks (usage-only swallowed, after-end passed), got %d", downstream)
	}

	mu.RLock()
	g := activeGatherer
	mu.RUnlock()
	if g == nil {
		t.Fatal("expect gatherer after setup")
	}
	mfs, err := g.Gather()
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, mf := range mfs {
		found[mf.GetName()] = true
	}
	for _, want := range []string{
		"txing_llm_requests_total", "txing_llm_errors_total", "txing_llm_ttft_seconds",
		"txing_llm_tokens_total", "txing_llm_streams_active",
	} {
		if !found[want] {
			t.Fatalf("metric %s not registered", want)
		}
	}
}

// TestDisabledNoop 未启用时 facade 全部 no-op
func TestDisabledNoop(t *testing.T) {
	disabled := enabledConfig()
	disabled.Enabled = false
	restore := SetConfigForTest(&disabled)
	defer restore()

	if LLM("ch", "m") != nil {
		t.Fatal("expect nil trace when disabled")
	}
	var t2 *llmTrace
	t2.End(nil)          // nil-safe
	_ = t2.WrapHook(nil) // nil-safe
	HTTPRequest("/api/x", "GET", 200)
	WSConnect()
	WSDisconnect()
}
