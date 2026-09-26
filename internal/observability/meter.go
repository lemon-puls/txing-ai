// Package observability 提供可插拔的指标埋点与进程内时序存储：
//   - 业务代码只依赖本包的类型化 facade（metrics.go），不直接接触底层实现
//   - Meter 是实现插槽：promMeter（prometheus/client_golang）| nopMeter |（未来）otel/remote-write
//   - sampler 周期性从 registry 采集快照写入环形缓冲（store.go），供管理后台直读
//
// 设计文档：docs/observability_design.md
package observability

// Counter 单调递增计数器；label 值按 NewCounter 声明的 labelNames 顺序位置传入
type Counter interface {
	Inc(labels ...string)
	Add(delta float64, labels ...string)
}

// Gauge 可任意变化的瞬时值
type Gauge interface {
	Set(v float64, labels ...string)
	Inc(labels ...string)
	Dec(labels ...string)
}

// Histogram 直方图（自动分桶统计分布）
type Histogram interface {
	Observe(v float64, labels ...string)
}

// Meter 指标实现插槽，可替换为 prometheus / no-op / 未来的 OTel、remote-write 适配器
type Meter interface {
	NewCounter(name, help string, labelNames []string) Counter
	NewGauge(name, help string, labelNames []string) Gauge
	NewHistogram(name, help string, buckets []float64, labelNames []string) Histogram
}
