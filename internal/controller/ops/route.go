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

		// 会话持久化（聊天记录）
		adminRouter.POST("/ops/chat/sessions/list", GetOpsSessionList)
		adminRouter.GET("/ops/chat/sessions/:id", GetOpsSessionDetail)
		adminRouter.DELETE("/ops/chat/sessions/:id", DeleteOpsSession)
		adminRouter.POST("/ops/chat/sessions/:id/append", AppendOpsSession)
	}

}
