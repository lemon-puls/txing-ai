package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
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
		log.Error("redis connect failed, panic", zap.Error(err))
		panic(err)
	}
	log.Info("redis connect success")
	return RDB
}
