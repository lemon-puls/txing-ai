package channel

import (
	"errors"
	"fmt"
	"math/rand"
	"txing-ai/internal/domain"
	"txing-ai/internal/global/logging/log"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

// 渠道选择相关的错误，供上层识别失败原因并给出用户提示
var (
	// ErrNoChannelFound 没有任何启用且支持该模型的渠道
	ErrNoChannelFound = errors.New("no channel found for model")
	// ErrNoAvailableChannel 有支持该模型的渠道，但所有渠道的映射条件均不满足当前请求参数
	ErrNoAvailableChannel = errors.New("no available channel for model with current parameters")
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
		return nil, "", fmt.Errorf("%w: %s", ErrNoChannelFound, model)
	}

	// 根据 mappingParams 过滤出最终满足条件的 channel
	filteredSequence := lo.Filter(*sequence, func(c domain.Channel, _ int) bool {
		return c.GetMappingModel(model, mappingParams) != ""
	})

	// 过滤后没有可用的 channel（例如所有渠道的模型映射条件都不满足当前请求参数），
	// 直接返回错误，避免 rand.Intn(0) 触发 "invalid argument to Intn" panic
	if len(filteredSequence) == 0 {
		log.Error("no available channel for model with current params",
			zap.String("model", model),
			zap.Any("mappingParams", mappingParams))
		return nil, "", fmt.Errorf("%w: %s, params: %v", ErrNoAvailableChannel, model, mappingParams)
	}

	return chooseFromFiltered(filteredSequence, model, mappingParams)
}

// chooseFromFiltered 从过滤后的 channel 列表中随机选择一个 channel，并返回映射后的模型
// filteredSequence 为空时返回错误，避免 rand.Intn(0) 触发 "invalid argument to Intn" panic
func chooseFromFiltered(filteredSequence []domain.Channel, model string, mappingParams map[string]interface{}) (*domain.Channel, string, error) {
	if len(filteredSequence) == 0 {
		return nil, "", fmt.Errorf("%w: %s", ErrNoAvailableChannel, model)
	}

	// TODO 后续优化为根据优先级和权重选择 以及实现重试机制
	// 从中随机选择一个 channel
	targetChannel := filteredSequence[rand.Intn(len(filteredSequence))]
	return &targetChannel, targetChannel.GetMappingModel(model, mappingParams), nil
}
