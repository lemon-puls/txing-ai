package agent

import (
	"github.com/gin-gonic/gin"
)

// Register 注册 agent 相关路由
func Register(r *gin.RouterGroup) {
	r.POST("/exec", Exec)
	// 添加基于 SSE 的智能体流式执行路由
	r.POST("/exec/stream", ExecStream)
}
