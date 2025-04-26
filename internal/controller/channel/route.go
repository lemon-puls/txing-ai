package channel

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	groupRouter := router.Group("/admin/channel", middleware.AuthMiddleware())
	{
		groupRouter.POST("", Create)
		groupRouter.PUT("/:id", Update)
		groupRouter.DELETE("/:id", Delete)
		groupRouter.GET("/:id", Get)
		groupRouter.GET("/list", List)
	}

}
