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
		// 基础 CRUD
		workflowGroup.POST("", Create)
		workflowGroup.PUT("/:id", Update)
		workflowGroup.DELETE("/:id", Delete)
		workflowGroup.GET("/:id", Get)
		workflowGroup.GET("", List)

		// 配置查询
		workflowGroup.GET("/models", GetModels)
		workflowGroup.GET("/tools", GetTools)

		// 校验
		workflowGroup.POST("/validate", ValidateTopology)
		workflowGroup.POST("/:id/validate", func(ctx *gin.Context) {
			ValidateById(ctx, resProvider)
		})

		// 执行
		workflowGroup.POST("/:id/run", func(ctx *gin.Context) {
			Run(ctx, resProvider)
		})

		// 版本管理
		workflowGroup.POST("/:id/versions", CreateVersion)
		workflowGroup.GET("/:id/versions", ListVersions)
		workflowGroup.GET("/:id/versions/:version", GetVersion)
		workflowGroup.POST("/:id/versions/publish", PublishVersion)
		workflowGroup.POST("/:id/versions/:version/rollback", RollbackVersion)

		// 模板管理
		workflowGroup.POST("/templates", CreateTemplate)
		workflowGroup.GET("/templates", ListTemplates)
		workflowGroup.POST("/templates/clone", CloneTemplate)
	}
}
