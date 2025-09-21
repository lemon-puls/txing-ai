package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	"txing-ai/internal/agent"
	"txing-ai/internal/utils"
)

func RegisterMiddleware(app *gin.Engine, db *gorm.DB, redis *redis.Client, cosClient *utils.COSClient, agentFactory agent.AgentFactory) {

	app.Use(LoggerWithZap(zap.L(), time.DateTime, false))

	app.Use(BuiltinMiddleWare(db, redis, cosClient, agentFactory))

	app.Use(RecoveryWithZap(zap.L(), false))

}
