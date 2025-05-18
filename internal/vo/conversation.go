package vo

import (
	"time"
)

// ConversationSimpleVO 会话基本信息
type ConversationSimpleVO struct {
	ID         int64     `json:"id"`         // 会话ID
	Name       string    `json:"name"`       // 会话标题
	Model      string    `json:"model"`      // 使用的模型ID
	Avatar     string    `json:"avatar"`     // 会话头像
	CreateTime time.Time `json:"createTime"` // 创建时间
	UpdateTime time.Time `json:"updateTime"` // 更新时间
	PresetId   int64     `json:"presetId"`   // 预设ID
}

// ConversationDetailVO 会话详情信息
type ConversationDetailVO struct {
	ID        int64  `json:"id"` // 会话ID
	UserID    int64  `json:"userId"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	EnableWeb bool   `json:"enableWeb"`
	Context   int    `json:"context"`

	// 可选的模型参数
	MaxTokens         *int     `json:"maxTokens,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"topP,omitempty"`
	TopK              *int     `json:"topK,omitempty"`
	PresencePenalty   *float32 `json:"presencePenalty,omitempty"`
	FrequencyPenalty  *float32 `json:"frequencyPenalty,omitempty"`
	RepetitionPenalty *float32 `json:"repetitionPenalty,omitempty"`

	Messages []MessageVO `json:"messages"`         // 消息列表
	Preset   *PresetVO   `json:"preset,omitempty"` // 预设信息
}

// 聊天消息
type MessageVO struct {
	Role             string  `json:"role"`
	Content          string  `json:"content"`
	ReasoningContent string  `json:"reasoningContent"`
	Name             *string `json:"name,omitempty"`
}
