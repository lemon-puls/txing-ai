package preset

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	groupRouter := router.Use(middleware.AuthMiddleware())
	{
		groupRouter.POST("/admin/preset", Create)
		groupRouter.PUT("/admin/preset/:id", Update)
		groupRouter.DELETE("/admin/preset/:id", Delete)
		groupRouter.GET("/admin/preset/:id", Get)
		groupRouter.GET("/admin/preset/list", List)
	}

}
