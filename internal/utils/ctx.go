package utils

import (
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

func GetCosClientFromContext(c *gin.Context) *COSClient {
	return c.MustGet("cos").(*COSClient)
}

func GetRoleFromContext(c *gin.Context) int8 {
	return c.MustGet("role").(int8)
}

func GetRDBFromContext(c *gin.Context) *redis.Client {
	return c.MustGet("redis").(*redis.Client)
}

func GetAgentFromContext(c *gin.Context) string {
	return c.MustGet("agent").(string)
}
