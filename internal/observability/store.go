package observability

import (
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"
)

// Store 进程内固定槽环形缓冲时序存储：
// sampler 周期写入（counter 记增量、gauge 记末值、histogram 记 count/sum/bucket 增量），
// 管理后台按窗口/分组查询。重启清零（设计取舍：长期统计走 DB 业务表聚合）
type Store struct {
	mu          sync.RWMutex
	stepSecs    int64              // 槽粒度（秒）
	capacity    int                // 槽总数 = retention / step
	slots       []*slotEntry       // ring：下标 = tickIndex % capacity
	series      map[uint64]*series // seriesID → 序列元信息
	tombstones  map[uint64]*series // 剪枝序列的累计基线（复活时沿用，避免历史值被误记为增量）
	maxSeries   int                // 序列数硬顶（基数防御）
	staleTicks  int64              // 连续 N 个 tick 未出现即剪枝
	lastTick    int64              // 已推进到的最新 tickIndex
	startedAt   int64              // 启动时间（unix 秒）
	dropped     uint64             // 因超 maxSeries 丢弃的写入次数
}

type slotEntry struct {
	tick int64     // 槽起始 tickIndex（写入时校验，防止 ring 覆盖后误读）
	data *slotData
}

type slotData struct {
	counters map[uint64]float64
	gauges   map[uint64]float64
	hists    map[uint64]*histDelta
}

// histDelta 槽内直方图增量：Buckets[i] 为本槽观测中落在 leBounds[i] 边界内的累计数
// （对 dto 累计值做差所得；跨槽/跨序列求和即为聚合窗口内的累计分布）
type histDelta struct {
	Count   float64
	Sum     float64
	Buckets []float64
}

type seriesKind int

const (
	kindCounter seriesKind = iota
	kindGauge
	kindHistogram
)

// series 序列元信息；id 由 metric 名 + 排序后标签内容寻址（fnv64），
// 序列被剪枝后再次出现会以相同 id 重建，历史槽数据自然延续
type series struct {
	id         uint64
	name       string
	labels     []string          // 排序后的 "k=v" 列表（key 计算/展示用）
	labelMap   map[string]string // 分组查询用
	kind       seriesKind
	leBounds   []float64 // histogram bucket 上界（含 +Inf 尾）；summary 类无 bucket 时为 nil
	lastValue  float64   // counter 当前值 / histogram SampleCount
	lastSum    float64   // histogram SampleSum
	lastBucket []float64 // histogram 各 bucket 累计值
	lastSeen   int64
}

func newSlotData() *slotData {
	return &slotData{
		counters: make(map[uint64]float64),
		gauges:   make(map[uint64]float64),
		hists:    make(map[uint64]*histDelta),
	}
}

func NewStore(step time.Duration, retention time.Duration, maxSeries int) *Store {
	stepSecs := int64(step.Seconds())
	if stepSecs <= 0 {
		stepSecs = 60
	}
	capacity := int(retention.Seconds()) / int(stepSecs)
	if capacity <= 0 {
		capacity = 1
	}
	if maxSeries <= 0 {
		maxSeries = 2000
	}
	return &Store{
		stepSecs:   stepSecs,
		capacity:   capacity,
		slots:      make([]*slotEntry, capacity),
		series:     make(map[uint64]*series),
		tombstones: make(map[uint64]*series),
		maxSeries:  maxSeries,
		staleTicks: 120, // 默认连续 2h（1m 槽）未出现即剪枝
		lastTick:   time.Now().Unix() / stepSecs,
		startedAt:  time.Now().Unix(),
	}
}

// StartedAt 进程内数据的起点（"自启动"口径）
func (s *Store) StartedAt() int64 {
	if s == nil {
		return 0
	}
	return s.startedAt
}

func seriesKey(name string, labels []string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	for _, l := range labels {
		_, _ = h.Write([]byte{'|'})
		_, _ = h.Write([]byte(l))
	}
	return h.Sum64()
}

