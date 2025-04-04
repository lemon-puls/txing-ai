package iface

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ResourceProvider interface {
	GetRedisClient() *redis.Client
	GetDB() *gorm.DB
}
