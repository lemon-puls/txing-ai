package chat

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {
	// 获取用户列表
	router.GET("/ws", middleware.AuthMiddleware(), Chat)
}
