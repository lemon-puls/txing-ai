package file

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

// Register 注册文件相关路由
func Register(router gin.IRouter) {
	// 文件上传和下载路由
	fileRouter := router.Group("/file")
	{
		// 上传文件接口，需要登录验证
		fileRouter.POST("/upload", middleware.AuthMiddleware(), Upload)

		// 下载文件接口，不需要登录验证
		fileRouter.GET("/download", middleware.AuthMiddleware(), Download)
	}
}
