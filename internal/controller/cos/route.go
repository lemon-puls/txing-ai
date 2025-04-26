package cos

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

// 注册COS相关路由
func Register(r *gin.RouterGroup) {
	// 获取预签名URL
	r.POST("/presigned-url", middleware.AuthMiddleware(), GetPresignedURL)
}
