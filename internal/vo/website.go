package vo

import (
	"time"
	"txing-ai/internal/domain"
)

// WebsiteVO 网站视图对象
type WebsiteVO struct {
	ID          int64     `json:"id"`                                                                    // 主键ID
	Name        string    `json:"name" example:"GitHub"`                                                 // 网站名称
	Description string    `json:"description" example:"全球最大的代码托管平台"`                                     // 网站描述
	Url         string    `json:"url" example:"https://github.com"`                                      // 网站地址
	Avatar      string    `json:"avatar" example:"https://github.githubassets.com/favicons/favicon.svg"` // 网站头像
	Tags        string    `json:"tags" example:"开发,代码,开源"`                                               // 标签数组
	Sort        int       `json:"sort" example:"0"`                                                      // 排序
	Status      int       `json:"status" example:"1"`                                                    // 状态 1启用 0禁用
	CreatedAt   time.Time `json:"createdAt"`                                                             // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`                                                             // 更新时间
}

// ToWebsiteVO 将 Website 转换为 WebsiteVO
func ToWebsiteVO(website domain.Website) WebsiteVO {

	return WebsiteVO{
		ID:          website.Id,
		Name:        website.Name,
		Description: website.Description,
		Url:         website.Url,
		Avatar:      website.Avatar,
		Tags:        website.Tags,
		Sort:        website.Sort,
		Status:      website.Status,
		CreatedAt:   website.CreateTime,
		UpdatedAt:   website.UpdateTime,
	}
}

// ToWebsiteVOs 将 Website 切片转换为 WebsiteVO 切片
func ToWebsiteVOs(websites []domain.Website) []WebsiteVO {
	vos := make([]WebsiteVO, len(websites))
	for i, website := range websites {
		vos[i] = ToWebsiteVO(website)
	}
	return vos
}

// GetFaviconVO 获取网站图标响应
type GetFaviconVO struct {
	Favicon string `json:"favicon" example:"https://github.githubassets.com/favicons/favicon.svg"` // 网站图标URL
}