// advance 推进 ring 到 now 所在槽（跨越的槽全部清零重建），并顺带剪枝 stale 序列
func (s *Store) advance(now int64) {
	cur := now / s.stepSecs
	if cur <= s.lastTick {
		return
	}
	gap := cur - s.lastTick
	if gap > int64(s.capacity) {
		gap = int64(s.capacity) // 休眠超过保留期：全部槽作废
	}
	for i := int64(1); i <= gap; i++ {
		t := cur - gap + i
		s.slots[t%int64(s.capacity)] = &slotEntry{tick: t, data: newSlotData()}
	}
	deadline := cur - s.staleTicks
	for id, ser := range s.series {
		if ser.lastSeen <= deadline {
			// 剪枝但保留累计基线（墓碑）：序列复活时不会把历史累计值误记为增量
			s.tombstones[id] = ser
			delete(s.series, id)
		}
	}
	if len(s.tombstones) > s.maxSeries*4 {
		s.tombstones = make(map[uint64]*series) // 墓碑超限兜底清空（极端基数爆炸场景）
	}
	s.lastTick = cur
}

func (s *Store) curSlotData() *slotData {
	entry := s.slots[s.lastTick%int64(s.capacity)]
	if entry == nil || entry.tick != s.lastTick {
		entry = &slotEntry{tick: s.lastTick, data: newSlotData()}
		s.slots[s.lastTick%int64(s.capacity)] = entry
	}
	return entry.data
}

// seriesFor 查找/创建序列（含墓碑复活：沿用历史累计基线）；
// 超过 maxSeries 硬顶时返回 nil（丢弃本次写入）。
// 新序列 lastValue 为零值，首个快照自然计为增量（registry 是随进程新建的，
// 计数器从 0 起步，不存在历史值虚增）
func (s *Store) seriesFor(id uint64, name string, labels []string, labelMap map[string]string, kind seriesKind) *series {
	if ser, ok := s.series[id]; ok {
		ser.lastSeen = s.lastTick
		return ser
	}
	if ser, ok := s.tombstones[id]; ok {
		delete(s.tombstones, id)
		ser.lastSeen = s.lastTick
		s.series[id] = ser
		return ser
	}
	if len(s.series) >= s.maxSeries {
		s.dropped++
		return nil
	}
	ser := &series{
		id:       id,
		name:     name,
		labels:   labels,
		labelMap: labelMap,
		kind:     kind,
		lastSeen: s.lastTick,
	}
	s.series[id] = ser
	return ser
}

// ApplyCounter 记录 counter 当前值（store 内部转槽内增量）
func (s *Store) ApplyCounter(name string, labels []string, value float64, now int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.advance(now)
	id := seriesKey(name, labels)
	ser := s.seriesFor(id, name, labels, labelMapOf(labels), kindCounter)
	if ser == nil {
		return
	}
	delta := value - ser.lastValue
	if delta < 0 {
		delta = value // 序列重建防御（理论不发生）
	}
	ser.lastValue = value
	s.curSlotData().counters[id] += delta
}

// ApplyGauge 记录 gauge 瞬时值（槽内取最后一次采样）
func (s *Store) ApplyGauge(name string, labels []string, value float64, now int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.advance(now)
	id := seriesKey(name, labels)
	ser := s.seriesFor(id, name, labels, labelMapOf(labels), kindGauge)
	if ser == nil {
		return
	}
	s.curSlotData().gauges[id] = value
}

