package app

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"txing-ai/internal/iface"
	"txing-ai/internal/tool/mcp"
)

func NewResourceProvider(redis *redis.Client, db *gorm.DB, mcpClientManager *mcp.MCPClientManager) iface.ResourceProvider {
	// 后续交给 di 注入
	return &resProvider{redis, db, mcpClientManager}
}

type resProvider struct {
	redis            *redis.Client
	db               *gorm.DB
	mcpClientManager *mcp.MCPClientManager
}

func (p *resProvider) GetRedisClient() *redis.Client {
	return p.GetRedisClient()
}

func (p *resProvider) GetDB() *gorm.DB {
	return p.db
}

func (p *resProvider) GetMCPClientManager() *mcp.MCPClientManager {
	return p.mcpClientManager
}

var _ iface.ResourceProvider = (*resProvider)(nil)
