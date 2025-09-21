package agent

import (
	"github.com/gin-gonic/gin"
)

// Register 注册 agent 相关路由
func Register(r *gin.RouterGroup) {
	r.POST("/exec", Exec)
}
