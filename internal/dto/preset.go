package dto

import "txing-ai/internal/utils/page"

// CreatePresetReq 创建预设请求
type CreatePresetReq struct {
	UserID      *int64 `json:"userId" example:"1"`                              // 用户ID
	Avatar      string `json:"avatar" example:"https://example.com/avatar.png"` // 预设头像
	Name        string `json:"name" binding:"required" example:"GPT助手"`         // 预设名称
	Description string `json:"description" example:"一个智能的GPT助手"`                // 预设描述
	Context     string `json:"context" example:"你是一个智能助手,能够帮助用户解决各种问题..."`      // 预设上下文
	Official    bool   `json:"official" example:"false"`                        // 是否官方预设
}

// UpdatePresetReq 更新预设请求
type UpdatePresetReq struct {
	UserID      *int64 `json:"userId" example:"1"`                              // 用户ID
	Avatar      string `json:"avatar" example:"https://example.com/avatar.png"` // 预设头像
	Name        string `json:"name" example:"GPT助手"`                            // 预设名称
	Description string `json:"description" example:"一个智能的GPT助手"`                // 预设描述
	Context     string `json:"context" example:"你是一个智能助手,能够帮助用户解决各种问题..."`      // 预设上下文
	Official    *bool  `json:"official" example:"false"`                        // 是否官方预设
}

// ListPresetReq 获取预设列表请求
type ListPresetReq struct {
	page.PageRequest
	Official *bool  `form:"official"`           // 是否官方预设
	UserID   *int64 `form:"userId" example:"1"` // 用户ID
	Name     string `form:"name"`               // 预设名称
}
