package app

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"txing-ai/internal/iface"
)

func NewResourceProvider(redis *redis.Client, db *gorm.DB) iface.ResourceProvider {
	// 后续交给 di 注入
	return &resProvider{redis, db}
}

type resProvider struct {
	redis *redis.Client
	db    *gorm.DB
}

func (p *resProvider) GetRedisClient() *redis.Client {
	return p.GetRedisClient()
}

func (p *resProvider) GetDB() *gorm.DB {
	return p.db
}

var _ iface.ResourceProvider = (*resProvider)(nil)
