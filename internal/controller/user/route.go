package user

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/middleware"
)

// Routes 初始化用户相关路由
// Initialize user-related routes
// Register 函数用于定义路由
func Register(r *gin.RouterGroup) {
	// 定义用户路由组
	// User route group
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", userRegister)
		userGroup.POST("/login", Login)
		userGroup.POST("/logout", middleware.AuthMiddleware(), Logout)
		userGroup.POST("/refresh", RefreshToken)
		userGroup.GET("/info", middleware.AuthMiddleware(), GetCurrentUser)
		// 用户登录路由
		// User login route
		userGroup.PUT("/profile", middleware.AuthMiddleware(), UpdateProfile)
		userGroup.PUT("/password", middleware.AuthMiddleware(), UpdatePassword)
		userGroup.POST("/reset-password", middleware.AuthMiddleware(), ResetPassword)
	}

	// 管理员路由组
	// Admin route group
	adminGroup := r.Group("/admin/user", middleware.AuthMiddleware())
	{
		adminGroup.GET("/list", List)
		adminGroup.PUT("/status/:id", ToggleUserStatus)
	}
}
