package aboutme

import (
	"github.com/gin-gonic/gin"

	"txing-ai/internal/middleware"
)

// Register 注册关于我页面相关路由
// /api/about   - 公开快照（无需鉴权）
// /api/admin/about/* - 管理后台 CRUD（需要超级管理员权限）
func Register(router gin.IRouter) {

	// === 公开接口 ===
	router.GET("/about", PublicSnapshot)

	// === 管理后台接口 ===
	adminGroup := router.Group("/admin/about", middleware.AuthMiddleware())
	{
		// Hero 单例
		adminGroup.GET("/hero", GetHero)
		adminGroup.PUT("/hero", UpsertHero)

		// Contact 单例
		adminGroup.GET("/contact", GetContact)
		adminGroup.PUT("/contact", UpsertContact)

		// Floating Icon
		adminGroup.GET("/floating-icon/list", ListFloatingIcons)
		adminGroup.POST("/floating-icon", CreateFloatingIcon)
		adminGroup.PUT("/floating-icon/:id", UpdateFloatingIcon)
		adminGroup.DELETE("/floating-icon/:id", DeleteFloatingIcon)

		// Reason
		adminGroup.GET("/reason/list", ListReasons)
		adminGroup.POST("/reason", CreateReason)
		adminGroup.PUT("/reason/:id", UpdateReason)
		adminGroup.DELETE("/reason/:id", DeleteReason)

		// Skill
		adminGroup.GET("/skill/list", ListSkills)
		adminGroup.POST("/skill", CreateSkill)
		adminGroup.PUT("/skill/:id", UpdateSkill)
		adminGroup.DELETE("/skill/:id", DeleteSkill)

		// Project
		adminGroup.GET("/project/list", ListProjects)
		adminGroup.GET("/project/:id", GetProject)
		adminGroup.POST("/project", CreateProject)
		adminGroup.PUT("/project/:id", UpdateProject)
		adminGroup.DELETE("/project/:id", DeleteProject)

		// Timeline
		adminGroup.GET("/timeline/list", ListTimelines)
		adminGroup.POST("/timeline", CreateTimeline)
		adminGroup.PUT("/timeline/:id", UpdateTimeline)
		adminGroup.DELETE("/timeline/:id", DeleteTimeline)
	}
}