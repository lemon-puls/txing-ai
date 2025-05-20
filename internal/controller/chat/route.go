package chat

import (
	"txing-ai/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	// WebSocket 连接
	router.GET("/ws", middleware.AuthMiddleware(), Chat)
	// 获取会话列表
	router.POST("/conversation/list", middleware.AuthMiddleware(), GetConversationList)

	// 获取会话详情
	router.GET("/conversations/:id", middleware.AuthMiddleware(), GetConversationDetail)

	// 删除会话
	router.DELETE("/conversations/:id", middleware.AuthMiddleware(), DeleteConversation)
}
