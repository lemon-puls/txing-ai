package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func BuiltinMiddleWare(db *gorm.DB, cache *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Set("redis", cache)
		c.Next()
	}
}
