package dto

import "txing-ai/internal/utils/page"

// RegisterReq 用户注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required" example:"zhangsan"`    // 用户名
	Password string `json:"password" binding:"required" example:"123456"`      // 密码
	Email    string `json:"email" binding:"required,email" example:"a@b.com"`  // 邮箱
	Phone    string `json:"phone" binding:"required" example:"13800138000"`    // 手机号
	Gender   int8   `json:"gender" binding:"oneof=0 1 2" example:"1"`          // 性别(0:未知 1:男 2:女)
	Age      int8   `json:"age" binding:"required,gte=0,lte=150" example:"18"` // 年龄
	Avatar   string `json:"avatar" example:"https://example.com/avatar.png"`   // 头像
}

// LoginReq 用户登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required" example:"zhangsan"` // 用户名
	Password string `json:"password" binding:"required" example:"123456"`   // 密码
}

// UpdateProfileReq 更新个人信息请求
type UpdateProfileReq struct {
	Email  string `json:"email" binding:"required,email" example:"a@b.com"`  // 邮箱
	Phone  string `json:"phone" binding:"required" example:"13800138000"`    // 手机号
	Gender int8   `json:"gender" binding:"oneof=0 1 2" example:"1"`          // 性别(0:未知 1:男 2:女)
	Age    int8   `json:"age" binding:"required,gte=0,lte=150" example:"18"` // 年龄
	Avatar string `json:"avatar" example:"https://example.com/avatar.png"`   // 头像
}

// UpdatePasswordReq 修改密码请求
type UpdatePasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"123456"` // 原密码
	NewPassword string `json:"newPassword" binding:"required" example:"654321"` // 新密码
}

// ResetPasswordReq 重置密码请求
type ResetPasswordReq struct {
	Email string `json:"email" binding:"required,email" example:"a@b.com"` // 邮箱
}

// ListUserReq 获取用户列表请求
type ListUserReq struct {
	page.PageRequest
	Username string `form:"username" example:"zhangsan"` // 用户名
	Status   *int8  `form:"status" example:"0"`          // 状态(0:正常 1:禁用)
}
