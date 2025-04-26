package dto

import "txing-ai/internal/utils/page"

// CreateChannelReq 创建渠道请求
// Create channel request
type CreateChannelReq struct {
	Name     string   `json:"name" binding:"required"`     // 渠道名称 Channel name
	Type     string   `json:"type" binding:"required"`     // 渠道类型 Channel type
	Priority int      `json:"priority"`                    // 优先级 Priority
	Weight   int      `json:"weight"`                      // 权重 Weight
	Models   []string `json:"models" binding:"required"`   // 支持的模型列表 Supported models
	Retry    int      `json:"retry"`                       // 重试次数 Retry times
	Secret   string   `json:"secret" binding:"required"`   // 密钥 Secret key
	Endpoint string   `json:"endpoint" binding:"required"` // 服务地址 Service endpoint
	Status   bool     `json:"status"`                      // 启用状态 Enable status
}

// UpdateChannelReq 更新渠道请求
// Update channel request
type UpdateChannelReq struct {
	Name     string   `json:"name"`     // 渠道名称 Channel name
	Type     string   `json:"type"`     // 渠道类型 Channel type
	Priority int      `json:"priority"` // 优先级 Priority
	Weight   int      `json:"weight"`   // 权重 Weight
	Models   []string `json:"models"`   // 支持的模型列表 Supported models
	Retry    int      `json:"retry"`    // 重试次数 Retry times
	Secret   string   `json:"secret"`   // 密钥 Secret key
	Endpoint string   `json:"endpoint"` // 服务地址 Service endpoint
	Status   bool     `json:"status"`   // 启用状态 Enable status
}

// ListChannelReq 渠道列表请求
// List channel request
type ListChannelReq struct {
	page.PageRequest
	Type   string `form:"type"`   // 渠道类型 Channel type
	Status *bool  `form:"status"` // 启用状态 Enable status
	Name   string `form:"name"`   // 渠道名称 Channel name
}
