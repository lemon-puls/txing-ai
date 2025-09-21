package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"txing-ai/internal/agent"
	"txing-ai/internal/utils"
)

func BuiltinMiddleWare(db *gorm.DB, cache *redis.Client, cosClient *utils.COSClient,
	agentFactory agent.AgentFactory) gin.HandlerFunc {
	// 创建消息限制工具类实例
	messageLimiter := utils.NewMessageLimiter(cache)
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Set("redis", cache)
		c.Set("cos", cosClient)
		c.Set("agentFactory", agentFactory)
		c.Set("messageLimiter", messageLimiter)
		c.Next()
	}
}
