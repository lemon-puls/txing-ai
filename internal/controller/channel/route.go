package channel

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	router.POST("", Create)
	router.PUT("/:id", Update)
	router.DELETE("/:id", Delete)
	router.GET("/:id", Get)
	router.GET("/list", List)
}
