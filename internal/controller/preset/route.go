package preset

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	router.POST("/admin/preset", Create)
	router.PUT("/admin/preset/:id", Update)
	router.DELETE("/admin/preset/:id", Delete)
	router.GET("/admin/preset/:id", Get)
	router.GET("/admin/preset/list", List)
}
