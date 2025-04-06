package volcengine

import (
	"context"
	"fmt"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"go.uber.org/zap"
	"io"
	"time"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

type ChatClient struct {
	Endpoint string
	ApiKey   string
	client   *arkruntime.Client
}

// 转换消息格式为适合 Ark Runtime 的格式
func (c *ChatClient) ConvertMessage(message []global.Message) []*model.ChatCompletionMessage {
	messages := make([]*model.ChatCompletionMessage, 0)
	for _, msg := range message {
		// 构建消息
		target := &model.ChatCompletionMessage{
			Content: &model.ChatCompletionMessageContent{
				StringValue: volcengine.String(msg.Content),
			},
		}
		// 设置消息类型
		switch msg.Role {
		case global.User:
			target.Role = model.ChatMessageRoleUser
		case global.System:
			target.Role = model.ChatMessageRoleSystem
		case global.Assistant:
			target.Role = model.ChatMessageRoleAssistant
		}
		// 添加到消息列表
		messages = append(messages, target)
	}
	return messages
}

func (c ChatClient) StreamChat(conf *adaptercommon.ChatConfig, callback global.Hook) error {
	ctx := context.Background()

	req := model.CreateChatCompletionRequest{
		Model:             conf.Model,
		Messages:          c.ConvertMessage(conf.Message),
		MaxTokens:         conf.MaxTokens,
		Temperature:       conf.Temperature,
		TopP:              conf.TopP,
		PresencePenalty:   conf.PresencePenalty,
		FrequencyPenalty:  conf.FrequencyPenalty,
		RepetitionPenalty: conf.RepetitionPenalty,
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Error("stream chat error", zap.Error(err))
		return err
	}
	defer stream.Close()

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Error("stream chat error", zap.Error(err))
			return err
		}

		if len(recv.Choices) > 0 {
			fmt.Print(recv.Choices[0].Delta.Content)
			chunk := &global.Chunk{Content: recv.Choices[0].Delta.Content}
			err := callback(chunk)
			if err != nil {
				log.Error("callback error", zap.Error(err))
				return err
			}
		}
	}
}

func NewChatClient(endpoint, apiKey string) *ChatClient {
	client := arkruntime.NewClientWithApiKey(
		apiKey,
		arkruntime.WithBaseUrl(endpoint),
		arkruntime.WithTimeout(2*time.Minute),
		arkruntime.WithRetryTimes(2),
	)
	return &ChatClient{
		Endpoint: endpoint,
		ApiKey:   apiKey,
		client:   client,
	}
}

var _ adaptercommon.ChatRequester = (*ChatClient)(nil)
