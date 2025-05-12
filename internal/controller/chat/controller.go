package chat

import (
	"strconv"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/service/chat"
	"txing-ai/internal/service/conversation"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"go.uber.org/zap"

	"github.com/jinzhu/copier"
	"github.com/samber/lo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary 建立聊天 WebSocket 连接
// @Description 建立用于实时聊天的 WebSocket 连接，支持发送聊天消息和停止生成。连接建立后，客户端可以发送聊天消息和停止指令，服务器会以流式响应的方式返回 AI 回复
// @Tags 聊天会话
// @Accept json
// @Produce json
// @Param id query int true "会话ID"
// @Param token query string false "用户令牌"
// @Success 101 {string} string "Switching Protocols 切换到 WebSocket 协议"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "未授权"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/chat/ws [get]
// @x-message-request {"type":"chat","content":"聊天内容","model":"模型标识","context":1,"enableWeb":false,"max_tokens":2048,"temperature":1.0,"top_p":0.7,"top_k":50,"presence_penalty":0.0,"frequency_penalty":0.0,"repetition_penalty":1.0}
// @x-message-stop {"type":"stop"}
// @x-message-response {"conversationId":123,"content":"AI回复内容","reasoning_content":"思考过程","end":false}
// @x-message-error {"type":"error","message":"错误信息"}
func Chat(c *gin.Context) {
	var webSocket *utils.WebSocket
	if webSocket = utils.NewWebSocket(c); webSocket == nil {
		log.Error("NewWebSocket failed")
		return
	}

	// 从 URL 参数中获取会话 ID
	id := c.Query("id")
	if id == "" {
		id = "-1"
	}

	conversationId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Error("invalid conversation id")
		panic("invalid conversation id")
	}

	// 从 URL 参数中获取用户令牌
	var userId int64 = -1
	userId, _ = utils.GetUIDFromContextAllowEmpty(c)

	db := utils.GetDBFromContext(c)

	// 获取预设 id
	presetId := c.Query("presetId")

	// 获取到 conversation 实例
	conversation := conversation.ExtractConversation(db, conversationId, userId, presetId)
	if conversation == nil {
		log.Error("ExtractConversation failed")
		return
	}

	// 实例化 Connection
	buf := utils.NewConnection(webSocket, userId != -1, 10)
	// 设置消息处理函数并且启动消息处理协程
	buf.Handle(func(msg *dto.WsMessageRequest) error {
		switch msg.Type {
		case global.MessageTypeChat:
			// 处理聊天消息
			// 1. 保存消息
			if err := conversation.HandleMessage(msg, db); err == nil {
				// 开启协程处理聊天，为了不阻塞当前协程，确保能继续接收并处理其他消息，例如停止消息
				// TODO 限制只能同时处理一个聊天请求
				go func() {
					// 捕获 panic 并记录日志
					defer func() {
						if err := recover(); err != nil {
							log.Error("chat panic", zap.Any("err", err))
						}
					}()
					// 2. 调用模型，返回响应结果
					content, reasoningContent := chat.HandleChat(c, buf, conversation, db)
					// 3. 保存响应结果
					conversation.SaveResponse(db, content, reasoningContent)
				}()
			}

		case global.MessageTypeStop:
			buf.Cancel()
		}
		return nil
	})
}

// @Summary 获取会话列表
// @Description 获取用户的会话列表
// @Tags 聊天
// @Accept json
// @Produce json
// @Param data body dto.ConversationListRequest true "请求参数"
// @Success 200 {object} utils.Response "成功"
// @Router /api/chat/conversation/list [post]
func GetConversationList(c *gin.Context) {
	// 获取当前用户ID
	userId := utils.GetUIDFromContext(c)

	// 解析分页参数
	var req dto.ConversationListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, global.CodeInvalidParams, err)
		return
	}

	// 设置默认分页大小
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	db := utils.GetDBFromContext(c)
	cosClient := utils.GetCosClientFromContext(c)

	// 使用游标分页查询
	result, err := page.GetCursorPageByMySQL[domain.Conversation](
		db,
		req.CursorPageBaseRequest,
		func(db *gorm.DB) {
			db.Where("user_id = ?", userId)
		},
		func(t *domain.Conversation) interface{} {
			return &t.CreateTime
		},
	)

	if err != nil {
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	pageVO, err := page.ConvertCursorPageVO[domain.Conversation, vo.ConversationSimpleVO](result)
	if err != nil {
		log.Error("ConvertCursorPageVO failed", zap.Error(err))
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	// 收集所有用到的模型名称
	modelNames := lo.Map(pageVO.Data, func(item vo.ConversationSimpleVO, _ int) string {
		return item.Model
	})
	modelNames = lo.Uniq(modelNames)

	// 批量查询模型信息
	var models []domain.Model
	if err := db.Where("name IN ?", modelNames).Find(&models).Error; err != nil {
		log.Error("query models failed", zap.Error(err))
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	// 构建模型名称到头像的映射
	modelAvatarMap := lo.SliceToMap(models, func(model domain.Model) (string, string) {
		return model.Name, model.Avatar
	})

	// 更新会话的模型头像
	pageVO.Data = lo.Map(pageVO.Data, func(item vo.ConversationSimpleVO, _ int) vo.ConversationSimpleVO {
		if avatar, exists := modelAvatarMap[item.Model]; exists {
			item.ModelAvatar, _ = cosClient.GenerateDownloadPresignedURL(avatar)
		}
		return item
	})

	utils.OkWithData(c, pageVO)
	return
}

// @Summary 获取会话详情
// @Description 获取指定会话的详细信息，包括基本信息和消息列表
// @Tags 聊天会话
// @Accept json
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} utils.Response{data=vo.ConversationDetailVO} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "未授权"
// @Failure 404 {object} utils.Response "会话不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/chat/conversations/{id} [get]
func GetConversationDetail(c *gin.Context) {
	// 获取当前用户ID
	userId := utils.GetUIDFromContext(c)
	//userId := int64(-1)
	// 获取会话ID
	id := c.Param("id")
	conversationId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, global.CodeInvalidParams, err)
		return
	}

	db := utils.GetDBFromContext(c)

	entity, err := conversation.QueryConversationById(db, conversationId)
	if err != nil {
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	if entity.UserID != userId {
		utils.ErrorWithCode(c, global.CodeNotPermission, err)
		return
	}

	// 字段复制
	result := &vo.ConversationDetailVO{}
	copier.Copy(result, entity)

	// 使用 lo 将  entity.FormattedMessage 转换为 vo.MessageVO 列表
	result.Messages = lo.Map(entity.FormattedMessage, func(item global.Message, _ int) vo.MessageVO {
		return vo.MessageVO{
			Role:             item.Role,
			Content:          item.Content,
			ReasoningContent: item.ReasoningContent,
			Name:             item.Name}
	})

	utils.OkWithData(c, result)
}
