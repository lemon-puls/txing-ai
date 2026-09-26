package dashboard

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

// Register 路由注册：管理后台控制台（/api/admin 前缀在 AuthMiddleware 中校验超管角色）
func Register(router gin.IRouter) {
	adminRouter := router.Group("/admin/dashboard", middleware.AuthMiddleware())
	{
		adminRouter.GET("/overview", Overview)
		adminRouter.GET("/trends", Trends)
		adminRouter.GET("/model-usage", ModelUsage)
		adminRouter.GET("/assistant-usage", AssistantUsage)
		adminRouter.GET("/channel-usage", ChannelUsage)
		adminRouter.GET("/timeseries", Timeseries)
		adminRouter.GET("/activities", Activities)
	}
}
