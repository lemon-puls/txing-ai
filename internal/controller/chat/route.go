package chat

import (
	"txing-ai/internal/iface"
	"txing-ai/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter, resProvider iface.ResourceProvider) {
	// WebSocket 连接
	router.GET("/ws", middleware.AuthMiddleware(), func(ctx *gin.Context) {
		Chat(ctx, resProvider)
	})
	// 获取会话列表
	router.POST("/conversation/list", middleware.AuthMiddleware(), GetConversationList)

	// 获取会话详情
	router.GET("/conversations/:id", middleware.AuthMiddleware(), GetConversationDetail)

	// 批量删除会话
	router.POST("/conversations/deletebatch", middleware.AuthMiddleware(), BatchDeleteConversations)
}
