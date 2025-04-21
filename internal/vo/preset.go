package vo

import (
	"time"
	"txing-ai/internal/domain"
)

// PresetVO 预设视图对象
type PresetVO struct {
	ID          int64     `json:"id"`                                              // 主键ID
	UserID      *int64    `json:"userId" example:"1"`                              // 用户ID
	Avatar      string    `json:"avatar" example:"https://example.com/avatar.png"` // 预设头像
	Name        string    `json:"name" example:"GPT助手"`                            // 预设名称
	Description string    `json:"description" example:"一个智能的GPT助手"`                // 预设描述
	Context     string    `json:"context" example:"你是一个智能助手..."`                   // 预设上下文
	Official    bool      `json:"official" example:"false"`                        // 是否官方预设
	CreatedAt   time.Time `json:"createdAt"`                                       // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`                                       // 更新时间
}

// ToPresetVO 将 Preset 转换为 PresetVO
func ToPresetVO(preset domain.Preset) PresetVO {
	return PresetVO{
		ID:          preset.Id,
		UserID:      preset.UserID,
		Avatar:      preset.Avatar,
		Name:        preset.Name,
		Description: preset.Description,
		Context:     preset.Context,
		Official:    preset.Official,
		CreatedAt:   preset.CreateTime,
		UpdatedAt:   preset.UpdateTime,
	}
}

// ToPresetVOs 将 Preset 切片转换为 PresetVO 切片
func ToPresetVOs(presets []domain.Preset) []PresetVO {
	vos := make([]PresetVO, len(presets))
	for i, preset := range presets {
		vos[i] = ToPresetVO(preset)
	}
	return vos
}
