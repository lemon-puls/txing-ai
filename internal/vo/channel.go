package vo

import (
	"github.com/samber/lo"
	"time"
	"txing-ai/internal/domain"
	"txing-ai/internal/global"
)

// ChannelVO 渠道视图对象
// Channel view object
type ChannelVO struct {
	Id       int64                 `json:"id"`        // 渠道ID Channel ID
	Name     string                `json:"name"`      // 渠道名称 Channel name
	Type     string                `json:"type"`      // 渠道类型 Channel type
	Priority int                   `json:"priority"`  // 优先级 Priority
	Weight   int                   `json:"weight"`    // 权重 Weight
	Models   []string              `json:"models"`    // 支持的模型列表 Supported models
	Retry    int                   `json:"retry"`     // 重试次数 Retry times
	Secret   string                `json:"secret"`    // 密钥 Secret key
	Endpoint string                `json:"endpoint"`  // 服务地址 Service endpoint
	Status   bool                  `json:"status"`    // 启用状态 Enable status
	CreateAt time.Time             `json:"create_at"` // 创建时间 Create time
	UpdateAt time.Time             `json:"update_at"` // 更新时间 Update time
	Mappings []global.ModelMapping `json:"mappings"`  // 模型映射关系 Mappings
}

// ChannelListVO 渠道列表视图对象
// Channel list view object
type ChannelListVO struct {
	Total   int64       `json:"total"`   // 总数 Total count
	Records []ChannelVO `json:"records"` // 记录列表 Record list
}

// ToChannelVO 将 Channel 实体转换为 VO
// Convert Channel entity to VO
func ToChannelVO(channel domain.Channel) ChannelVO {
	return ChannelVO{
		Id:       channel.GetId(),
		Name:     channel.Name,
		Type:     channel.Type,
		Priority: channel.Priority,
		Weight:   channel.Weight,
		Models:   channel.Models,
		Retry:    channel.Retry,
		Secret:   channel.Secret,
		Endpoint: channel.Endpoint,
		Status:   channel.Status,
		CreateAt: channel.CreateTime,
		UpdateAt: channel.UpdateTime,
		Mappings: channel.Mappings,
	}
}

func ToChannelVOs(users []domain.Channel) []ChannelVO {
	channelVos := lo.Map(users, func(channel domain.Channel, _ int) ChannelVO {
		return ToChannelVO(channel)
	})

	return channelVos
}
