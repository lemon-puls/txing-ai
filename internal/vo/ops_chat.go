package vo

import (
	"time"
	"txing-ai/internal/domain"
)

// OpsChatSessionSimpleVO 运营助手会话列表项
type OpsChatSessionSimpleVO struct {
	Id         int64     `json:"id"`         // 会话ID
	Title      string    `json:"title"`      // 会话标题
	CreateTime time.Time `json:"createTime"` // 创建时间
	UpdateTime time.Time `json:"updateTime"` // 更新时间
}

// OpsChatSessionDetailVO 运营助手会话详情（含完整消息，用于刷新后回放）
type OpsChatSessionDetailVO struct {
	Id         int64                   `json:"id"`
	Title      string                  `json:"title"`
	Model      string                  `json:"model"`
	Messages   []domain.OpsChatMessage `json:"messages"`
	CreateTime time.Time               `json:"createTime"`
	UpdateTime time.Time               `json:"updateTime"`
}
