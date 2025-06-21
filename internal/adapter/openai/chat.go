package openai

import (
	"context"
	"fmt"
	"io"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type ChatClient struct {
	Endpoint string
	ApiKey   string
	client   *openai.Client
}

func (c ChatClient) StreamChat(ctx context.Context, conf *adaptercommon.ChatConfig, callback global.Hook) error {
	// 构建消息
	messages := make([]openai.ChatCompletionMessage, 0, len(conf.Message))
	for _, msg := range conf.Message {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 构建请求
	req := openai.ChatCompletionRequest{
		Model:    conf.Model,
		Messages: messages,
		Stream:   true,
	}

	// 设置可选参数
	if conf.MaxTokens != nil {
		req.MaxTokens = *conf.MaxTokens
	}
	if conf.Temperature != nil {
		req.Temperature = *conf.Temperature
	}
	if conf.TopP != nil {
		req.TopP = *conf.TopP
	}
	if conf.PresencePenalty != nil {
		req.PresencePenalty = *conf.PresencePenalty
	}
	if conf.FrequencyPenalty != nil {
		req.FrequencyPenalty = *conf.FrequencyPenalty
	}

	// 创建流式请求
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Error("create stream error", zap.Error(err))
		return fmt.Errorf("create stream error: %v", err)
	}
	defer stream.Close()

	// 处理流式响应
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				log.Error("stream recv error", zap.Error(err))
				return fmt.Errorf("stream recv error: %v", err)
			}

			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				content := choice.Delta.Content
				reasoningContent := choice.Delta.ReasoningContent
				if content != "" || reasoningContent != "" {
					chunk := &global.Chunk{
						Content:          choice.Delta.Content,
						ReasoningContent: choice.Delta.ReasoningContent,
					}
					if err := callback(chunk); err != nil {
						log.Error("callback error", zap.Error(err))
						return fmt.Errorf("callback error: %v", err)
					}
				}
			}
		}
	}
}

func NewChatClient(endpoint, apiKey string) *ChatClient {
	config := openai.DefaultConfig(apiKey)
	if endpoint != "" {
		config.BaseURL = endpoint
	}
	client := openai.NewClientWithConfig(config)
	return &ChatClient{
		Endpoint: endpoint,
		ApiKey:   apiKey,
		client:   client,
	}
}

var _ adaptercommon.ChatRequester = (*ChatClient)(nil)
