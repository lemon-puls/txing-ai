package utils

import (
	"txing-ai/internal/enum"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func GetDBFromContext(c *gin.Context) *gorm.DB {
	return c.MustGet("db").(*gorm.DB)
}

func GetUIDFromContext(c *gin.Context) int64 {
	return c.MustGet("userId").(int64)
}

// 从 context 中获取 userId， 允许为空
func GetUIDFromContextAllowEmpty(c *gin.Context) (int64, bool) {
	uid, ok := c.Get("userId")
	if ok {
		return uid.(int64), true
	}
	return -1, false
}

func GetCosClientFromContext(c *gin.Context) *COSClient {
	return c.MustGet("cos").(*COSClient)
}

func GetMessageLimiterFromContext(c *gin.Context) *MessageLimiter {
	return c.MustGet("messageLimiter").(*MessageLimiter)
}

func GetRoleFromContext(c *gin.Context) int8 {
	return c.MustGet("role").(int8)
}

func GetIsAdminFromContext(c *gin.Context) bool {
	return c.MustGet("role").(int8) == enum.UserTypeSuper
}

func GetRDBFromContext(c *gin.Context) *redis.Client {
	return c.MustGet("redis").(*redis.Client)
}

func GetAgentFromContext(c *gin.Context) string {
	return c.MustGet("agent").(string)
}
