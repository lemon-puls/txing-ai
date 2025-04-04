package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"txing-ai/internal/global"
)

func NewRedisClient(cfg *global.RedisConfig, ctx context.Context) *redis.Client {

	// 实例化 redis 客户端
	RDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	// 验证连接
	if _, err := RDB.Ping(ctx).Result(); err != nil {
		zap.L().Error("redis connect failed, panic", zap.Error(err))
		panic(err)
	}
	zap.L().Info("redis connect success")
	return RDB
}
