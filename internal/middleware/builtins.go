package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"txing-ai/internal/utils"
)

func BuiltinMiddleWare(db *gorm.DB, cache *redis.Client, cosClient *utils.COSClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Set("redis", cache)
		c.Set("cos", cosClient)
		c.Next()
	}
}
