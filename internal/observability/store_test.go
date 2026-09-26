package observability

import (
	"math"
	"testing"
	"time"
)

func TestStoreCounterDelta(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	now := time.Now().Unix()
	base := now - now%60 // 对齐槽起点，保证断言确定性
	labels := []string{"model=x"}

	// 新序列首个快照自然计为增量（registry 随进程新建，从 0 起步）
	st.ApplyCounter("c_total", labels, 10, base)
	// 同槽第二次：delta = 15-10 = 5
	st.ApplyCounter("c_total", labels, 15, base)

	sums := st.Summarize("c_total", "", time.Unix(base-60, 0), time.Unix(base+60, 0))
	if len(sums) != 1 || sums[0].Sum != 15 {
		t.Fatalf("expect same-slot delta 15 (10+5), got %+v", sums)
	}

	// 下一槽：再增 20
	next := base + 60
	st.ApplyCounter("c_total", labels, 35, next)
	points := st.Query(QueryOptions{Metric: "c_total", From: time.Unix(base, 0), To: time.Unix(next, 0), Mode: QuerySum})
	if len(points) != 1 {
		t.Fatalf("expect 1 series, got %d", len(points))
	}
	if len(points[0].Points) != 2 {
		t.Fatalf("expect 2 points, got %d", len(points[0].Points))
	}
	if points[0].Points[0].Value != 15 || points[0].Points[1].Value != 20 {
		t.Fatalf("expect [15,20], got %+v", points[0].Points)
	}
}

func TestStoreCounterRebuildGuard(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	now := time.Now().Unix()
	labels := []string{"ch=a"}

	st.ApplyCounter("r_total", labels, 100, now)
	// 序列重建（counter 回退）：delta = 当前值
	st.ApplyCounter("r_total", labels, 30, now)

	sums := st.Summarize("r_total", "", time.Unix(now-60, 0), time.Unix(now+60, 0))
	if sums[0].Sum != 130 {
		t.Fatalf("expect 100+30=130 on rebuild, got %v", sums[0].Sum)
	}
}

func TestStoreGauge(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	now := time.Now().Unix()

	st.ApplyGauge("g", nil, 42, now)
	st.ApplyGauge("g", nil, 55, now) // 同槽覆盖

	out := st.Query(QueryOptions{Metric: "g", From: time.Unix(now-60, 0), To: time.Unix(now+60, 0), Mode: QueryLast})
	if len(out) != 1 || len(out[0].Points) != 1 || out[0].Points[0].Value != 55 {
		t.Fatalf("expect gauge last-wins 55, got %+v", out)
	}
}

func TestStoreHistogramQuantile(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	now := time.Now().Unix()
	base := now - now%60 // 对齐槽起点，保证断言确定性
	labels := []string{"route=/api/x"}
	// buckets: [0.1, 0.5, 1, +Inf]，累计 [4, 9, 10, 10]，sum 3.0
	bounds := []float64{0.1, 0.5, 1, math.Inf(1)}
	cum := []float64{4, 9, 10, 10}
	st.ApplyHistogram("h_seconds", labels, 10, 3.0, cum, bounds, base)

	out := st.Query(QueryOptions{Metric: "h_seconds", From: time.Unix(base, 0), To: time.Unix(base+60, 0), Mode: QueryHistMean})
	if out[0].Points[0].Value != 0.3 {
		t.Fatalf("expect mean 0.3, got %v", out[0].Points[0].Value)
	}

	out = st.Query(QueryOptions{Metric: "h_seconds", From: time.Unix(base, 0), To: time.Unix(base+60, 0), Mode: QueryHistQuantile, Quantile: 0.95})
	// rank = 9.5，落在 bucket [0.5,1)：0.5 + (9.5-9)/(10-9) * (1-0.5) = 0.75
	if got := out[0].Points[0].Value; math.Abs(got-0.75) > 1e-9 {
		t.Fatalf("expect p95 0.75, got %v", got)
	}

	sums := st.Summarize("h_seconds", "", time.Unix(base-60, 0), time.Unix(base+60, 0))
	if sums[0].P95 < 0.74 || sums[0].P95 > 0.76 {
		t.Fatalf("summarize p95 expect ~0.75, got %v", sums[0].P95)
	}
}

