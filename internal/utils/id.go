package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// RandomNumber 生成指定长度的随机数字字符串
func RandomNumber(length int) string {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成随机数
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte(rand.Intn(10) + '0') // 生成 0-9 的随机数字
	}

	return string(result)
}

// GenerateUniqueID 生成带前缀的唯一ID
// 返回格式：前缀_时间戳毫秒_4位随机数，如：1678956123456
func GenerateUniqueID() string {
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒时间戳
	return fmt.Sprintf("%d", timestamp)
}
