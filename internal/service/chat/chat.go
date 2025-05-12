package chat

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math/rand"
	"txing-ai/internal/adapter"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/service/channel"
	"txing-ai/internal/utils"
)

const defaultRespMessage = "Sorry, I don't understand your message."

// 通过回调让下层把大模型响应消息块即时通过 chan 传送给上层处理
type partialChunk struct {
	Chunk *global.Chunk
	End   bool
	Err   error
}

// 处理聊天（调用大模型发送消息，并且响应结果）
func HandleChat(ctx context.Context, conn *utils.Connection, conversation *domain.Conversation, db *gorm.DB) (content, reasoningContent string) {

	// 创建响应缓冲区
	buffer := utils.NewChatRespBuffer()

	// 设置 ctx
	ctxWithCancel, cancel := context.WithCancel(ctx)

	conn.SetCancelFunc(cancel)

	// 开启聊天
	err := execChat(ctxWithCancel, conn, conversation, buffer, db)

	if err != nil {
		log.Error("execChat failed", zap.Error(err))
		return err.Error(), ""
	}

	if buffer.IsEmpty() {
		// 没有任何响应，返回默认消息
		err := conn.Send(dto.WsMessageResponse{
			Content:        defaultRespMessage,
			End:            true,
			ConversationId: conversation.Id,
		})
		if err != nil {
			return
		}
		return defaultRespMessage, ""
	}

	// 发送消息结束标志
	conn.Send(dto.WsMessageResponse{
		End:            true,
		ConversationId: conversation.Id,
	})

	return buffer.GetOrDefault(defaultRespMessage)
}

// 开启聊天
func execChat(ctx context.Context, conn *utils.Connection, conversation *domain.Conversation, buffer *utils.ChatRespBuffer, db *gorm.DB) error {
	// 创建 channel 用于接收大模型的响应
	chunkChan := make(chan partialChunk, 20)
	defer close(chunkChan)

	// 启动协程， 调用大模型发送消息，并将响应写入 chan
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("execChat panic", zap.Any("err", err))
			}
		}()

		err := NewChatRequest(
			ctx,
			db,
			&adaptercommon.ChatConfig{
				Model:             conversation.Model,
				Message:           conversation.GetChatMessages(),
				EnableWeb:         conversation.EnableWeb,
				MaxTokens:         conversation.MaxTokens,
				Temperature:       conversation.Temperature,
				TopP:              conversation.TopP,
				TopK:              conversation.TopK,
				FrequencyPenalty:  conversation.FrequencyPenalty,
				PresencePenalty:   conversation.PresencePenalty,
				RepetitionPenalty: conversation.RepetitionPenalty,
			},
			func(chunk *global.Chunk) error {
				chunkChan <- partialChunk{Chunk: chunk, End: false, Err: nil}
				return nil
			},
		)

		// 发送结束标志
		chunkChan <- partialChunk{End: true, Err: err}
	}()

	// 循环从 chan 接收大模型的响应，并将响应添加到 buffer 中以及发送给客户端
	for {
		select {
		case data := <-chunkChan:
			if data.Err != nil {
				log.Error("execChat failed", zap.Error(data.Err))
				return data.Err
			}

			if data.End { // 结束标志
				return nil
			}

			content, reasoningContent := buffer.WriteChunk(data.Chunk)

			err := conn.Send(dto.WsMessageResponse{
				Content:          content,
				ReasoningContent: reasoningContent,
				End:              false,
				ConversationId:   conversation.Id,
			})
			if err != nil {
				log.Error("failed to send message to client", zap.Error(err))
				return nil
			}
		}
	}
}

func NewChatRequest(ctx context.Context, db *gorm.DB, chatConfig *adaptercommon.ChatConfig, hook global.Hook) error {

	// 获取所有支持该模型的 channel
	sequence := channel.GetAllChannelsByModel(db, chatConfig.Model)
	// 判断是否有支持该模型的 channel
	if len(*sequence) == 0 {
		log.Error("no channel found for model ", zap.String("model", chatConfig.Model))
		return errors.New("no channel found for model " + chatConfig.Model)
	}
	// TODO 后续优化为根据优先级和权重选择 以及实现重试机制
	// 从中随机选择一个 channel
	targetChannel := (*sequence)[rand.Intn(len(*sequence))]

	err := adapter.NewChatRequest(ctx, &targetChannel, chatConfig, hook)

	return err
}
