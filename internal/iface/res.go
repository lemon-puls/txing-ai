package iface

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"txing-ai/internal/tool/mcp"
)

type ResourceProvider interface {
	GetRedisClient() *redis.Client
	GetDB() *gorm.DB
	GetMCPClientManager() *mcp.MCPClientManager
}
