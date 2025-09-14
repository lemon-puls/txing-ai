package dto

import "txing-ai/internal/utils/page"

// CreateWebsiteReq 创建网站请求
type CreateWebsiteReq struct {
	Name        string `json:"name" binding:"required,min=2,max=50" example:"GitHub"`                 // 网站名称
	Description string `json:"description" binding:"required,max=200" example:"全球最大的代码托管平台"`          // 网站描述
	Url         string `json:"url" binding:"required,url" example:"https://github.com"`               // 网站地址
	Avatar      string `json:"avatar" example:"https://github.githubassets.com/favicons/favicon.svg"` // 网站头像
	Tags        string `json:"tags" binding:"required" example:"开发,代码,开源"`                            // 标签数组
	Sort        int    `json:"sort" example:"0"`                                                      // 排序
	Status      int    `json:"status" binding:"oneof=0 1" example:"1"`                                // 状态 1启用 0禁用
}

// UpdateWebsiteReq 更新网站请求
type UpdateWebsiteReq struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=50" example:"GitHub"`                // 网站名称
	Description string `json:"description" binding:"omitempty,max=200" example:"全球最大的代码托管平台"`         // 网站描述
	Url         string `json:"url" binding:"omitempty,url" example:"https://github.com"`              // 网站地址
	Avatar      string `json:"avatar" example:"https://github.githubassets.com/favicons/favicon.svg"` // 网站头像
	Tags        string `json:"tags" binding:"omitempty" example:"开发,代码,开源"`                           // 标签数组
	Sort        *int   `json:"sort" example:"0"`                                                      // 排序
	Status      *int   `json:"status" binding:"omitempty,oneof=0 1" example:"1"`                      // 状态 1启用 0禁用
}

// ListWebsiteReq 获取网站列表请求
type ListWebsiteReq struct {
	page.PageRequest
	Name   string `form:"name" example:"GitHub"` // 网站名称
	Tag    string `form:"tag" example:"开发"`      // 标签
	Status *int   `form:"status" example:"1"`    // 状态
}

// GetFaviconReq 获取网站图标请求
type GetFaviconReq struct {
	Url string `json:"url" binding:"required,url" example:"https://github.com"` // 网站地址
}
