package dto

import "txing-ai/internal/utils/page"

// GetUserListRequest 获取用户列表请求
type UserListRequest struct {
	page.PageRequest
	UserName string `json:"userName" form:"userName"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}
