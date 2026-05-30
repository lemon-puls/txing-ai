package workflow

import (
	"txing-ai/internal/iface"
	"txing-ai/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, resProvider iface.ResourceProvider) {
	// 工作流路由组
	workflowGroup := r.Group("/workflow").Use(middleware.AuthMiddleware())
	{
		workflowGroup.POST("", Create)
		workflowGroup.PUT("/:id", Update)
		workflowGroup.DELETE("/:id", Delete)
		workflowGroup.GET("/:id", Get)
		workflowGroup.GET("", List)
		workflowGroup.GET("/models", GetModels)
		workflowGroup.GET("/tools", GetTools)
		workflowGroup.POST("/validate", ValidateTopology)
		workflowGroup.POST("/:id/validate", func(ctx *gin.Context) {
			ValidateById(ctx, resProvider)
		})
		workflowGroup.POST("/:id/run", func(ctx *gin.Context) {
			Run(ctx, resProvider)
		})
	}
}
