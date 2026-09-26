package dashboard

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	statssvc "txing-ai/internal/service/stats"
	"txing-ai/internal/utils"
	"txing-ai/internal/observability"
	"txing-ai/internal/vo"
)

// getDB 获取 gin context 中的 DB（BuiltinMiddleWare 注入）
func getDB(ctx *gin.Context) *gorm.DB {
	return utils.GetDBFromContext[*gorm.DB](ctx)
}

func formatTrend(today, yesterday int64) *float64 {
	if yesterday <= 0 {
		return nil
	}
	v := (float64(today) - float64(yesterday)) / float64(yesterday) * 100
	return &v
}

func sinceStartStr() string {
	if ts := observability.SinceStart(); ts > 0 {
		return time.Unix(ts, 0).Format(time.RFC3339)
	}
	return ""
}

// Overview 控制台总览卡片
// @Summary 控制台总览
// @Description 统计卡片：今日对话数/今日活跃用户（DB 口径，对比昨日同期）+ LLM 平均响应/Token 消耗（进程内运行时指标口径）
// @Tags 控制台
// @Produce json
// @Success 200 {object} utils.Response{data=vo.DashboardOverviewVO}
// @Router /api/admin/dashboard/overview [get]
func Overview(ctx *gin.Context) {
	db := getDB(ctx)
	now := time.Now()
	cards := make([]vo.DashboardCard, 0, 4)

	// 今日对话数（DB）
	if pair, err := statssvc.ConversationToday(db, now); err == nil {
		cards = append(cards, vo.DashboardCard{
			Key: "conversations_today", Title: "今日对话数",
			Value: strconv.FormatInt(pair.Today, 10),
			TrendPercent: formatTrend(pair.Today, pair.Yesterday),
			Note: "对比昨日同期",
		})
	}

	// 今日活跃用户（DB）
	if pair, err := statssvc.ActiveUsersToday(db, now); err == nil {
		cards = append(cards, vo.DashboardCard{
			Key: "active_users_today", Title: "今日活跃用户",
			Value: strconv.FormatInt(pair.Today, 10),
			TrendPercent: formatTrend(pair.Today, pair.Yesterday),
			Note: "对比昨日同期",
		})
	}

	// LLM 平均响应（近 1h，环形缓冲）
	if sums := observability.Summarize("txing_llm_request_duration_seconds", "", now.Add(-time.Hour), now); len(sums) > 0 && sums[0].Sum > 0 {
		cards = append(cards, vo.DashboardCard{
			Key: "llm_latency_1h", Title: "LLM 平均响应",
			Value: strconv.FormatFloat(sums[0].Mean*1000, 'f', 0, 64),
			Unit:  "ms",
			Note:  fmt.Sprintf("近 1 小时 %d 次 · P95 %.0fms", int64(sums[0].Sum), sums[0].P95*1000),
		})
	} else {
		cards = append(cards, vo.DashboardCard{
			Key: "llm_latency_1h", Title: "LLM 平均响应",
			Value: "—", Note: "近 1 小时暂无调用",
		})
	}

	// Token 消耗（自启动累计）
	var tokens float64
	for _, s := range observability.Summarize("txing_llm_tokens_total", "direction",
		time.Unix(observability.SinceStart(), 0), now) {
		tokens += s.Sum
	}
	cards = append(cards, vo.DashboardCard{
		Key: "tokens_since_start", Title: "Token 消耗",
		Value: formatCount(tokens),
		Note:  "自启动累计",
	})

	utils.OkWithData(ctx, vo.DashboardOverviewVO{
		Cards:      cards,
		SinceStart: sinceStartStr(),
	})
}

// Trends 对话趋势
// @Summary 对话趋势
// @Description 按日（range=365 时按月）聚合的对话数与活跃用户数
// @Tags 控制台
// @Produce json
// @Param range query int false "天数：7/30/365（默认 7）" default(7)
// @Success 200 {object} utils.Response{data=vo.DashboardTrendVO}
// @Router /api/admin/dashboard/trends [get]
func Trends(ctx *gin.Context) {
	days, _ := strconv.Atoi(ctx.DefaultQuery("range", "7"))
	if days != 30 && days != 365 {
		days = 7
	}
	points, granularity, err := statssvc.ConversationTrend(getDB(ctx), days, time.Now())
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询对话趋势失败", err)
		return
	}
	out := vo.DashboardTrendVO{Granularity: granularity, Points: make([]vo.DashboardTrendPoint, 0, len(points))}
	for _, p := range points {
		out.Points = append(out.Points, vo.DashboardTrendPoint{
			Date: p.Date, Conversations: p.Conversations, ActiveUsers: p.ActiveUsers,
		})
	}
	utils.OkWithData(ctx, out)
}

