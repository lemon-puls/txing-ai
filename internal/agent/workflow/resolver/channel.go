package resolver

import (
	"errors"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	channelservice "txing-ai/internal/service/channel"

	"gorm.io/gorm"
)

// ChannelModelResolver 基于渠道服务的模型解析器实现
type ChannelModelResolver struct {
	db *gorm.DB
}

// NewChannelModelResolver 创建基于渠道服务的模型解析器
func NewChannelModelResolver(db *gorm.DB) *ChannelModelResolver {
	return &ChannelModelResolver{db: db}
}

// Resolve 根据模型名称解析对应的端点和密钥
func (r *ChannelModelResolver) Resolve(modelName string) (*types.ModelInfo, error) {
	if modelName == "" {
		return nil, errors.New("模型名称为空")
	}

	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}

	channel, mappingModel, err := channelservice.ChooseChannelAndModel(r.db, modelName, mappingParams)
	if err != nil {
		return nil, err
	}

	return &types.ModelInfo{
		Endpoint: channel.GetEndpoint(),
		APIKey:   channel.GetRandomSecret(),
		Model:    mappingModel,
	}, nil
}
