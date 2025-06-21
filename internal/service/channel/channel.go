package channel

import (
	"txing-ai/internal/domain"
	"txing-ai/internal/global/logging/log"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

// 定义渠道序列
type Sequence = *[]domain.Channel

// GetAllChannelsByModel 查询支持指定模型的所有渠道
// TODO 提前加载到内存中，减少查询次数
func GetAllChannelsByModel(db *gorm.DB, model string) Sequence {
	var channels []domain.Channel
	result := make([]domain.Channel, 0)

	// 查询所有启用的渠道
	if err := db.Find(&channels).Error; err != nil {
		log.Error("query all channels error", zap.Error(err))
		return &result
	}

	// 使用 lo.Filter 过滤支持指定模型的渠道
	result = lo.Filter(channels, func(channel domain.Channel, _ int) bool {
		return lo.Contains(channel.Models, model)
	})

	return &result
}
