package vo

import "time"

// ConversationSimpleVO 会话基本信息
type ConversationSimpleVO struct {
	ID         int64     `json:"id"`         // 会话ID
	Name       string    `json:"name"`       // 会话标题
	Model      string    `json:"model"`      // 使用的模型ID
	CreateTime time.Time `json:"createTime"` // 创建时间
	UpdateTime time.Time `json:"updateTime"` // 更新时间
}
