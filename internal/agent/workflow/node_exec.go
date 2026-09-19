package workflow

import (
	"context"
	// [TOOL-NODE-DISABLED] 工具节点停用后以下导入暂无使用，重新启用时恢复
	// "encoding/json"
	"fmt"

	// "github.com/cloudwego/eino/components/tool"
	// "go.uber.org/zap"

	// "txing-ai/internal/agent/workflow/parallel"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	// "txing-ai/internal/global/logging/log"
)

// ExecuteLLMNodeInParallel 在并行上下文中执行 LLM 节点（实现 parallel.NodeExecutor 接口）
// ExecuteLLMNodeInParallel executes an LLM node in parallel context with real LLM calls.
// 实际执行委托给共享执行核心 ExecuteLLM，与 LLM/Agent 节点共用同一套逻辑
// （模型解析、工具绑定、带重试的多轮工具循环），能力保持对齐。
func (e *WorkflowAgent) ExecuteLLMNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	modelConfig := node.Data.ModelConfig
	if modelConfig == nil {
		return "", fmt.Errorf("LLM 节点缺少模型配置")
	}

	nodeMaxTokens := 8192
	if modelConfig.MaxTokens > 0 {
		nodeMaxTokens = modelConfig.MaxTokens
	}
	nodeTemperature := float32(0.7)
	if modelConfig.Temperature > 0 {
		nodeTemperature = float32(modelConfig.Temperature)
	}
	llmMaxToolRounds := modelConfig.MaxToolRounds
	if llmMaxToolRounds <= 0 {
		llmMaxToolRounds = 5
	}

	cfg := &LLMExecConfig{
		NodeID:          node.Id,
		NodeLabel:       node.Data.Label,
		NodeType:        "llm",
		ModelResolver:   e.modelResolver,
		ModelName:       modelConfig.Model,
		DefaultEndpoint: e.endpoint,
		DefaultAPIKey:   e.apiKey,
		DefaultModel:    e.model,
		MaxTokens:       nodeMaxTokens,
		Temperature:     nodeTemperature,
		SystemPrompt:    modelConfig.SystemPrompt,
		AllTools:        e.tools,
		ToolNames:       modelConfig.Tools,
		MaxToolRounds:   llmMaxToolRounds,
		Retry:           modelConfig.Retry,
		// 并行分支输出由 MergeResults 统一汇总，不作为最终内容推送
		// Branch outputs are merged by MergeResults; do not push them as final content.
		EmitFinalContent: false,
	}

	// 包装 callback 注入节点信息，使并行分支内 LLM 节点的工具调用
	// 可被前端与持久化日志正确跟踪（与主流程 LLM/Agent 节点一致）
	llmCallback := callback
	if callback != nil {
		llmCallback = func(chunk *global.Chunk) error {
			chunk.NodeId = node.Id
			chunk.NodeType = "llm"
			chunk.NodeLabel = node.Data.Label
			return callback(chunk)
		}
	}

	return ExecuteLLM(ctx, cfg, input, llmCallback)
}

/* [TOOL-NODE-DISABLED] 工具节点已停用（实现 parallel.NodeExecutor 接口）。
   工具调用能力改由 LLM 节点绑定工具提供。重新启用时取消本块注释，并恢复
   parallel.NodeExecutor 接口中的 ExecuteToolNodeInParallel 方法声明。
   Original implementation of tool-node execution in parallel context.
// ExecuteToolNodeInParallel 在并行上下文中执行工具节点（实现 parallel.NodeExecutor 接口）
// ExecuteToolNodeInParallel executes a tool node in parallel context
func (e *WorkflowAgent) ExecuteToolNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	toolConfig := node.Data.ToolConfig
	if toolConfig == nil {
		return input, nil
	}

	toolName := toolConfig.ToolName
	if toolName == "" && len(toolConfig.Tools) > 0 {
		toolName = toolConfig.Tools[0]
	}

	if toolName == "" {
		log.Warn("工具节点未配置工具名称", zap.String("nodeId", node.Id))
		return input, nil
	}

	// 查找指定工具
	toolTools := e.getToolsByNames([]string{toolName})
	if len(toolTools) == 0 {
		log.Warn("工具节点找不到指定工具", zap.String("nodeId", node.Id), zap.String("toolName", toolName))
		return input, nil
	}

	// 获取工具参数，并替换变量占位符
	toolParams := parallel.ReplaceVarsInParams(toolConfig.Params, input, input)

	// 构建工具参数：如果有输入，将输入内容合并到参数中
	paramsJSON, _ := json.Marshal(toolParams)
	if input != "" {
		var paramsMap map[string]interface{}
		if err := json.Unmarshal(paramsJSON, &paramsMap); err == nil {
			if _, exists := paramsMap["toolInput"]; !exists {
				paramsMap["toolInput"] = input
				paramsJSON, _ = json.Marshal(paramsMap)
			}
		}
	}

	// 断言为 InvokableTool（直接执行）
	invokableTool, ok := toolTools[0].(tool.InvokableTool)
	if !ok {
		log.Warn("工具不支持直接执行", zap.String("nodeId", node.Id), zap.String("toolName", toolName))
		return input, nil
	}

	// 执行工具
	result, err := invokableTool.InvokableRun(ctx, string(paramsJSON))
	if err != nil {
		log.Error("工具执行失败", zap.String("nodeId", node.Id), zap.Error(err))
		return "", err
	}

	log.Info("工具执行成功", zap.String("nodeId", node.Id), zap.String("toolName", toolName))
	return result, nil
}
*/
