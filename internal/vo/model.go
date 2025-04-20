package vo

import (
	"time"
	"txing-ai/internal/domain"
)

// ModelVO 模型视图对象
type ModelVO struct {
	ID          int64     `json:"id"`                                              // 主键ID
	Name        string    `json:"name" example:"gpt-3.5-turbo"`                    // 模型名称
	Description string    `json:"description" example:"GPT-3.5 Turbo模型"`         // 模型描述
	Default     bool      `json:"default" example:"false"`                         // 是否为默认模型
	HighContext bool      `json:"high_context" example:"false"`                    // 是否支持高上下文
	Avatar      string    `json:"avatar" example:"https://example.com/avatar.png"` // 模型头像
	Tag         string    `json:"tag" example:"GPT,对话"`                          // 模型标签
	CreatedAt   time.Time `json:"created_at"`                                      // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                                      // 更新时间
}

// ToModelVO 将 Model 转换为 ModelVO
func ToModelVO(model domain.Model) ModelVO {
	return ModelVO{
		ID:          model.Id,
		Name:        model.Name,
		Description: model.Description,
		Default:     model.Default,
		HighContext: model.HighContext,
		Avatar:      model.Avatar,
		Tag:         model.Tag,
		CreatedAt:   model.CreateTime,
		UpdatedAt:   model.UpdateTime,
	}
}

// ToModelVOs 将 Model 切片转换为 ModelVO 切片
func ToModelVOs(models []domain.Model) []ModelVO {
	vos := make([]ModelVO, len(models))
	for i, model := range models {
		vos[i] = ToModelVO(model)
	}
	return vos
}
