package user

import (
	"github.com/gin-gonic/gin"
)

func Register(router gin.IRouter) {
	// 获取用户列表
	router.GET("/list", GetUserList)
}
