package vo

import "time"

// DashboardCard 控制台统计卡片（卡片列表由响应驱动，前端不写死）
type DashboardCard struct {
	Key          string   `json:"key"`
	Title        string   `json:"title"`
	Value        string   `json:"value"`
	Unit         string   `json:"unit,omitempty"`
	TrendPercent *float64 `json:"trendPercent,omitempty"` // 与昨日同期对比（%），无对比口径时为 null
	Note         string   `json:"note,omitempty"`         // 口径说明
}

// DashboardOverviewVO 控制台总览
type DashboardOverviewVO struct {
	Cards      []DashboardCard `json:"cards"`
	SinceStart string          `json:"sinceStart,omitempty"` // 进程内运行时数据起点（ISO 时间）
}

// DashboardTrendPoint 对话趋势单点
type DashboardTrendPoint struct {
	Date          string `json:"date"`
	Conversations int64  `json:"conversations"`
	ActiveUsers   int64  `json:"activeUsers"`
}

// DashboardTrendVO 对话趋势（按日/月聚合）
type DashboardTrendVO struct {
	Granularity string                `json:"granularity"` // day | month
	Points      []DashboardTrendPoint `json:"points"`
}

// DashboardNameCount 通用名称-数量（模型占比/助手排行）
type DashboardNameCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// DashboardTimeseriesPoint 运行时时序单点
type DashboardTimeseriesPoint struct {
	Ts    int64   `json:"ts"` // unix 秒（槽起始）
	Value float64 `json:"value"`
}

// DashboardTimeseriesSeries 单条时序序列
type DashboardTimeseriesSeries struct {
	Metric string                     `json:"metric"` // 指标名（图表标题/分组用）
	Key    string                     `json:"key"`    // 分组值（如渠道名）
	Points []DashboardTimeseriesPoint `json:"points"`
}

// DashboardTimeseriesVO 运行时时序
type DashboardTimeseriesVO struct {
	StepSeconds int64                       `json:"stepSeconds"`
	Series      []DashboardTimeseriesSeries `json:"series"`
}

// DashboardChannelUsageItem 渠道用量行
type DashboardChannelUsageItem struct {
	Channel      string  `json:"channel"`
	Requests     int64   `json:"requests"`
	Errors       int64   `json:"errors"`
	ErrorRate    float64 `json:"errorRate"`
	AvgLatencyMs float64 `json:"avgLatencyMs"`
	P95LatencyMs float64 `json:"p95LatencyMs"`
}

// DashboardChannelUsageVO LLM 渠道用量（进程内环形缓冲口径）
type DashboardChannelUsageVO struct {
	SinceStart string                      `json:"sinceStart,omitempty"`
	Items      []DashboardChannelUsageItem `json:"items"`
}

// DashboardActivity 最近活动行
type DashboardActivity struct {
	Time   time.Time `json:"time"`
	User   string    `json:"user"`
	Action string    `json:"action"`
	Detail string    `json:"detail"`
	Status string    `json:"status"`
}
