package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	agent "txing-ai/internal/agent/agent"
	"txing-ai/internal/utils"
)

func RegisterMiddleware(app *gin.Engine, db *gorm.DB, redis *redis.Client, cosClient *utils.COSClient, agentFactory agent.AgentFactory) {

	// 指标中间件放首位：panic 被 Recovery abort 后仍能按最终状态码计数，时延含全部中间件开销
	app.Use(Metrics())

	app.Use(LoggerWithZap(zap.L(), time.DateTime, false))

	app.Use(BuiltinMiddleWare(db, redis, cosClient, agentFactory))

	app.Use(RecoveryWithZap(zap.L(), false))

}
