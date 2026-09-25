package wiki

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

// Register 路由注册：公开问答 + 管理后台（/api/admin 前缀在 AuthMiddleware 中校验超管角色）
func Register(router gin.IRouter) {

	// 公开路由：面试官匿名问答（免登录，IP 限流 + 日配额 + 总开关，对齐 /api/about 范式）
	router.POST("/wiki/ask", Ask)

	// 管理端路由
	adminRouter := router.Group("/admin", middleware.AuthMiddleware())
	{
		// 源管理（Raw 层）
		adminRouter.POST("/wiki/sources/md", CreateMDSource)
		adminRouter.POST("/wiki/sources/url", CreateURLSource)
		adminRouter.GET("/wiki/sources/list", ListSources)
		adminRouter.DELETE("/wiki/sources/:id", DeleteSource)
		adminRouter.POST("/wiki/sources/:id/ingest", TriggerIngest)

		// 草稿审核
		adminRouter.GET("/wiki/drafts/list", ListDrafts)
		adminRouter.PUT("/wiki/drafts/:id", UpdateDraft)
		adminRouter.POST("/wiki/drafts/confirm_all", ConfirmAllDrafts)
		adminRouter.POST("/wiki/drafts/:id/confirm", ConfirmDraft)
		adminRouter.DELETE("/wiki/drafts/:id", DeleteDraft)

		// 已发布页管理
		adminRouter.GET("/wiki/pages/list", ListPublished)
		adminRouter.GET("/wiki/pages/:id", GetPublished)
		adminRouter.PUT("/wiki/pages/:id", UpdatePublished)
		adminRouter.POST("/wiki/pages/:id/offline", OfflinePage)

		// 知识图谱（可视化）
		adminRouter.GET("/wiki/graph", Graph)

		// 导出（md 资产退出保险）
		adminRouter.GET("/wiki/export", ExportAll)
	}
}
