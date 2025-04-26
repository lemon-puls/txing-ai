package domain

import (
	"math/rand"
	"strings"
	"txing-ai/internal/iface"
)

type Channel struct {
	BaseModel
	Name     string   `gorm:"type:varchar(255);not null;comment:渠道名称" json:"name"`
	Type     string   `gorm:"type:varchar(50);not null;comment:渠道类型" json:"type"`
	Priority int      `gorm:"type:int;default:0;comment:优先级" json:"priority"`
	Weight   int      `gorm:"type:int;default:0;comment:权重" json:"weight"`
	Models   []string `gorm:"type:json;serializer:json;comment:支持的模型列表" json:"models"`
	Retry    int      `gorm:"type:int;default:3;comment:重试次数" json:"retry"`
	Secret   string   `gorm:"type:varchar(255);comment:密钥" json:"secret"`
	Endpoint string   `gorm:"type:varchar(255);not null;comment:服务地址" json:"endpoint"`
	Status   bool     `gorm:"type:int;default:1;comment:启用状态(0: 禁用 1: 启用)" json:"status"`
}

func (c *Channel) GetId() int64 {
	return c.Id
}

// 获取Channel类型的函数
func (c *Channel) GetType() string {
	// 返回Channel的Type属性
	return c.Type
}

func (c *Channel) GetRetry() int {
	return c.Retry
}

// 随机获取 Secret
func (c *Channel) GetRandomSecret() string {
	split := strings.Split(c.Secret, "\n")
	if len(split) == 0 {
		return ""
	}
	// 生成随机 index
	index := rand.Intn(len(split))
	return split[index]
}

func (c *Channel) GetEndpoint() string {
	return c.Endpoint
}

var _ iface.ChannelConfig = (*Channel)(nil)
