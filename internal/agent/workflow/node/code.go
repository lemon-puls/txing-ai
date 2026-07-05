package node

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// ExecuteCodeNode 执行代码节点
func ExecuteCodeNode(
	ctx context.Context,
	nodeId string,
	nodeLabel string,
	config *types.CodeConfig,
	input *schema.Message,
	callback func(chunk *global.Chunk) error,
) (*schema.Message, error) {
	execLog := &types.NodeExecutionLog{
		NodeID:    nodeId,
		NodeType:  "code",
		NodeLabel: nodeLabel,
		StartTime: time.Now().UnixMilli(),
	}

	if callback != nil {
		callback(&global.Chunk{
			NodeId:     nodeId,
			NodeType:   "code",
			NodeLabel:  nodeLabel,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 执行代码...", nodeLabel),
		})
	}

	log.Info("执行代码节点", zap.String("nodeId", nodeId), zap.String("language", config.Language))

	// 获取输入内容
	inputContent := ""
	if input != nil {
		inputContent = input.Content
	}
	execLog.Input = inputContent

	// 准备代码（替换变量）
	// 支持 {{input}}、{{output}}、{{input.xxx}}、{{output.xxx}}
	// Supports {{input}}, {{output}}, {{input.xxx}}, {{output.xxx}}
	code := config.Code
	code = strings.ReplaceAll(code, "{{input}}", inputContent)
	code = strings.ReplaceAll(code, "{{output}}", inputContent)
	code = types.ReplaceNestedVars(code, inputContent, inputContent)

	// 设置超时
	timeout := 30
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 根据语言执行代码
	var result string
	var err error

	switch config.Language {
	case "javascript", "js":
		result, err = executeJavaScript(ctx, code)
	case "python", "py":
		result, err = executePython(ctx, code)
	default:
		err = fmt.Errorf("不支持的编程语言: %s", config.Language)
	}

	if err != nil {
		log.Error("代码执行失败", zap.String("nodeId", nodeId), zap.Error(err))
		execLog.Status = "failed"
		execLog.Error = err.Error()
		execLog.EndTime = time.Now().UnixMilli()
		execLog.Duration = execLog.EndTime - execLog.StartTime
		types.SendExecutionLog(callback, execLog)
		return nil, err
	}

	execLog.Status = "completed"
	execLog.Output = result
	execLog.EndTime = time.Now().UnixMilli()
	execLog.Duration = execLog.EndTime - execLog.StartTime
	types.SendExecutionLog(callback, execLog)

	if callback != nil {
		callback(&global.Chunk{
			NodeId:     nodeId,
			NodeType:   "code",
			NodeLabel:  nodeLabel,
			NodeStatus: "completed",
			ShowMsg:    fmt.Sprintf("[%s] 代码执行完成", nodeLabel),
		})
	}

	return schema.AssistantMessage(result, nil), nil
}

// executeJavaScript 执行 JavaScript 代码
func executeJavaScript(ctx context.Context, code string) (string, error) {
	// 使用 Node.js 执行
	cmd := exec.CommandContext(ctx, "node", "-e", code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("JavaScript 执行错误: %w\n%s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// executePython 执行 Python 代码
func executePython(ctx context.Context, code string) (string, error) {
	// 使用 Python 执行
	cmd := exec.CommandContext(ctx, "python3", "-c", code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("Python 执行错误: %w\n%s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}
