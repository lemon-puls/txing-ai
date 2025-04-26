package model

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	adminRouter := router.Group("/admin", middleware.AuthMiddleware())
	{
		adminRouter.POST("/model", Create)
		adminRouter.PUT("/model/:id", Update)
		adminRouter.DELETE("/model/:id", Delete)
		adminRouter.GET("/model/:id", Get)
		adminRouter.GET("/model/list", List)
	}

}
