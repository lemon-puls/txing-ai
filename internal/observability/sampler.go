package observability

import (
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

// StartSampler 启动采样协程：周期 Gather registry 快照写入环形缓冲。
// 无论 enabled 与否都会启动（开销可忽略），每 tick 检查开关：
// 关闭期间跳过采样、缓冲数据保留，重新开启后继续写入。
// 返回值用于停机收尾（触发一次 best-effort 收尾采样）
func StartSampler() func() {
	stop := make(chan struct{})
	done := make(chan struct{})

	mu.Lock()
	if activeStore == nil {
		activeStore = NewStore(slotIntervalConfig(), retentionConfig(), maxSeriesConfig())
	}
	store := activeStore
	gatherer := activeGatherer // 已持写锁，直接读（gathererSnapshot 会重入加锁死锁）
	mu.Unlock()

	if gatherer == nil {
		// meter 为 nop（Setup 未构建 prom 实现）：无需采样
		close(done)
		return func() {}
	}

	sample := func() {
		defer func() {
			if r := recover(); r != nil {
				zap.L().Error("observability sampler panic", zap.Any("err", r))
			}
		}()
		sampleOnce(gatherer, store)
	}

	go func() {
		defer close(done)
		interval := sampleIntervalConfig()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		sample() // 首次立即采样，建立 counter 基线（首槽 delta 为 0）
		for {
			select {
			case <-ticker.C:
				if Enabled() {
					sample()
				}
			case <-stop:
				if Enabled() {
					sample() // 收尾采样 best-effort
				}
				return
			}
		}
	}()

	return func() {
		close(stop)
		<-done
	}
}

// SampleNow 立即执行一次采样（Gather → 环形缓冲）；
// 供停机收尾、测试与运维手动触发使用
func SampleNow() {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("observability sampler panic", zap.Any("err", r))
		}
	}()
	g := gathererSnapshot()
	mu.RLock()
	st := activeStore
	mu.RUnlock()
	if g == nil || st == nil {
		return
	}
	sampleOnce(g, st)
}

// sampleOnce 一次 Gather → 解析 dto → 写 store
func sampleOnce(g prometheus.Gatherer, st *Store) {
	mfs, err := g.Gather()
	if err != nil {
		zap.L().Error("observability gather error", zap.Error(err))
		return
	}
	now := time.Now().Unix()
	for _, mf := range mfs {
		name := mf.GetName()
		for _, m := range mf.GetMetric() {
			labels := dtoLabels(m.GetLabel())
			switch mf.GetType() {
			case dto.MetricType_COUNTER:
				st.ApplyCounter(name, labels, m.GetCounter().GetValue(), now)
			case dto.MetricType_GAUGE, dto.MetricType_UNTYPED:
				st.ApplyGauge(name, labels, m.GetGauge().GetValue(), now)
			case dto.MetricType_HISTOGRAM:
				h := m.GetHistogram()
				cum, bounds := dtoBuckets(h.GetBucket())
				st.ApplyHistogram(name, labels, float64(h.GetSampleCount()), h.GetSampleSum(), cum, bounds, now)
			case dto.MetricType_SUMMARY:
				// summary（如 go_gc_duration_seconds）：按 count/sum 记，无 bucket 分位
				h := m.GetSummary()
				st.ApplyHistogram(name, labels, float64(h.GetSampleCount()), h.GetSampleSum(), nil, nil, now)
			}
		}
	}
}

// dtoLabels 转 "k=v" 并按字典序排序（保证序列 key 稳定）
func dtoLabels(pairs []*dto.LabelPair) []string {
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.GetName()+"="+p.GetValue())
	}
	sort.Strings(out)
	return out
}

// dtoBuckets 提取累计 bucket 值与上界（含 +Inf 尾）
func dtoBuckets(buckets []*dto.Bucket) (cum []float64, bounds []float64) {
	if len(buckets) == 0 {
		return nil, nil
	}
	cum = make([]float64, 0, len(buckets))
	bounds = make([]float64, 0, len(buckets))
	for _, b := range buckets {
		cum = append(cum, float64(b.GetCumulativeCount()))
		bounds = append(bounds, b.GetUpperBound())
	}
	return cum, bounds
}
