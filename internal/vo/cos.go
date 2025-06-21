package vo

// 获取预签名URL响应
type GetPresignedURLVO struct {
	URL      string `json:"url"`      // 预签名URL
	Key      string `json:"key"`      // 对象键
	ExpireAt int64  `json:"expireAt"` // 过期时间戳
}
