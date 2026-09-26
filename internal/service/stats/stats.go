// Package stats 管理后台控制台的 DB 业务统计聚合层：
// 对话量/活跃用户/模型占比/助手排行/最近活动等长期运营数据来自 MySQL 业务表，
// 与运行时指标（observability 环形缓冲）互补
package stats

import (
	"time"

	"gorm.io/gorm"
)

// ComparePair 今日 vs 昨日同期时段（对比口径更公平）
type ComparePair struct {
	Today     int64 `json:"today"`
	Yesterday int64 `json:"yesterdaySamePeriod"`
}

// TrendPoint 单日/单月趋势点
type TrendPoint struct {
	Date         string `json:"date"`
	Conversations int64 `json:"conversations"`
	ActiveUsers   int64 `json:"activeUsers"`
}

// NameCount 通用名称-数量统计（模型占比/助手排行）
type NameCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// Activity 最近活动条目
type Activity struct {
	Time   time.Time `json:"time"`
	User   string    `json:"user"`
	Action string    `json:"action"`
	Detail string    `json:"detail"`
	Status string    `json:"status"`
}

// ConversationToday 今日对话数（对比昨日同期）
func ConversationToday(db *gorm.DB, now time.Time) (ComparePair, error) {
	var res ComparePair
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yStart := todayStart.AddDate(0, 0, -1)

	err := db.Table("conversation").
		Where("create_time >= ? AND create_time < ?", todayStart, now).
		Where("delete_time IS NULL").
		Count(&res.Today).Error
	if err != nil {
		return res, err
	}
	err = db.Table("conversation").
		Where("create_time >= ? AND create_time < ?", yStart, yStart.Add(now.Sub(todayStart))).
		Where("delete_time IS NULL").
		Count(&res.Yesterday).Error
	return res, err
}

// ActiveUsersToday 今日活跃用户数（对比昨日同期）
func ActiveUsersToday(db *gorm.DB, now time.Time) (ComparePair, error) {
	var res ComparePair
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yStart := todayStart.AddDate(0, 0, -1)

	err := db.Table("conversation").
		Where("create_time >= ? AND create_time < ?", todayStart, now).
		Where("delete_time IS NULL").
		Distinct("user_id").Count(&res.Today).Error
	if err != nil {
		return res, err
	}
	err = db.Table("conversation").
		Where("create_time >= ? AND create_time < ?", yStart, yStart.Add(now.Sub(todayStart))).
		Where("delete_time IS NULL").
		Distinct("user_id").Count(&res.Yesterday).Error
	return res, err
}

// ConversationTrend 对话趋势（按日聚合；days>=360 时按月聚合）
func ConversationTrend(db *gorm.DB, days int, now time.Time) ([]TrendPoint, string, error) {
	monthly := days >= 360
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days + 1)

	type row struct {
		Bucket string
		Cnt    int64
		Users  int64
	}
	var rows []row

	query := db.Table("conversation").
		Where("create_time >= ?", from).
		Where("delete_time IS NULL")
	if monthly {
		query = query.Select("DATE_FORMAT(create_time, '%Y-%m') AS bucket, COUNT(*) AS cnt, COUNT(DISTINCT user_id) AS users")
	} else {
		query = query.Select("DATE(create_time) AS bucket, COUNT(*) AS cnt, COUNT(DISTINCT user_id) AS users")
	}
	if err := query.Group("bucket").Order("bucket ASC").Find(&rows).Error; err != nil {
		return nil, "", err
	}

	granularity := "day"
	if monthly {
		granularity = "month"
	}
	points := make([]TrendPoint, 0, len(rows))
	for _, r := range rows {
		points = append(points, TrendPoint{Date: r.Bucket, Conversations: r.Cnt, ActiveUsers: r.Users})
	}
	return points, granularity, nil
}

// ModelUsage 模型使用占比（近 N 天，按会话数）
func ModelUsage(db *gorm.DB, days int, now time.Time) ([]NameCount, error) {
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days + 1)

	type row struct {
		Name string
		Cnt  int64
	}
	var rows []row
	err := db.Table("conversation").
		Select("model AS name, COUNT(*) AS cnt").
		Where("create_time >= ?", from).
		Where("delete_time IS NULL").
		Group("model").Order("cnt DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]NameCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, NameCount{Name: r.Name, Count: r.Cnt})
	}
	return out, nil
}

// AssistantUsage 助手（预设）使用排行 TOP N（近 N 天）
func AssistantUsage(db *gorm.DB, days, limit int, now time.Time) ([]NameCount, error) {
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days + 1)

	type row struct {
		Name string
		Cnt  int64
	}
	var rows []row
	err := db.Table("conversation").
		Select("preset.name AS name, COUNT(*) AS cnt").
		Joins("JOIN preset ON preset.id = conversation.preset_id").
		Where("conversation.create_time >= ?", from).
		Where("conversation.delete_time IS NULL").
		Where("conversation.preset_id IS NOT NULL").
		Where("preset.delete_time IS NULL").
		Group("preset.name").Order("cnt DESC").Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]NameCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, NameCount{Name: r.Name, Count: r.Cnt})
	}
	return out, nil
}

// RecentActivities 最近活动：会话创建 + 工作流执行按时间归并
func RecentActivities(db *gorm.DB, limit int) ([]Activity, error) {
	type convRow struct {
		Time     time.Time
		Username string
		Model    string
	}
	var convs []convRow
	if err := db.Table("conversation").
		Select("conversation.create_time AS time, user.username AS username, conversation.model AS model").
		Joins("LEFT JOIN user ON user.id = conversation.user_id").
		Where("conversation.delete_time IS NULL").
		Order("conversation.create_time DESC").Limit(limit).
		Find(&convs).Error; err != nil {
		return nil, err
	}

	type wfRow struct {
		Time     time.Time
		Username string
		Name     string
		Status   string
	}
	var wfs []wfRow
	if err := db.Table("workflow_executions").
		Select("workflow_executions.create_time AS time, user.username AS username, agent_flows.name AS name, workflow_executions.status AS status").
		Joins("LEFT JOIN conversation ON conversation.id = workflow_executions.conversation_id").
		Joins("LEFT JOIN user ON user.id = conversation.user_id").
		Joins("LEFT JOIN agent_flows ON agent_flows.id = workflow_executions.workflow_id").
		Where("workflow_executions.delete_time IS NULL").
		Order("workflow_executions.create_time DESC").Limit(limit).
		Find(&wfs).Error; err != nil {
		return nil, err
	}

	acts := make([]Activity, 0, len(convs)+len(wfs))
	for _, c := range convs {
		user := c.Username
		if user == "" {
			user = "匿名用户"
		}
		acts = append(acts, Activity{
			Time: c.Time, User: user, Action: "创建对话", Detail: c.Model, Status: "completed",
		})
	}
	for _, w := range wfs {
		user := w.Username
		if user == "" {
			user = "匿名用户"
		}
		acts = append(acts, Activity{
			Time: w.Time, User: user, Action: "执行工作流", Detail: w.Name, Status: w.Status,
		})
	}

	// 归并排序取前 limit 条
	sortSlice(acts)
	if len(acts) > limit {
		acts = acts[:limit]
	}
	return acts, nil
}

func sortSlice(acts []Activity) {
	for i := 1; i < len(acts); i++ {
		for j := i; j > 0 && acts[j].Time.After(acts[j-1].Time); j-- {
			acts[j], acts[j-1] = acts[j-1], acts[j]
		}
	}
}
