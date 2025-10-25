package channel

import (
	"errors"
	"math/rand"
	"txing-ai/internal/domain"
	"txing-ai/internal/global/logging/log"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

// 定义渠道序列
type Sequence = *[]domain.Channel

// getAllChannelsByModel 查询支持指定模型的所有渠道
// TODO 提前加载到内存中，减少查询次数
func getAllChannelsByModel(db *gorm.DB, model string, mappingParams map[string]interface{}) Sequence {
	var channels []domain.Channel
	result := make([]domain.Channel, 0)

	// 查询所有启用的渠道
	if err := db.Where("status = ?", 1).Find(&channels).Error; err != nil {
		log.Error("query all channels error", zap.Error(err))
		return &result
	}

	// 使用 lo.Filter 过滤支持指定模型的渠道
	result = lo.Filter(channels, func(channel domain.Channel, _ int) bool {
		return lo.Contains(channel.Models, model)
	})

	return &result
}

// 指定模型，返回选用的渠道
func ChooseChannelAndModel(db *gorm.DB, model string, mappingParams map[string]interface{}) (channel *domain.Channel, mappingModel string, error error) {
	sequence := getAllChannelsByModel(db, model, mappingParams)

	// 判断是否有支持该模型的 channel
	if len(*sequence) == 0 {
		log.Error("no channel found for model ", zap.String("model", model))
		return nil, "", errors.New("no channel found for model " + model)
	}

	// 根据 mappingParams 过滤出最终满足条件的 channel
	filteredSequence := lo.Filter(*sequence, func(c domain.Channel, _ int) bool {
		return c.GetMappingModel(model, mappingParams) != ""
	})

	// TODO 后续优化为根据优先级和权重选择 以及实现重试机制
	// 从中随机选择一个 channel
	targetChannel := filteredSequence[rand.Intn(len(filteredSequence))]
	return &targetChannel, targetChannel.GetMappingModel(model, mappingParams), nil
}
