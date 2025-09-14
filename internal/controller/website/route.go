package website

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	// 用户端路由
	router.GET("/websites/list", UserList)

	// 管理端路由
	adminRouter := router.Group("/admin", middleware.AuthMiddleware())
	{
		adminRouter.POST("/websites", Create)
		adminRouter.PUT("/websites/:id", Update)
		adminRouter.DELETE("/websites/:id", Delete)
		adminRouter.GET("/websites/:id", Get)
		adminRouter.GET("/websites/list", List)
		adminRouter.POST("/websites/favicon", GetFavicon)
	}

}
