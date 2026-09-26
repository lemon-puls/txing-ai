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

func (c ChatClient) StreamChat(ctx context.Context, conf *adaptercommon.ChatConfig, callback global.Hook) error {

	req := model.BotChatCompletionRequest{
		BotId:    conf.Model,
		Messages: c.ConvertMessage(conf.Message),
	}

	if conf.MaxTokens != nil {
		req.MaxTokens = *conf.MaxTokens
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
	if conf.RepetitionPenalty != nil {
		req.RepetitionPenalty = *conf.RepetitionPenalty
	}
	if conf.Temperature != nil {
		req.Temperature = *conf.Temperature
	}

	stream, err := c.client.CreateBotChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("stream chat error: %v\n", err)
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
			// 如果是 context canceled，则返回 nil，表示结束
			if err == context.Canceled {
				return nil
			}
			return err
		}

		// token 用量（best-effort）：Bot API 的 usage 可能挂在 Usage 或 BotUsage.ModelUsage，
		// 两种都读，谁有值用谁；不主动设 stream_options（Bot 端点行为未验证）
		if usage := extractUsage(&recv); usage != nil {
			if err := callback(usage); err != nil {
				log.Error("callback error", zap.Error(err))
				return err
			}
		}

		if len(recv.Choices) > 0 {
			var reasoningContent string
			if recv.Choices[0].Delta.ReasoningContent != nil {
				reasoningContent = *recv.Choices[0].Delta.ReasoningContent
			}

			chunk := &global.Chunk{
				Content:          recv.Choices[0].Delta.Content,
				ReasoningContent: reasoningContent,
			}
			err := callback(chunk)
			if err != nil {
				log.Error("callback error", zap.Error(err))
				return err
			}
		}
		// TODO 处理网页引用信息
		if recv.References != nil {
			for _, ref := range recv.References {
				fmt.Printf("reference url: %s\n", ref.Url)
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

// extractUsage 提取 token 用量：优先 Usage（标准 chat 端点），
// 回退汇总 BotUsage.ModelUsage（Bot 端点）；无值返回 nil
func extractUsage(recv *model.BotChatCompletionStreamResponse) *global.Chunk {
	if recv == nil {
		return nil
	}
	if recv.Usage != nil && recv.Usage.TotalTokens > 0 {
		return &global.Chunk{Usage: &global.UsageInfo{
			PromptTokens:     int64(recv.Usage.PromptTokens),
			CompletionTokens: int64(recv.Usage.CompletionTokens),
			TotalTokens:      int64(recv.Usage.TotalTokens),
		}}
	}
	if recv.BotUsage != nil {
		var prompt, completion, total int64
		for _, mu := range recv.BotUsage.ModelUsage {
			if mu == nil {
				continue
			}
			prompt += int64(mu.PromptTokens)
			completion += int64(mu.CompletionTokens)
			total += int64(mu.TotalTokens)
		}
		if total > 0 {
			return &global.Chunk{Usage: &global.UsageInfo{
				PromptTokens:     prompt,
				CompletionTokens: completion,
				TotalTokens:      total,
			}}
		}
	}
	return nil
}

var _ adaptercommon.ChatRequester = (*ChatClient)(nil)
