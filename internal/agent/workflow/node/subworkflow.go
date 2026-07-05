package node

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// SubWorkflowExecutor 子工作流执行器接口
type SubWorkflowExecutor interface {
	ExecuteSubWorkflow(ctx context.Context, workflowID int64, input string, callback func(chunk *global.Chunk) error) (string, error)
}

// ExecuteSubWorkflowNode 执行子工作流节点
func ExecuteSubWorkflowNode(
	ctx context.Context,
	nodeId string,
	nodeLabel string,
	config *types.SubWorkflowConfig,
	input *schema.Message,
	executor SubWorkflowExecutor,
	callback func(chunk *global.Chunk) error,
) (*schema.Message, error) {
	execLog := &types.NodeExecutionLog{
		NodeID:    nodeId,
		NodeType:  "subworkflow",
		NodeLabel: nodeLabel,
		StartTime: time.Now().UnixMilli(),
	}

	if callback != nil {
		callback(&global.Chunk{
			NodeId:     nodeId,
			NodeType:   "subworkflow",
			NodeLabel:  nodeLabel,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 执行子工作流...", nodeLabel),
		})
	}

	log.Info("执行子工作流节点", zap.String("nodeId", nodeId), zap.Int64("workflowId", config.WorkflowID))

	// 获取输入内容
	inputContent := ""
	if input != nil {
		inputContent = input.Content
	}
	execLog.Input = inputContent

	// 准备子工作流输入
	subInput := config.Input
	if subInput == "" {
		subInput = inputContent
	} else {
		subInput = strings.ReplaceAll(subInput, "{{input}}", inputContent)
		subInput = strings.ReplaceAll(subInput, "{{output}}", inputContent)
	}

	// 设置超时
	timeout := 60
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行子工作流
	result, err := executor.ExecuteSubWorkflow(ctx, config.WorkflowID, subInput, callback)
	if err != nil {
		log.Error("子工作流执行失败", zap.String("nodeId", nodeId), zap.Error(err))
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
			NodeType:   "subworkflow",
			NodeLabel:  nodeLabel,
			NodeStatus: "completed",
			ShowMsg:    fmt.Sprintf("[%s] 子工作流执行完成", nodeLabel),
		})
	}

	return schema.AssistantMessage(result, nil), nil
}
