package dto

import "txing-ai/internal/utils/page"

// CreateModelReq 创建模型请求
type CreateModelReq struct {
	Name        string `json:"name" binding:"required" example:"gpt-3.5-turbo"` // 模型名称
	Description string `json:"description" example:"GPT-3.5 Turbo模型"`           // 模型描述
	Default     bool   `json:"default" example:"false"`                         // 是否为默认模型
	HighContext bool   `json:"high_context" example:"false"`                    // 是否支持高上下文
	Avatar      string `json:"avatar" example:"https://example.com/avatar.png"` // 模型头像
	Tag         string `json:"tag" example:"GPT,对话"`                            // 模型标签
}

// UpdateModelReq 更新模型请求
type UpdateModelReq struct {
	Name        string `json:"name" binding:"required" example:"gpt-3.5-turbo"` // 模型名称
	Description string `json:"description" example:"GPT-3.5 Turbo模型"`           // 模型描述
	Default     bool   `json:"default" example:"false"`                         // 是否为默认模型
	HighContext bool   `json:"high_context" example:"false"`                    // 是否支持高上下文
	Avatar      string `json:"avatar" example:"https://example.com/avatar.png"` // 模型头像
	Tag         string `json:"tag" example:"GPT,对话"`                            // 模型标签
}

// ListModelReq 获取模型列表请求
type ListModelReq struct {
	page.PageRequest
	Tag     string `form:"tag" example:"GPT"` // 标签
	Default *bool  `form:"default"`           // 是否为默认模型
}
