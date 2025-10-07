package domain

import (
	"github.com/samber/lo"
	"math/rand"
	"strings"
	"txing-ai/internal/global"
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
	// 模型映射关系，使用 JSON 格式存储
	// 配置示例：
	// {
	// 	"mappings": [
	// 	  {
	// 		"sourceModel": "deepseek-r1",
	// 		"conditions": [
	// 		  {
	// 			"targetModel": "deepseek-r1-250120",
	// 			"conditions": {
	// 			  "enableWeb": true
	// 			}
	// 		  },
	// 		  {
	// 			"targetModel": "deepseek-r1-250121",
	// 			"conditions": {
	// 			  "enableWeb": false
	// 			}
	// 		  }
	// 		]
	// 	  }
	// 	]
	// }
	Mappings []global.ModelMapping `gorm:"type:longtext;serializer:json;comment:模型映射关系" json:"mappings"`
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

// GetMappingModel 获取映射后的模型
func (c *Channel) GetMappingModel(model string, mappingParams map[string]interface{}) string {
	// 遍历映射规则
	for _, mapping := range c.Mappings {
		if mapping.SourceModel != model {
			continue
		}

		// 遍历条件列表
		for _, condition := range mapping.Conditions {
			// 检查所有条件是否满足
			allConditionsMet := true

			// 收集遍历过的条件 key
			checkedKeys := make(map[string]struct{})

			for key, expectedValue := range condition.Conditions {
				// 保存遍历过的条件 key
				checkedKeys[key] = struct{}{}

				actualValue, exists := mappingParams[key]
				if allConditionsMet && (!exists || actualValue != expectedValue) {
					allConditionsMet = false
				}
			}

			// 过滤出 mappingParams 中没有遍历过的条件
			// 使用 lo.PickBy 过滤出未校验过的参数键值对
			uncheckedParams := lo.PickBy(mappingParams, func(k string, v interface{}) bool {
				_, ok := checkedKeys[k]
				return !ok
			})
			for key := range uncheckedParams {
				// 判断是否是默认值
				val, exists := mappingParams[key]
				defaultVal, existsDefault := global.ChannelModelMappingConditionDefaultVal[key]
				if exists {
					if !existsDefault || val != defaultVal {
						allConditionsMet = false
					}
				}
			}

			// 如果所有条件都满足，返回目标模型
			if allConditionsMet {
				return condition.TargetModel
			}
		}
	}

	// 通过映射关系没有找到匹配的模型，判断 mappingParams 中的条件是否均为默认值，如果是则返回原始模型，否则返回空字符串
	for key, val := range mappingParams {
		defaultVal, existsDefault := global.ChannelModelMappingConditionDefaultVal[key]
		if existsDefault && val != defaultVal {
			return ""
		}
	}

	// 如果没有找到匹配的映射规则，返回原始模型
	return model
}

var _ iface.ChannelConfig = (*Channel)(nil)
