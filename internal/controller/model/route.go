package model

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	router.POST("/admin/model", Create)
	router.PUT("/admin/model/:id", Update)
	router.DELETE("/admin/model/:id", Delete)
	router.GET("/admin/model/:id", Get)
	router.GET("/admin/model/list", List)
}
