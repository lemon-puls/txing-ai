package model

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	router.GET("/model/:id", Get)
	router.GET("/model/list", List)

	adminRouter := router.Group("/admin", middleware.AuthMiddleware())
	{
		adminRouter.POST("/model", Create)
		adminRouter.PUT("/model/:id", Update)
		adminRouter.DELETE("/model/:id", Delete)
	}

}
