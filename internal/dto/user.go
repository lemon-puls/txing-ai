package dto

import "txing-ai/internal/utils/page"

// RegisterReq 注册请求
// Register request
type RegisterReq struct {
	Username  string `json:"username" binding:"required,min=2,max=32" example:"johndoe"`     // 用户名 Username
	Password  string `json:"password" binding:"required,min=6,max=32" example:"password123"` // 密码 Password
	Email     string `json:"email" binding:"required,email" example:"john@example.com"`      // 邮箱 Email
	Captcha   string `json:"captcha" binding:"required,len=4" example:"1234"`                // 验证码 Captcha
	CaptchaId string `json:"captchaId" binding:"required" example:"abc123"`                  // 验证码ID Captcha ID
	Phone     string `json:"phone" example:"13800138000"`                                    // 手机号 Phone number
}

// LoginReq 登录请求
// Login request
type LoginReq struct {
	Username  string `json:"username" binding:"required" example:"johndoe"`     // 用户名 Username
	Password  string `json:"password" binding:"required" example:"password123"` // 密码 Password
	Captcha   string `json:"captcha" binding:"required,len=4" example:"1234"`   // 验证码 Captcha
	CaptchaId string `json:"captchaId" binding:"required" example:"abc123"`     // 验证码ID Captcha ID
}

// UpdateProfileReq 更新个人信息请求
// Update profile request
type UpdateProfileReq struct {
	Email  string `json:"email" binding:"omitempty,email" example:"john@example.com"` // 邮箱 Email
	Phone  string `json:"phone" example:"13800138000"`                                // 手机号 Phone number
	Gender int8   `json:"gender" example:"1"`                                         // 性别 Gender: 0-未知 1-男 2-女
	Age    int8   `json:"age" example:"18"`                                           // 年龄 Age
	Avatar string `json:"avatar" example:"https://example.com/avatar.jpg"`            // 头像 Avatar URL
}

// UpdatePasswordReq 修改密码请求
// Update password request
type UpdatePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"` // 原密码 Old password
	NewPassword string `json:"new_password" binding:"required" example:"newpass123"` // 新密码 New password
}

// ResetPasswordReq 重置密码请求
// Reset password request
type ResetPasswordReq struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"` // 邮箱 Email
}

// ListUserReq 获取用户列表请求
// List users request
type ListUserReq struct {
	page.PageRequest
	Username string `form:"username" example:"john"` // 用户名 Username
	Status   *int   `form:"status" example:"0"`      // 状态 Status: 0-正常 1-禁用
}
