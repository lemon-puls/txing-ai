package vo

import (
	"time"
	"txing-ai/internal/domain"
)

// UserVO 用户视图对象
type UserVO struct {
	ID        int64     `json:"id"`                                              // 主键ID
	Username  string    `json:"username" example:"zhangsan"`                     // 用户名
	Email     string    `json:"email" example:"a@b.com"`                         // 邮箱
	Phone     string    `json:"phone" example:"13800138000"`                     // 手机号
	Gender    int8      `json:"gender" example:"1"`                              // 性别(0:未知 1:男 2:女)
	Status    int8      `json:"status" example:"0"`                              // 状态(0:正常 1:禁用)
	Role      int8      `json:"role" example:"0"`                                // 角色(0:普通用户 1:超管)
	Age       int8      `json:"age" example:"18"`                                // 年龄
	Avatar    string    `json:"avatar" example:"https://example.com/avatar.png"` // 头像
	CreatedAt time.Time `json:"createdAt"`                                       // 创建时间
	UpdatedAt time.Time `json:"updatedAt"`                                       // 更新时间
}

// LoginVO 登录响应视图对象
type LoginVO struct {
	UserVO
	// access token
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT token
	// refresh token
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT refresh token

}

// ToUserVO 将 User 转换为 UserVO
func ToUserVO(user domain.User) UserVO {
	return UserVO{
		ID:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Gender:    user.Gender,
		Status:    user.Status,
		Role:      user.Role,
		Age:       user.Age,
		Avatar:    user.Avatar,
		CreatedAt: user.CreateTime,
		UpdatedAt: user.UpdateTime,
	}
}

// ToLoginVO 将 User 和 token 转换为 LoginVO
func ToLoginVO(user domain.User, token string, refreshToken string) LoginVO {
	return LoginVO{
		UserVO:       ToUserVO(user),
		Token:        token,
		RefreshToken: refreshToken,
	}
}

// ToUserVOs 将 User 切片转换为 UserVO 切片
func ToUserVOs(users []domain.User) []UserVO {
	vos := make([]UserVO, len(users))
	for i, user := range users {
		vos[i] = ToUserVO(user)
	}
	return vos
}