// ApplyHistogram 记录 histogram 累计快照（count/sum/各 bucket 累计值），
// store 内部转槽内增量；bounds 为 bucket 上界（含 +Inf，与 buckets 一一对齐），
// summary 类指标（无 bucket）传 nil
func (s *Store) ApplyHistogram(name string, labels []string, count, sum float64, cumBuckets []float64, bounds []float64, now int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.advance(now)
	id := seriesKey(name, labels)
	ser := s.seriesFor(id, name, labels, labelMapOf(labels), kindHistogram)
	if ser == nil {
		return
	}
	if len(ser.leBounds) == 0 && len(bounds) > 0 {
		ser.leBounds = bounds
	}
	countDelta := count - ser.lastValue
	if countDelta < 0 {
		countDelta = count
	}
	sumDelta := sum - ser.lastSum
	if sumDelta < 0 {
		sumDelta = 0
	}
	ser.lastValue = count
	ser.lastSum = sum

	nb := len(ser.leBounds)
	var bucketDeltas []float64
	if nb > 0 && len(cumBuckets) == nb && ser.lastBucket != nil {
		bucketDeltas = make([]float64, nb)
		for i := 0; i < nb; i++ {
			d := cumBuckets[i] - ser.lastBucket[i]
			if d < 0 {
				d = cumBuckets[i]
			}
			bucketDeltas[i] = d
		}
	} else if nb > 0 && len(cumBuckets) == nb {
		// 首次见到 bucket：增量即当前累计值
		bucketDeltas = make([]float64, nb)
		copy(bucketDeltas, cumBuckets)
	}
	if len(cumBuckets) == nb && nb > 0 {
		ser.lastBucket = append(ser.lastBucket[:0], cumBuckets...)
	} else if nb == 0 {
		ser.lastBucket = nil
	}

	data := s.curSlotData()
	hd, ok := data.hists[id]
	if !ok {
		hd = &histDelta{}
		data.hists[id] = hd
	}
	hd.Count += countDelta
	hd.Sum += sumDelta
	if len(bucketDeltas) > 0 {
		if len(hd.Buckets) != len(bucketDeltas) {
			if hd.Buckets == nil {
				hd.Buckets = make([]float64, len(bucketDeltas))
			} else {
				// bucket 数变化（理论不发生）：放弃旧数据重对齐
				hd.Buckets = make([]float64, len(bucketDeltas))
			}
		}
		for i, d := range bucketDeltas {
			hd.Buckets[i] += d
		}
	}
}

func labelMapOf(labels []string) map[string]string {
	m := make(map[string]string, len(labels))
	for _, l := range labels {
		for i := 0; i < len(l); i++ {
			if l[i] == '=' {
				m[l[:i]] = l[i+1:]
				break
			}
		}
	}
	return m
}

// ---- 查询 ----

type Point struct {
	Ts    int64   `json:"ts"` // 槽起始 unix 秒
	Value float64 `json:"value"`
}

type SeriesOut struct {
	Key    string            `json:"key"` // 分组值或 "total"
	Labels map[string]string `json:"labels,omitempty"`
	Points []Point           `json:"points"`
}

type QueryMode int

const (
	QuerySum        QueryMode = iota // counter：每槽跨序列求和（0 填充缺失槽）
	QueryLast                        // gauge：每槽取值（跨序列求和；缺失槽跳过）
	QueryHistMean                    // histogram：每槽 sum/count
	QueryHistQuantile                // histogram：每槽分位数线性插值
)

type QueryOptions struct {
	Metric   string
	From, To time.Time
	GroupBy  string // "" = 全部序列合并为 "total"
	TopN     int    // GroupBy 时按窗口总量取前 N，其余并入 "other"（<=0 不截断）
	Mode     QueryMode
	Quantile float64 // QueryHistQuantile 时生效（0-1）
}

// aggAcc 查询聚合累加器：counter/gauge 直接累加 value；histogram 合并 count/sum/buckets
type aggAcc struct {
	val     float64
	count   float64
	sum     float64
	buckets []float64
	bounds  []float64
}

