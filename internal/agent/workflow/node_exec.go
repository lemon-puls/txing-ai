package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow/parallel"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// ExecuteLLMNodeInParallel 在并行上下文中执行 LLM 节点（实现 parallel.NodeExecutor 接口）
// ExecuteLLMNodeInParallel executes an LLM node in parallel context with real LLM calls
func (e *WorkflowAgent) ExecuteLLMNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	modelConfig := node.Data.ModelConfig
	if modelConfig == nil {
		return "", fmt.Errorf("LLM 节点缺少模型配置")
	}

	nodeModelName := modelConfig.Model
	nodeMaxTokens := 8192
	if modelConfig.MaxTokens > 0 {
		nodeMaxTokens = modelConfig.MaxTokens
	}
	nodeTemperature := float32(0.7)
	if modelConfig.Temperature > 0 {
		nodeTemperature = float32(modelConfig.Temperature)
	}
	systemPrompt := modelConfig.SystemPrompt
	llmToolNames := modelConfig.Tools
	llmMaxToolRounds := modelConfig.MaxToolRounds
	if llmMaxToolRounds <= 0 {
		llmMaxToolRounds = 5
	}

	// 解析节点模型信息
	var nodeEndpoint, nodeAPIKey, nodeModel string
	if e.modelResolver != nil {
		info, err := e.modelResolver.Resolve(nodeModelName)
		if err == nil {
			nodeEndpoint = info.Endpoint
			nodeAPIKey = info.APIKey
			nodeModel = info.Model
		} else if nodeModelName != "" {
			nodeEndpoint, nodeAPIKey, nodeModel = nodeModelName, "", nodeModelName
		}
	} else if nodeModelName != "" {
		nodeEndpoint, nodeAPIKey, nodeModel = nodeModelName, "", nodeModelName
	}
	if nodeEndpoint == "" {
		nodeEndpoint = e.endpoint
	}
	if nodeAPIKey == "" {
		nodeAPIKey = e.apiKey
	}
	if nodeModel == "" {
		nodeModel = e.model
	}

	// 创建节点专属模型
	nodeChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     nodeEndpoint,
		Model:       nodeModel,
		APIKey:      nodeAPIKey,
		MaxTokens:   &nodeMaxTokens,
		Temperature: &nodeTemperature,
	})
	if err != nil {
		return "", fmt.Errorf("创建节点模型失败: %w", err)
	}

	// 绑定工具（如果配置了）
	var llmToolNode *compose.ToolsNode
	if len(llmToolNames) > 0 {
		llmNodeTools := e.getToolsByNames(llmToolNames)
		if len(llmNodeTools) > 0 {
			nodeToolInfos := make([]*schema.ToolInfo, 0, len(llmNodeTools))
			for _, t := range llmNodeTools {
				info, err := t.Info(ctx)
				if err != nil {
					continue
				}
				nodeToolInfos = append(nodeToolInfos, info)
			}
			if err := nodeChatModel.BindTools(nodeToolInfos); err != nil {
				log.Warn("LLM 节点绑定工具失败", zap.String("nodeId", node.Id), zap.Error(err))
			} else {
				llmToolNode, _ = compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
					Tools:               llmNodeTools,
					ExecuteSequentially: true,
				})
			}
		}
	}

	// 构建消息
	var messages []*schema.Message
	if systemPrompt != "" {
		messages = append(messages, schema.SystemMessage(systemPrompt))
	}
	if input != "" {
		messages = append(messages, schema.UserMessage(input))
	}

	// 首次 LLM 调用
	response, genErr := nodeChatModel.Generate(ctx, messages)
	if genErr != nil {
		return "", fmt.Errorf("LLM 调用失败: %w", genErr)
	}

	// 工具调用循环
	for round := 0; round < llmMaxToolRounds; round++ {
		if len(response.ToolCalls) == 0 {
			break
		}

		// 发送工具调用开始消息
		if callback != nil {
			for _, tc := range response.ToolCalls {
				callback(&global.Chunk{
					NodeId:     node.Id,
					NodeType:   "llm",
					NodeLabel:  node.Data.Label,
					ToolCallId: tc.ID,
					ToolName:   tc.Function.Name,
					ToolParams: tc.Function.Arguments,
					ToolStatus: "running",
					ShowMsg:    fmt.Sprintf("[%s] 调用工具: %s", node.Data.Label, tc.Function.Name),
				})
			}
		}

		messages = append(messages, response)

		if llmToolNode != nil {
			toolResults, toolErr := llmToolNode.Invoke(ctx, response)
			if toolErr != nil {
				messages = append(messages, schema.ToolMessage("工具执行失败: "+toolErr.Error(), response.ToolCalls[0].ID))
			} else {
				if callback != nil {
					for _, tr := range toolResults {
						callback(&global.Chunk{
							NodeId:     node.Id,
							NodeType:   "llm",
							NodeLabel:  node.Data.Label,
							ToolCallId: tr.ToolCallID,
							ToolName:   tr.ToolName,
							ToolResult: tr.Content,
							ToolStatus: "completed",
							ShowMsg:    fmt.Sprintf("[%s] 工具 %s 执行完成", node.Data.Label, tr.ToolName),
						})
					}
				}
				messages = append(messages, toolResults...)
			}
		}

		response, genErr = nodeChatModel.Generate(ctx, messages)
		if genErr != nil {
			break
		}
	}

	result := ""
	if response != nil {
		result = response.Content
	}
	return result, nil
}

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