func TestStoreRingRotation(t *testing.T) {
	// 容量 2 个槽
	st := NewStore(time.Minute, 2*time.Minute, 100)
	now := time.Now().Unix()
	base := now - now%60 // 对齐槽起点

	st.ApplyCounter("q_total", nil, 1, base)     // 基线
	st.ApplyCounter("q_total", nil, 2, base+60)  // delta 1
	st.ApplyCounter("q_total", nil, 3, base+120) // delta 1
	st.ApplyCounter("q_total", nil, 5, base+180) // delta 2（并覆盖 base 槽）

	// ring 容量 2：只剩 base+120(1) 与 base+180(2)，更早的槽已被覆盖
	sums := st.Summarize("q_total", "", time.Unix(base, 0), time.Unix(base+180, 0))
	if sums[0].Sum != 3 {
		t.Fatalf("expect rotated sum 3, got %v", sums[0].Sum)
	}
}

func TestStoreStalePrune(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	st.staleTicks = 2
	now := time.Now().Unix()

	st.ApplyCounter("p_total", []string{"a=1"}, 5, now)
	// 推进 3 个槽（超过 staleTicks=2）触发剪枝
	st.ApplyCounter("other_total", nil, 1, now+180)

	st.mu.RLock()
	_, alive := st.series[seriesKey("p_total", []string{"a=1"})]
	st.mu.RUnlock()
	if alive {
		t.Fatal("expect stale series pruned")
	}
}

// TestStoreTombstoneRevive 剪枝后复活：沿用墓碑基线，历史累计不计为增量
func TestStoreTombstoneRevive(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	st.staleTicks = 2
	now := time.Now().Unix()
	labels := []string{"a=1"}
	key := seriesKey("t_total", labels)

	st.ApplyCounter("t_total", labels, 100, now) // delta 100
	st.ApplyCounter("other_total", nil, 1, now+180) // 触发剪枝 → t_total 进墓碑
	st.mu.RLock()
	_, alive := st.series[key]
	st.mu.RUnlock()
	if alive {
		t.Fatal("expect pruned before revive")
	}

	st.ApplyCounter("t_total", labels, 103, now+180) // 复活：delta = 103-100 = 3

	sums := st.Summarize("t_total", "", time.Unix(now, 0), time.Unix(now+180, 0))
	if len(sums) != 1 || sums[0].Sum != 103 {
		t.Fatalf("expect 100+3=103 with tombstone baseline, got %+v", sums)
	}
}

func TestStoreMaxSeriesCap(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 2)
	now := time.Now().Unix()

	st.ApplyCounter("m_total", []string{"a=1"}, 1, now)
	st.ApplyCounter("m_total", []string{"a=2"}, 1, now)
	st.ApplyCounter("m_total", []string{"a=3"}, 1, now) // 超顶丢弃

	st.mu.RLock()
	n := len(st.series)
	dropped := st.dropped
	st.mu.RUnlock()
	if n != 2 || dropped != 1 {
		t.Fatalf("expect 2 series + 1 dropped, got %d series, %d dropped", n, dropped)
	}
}

func TestStoreGroupByTopN(t *testing.T) {
	st := NewStore(time.Minute, time.Hour, 100)
	now := time.Now().Unix()
	base := now - now%60 // 对齐槽起点，保证断言确定性

	for i := 0; i < 5; i++ {
		st.ApplyCounter("by_total", []string{"ch=c" + string(rune('0'+i))}, float64(10-i), base)
	}
	out := st.Query(QueryOptions{
		Metric: "by_total", From: time.Unix(base, 0), To: time.Unix(base+60, 0),
		GroupBy: "ch", TopN: 2, Mode: QuerySum,
	})
	if len(out) != 3 { // top2 + other
		t.Fatalf("expect 3 groups, got %d", len(out))
	}
	if out[0].Key != "c0" || out[1].Key != "c1" || out[2].Key != "other" {
		t.Fatalf("unexpected group order/keys: %+v", out)
	}
	if out[2].Points[0].Value != 8+7+6 { // c2+c3+c4
		t.Fatalf("expect other=21, got %v", out[2].Points[0].Value)
	}
}