// Query 按窗口查询时序；counter/histogram 模式缺失槽补 0 点，gauge 模式跳过
func (s *Store) Query(opt QueryOptions) []SeriesOut {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	step := s.stepSecs
	fromTick := opt.From.Unix() / step
	toTick := opt.To.Unix() / step
	if toTick > s.lastTick {
		toTick = s.lastTick
	}
	if fromTick < s.lastTick-int64(s.capacity)+1 {
		fromTick = s.lastTick - int64(s.capacity) + 1
	}
	if fromTick > toTick {
		return nil
	}

	// 目标序列集合（按 metric 名过滤）
	targets := make([]*series, 0, 8)
	for _, ser := range s.series {
		if ser.name == opt.Metric {
			targets = append(targets, ser)
		}
	}

	groupOf := func(ser *series) string {
		if opt.GroupBy == "" {
			return "total"
		}
		return ser.labelMap[opt.GroupBy]
	}

	// 窗口总量（TopN 排名用）
	totals := make(map[string]float64)
	// groups[tick][groupKey] → 累加器
	groups := make(map[int64]map[string]*aggAcc)

	readSlot := func(t int64) *slotData {
		entry := s.slots[t%int64(s.capacity)]
		if entry == nil || entry.tick != t {
			return nil
		}
		return entry.data
	}

	for t := fromTick; t <= toTick; t++ {
		data := readSlot(t)
		if data == nil {
			continue
		}
		byGroup, ok := groups[t]
		if !ok {
			byGroup = make(map[string]*aggAcc)
			groups[t] = byGroup
		}
		for _, ser := range targets {
			gk := groupOf(ser)
			acc, ok := byGroup[gk]
			if !ok {
				acc = &aggAcc{}
				byGroup[gk] = acc
			}
			switch ser.kind {
			case kindCounter:
				if v, ok := data.counters[ser.id]; ok {
					acc.val += v
					acc.count += v
				}
			case kindGauge:
				if v, ok := data.gauges[ser.id]; ok {
					acc.val += v
					acc.count++
				}
			case kindHistogram:
				if hd, ok := data.hists[ser.id]; ok {
					acc.count += hd.Count
					acc.sum += hd.Sum
					if len(hd.Buckets) > 0 && len(acc.bounds) == 0 {
						acc.bounds = ser.leBounds
					}
					if len(hd.Buckets) > 0 {
						if len(acc.buckets) != len(hd.Buckets) {
							grown := make([]float64, len(hd.Buckets))
							copy(grown, acc.buckets)
							acc.buckets = grown
						}
						for i, b := range hd.Buckets {
							acc.buckets[i] += b
						}
					}
				}
			}
		}
	}

	// 汇总窗口总量
	for _, byGroup := range groups {
		for gk, acc := range byGroup {
			switch opt.Mode {
			case QueryHistMean, QueryHistQuantile:
				totals[gk] += acc.count
			default:
				totals[gk] += acc.count
			}
		}
	}

	// TopN 保留集
	keep := make(map[string]bool)
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return totals[keys[i]] > totals[keys[j]] })
	hasOther := false
	for i, k := range keys {
		if opt.TopN > 0 && i >= opt.TopN {
			hasOther = true
			continue
		}
		keep[k] = true
	}

	zeroFill := opt.Mode == QuerySum || opt.Mode == QueryHistMean || opt.Mode == QueryHistQuantile
	outKeys := make([]string, 0, len(keep)+1)
	for _, k := range keys {
		if keep[k] {
			outKeys = append(outKeys, k)
		}
	}
	if hasOther {
		outKeys = append(outKeys, "other")
	}

	result := make([]SeriesOut, 0, len(outKeys))
	for _, gk := range outKeys {
		so := SeriesOut{Key: gk, Points: make([]Point, 0, toTick-fromTick+1)}
		for t := fromTick; t <= toTick; t++ {
			byGroup, ok := groups[t]
			if !ok {
				if zeroFill {
					so.Points = append(so.Points, Point{Ts: t * step, Value: 0})
				}
				continue
			}
			var acc *aggAcc
			if hasOther && !keep[gk] {
				// "other"：合并所有未保留分组
				merged := &aggAcc{}
				for k, a := range byGroup {
					if keep[k] {
						continue
					}
					mergeAcc(merged, a)
				}
				acc = merged
			} else {
				acc = byGroup[gk]
			}
			if acc == nil || (acc.count == 0 && acc.val == 0 && !zeroFill) {
				if zeroFill {
					so.Points = append(so.Points, Point{Ts: t * step, Value: 0})
				}
				continue
			}
			so.Points = append(so.Points, Point{Ts: t * step, Value: acc.value(opt)})
		}
		result = append(result, so)
	}
	return result
}

func mergeAcc(dst, src *aggAcc) {
	dst.val += src.val
	dst.count += src.count
	dst.sum += src.sum
	if len(src.buckets) > 0 {
		if len(dst.buckets) != len(src.buckets) {
			grown := make([]float64, len(src.buckets))
			copy(grown, dst.buckets)
			dst.buckets = grown
		}
		for i, b := range src.buckets {
			dst.buckets[i] += b
		}
	}
	if len(dst.bounds) == 0 {
		dst.bounds = src.bounds
	}
}

