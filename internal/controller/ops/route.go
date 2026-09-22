package ops

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	// 管理端路由（/api/admin 前缀在 AuthMiddleware 中校验超管角色）
	adminRouter := router.Group("/admin", middleware.AuthMiddleware())
	{
		adminRouter.POST("/ops/chat/stream", ChatStream)
	}

}
