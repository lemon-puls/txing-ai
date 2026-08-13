package chat

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"txing-ai/internal/adapter"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/service/channel"
	"txing-ai/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const defaultRespMessage = "Sorry, I don't understand your message."

const defaultErrRespMessage = "Sorry, System error, please try again later."

const exceedMessageLimit = "您今日的消息次数已达上限（%d条），请明天再试"

// friendlyChatErrorMessage 将底层错误转换为用户可理解的提示信息，
// 让用户了解失败原因并获得操作指引；无法识别的错误保持通用提示
func friendlyChatErrorMessage(err error, model string) string {
	switch {
	case errors.Is(err, channel.ErrNoChannelFound):
		return fmt.Sprintf("当前模型「%s」暂无可用的服务渠道，请更换其他模型重试，或联系管理员在「渠道管理」中为该模型配置已启用的渠道。", model)
	case errors.Is(err, channel.ErrNoAvailableChannel):
		return fmt.Sprintf("当前请求参数（如「联网搜索」）未匹配到「%s」的可用的渠道映射，请关闭「联网搜索」后重试，或联系管理员检查渠道的模型映射配置。", model)
	default:
		return defaultErrRespMessage
	}
}

// 通过回调让下层把大模型响应消息块即时通过 chan 传送给上层处理
type partialChunk struct {
	Chunk *global.Chunk
	End   bool
	Err   error
}

// 处理聊天（调用大模型发送消息，并且响应结果）
func HandleChat(ctx *gin.Context, conn *utils.Connection, conversation *domain.Conversation, db *gorm.DB) (content, reasoningContent string) {

	uid, exists := utils.GetUIDFromContextAllowEmpty(ctx)

	var role int8
	if exists {
		role = utils.GetRoleFromContext(ctx)
	}

	messageLimiter := utils.GetMessageLimiterFromContext[*utils.MessageLimiter](ctx)

	// 检查消息限制
	allowed, err := messageLimiter.CheckAndIncrement(ctx, uid, role, utils.BusinessTypeChat)
	if err != nil {
		log.Error("check message limit error", zap.Error(err))
		conn.Send(dto.WsMessageResponse{
			Content:        defaultErrRespMessage,
			End:            true,
			ConversationId: conversation.Id,
		})
		return defaultErrRespMessage, ""
	}

	// 如果不允许发送消息，返回提示信息
	if !allowed {
		limitMessage := fmt.Sprintf(exceedMessageLimit, utils.BusinessUseLimits[utils.BusinessTypeChat])
		conn.Send(dto.WsMessageResponse{
			Content:        limitMessage,
			End:            true,
			ConversationId: conversation.Id,
		})
		return limitMessage, ""
	}

	// 创建响应缓冲区
	buffer := utils.NewChatRespBuffer()

	// 设置 ctx
	ctxWithCancel, cancel := context.WithCancel(ctx)

	conn.SetCancelFunc(cancel)

	// 开启聊天
	err = execChat(ctxWithCancel, conn, conversation, buffer, db)

	if err != nil {
		log.Error("execChat failed", zap.Error(err))
		errMessage := friendlyChatErrorMessage(err, conversation.Model)
		conn.Send(dto.WsMessageResponse{
			Content:        errMessage,
			End:            true,
			ConversationId: conversation.Id,
		})
		return errMessage, ""
	}

	if buffer.IsEmpty() {
		// 没有任何响应，返回默认消息
		err := conn.Send(dto.WsMessageResponse{
			Content:        defaultRespMessage,
			End:            true,
			ConversationId: conversation.Id,
		})
		if err != nil {
			return defaultErrRespMessage, ""
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
			// panic 恢复后也必须通知主循环结束，否则主循环会一直阻塞在 <-chunkChan 上，
			// 导致客户端连接挂起、协程泄漏
			if r := recover(); r != nil {
				log.Error("execChat panic", zap.Any("err", r))
				chunkChan <- partialChunk{End: true, Err: fmt.Errorf("execChat panic: %v", r)}
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

	// 构建映射参数
	mappingParams := map[string]interface{}{
		"enableWeb": chatConfig.EnableWeb,
		// 可以添加更多参数，例如：
		// "type": "model",
		// 等等
	}

	// 获取所有支持该模型的 channel
	targetChannel, mappingModel, err := channel.ChooseChannelAndModel(db, chatConfig.Model, mappingParams)
	if err != nil {
		log.Error("choose channel failed", zap.Error(err))
		return err
	}
	chatConfig.Model = mappingModel

	err = adapter.NewChatRequest(ctx, targetChannel, chatConfig, hook)

	return err
}
