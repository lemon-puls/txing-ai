package chat

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/service/chat"
	"txing-ai/internal/service/conversation"
	"txing-ai/internal/utils"
)

// 发起聊天
func Chat(c *gin.Context) {

	var webSocket *utils.WebSocket
	if webSocket = utils.NewWebSocket(c); webSocket == nil {
		log.Error("NewWebSocket failed")
		return
	}

	// TODO 从 websocket 连接读取认证信息
	var id int64 = -1
	var userId int64 = -1

	db := utils.GetDBFromContext(c)

	// 获取到 conversation 实例
	conversation := conversation.ExtractConversation(db, id, userId)
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
			// TODO 处理聊天消息
			// 1. 保存消息
			if err := conversation.HandleMessage(msg, db); err == nil {
				// 2. 调用模型，返回响应结果
				content, reasoningContent := chat.HandleChat(buf, conversation, db)
				// 3. 保存响应结果
				conversation.SaveResponse(db, content, reasoningContent)
			}

		case global.MessageTypeStop:
		}
		return nil
	})
}
