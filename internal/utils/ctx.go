package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func GetDBFromContext(c *gin.Context) *gorm.DB {
	return c.MustGet("db").(*gorm.DB)
}

func GetRDBFromContext(c *gin.Context) *redis.Client {
	return c.MustGet("redis").(*redis.Client)
}

func GetUserFromContext(c *gin.Context) string {
	return c.MustGet("user").(string)
}

func GetAdminFromContext(c *gin.Context) bool {
	return c.MustGet("admin").(bool)
}

func GetAgentFromContext(c *gin.Context) string {
	return c.MustGet("agent").(string)
}
