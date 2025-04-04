package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

func RegisterMiddleware(app *gin.Engine, db *gorm.DB, redis *redis.Client) {

	app.Use(BuiltinMiddleWare(db, redis))

	app.Use(RecoveryWithZap(zap.L(), false))

	app.Use(LoggerWithZap(zap.L(), time.DateTime, false))
}
