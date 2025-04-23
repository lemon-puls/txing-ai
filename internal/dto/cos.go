package dto

// 获取预签名URL请求
type GetPresignedURLReq struct {
	Key  string `json:"key" binding:"required"`  // 对象键
	Type string `json:"type" binding:"required"` // 类型：upload/download
}