// ModelUsage 模型使用占比
// @Summary 模型使用占比
// @Description 近 N 天按模型聚合的会话数占比
// @Tags 控制台
// @Produce json
// @Param days query int false "统计天数（默认 30）" default(30)
// @Success 200 {object} utils.Response{data=[]vo.DashboardNameCount}
// @Router /api/admin/dashboard/model-usage [get]
func ModelUsage(ctx *gin.Context) {
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "30"))
	if days <= 0 || days > 365 {
		days = 30
	}
	items, err := statssvc.ModelUsage(getDB(ctx), days, time.Now())
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询模型占比失败", err)
		return
	}
	utils.OkWithData(ctx, convertNameCounts(items))
}

// AssistantUsage 助手使用排行
// @Summary 助手使用排行
// @Description 近 N 天预设（助手）会话数 TOP N
// @Tags 控制台
// @Produce json
// @Param days query int false "统计天数（默认 30）" default(30)
// @Param limit query int false "返回条数（默认 10）" default(10)
// @Success 200 {object} utils.Response{data=[]vo.DashboardNameCount}
// @Router /api/admin/dashboard/assistant-usage [get]
func AssistantUsage(ctx *gin.Context) {
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "30"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if days <= 0 || days > 365 {
		days = 30
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	items, err := statssvc.AssistantUsage(getDB(ctx), days, limit, time.Now())
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询助手排行失败", err)
		return
	}
	utils.OkWithData(ctx, convertNameCounts(items))
}

// ChannelUsage LLM 渠道用量
// @Summary LLM 渠道用量
// @Description 按渠道聚合的 LLM 请求数/错误率/延迟（进程内口径，自启动起）
// @Tags 控制台
// @Produce json
// @Success 200 {object} utils.Response{data=vo.DashboardChannelUsageVO}
// @Router /api/admin/dashboard/channel-usage [get]
func ChannelUsage(ctx *gin.Context) {
	now := time.Now()
	from := time.Unix(observability.SinceStart(), 0)
	if from.IsZero() {
		from = now.Add(-time.Hour)
	}

	reqs := observability.Summarize("txing_llm_requests_total", "channel", from, now)
	errs := observability.Summarize("txing_llm_errors_total", "channel", from, now)
	lats := observability.Summarize("txing_llm_request_duration_seconds", "channel", from, now)

	errMap := make(map[string]float64, len(errs))
	for _, e := range errs {
		errMap[e.Key] = e.Sum
	}
	latMap := make(map[string]observability.AggSummary, len(lats))
	for _, l := range lats {
		latMap[l.Key] = l
	}

	items := make([]vo.DashboardChannelUsageItem, 0, len(reqs))
	for _, r := range reqs {
		if r.Key == "" {
			continue
		}
		lat := latMap[r.Key]
		item := vo.DashboardChannelUsageItem{
			Channel:  channelDisplayName(r.Key),
			Requests: int64(r.Sum),
			Errors:   int64(errMap[r.Key]),
		}
		if item.Requests > 0 {
			item.ErrorRate = float64(item.Errors) / float64(item.Requests) * 100
		}
		if lat.Sum > 0 {
			item.AvgLatencyMs = lat.Mean * 1000
			item.P95LatencyMs = lat.P95 * 1000
		}
		items = append(items, item)
	}
	// 渠道选择失败（无渠道可用）单独成行
	if e, ok := errMap["none"]; ok && e > 0 {
		items = append(items, vo.DashboardChannelUsageItem{
			Channel: channelDisplayName("none"), Errors: int64(e),
		})
	}

	utils.OkWithData(ctx, vo.DashboardChannelUsageVO{
		SinceStart: sinceStartStr(),
		Items:      items,
	})
}

func channelDisplayName(key string) string {
	if key == "none" {
		return "未匹配到渠道"
	}
	return key
}