func (a *aggAcc) value(opt QueryOptions) float64 {
	switch opt.Mode {
	case QueryHistMean:
		if a.count > 0 {
			return a.sum / a.count
		}
		return 0
	case QueryHistQuantile:
		return quantile(opt.Quantile, a.count, a.buckets, a.bounds)
	case QueryLast:
		return a.val
	default: // QuerySum
		return a.val
	}
}

// quantile Prometheus histogram_quantile 同源线性插值。
// buckets 为聚合窗口内各 le 边界处的累计观测数（含 +Inf 尾，buckets[len-1] == count），
// bounds 与之一一对齐
func quantile(q, count float64, buckets, bounds []float64) float64 {
	if count <= 0 || len(buckets) == 0 || len(bounds) != len(buckets) {
		return 0
	}
	rank := q * count
	var prevCum float64
	for i, c := range buckets {
		if c < rank {
			prevCum = c
			continue
		}
		if math.IsInf(bounds[i], 1) {
			// 落入 +Inf 桶：返回最后一个有限上界
			if i > 0 {
				return bounds[i-1]
			}
			return 0
		}
		if i == 0 {
			if c > 0 {
				return bounds[0] * rank / c
			}
			return 0
		}
		if c == prevCum {
			return bounds[i-1]
		}
		frac := (rank - prevCum) / (c - prevCum)
		return bounds[i-1] + frac*(bounds[i]-bounds[i-1])
	}
	if n := len(bounds); n > 0 {
		return bounds[n-1]
	}
	return 0
}

// ---- 窗口汇总（无时间轴，供卡片/渠道表用）----

type AggSummary struct {
	Key  string  `json:"key"`
	Sum  float64 `json:"sum"`  // counter 窗口总增量 / gauge 末值 / histogram 样本数
	Mean float64 `json:"mean"` // histogram：sum/count
	P95  float64 `json:"p95"`  // histogram：分位数插值（counter/gauge 为 0）
}

// Summarize 按分组汇总窗口数据（histogram 指标给出 mean/p95，counter 给出总量）
func (s *Store) Summarize(metric, groupBy string, from, to time.Time) []AggSummary {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	step := s.stepSecs
	fromTick := from.Unix() / step
	toTick := to.Unix() / step
	if toTick > s.lastTick {
		toTick = s.lastTick
	}
	if fromTick < s.lastTick-int64(s.capacity)+1 {
		fromTick = s.lastTick - int64(s.capacity) + 1
	}
	if fromTick > toTick {
		return nil
	}

	type acc struct {
		val     float64
		count   float64
		sum     float64
		buckets []float64
		bounds  []float64
	}
	accs := make(map[string]*acc)
	for _, ser := range s.series {
		if ser.name != metric {
			continue
		}
		gk := "total"
		if groupBy != "" {
			gk = ser.labelMap[groupBy]
		}
		a, ok := accs[gk]
		if !ok {
			a = &acc{}
			accs[gk] = a
		}
		for t := fromTick; t <= toTick; t++ {
			entry := s.slots[t%int64(s.capacity)]
			if entry == nil || entry.tick != t {
				continue
			}
			switch ser.kind {
			case kindCounter:
				if v, ok := entry.data.counters[ser.id]; ok {
					a.val += v
				}
			case kindGauge:
				if v, ok := entry.data.gauges[ser.id]; ok {
					a.val = v
				}
			case kindHistogram:
				if hd, ok := entry.data.hists[ser.id]; ok {
					a.count += hd.Count
					a.sum += hd.Sum
					if len(hd.Buckets) > 0 {
						if len(a.bounds) == 0 {
							a.bounds = ser.leBounds
						}
						if len(a.buckets) != len(hd.Buckets) {
							grown := make([]float64, len(hd.Buckets))
							copy(grown, a.buckets)
							a.buckets = grown
						}
						for i, b := range hd.Buckets {
							a.buckets[i] += b
						}
					}
				}
			}
		}
	}

	out := make([]AggSummary, 0, len(accs))
	for gk, a := range accs {
		su := AggSummary{Key: gk, Sum: a.val}
		if a.count > 0 {
			su.Sum = a.count
			su.Mean = a.sum / a.count
			su.P95 = quantile(0.95, a.count, a.buckets, a.bounds)
		}
		out = append(out, su)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sum > out[j].Sum })
	return out
}
