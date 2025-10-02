package eino_openai

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"io"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

type ChatClient struct {
	Endpoint string
	ApiKey   string
}

func (c ChatClient) StreamChat(ctx context.Context, conf *adaptercommon.ChatConfig, callback global.Hook) error {
	// 构建消息
	messages := make([]*schema.Message, 0, len(conf.Message))
	for _, msg := range conf.Message {
		messages = append(messages, &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.Content,
		})
	}

	chatModelConfig := &openai.ChatModelConfig{
		Model:   conf.Model, // 使用的模型版本
		APIKey:  c.ApiKey,   // OpenAI API 密钥
		BaseURL: c.Endpoint, // OpenAI API 地址
	}

	// 设置可选参数
	if conf.MaxTokens != nil {
		chatModelConfig.MaxTokens = conf.MaxTokens
	}
	if conf.Temperature != nil {
		chatModelConfig.Temperature = conf.Temperature
	}
	if conf.TopP != nil {
		chatModelConfig.TopP = conf.TopP
	}
	if conf.PresencePenalty != nil {
		chatModelConfig.PresencePenalty = conf.PresencePenalty
	}
	if conf.FrequencyPenalty != nil {
		chatModelConfig.FrequencyPenalty = conf.FrequencyPenalty
	}

	chatModel, err := openai.NewChatModel(ctx, chatModelConfig)
	if err != nil {
		log.Error("create chat model error", zap.Error(err))
		return err
	}
	streamResult, err := chatModel.Stream(ctx, messages)

	if err != nil {
		log.Error("stream chat model error", zap.Error(err))
		return err
	}

	// 处理流式响应
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			message, err := streamResult.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				log.Error("stream recv error", zap.Error(err))
				return fmt.Errorf("stream recv error: %v", err)
			}

			content := message.Content
			reasoningContent := message.ReasoningContent
			if content != "" || reasoningContent != "" {
				chunk := &global.Chunk{
					Content:          content,
					ReasoningContent: reasoningContent,
				}
				if err := callback(chunk); err != nil {
					log.Error("callback error", zap.Error(err))
					return fmt.Errorf("callback error: %v", err)
				}
			}
		}
	}
}

func NewChatClient(endpoint, apiKey string) *ChatClient {
	return &ChatClient{
		Endpoint: endpoint,
		ApiKey:   apiKey,
	}
}

var _ adaptercommon.ChatRequester = (*ChatClient)(nil)