// Timeseries 运行时时序
// @Summary 运行时监控时序
// @Description 进程内运行时指标时序：metric=http（QPM/时延）| runtime（协程/内存）| llm（QPM/时延/TTFT/Token）
// @Tags 控制台
// @Produce json
// @Param metric query string false "http | runtime | llm（默认 http）" default(http)
// @Param window query string false "1h | 6h | 24h（默认 1h）" default(1h)
// @Success 200 {object} utils.Response{data=vo.DashboardTimeseriesVO}
// @Router /api/admin/dashboard/timeseries [get]
func Timeseries(ctx *gin.Context) {
	metric := ctx.DefaultQuery("metric", "http")
	window, _ := time.ParseDuration(ctx.DefaultQuery("window", "1h"))
	if window <= 0 {
		window = time.Hour
	}
	now := time.Now()
	from := now.Add(-window)

	out := vo.DashboardTimeseriesVO{StepSeconds: 60, Series: []vo.DashboardTimeseriesSeries{}}
	conv := func(name string, series []observability.SeriesOut) []vo.DashboardTimeseriesSeries {
		res := make([]vo.DashboardTimeseriesSeries, 0, len(series))
		for _, s := range series {
			pts := make([]vo.DashboardTimeseriesPoint, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, vo.DashboardTimeseriesPoint{Ts: p.Ts, Value: p.Value})
			}
			res = append(res, vo.DashboardTimeseriesSeries{Metric: name, Key: s.Key, Points: pts})
		}
		return res
	}
	millis := func(name string, series []observability.SeriesOut) []vo.DashboardTimeseriesSeries {
		res := conv(name, series)
		for i := range res {
			for j := range res[i].Points {
				res[i].Points[j].Value *= 1000
			}
		}
		return res
	}

	switch metric {
	case "runtime":
		out.Series = append(out.Series,
			conv("goroutines", observability.Query(observability.QueryOptions{
				Metric: "go_goroutines", From: from, To: now, Mode: observability.QueryLast,
			}))...)
		mbSeries := conv("memory_mb", observability.Query(observability.QueryOptions{
			Metric: "process_resident_memory_bytes", From: from, To: now, Mode: observability.QueryLast,
		}))
		for i := range mbSeries {
			for j := range mbSeries[i].Points {
				mbSeries[i].Points[j].Value = mbSeries[i].Points[j].Value / 1024 / 1024
			}
		}
		out.Series = append(out.Series, mbSeries...)
	case "llm":
		out.Series = append(out.Series,
			conv("llm_qpm", observability.Query(observability.QueryOptions{
				Metric: "txing_llm_requests_total", From: from, To: now, Mode: observability.QuerySum,
			}))...)
		out.Series = append(out.Series,
			millis("llm_latency_ms", observability.Query(observability.QueryOptions{
				Metric: "txing_llm_request_duration_seconds", From: from, To: now, Mode: observability.QueryHistMean,
			}))...)
		out.Series = append(out.Series,
			millis("llm_ttft_ms", observability.Query(observability.QueryOptions{
				Metric: "txing_llm_ttft_seconds", From: from, To: now, Mode: observability.QueryHistMean,
			}))...)
	default: // http
		out.Series = append(out.Series,
			conv("http_qpm", observability.Query(observability.QueryOptions{
				Metric: "txing_http_requests_total", From: from, To: now, Mode: observability.QuerySum,
			}))...)
		out.Series = append(out.Series,
			millis("http_latency_ms", observability.Query(observability.QueryOptions{
				Metric: "txing_http_request_duration_seconds", From: from, To: now, Mode: observability.QueryHistMean,
			}))...)
	}

	utils.OkWithData(ctx, out)
}

// Activities 最近活动
// @Summary 最近活动
// @Description 最近的会话创建与工作流执行（按时间倒序）
// @Tags 控制台
// @Produce json
// @Param limit query int false "返回条数（默认 20）" default(20)
// @Success 200 {object} utils.Response{data=[]vo.DashboardActivity}
// @Router /api/admin/dashboard/activities [get]
func Activities(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := statssvc.RecentActivities(getDB(ctx), limit)
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询最近活动失败", err)
		return
	}
	out := make([]vo.DashboardActivity, 0, len(items))
	for _, a := range items {
		out = append(out, vo.DashboardActivity{
			Time: a.Time, User: a.User, Action: a.Action, Detail: a.Detail, Status: a.Status,
		})
	}
	utils.OkWithData(ctx, out)
}

func convertNameCounts(items []statssvc.NameCount) []vo.DashboardNameCount {
	out := make([]vo.DashboardNameCount, 0, len(items))
	for _, it := range items {
		out = append(out, vo.DashboardNameCount{Name: it.Name, Count: it.Count})
	}
	return out
}

// formatCount 大数友好显示（1234 → 1,234；1234567 → 1.23M）
func formatCount(v float64) string {
	switch {
	case math.Abs(v) >= 1_000_000:
		return strconv.FormatFloat(v/1_000_000, 'f', 2, 64) + "M"
	case math.Abs(v) >= 10_000:
		return strconv.FormatFloat(v/1000, 'f', 1, 64) + "k"
	default:
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
}
