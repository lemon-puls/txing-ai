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

// 通过回调让下层把大模型响应消息块即时传送给会话层处理
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

	// 设置 ctx（兼容停止生成等场景）；流的生命周期不依赖连接
	_, cancel := context.WithCancel(ctx)
	defer cancel()
	conn.SetCancelFunc(cancel)

	// 启动（或替换）会话流：LLM 生成与 WS 连接解耦，
	// 客户端断开后生成继续并在服务端累积，刷新后可 resume 恢复
	session := streamManager.Start(db, conversation, buildChatConfig(conversation))

	// 绑定当前连接为消费者（先回放快照，再持续推送增量）
	if err := session.Attach(conn); err != nil {
		log.Error("attach stream consumer failed", zap.Error(err))
	}

	// 等待流结束；流完成后保存最终内容，无论客户端是否还在
	<-session.Done()

	content, reasoningContent = session.Content()
	if session.Err() != nil {
		log.Error("chat stream failed", zap.Error(session.Err()))
		if content == "" && reasoningContent == "" {
			// 无任何输出，回退为友好错误提示（写协程已向客户端发送提示）
			return friendlyChatErrorMessage(session.Err(), conversation.Model), ""
		}
	}
	return content, reasoningContent
}

// HandleResume 恢复进行中的流式输出（客户端刷新页面重连后调用）。
// 流不存在或已结束时返回 active=false 的应答，前端据此更新界面
func HandleResume(conn *utils.Connection, conversation *domain.Conversation) {
	session := streamManager.Get(conversation.Id)
	if session == nil {
		// 没有进行中的流：告知前端无需恢复
		_ = conn.Send(dto.WsMessageResponse{
			Type:           MsgTypeResume,
			Active:         false,
			ConversationId: conversation.Id,
		})
		return
	}

	// 回放快照并绑定消费者（流已结束时仅回放最终快照，active=false）
	if err := session.Attach(conn); err != nil {
		log.Error("resume attach failed", zap.Error(err))
		return
	}

	// 等待流结束（消费者写协程负责推送增量与结束标志）
	<-session.Done()
}

// CancelStream 取消会话的流式生成（用户点击停止生成）
func CancelStream(convId int64) {
	streamManager.Cancel(convId)
}

// buildChatConfig 从会话构建 LLM 请求配置
func buildChatConfig(conversation *domain.Conversation) *adaptercommon.ChatConfig {
	return &adaptercommon.ChatConfig{
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
