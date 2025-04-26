package preset

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

func Register(router gin.IRouter) {

	router.Use(middleware.AuthMiddleware())

	router.POST("/preset", Create)
	router.PUT("/preset/:id", Update)
	router.DELETE("/preset/:id", Delete)
	router.GET("/preset/:id", Get)
	router.GET("/preset/list", List)

}
