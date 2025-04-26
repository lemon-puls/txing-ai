package preset

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	group := router.Group("/preset", middleware.AuthMiddleware())

	group.POST("", Create)
	group.PUT("/:id", Update)
	group.DELETE("/:id", Delete)
	group.GET("/:id", Get)
	group.GET("/list", List)

}
