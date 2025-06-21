package preset

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	group := router.Group("/preset")

	group.POST("", middleware.AuthMiddleware(), Create)
	group.PUT("/:id", middleware.AuthMiddleware(), Update)
	group.DELETE("/:id", middleware.AuthMiddleware(), Delete)
	group.GET("/:id", middleware.AuthMiddleware(), Get)
	group.GET("/list", List)

}
