package captcha

import (
	"github.com/gin-gonic/gin"
)

// Register 注册验证码相关路由
// Register captcha routes
func Register(r *gin.RouterGroup) {
	r.GET("", Generate)
}
