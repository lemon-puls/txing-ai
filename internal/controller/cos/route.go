package cos

import (
	"github.com/gin-gonic/gin"
)

// 注册COS相关路由
func Register(r *gin.RouterGroup) {
	// 获取预签名URL
	r.POST("/presigned-url", GetPresignedURL)
}
