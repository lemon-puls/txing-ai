package channel

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	router.POST("/admin/channel", Create)
	router.PUT("/admin/channel/:id", Update)
	router.DELETE("/admin/channel/:id", Delete)
	router.GET("/admin/channel/:id", Get)
	router.GET("/admin/channel/list", List)
}
