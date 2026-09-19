package workflow

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// calculateRetryDelay 计算重试延迟
func calculateRetryDelay(retryConfig *types.RetryConfig, attempt int) time.Duration {
	if retryConfig == nil {
		return 0
	}
	baseDelay := time.Duration(retryConfig.RetryDelay) * time.Millisecond
	if baseDelay == 0 {
		baseDelay = 1000 * time.Millisecond
	}

	switch retryConfig.BackoffType {
	case "exponential":
		return time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
	default: // "fixed"
		return baseDelay
	}
}

// getMaxRetries 获取最大重试次数
func getMaxRetries(retryConfig *types.RetryConfig) int {
	if retryConfig == nil {
		return 0
	}
	return retryConfig.MaxRetries
}

// executeWithRetry 带重试的执行函数
func executeWithRetry(retryConfig *types.RetryConfig, fn func() error) error {
	maxRetries := getMaxRetries(retryConfig)
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := calculateRetryDelay(retryConfig, attempt-1)
			log.Info("重试执行",
				zap.Int("attempt", attempt),
				zap.Int("maxRetries", maxRetries),
				zap.Duration("delay", delay))
			time.Sleep(delay)
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// 上下文已取消/超时（用户停止生成、执行超时）：重试无意义，
		// 立即返回，让取消信号尽快传播到上层，避免停止后仍在空转等待重试
		if errors.Is(lastErr, context.Canceled) || errors.Is(lastErr, context.DeadlineExceeded) {
			return lastErr
		}

		log.Warn("执行失败",
			zap.Int("attempt", attempt+1),
			zap.Error(lastErr))
	}

	return fmt.Errorf("执行失败（已重试 %d 次）: %w", maxRetries, lastErr)
}

// nodeStatusForError 节点执行出错时的展示状态：
// 用户停止生成（上下文取消）导致的中断标记为 interrupted（"已中断"），
// 而不是 failed（红色"失败"），避免主动中断被误展示为执行失败
func nodeStatusForError(err error) string {
	if errors.Is(err, context.Canceled) {
		return "interrupted"
	}
	return "failed"
}

// nodeStatusCallback 创建节点状态回调，包装原始 callback 发送 running/completed/failed 状态
func nodeStatusCallback(callback func(chunk *global.Chunk) error, nodeId, nodeType, nodeLabel string) func(status string) {
	return func(status string) {
		if callback == nil {
			return
		}
		callback(&global.Chunk{
			NodeId:     nodeId,
			NodeType:   nodeType,
			NodeLabel:  nodeLabel,
			NodeStatus: status,
		})
	}
}
