package channel

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	groupRouter := router.Use(middleware.AuthMiddleware())
	{
		groupRouter.POST("/admin/channel", Create)
		groupRouter.PUT("/admin/channel/:id", Update)
		groupRouter.DELETE("/admin/channel/:id", Delete)
		groupRouter.GET("/admin/channel/:id", Get)
		groupRouter.GET("/admin/channel/list", List)
	}

}
