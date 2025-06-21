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

const (
	// 每日消息限制
	DailyMessageLimit = 5
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
func (m *MessageLimiter) CheckAndIncrement(ctx *gin.Context, uid int64, role int8) (bool, error) {
	// 如果是管理员，直接返回 true
	if role == enum.UserTypeSuper {
		return true, nil
	}

	// 生成 Redis key
	var key string
	if uid > 0 {
		// 已登录用户使用 uid 作为限制依据
		key = fmt.Sprintf("%suser:%d:%s", MessageLimitKeyPrefix, uid, time.Now().Format("2006-01-02"))
	} else {
		// 未登录用户使用 IP 作为限制依据
		ip := ctx.ClientIP()
		key = fmt.Sprintf("%sip:%s:%s", MessageLimitKeyPrefix, ip, time.Now().Format("2006-01-02"))
	}

	// 获取当前计数
	count, err := m.rdb.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		log.Error("get message count error", zap.Error(err))
		return false, err
	}

	// 如果计数不存在，说明是今天第一次发送消息
	if err == redis.Nil {
		// 设置计数为 1，并设置过期时间
		err = m.rdb.Set(ctx, key, 1, MessageLimitKeyExpiration).Err()
		if err != nil {
			log.Error("set message count error", zap.Error(err))
			return false, err
		}
		return true, nil
	}

	// 如果已经达到限制，返回 false
	if count >= DailyMessageLimit {
		return false, nil
	}

	// 增加计数
	err = m.rdb.Incr(ctx, key).Err()
	if err != nil {
		log.Error("increment message count error", zap.Error(err))
		return false, err
	}

	return true, nil
}

// GetRemainingCount 获取剩余消息次数
func (m *MessageLimiter) GetRemainingCount(ctx *gin.Context, uid int64, role int8) (int, error) {
	// 如果是管理员，返回 -1 表示无限制
	if role == enum.UserTypeSuper {
		return -1, nil
	}

	// 生成 Redis key
	var key string
	if uid > 0 {
		// 已登录用户使用 uid 作为限制依据
		key = fmt.Sprintf("%suser:%d:%s", MessageLimitKeyPrefix, uid, time.Now().Format("2006-01-02"))
	} else {
		// 未登录用户使用 IP 作为限制依据
		ip := ctx.ClientIP()
		key = fmt.Sprintf("%sip:%s:%s", MessageLimitKeyPrefix, ip, time.Now().Format("2006-01-02"))
	}

	// 获取当前计数
	count, err := m.rdb.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		log.Error("get message count error", zap.Error(err))
		return 0, err
	}

	// 如果计数不存在，返回完整限制次数
	if err == redis.Nil {
		return DailyMessageLimit, nil
	}

	// 返回剩余次数
	return DailyMessageLimit - count, nil
}
