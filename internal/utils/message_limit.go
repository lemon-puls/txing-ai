package utils

import (
	"fmt"
	"time"
	"txing-ai/internal/enum"
	"txing-ai/internal/global/logging/log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 业务类型常量
const (
	// 其他业务类型可以在这里添加，例如：
	BusinessTypeChat   = "chat"   // 聊天消息
	BusinessTypeResume = "resume" // 简历优化功能
)

// 业务限制配置
var BusinessUseLimits = map[string]int{
	BusinessTypeChat:   5, // 聊天消息每日限制
	BusinessTypeResume: 2, // 智能代理每日限制
}

const (
	// Redis key 前缀
	MessageLimitKeyPrefix = "message_limit:"
	// Redis key 过期时间（24小时）
	MessageLimitKeyExpiration = 24 * time.Hour
)

// MessageLimiter 消息限制工具类
type MessageLimiter struct {
	rdb *redis.Client
}

// NewMessageLimiter 创建消息限制工具类实例
func NewMessageLimiter(rdb *redis.Client) *MessageLimiter {
	return &MessageLimiter{
		rdb: rdb,
	}
}

// CheckAndIncrement 检查并增加消息计数
// 如果用户是管理员，直接返回 true
// 如果用户是普通用户，检查是否超过每日限制
// businessType 参数指定业务类型，不同业务类型有不同的限制次数
func (m *MessageLimiter) CheckAndIncrement(ctx *gin.Context, uid int64, role int8, businessType string) (bool, error) {
	// 如果是管理员，直接返回 true
	if role == enum.UserTypeSuper {
		return true, nil
	}

	// 如果业务类型为空，使用默认业务类型
	if businessType == "" {
		businessType = BusinessTypeResume
	}

	// 获取该业务类型的限制次数
	limit, exists := BusinessUseLimits[businessType]
	if !exists {
		// 如果业务类型不存在，使用默认业务类型的限制
		limit = BusinessUseLimits[BusinessTypeResume]
	}

	// 生成 Redis key，加入业务类型
	var key string
	if uid > 0 {
		// 已登录用户使用 uid 作为限制依据
		key = fmt.Sprintf("%suser:%d:%s:%s", MessageLimitKeyPrefix, uid, businessType, time.Now().Format("2006-01-02"))
	} else {
		// 未登录用户使用 IP 作为限制依据
		ip := ctx.ClientIP()
		key = fmt.Sprintf("%sip:%s:%s:%s", MessageLimitKeyPrefix, ip, businessType, time.Now().Format("2006-01-02"))
	}

	// 获取当前计数
	count, err := m.rdb.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		log.Error("get message count error", zap.Error(err), zap.String("business_type", businessType))
		return false, err
	}

	// 如果计数不存在，说明是今天第一次发送消息
	if err == redis.Nil {
		// 设置计数为 1，并设置过期时间
		err = m.rdb.Set(ctx, key, 1, MessageLimitKeyExpiration).Err()
		if err != nil {
			log.Error("set message count error", zap.Error(err), zap.String("business_type", businessType))
			return false, err
		}
		return true, nil
	}

	// 如果已经达到限制，返回 false
	if count >= limit {
		return false, nil
	}

	// 增加计数
	err = m.rdb.Incr(ctx, key).Err()
	if err != nil {
		log.Error("increment message count error", zap.Error(err), zap.String("business_type", businessType))
		return false, err
	}

	return true, nil
}

// GetRemainingCount 获取剩余消息次数
// businessType 参数指定业务类型，不同业务类型有不同的限制次数
func (m *MessageLimiter) GetRemainingCount(ctx *gin.Context, uid int64, role int8, businessType string) (int, error) {
	// 如果是管理员，返回 -1 表示无限制
	if role == enum.UserTypeSuper {
		return -1, nil
	}

	// 如果业务类型为空，使用默认业务类型
	if businessType == "" {
		businessType = BusinessTypeResume
	}

	// 获取该业务类型的限制次数
	limit, exists := BusinessUseLimits[businessType]
	if !exists {
		// 如果业务类型不存在，使用默认业务类型的限制
		limit = BusinessUseLimits[BusinessTypeResume]
	}

	// 生成 Redis key，加入业务类型
	var key string
	if uid > 0 {
		// 已登录用户使用 uid 作为限制依据
		key = fmt.Sprintf("%suser:%d:%s:%s", MessageLimitKeyPrefix, uid, businessType, time.Now().Format("2006-01-02"))
	} else {
		// 未登录用户使用 IP 作为限制依据
		ip := ctx.ClientIP()
		key = fmt.Sprintf("%sip:%s:%s:%s", MessageLimitKeyPrefix, ip, businessType, time.Now().Format("2006-01-02"))
	}

	// 获取当前计数
	count, err := m.rdb.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		log.Error("get message count error", zap.Error(err), zap.String("business_type", businessType))
		return 0, err
	}

	// 如果计数不存在，返回完整限制次数
	if err == redis.Nil {
		return limit, nil
	}

	// 返回剩余次数
	return limit - count, nil
}
