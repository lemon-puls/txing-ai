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
	Role             string  `json:"role" example:"assistant"`               // 消息角色
	Content          string  `json:"content" example:"消息内容"`                // 消息内容
	ReasoningContent string  `json:"reasoningContent,omitempty" example:""`   // 思考过程内容
	Name             *string `json:"name,omitempty"`                          // 消息名称
	// 工作流相关字段
	WorkflowStatus string   `json:"workflowStatus,omitempty" example:"completed"` // 工作流状态 completed/failed
	Artifacts      string   `json:"artifacts,omitempty"`                           // 产物 JSON
	AppName        string   `json:"appName,omitempty" example:"AI 应用名称"`         // 应用名称
	Files          []string `json:"files,omitempty"`                               // 用户上传的文件名列表
	ExecutionLogs  string   `json:"executionLogs,omitempty"`                       // 节点执行日志 JSON
	// 多模态相关字段
	Images      []string        `json:"images,omitempty"`      // 图片 URL 列表
	Attachments []AttachmentVO  `json:"attachments,omitempty"` // 文件附件列表
}

// AttachmentVO 文件附件信息
type AttachmentVO struct {
	FileName string `json:"fileName"` // 文件名
	FileURL  string `json:"fileUrl"`  // 文件 URL
	FileType string `json:"fileType"` // 文件类型
	FileSize int64  `json:"fileSize"` // 文件大小
}
