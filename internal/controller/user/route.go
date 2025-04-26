package user

import (
	"github.com/gin-gonic/gin"
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
		userGroup.POST("/logout", Logout)
		userGroup.GET("/refresh", RefreshToken)
		userGroup.GET("/info", GetCurrentUser)
		// 用户登录路由
		// User login route
		userGroup.PUT("/profile", UpdateProfile)
		userGroup.PUT("/password", UpdatePassword)
		userGroup.POST("/reset-password", ResetPassword)
	}

	// 管理员路由组
	// Admin route group
	adminGroup := r.Group("/admin/user")
	{
		adminGroup.GET("/list", List)
		adminGroup.PUT("/status/:id", ToggleUserStatus)
	}
}
