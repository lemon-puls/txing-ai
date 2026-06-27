package tool

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

// SafeInvokableTool 包装 InvokableTool，将执行错误转为正常返回的错误信息
// 避免工具报错导致整个工作流中断，让 LLM 看到错误后自行决策后续步骤
type SafeInvokableTool struct {
	inner tool.InvokableTool
	info  *schema.ToolInfo
}

// SafeStreamableTool 包装 StreamableTool，将执行错误转为正常返回的错误信息
type SafeStreamableTool struct {
	inner tool.StreamableTool
	info  *schema.ToolInfo
}

// WrapSafe 将 BaseTool 包装为容错版本
// 如果工具实现了 InvokableTool 或 StreamableTool，会被包装为捕获错误的版本
func WrapSafe(bt tool.BaseTool) tool.BaseTool {
	info, err := bt.Info(context.Background())
	if err != nil {
		log.Warn("获取工具信息失败，跳过安全包装", zap.Error(err))
		return bt
	}

	if st, ok := bt.(tool.StreamableTool); ok {
		return &SafeStreamableTool{inner: st, info: info}
	}

	if it, ok := bt.(tool.InvokableTool); ok {
		return &SafeInvokableTool{inner: it, info: info}
	}

	// 既不是 InvokableTool 也不是 StreamableTool，原样返回
	return bt
}

// WrapSafeTools 批量包装工具列表
func WrapSafeTools(tools []tool.BaseTool) []tool.BaseTool {
	result := make([]tool.BaseTool, len(tools))
	for i, bt := range tools {
		result[i] = WrapSafe(bt)
	}
	return result
}

// --- SafeInvokableTool ---

func (s *SafeInvokableTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return s.info, nil
}

func (s *SafeInvokableTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	result, err := s.inner.InvokableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		log.Warn("工具执行出错，返回错误信息给 LLM 继续处理",
			zap.String("tool", s.info.Name),
			zap.Error(err))
		return fmt.Sprintf("工具执行失败: %s。请根据已有信息继续完成任务，或尝试其他方式。", err.Error()), nil
	}
	return result, nil
}

// --- SafeStreamableTool ---

func (s *SafeStreamableTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return s.info, nil
}

func (s *SafeStreamableTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error) {
	result, err := s.inner.StreamableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		log.Warn("流式工具执行出错，返回错误信息给 LLM 继续处理",
			zap.String("tool", s.info.Name),
			zap.Error(err))
		errMsg := fmt.Sprintf("工具执行失败: %s。请根据已有信息继续完成任务，或尝试其他方式。", err.Error())
		return schema.StreamReaderFromArray([]string{errMsg}), nil
	}
	return result, nil
}
