package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
	mytool "txing-ai/internal/tool"
)

// WorkflowAgent 工作流智能体
type WorkflowAgent struct {
	*BaseAgent
	tools         []tool.BaseTool
	topology      string
	modelResolver ModelResolver
	resProvider   iface.ResourceProvider
	endpoint      string // LLM 调用端点 / LLM endpoint
	apiKey        string // LLM API 密钥 / LLM API key
	model         string // 默认模型名 / Default model name
}

// WorkflowAgentState 工作流状态
type WorkflowAgentState struct {
	Messages []*schema.Message
}

// NewWorkflowAgent 创建一个新的工作流智能体
func NewWorkflowAgent(res iface.ResourceProvider, topology string, modelResolver ModelResolver) *WorkflowAgent {
	baseAgent := NewBaseAgent("WorkflowAgent", "A dynamic workflow agent based on JSON topology")
	baseAgent.SetSystemPrompt("You are a helpful AI assistant executing a workflow.")

	return &WorkflowAgent{
		BaseAgent:     baseAgent,
		tools:         mytool.ProvideTools(res),
		topology:      topology,
		modelResolver: modelResolver,
		resProvider:   res,
	}
}

// getToolsByNames 根据名称列表获取工具
func (a *WorkflowAgent) getToolsByNames(names []string) []tool.BaseTool {
	if len(names) == 0 {
		return a.tools
	}

	toolMap := make(map[string]tool.BaseTool)
	for _, t := range a.tools {
		info, _ := t.Info(context.Background())
		if info != nil {
			toolMap[info.Name] = t
		}
	}

	var result []tool.BaseTool
	for _, name := range names {
		if t, ok := toolMap[name]; ok {
			result = append(result, t)
		}
	}
	return result
}

// calculateRetryDelay 计算重试延迟
func calculateRetryDelay(retryConfig *RetryConfig, attempt int) time.Duration {
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
func getMaxRetries(retryConfig *RetryConfig) int {
	if retryConfig == nil {
		return 0
	}
	return retryConfig.MaxRetries
}

// executeWithRetry 带重试的执行函数
func executeWithRetry(retryConfig *RetryConfig, fn func() error) error {
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

		log.Warn("执行失败",
			zap.Int("attempt", attempt+1),
			zap.Error(lastErr))
	}

	return fmt.Errorf("执行失败（已重试 %d 次）: %w", maxRetries, lastErr)
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

// resolveModelInfo 解析模型信息，支持节点级别覆盖
func (a *WorkflowAgent) resolveModelInfo(nodeModel string, defaultEndpoint, defaultAPIKey, defaultModel string) (endpoint, apiKey, model string) {
	endpoint = defaultEndpoint
	apiKey = defaultAPIKey
	model = defaultModel

	// 如果节点指定了模型，尝试使用 ModelResolver 解析
	if nodeModel != "" && a.modelResolver != nil {
		info, err := a.modelResolver.Resolve(nodeModel)
		if err != nil {
			log.Warn("解析节点模型失败，使用默认模型",
				zap.String("nodeModel", nodeModel),
				zap.Error(err))
		} else {
			endpoint = info.Endpoint
			apiKey = info.APIKey
			model = info.Model
		}
	} else if nodeModel != "" {
		// 没有 ModelResolver 但节点指定了模型，仅覆盖模型名称
		model = nodeModel
	}

	return endpoint, apiKey, model
}

// BuildGraph 构建执行图（简化版本，使用 DAG 模式）
func (a *WorkflowAgent) BuildGraph(ctx context.Context, endpoint, apiKey, model string, callback func(chunk *global.Chunk) error) (*compose.Graph[[]*schema.Message, *schema.Message], error) {
	var topo Topology
	if err := json.Unmarshal([]byte(a.topology), &topo); err != nil {
		return nil, fmt.Errorf("解析拓扑图失败: %w", err)
	}

	// 使用拓扑配置中的默认模型（如果有）
	if topo.Config != nil && topo.Config.DefaultModel != "" {
		if a.modelResolver != nil {
			info, err := a.modelResolver.Resolve(topo.Config.DefaultModel)
			if err != nil {
				log.Warn("解析工作流默认模型失败，使用传入的默认模型",
					zap.String("configModel", topo.Config.DefaultModel),
					zap.Error(err))
			} else {
				endpoint = info.Endpoint
				apiKey = info.APIKey
				model = info.Model
			}
		} else {
			model = topo.Config.DefaultModel
		}
	}

	// 最大执行步数
	maxRunSteps := 30
	if topo.Config != nil && topo.Config.MaxRunSteps > 0 {
		maxRunSteps = topo.Config.MaxRunSteps
	}
	a.SetMaxRunSteps(maxRunSteps)

	// 默认模型配置
	defaultMaxTokens := 8192
	defaultTemperature := float32(0.7)

	// 创建默认的聊天模型
	defaultChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     endpoint,
		Model:       model,
		APIKey:      apiKey,
		MaxTokens:   &defaultMaxTokens,
		Temperature: &defaultTemperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建模型失败: %w", err)
	}

	// 绑定所有工具到默认模型
	toolInfos := make([]*schema.ToolInfo, 0, len(a.tools))
	for _, t := range a.tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolInfos = append(toolInfos, info)
	}
	if err := defaultChatModel.BindTools(toolInfos); err != nil {
		log.Warn("绑定工具失败", zap.Error(err))
	}

	// 使用 DAG 构建工作流
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *WorkflowAgentState {
			return &WorkflowAgentState{Messages: make([]*schema.Message, 0)}
		}))

	var startNodeId, endNodeId string

	// 添加节点
	// 收集所有并行组覆盖的分支节点，避免重复构建
	// Collect all branch nodes covered by parallel groups to avoid duplicate building
	parallelExecutorForScan := NewParallelExecutor(a, 10, "", "", "")
	parallelBranchNodes := make(map[string]struct{})
	for _, node := range topo.Nodes {
		if node.Data.NodeType == "parallel" {
			groups, _ := parallelExecutorForScan.IdentifyParallelGroups(&topo)
			for _, group := range groups {
				if group.ParallelID == node.Id {
					for _, branch := range group.Branches {
						for _, n := range branch {
							parallelBranchNodes[n.Id] = struct{}{}
						}
					}
					break
				}
			}
		}
	}

	for _, node := range topo.Nodes {
		nodeId := node.Id

		// 跳过已被并行组处理的分支节点，添加占位节点以保证边的引用存在
		// Skip branch nodes handled by parallel groups, but add placeholder nodes so edge references resolve
		if _, skip := parallelBranchNodes[nodeId]; skip {
			// 占位节点：直接透传输入（实际执行由 parallel 节点完成）
			placeholderId := nodeId
			graph.AddLambdaNode(placeholderId, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
				if len(input) > 0 {
					return input[len(input)-1], nil
				}
				return schema.UserMessage(""), nil
			}))
			continue
		}

		switch node.Data.NodeType {
		case "start":
			startNodeId = nodeId
			statusCb := nodeStatusCallback(callback, nodeId, "start", node.Data.Label)
			// 开始节点：将输入消息转换为消息列表
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "start",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCb("running")
				log.Info("Executing start node", zap.String("nodeId", nodeId))
				var result *schema.Message
				if len(input) > 0 {
					result = input[len(input)-1]
					execLog.Input = input[len(input)-1].Content
				} else {
					result = schema.UserMessage("")
				}
				execLog.Status = "completed"
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCb("completed")
				return result, nil
			}))

		case "end":
			endNodeId = nodeId
			statusCb := nodeStatusCallback(callback, nodeId, "end", node.Data.Label)
			// 结束节点：直接返回输入
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "end",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCb("running")
				log.Info("Executing end node", zap.String("nodeId", nodeId))
				if input != nil {
					execLog.Input = input.Content
				}
				execLog.Status = "completed"
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCb("completed")
				return input, nil
			}))

		case "llm":
			// LLM 节点：根据配置创建模型，支持绑定工具（Function Calling 多轮调用）
			nodeModelName := ""
			nodeMaxTokens := defaultMaxTokens
			var nodeTemperature float32 = 0.7
			systemPrompt := ""
			var retryConfig *RetryConfig
			var llmToolNames []string
			llmMaxToolRounds := 5 // 默认最大工具调用轮次

			if node.Data.ModelConfig != nil {
				mc := node.Data.ModelConfig
				if mc.Model != "" {
					nodeModelName = mc.Model
				}
				if mc.MaxTokens > 0 {
					nodeMaxTokens = mc.MaxTokens
				}
				if mc.Temperature > 0 {
					nodeTemperature = float32(mc.Temperature)
				}
				systemPrompt = mc.SystemPrompt
				retryConfig = mc.Retry
				llmToolNames = mc.Tools
				if mc.MaxToolRounds > 0 {
					llmMaxToolRounds = mc.MaxToolRounds
				}
			}

			// 解析节点模型信息（支持节点级别覆盖）
			nodeEndpoint, nodeAPIKey, nodeModel := a.resolveModelInfo(nodeModelName, endpoint, apiKey, model)

			// 创建节点专属的模型
			nodeChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
				BaseURL:     nodeEndpoint,
				Model:       nodeModel,
				APIKey:      nodeAPIKey,
				MaxTokens:   &nodeMaxTokens,
				Temperature: &nodeTemperature,
			})
			if err != nil {
				log.Error("创建节点模型失败", zap.String("nodeId", nodeId), zap.Error(err))
				// 使用默认模型
				nodeChatModel = defaultChatModel
			}

			// 如果配置了工具，绑定到模型
			var llmNodeTools []tool.BaseTool
			var llmToolNode *compose.ToolsNode
			if len(llmToolNames) > 0 {
				llmNodeTools = a.getToolsByNames(llmToolNames)
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
						log.Warn("LLM 节点绑定工具失败", zap.String("nodeId", nodeId), zap.Error(err))
					} else {
						log.Info("LLM 节点绑定工具成功", zap.String("nodeId", nodeId), zap.Int("toolCount", len(nodeToolInfos)))
						// 创建工具执行器
						llmToolNode, err = compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
							Tools:               llmNodeTools,
							ExecuteSequentially: true,
						})
						if err != nil {
							log.Error("创建 LLM 节点工具执行器失败", zap.String("nodeId", nodeId), zap.Error(err))
						}
					}
				}
			}

			// 捕获变量供闭包使用
			llmBoundToolNode := llmToolNode
			llmBoundMaxRounds := llmMaxToolRounds

			// 创建 LLM 节点 Lambda
			statusCbLLM := nodeStatusCallback(callback, nodeId, "llm", node.Data.Label)
			llmRetryCfg := retryConfig
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "llm",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbLLM("running")
				log.Info("Executing LLM node", zap.String("nodeId", nodeId), zap.String("model", nodeModel))

				// 构建消息列表
				var messages []*schema.Message
				if systemPrompt != "" {
					messages = append(messages, schema.SystemMessage(systemPrompt))
				}
				if input != nil && input.Content != "" {
					messages = append(messages, input)
					execLog.Input = input.Content
				}

				var response *schema.Message
				// 带重试的首次执行
				execErr := executeWithRetry(llmRetryCfg, func() error {
					var genErr error
					response, genErr = nodeChatModel.Generate(ctx, messages)
					return genErr
				})
				if execErr != nil {
					log.Error("LLM generate error", zap.Error(execErr))
					execLog.Status = "failed"
					execLog.Error = execErr.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbLLM("failed")
					return nil, execErr
				}

				// 如果绑定了工具，执行多轮工具调用循环
				if llmBoundToolNode != nil {
					for round := 0; round < llmBoundMaxRounds; round++ {
						// 检查是否有工具调用
						if len(response.ToolCalls) == 0 {
							break
						}

						log.Info("LLM 节点工具调用", zap.String("nodeId", nodeId), zap.Int("round", round+1), zap.Int("toolCallCount", len(response.ToolCalls)))

						// 发送每个工具调用的详细信息
						if callback != nil {
							for _, tc := range response.ToolCalls {
								callback(&global.Chunk{
									NodeId:     nodeId,
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

						// 将 assistant 消息（含 ToolCalls）加入消息列表
						messages = append(messages, response)

						// 执行工具调用
						toolResults, toolErr := llmBoundToolNode.Invoke(ctx, response)
						if toolErr != nil {
							log.Error("LLM 节点工具执行失败", zap.String("nodeId", nodeId), zap.Error(toolErr))
							if callback != nil {
								callback(&global.Chunk{
									NodeId:     nodeId,
									NodeType:   "llm",
									NodeLabel:  node.Data.Label,
									ToolStatus: "failed",
									ShowMsg:    fmt.Sprintf("[%s] 工具执行失败: %s", node.Data.Label, toolErr.Error()),
								})
							}
							messages = append(messages, schema.ToolMessage("工具执行失败: "+toolErr.Error(), response.ToolCalls[0].ID))
						} else {
							// 发送工具执行结果
							if callback != nil {
								for _, tr := range toolResults {
									callback(&global.Chunk{
										NodeId:     nodeId,
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

						if callback != nil {
							callback(&global.Chunk{
								NodeId:    nodeId,
								NodeType:  "llm",
								NodeLabel: node.Data.Label,
								ShowMsg:   fmt.Sprintf("[%s] 继续思考... (第%d轮)", node.Data.Label, round+1),
							})
						}

						// 再次调用 LLM
						execErr = executeWithRetry(llmRetryCfg, func() error {
							var genErr error
							response, genErr = nodeChatModel.Generate(ctx, messages)
							return genErr
						})
						if execErr != nil {
							log.Error("LLM 多轮调用 generate error", zap.Error(execErr), zap.Int("round", round+1))
							break
						}
					}
				}

				// 回调最终结果
				if callback != nil && response.Content != "" {
					callback(&global.Chunk{Content: response.Content, ShowMsg: fmt.Sprintf("[%s] 思考中...", node.Data.Label)})
				}

				execLog.Status = "completed"
				execLog.Output = response.Content
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbLLM("completed")
				return response, nil
			}))

		case "tool":
			// 工具节点：直接执行工具（不经过 LLM，不消耗 Token）
			var toolName string
			var toolParams map[string]interface{}
			var toolRetryConfig *RetryConfig
			if node.Data.ToolConfig != nil {
				tc := node.Data.ToolConfig
				// 优先使用 ToolName（新配置），兼容 Tools[0]（旧配置）
				if tc.ToolName != "" {
					toolName = tc.ToolName
				} else if len(tc.Tools) > 0 {
					toolName = tc.Tools[0]
				}
				toolParams = tc.Params
				toolRetryConfig = tc.Retry
			}

			if toolName == "" {
				log.Warn("工具节点未配置工具名称，跳过", zap.String("nodeId", nodeId))
				continue
			}

			// 查找指定工具
			toolTools := a.getToolsByNames([]string{toolName})
			if len(toolTools) == 0 {
				log.Warn("工具节点找不到指定工具", zap.String("nodeId", nodeId), zap.String("toolName", toolName))
				continue
			}
			targetTool := toolTools[0]

			// 断言为 InvokableTool（直接执行）
			invokableTool, ok := targetTool.(tool.InvokableTool)
			if !ok {
				log.Warn("工具不支持直接执行（未实现 InvokableTool 接口）", zap.String("nodeId", nodeId), zap.String("toolName", toolName))
				continue
			}

			// 序列化工具参数
			paramsJSON := "{}"
			if len(toolParams) > 0 {
				if pBytes, err := json.Marshal(toolParams); err == nil {
					paramsJSON = string(pBytes)
				}
			}

			// 捕获变量供闭包使用
			boundParamsJSON := paramsJSON
			boundToolName := toolName

			statusCbTool := nodeStatusCallback(callback, nodeId, "tool", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "tool",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbTool("running")
				log.Info("Executing Tool node (direct)", zap.String("nodeId", nodeId), zap.String("toolName", boundToolName))

				// 构建工具参数：如果有输入，将输入内容合并到参数中
				actualParamsJSON := boundParamsJSON
				if input != nil && input.Content != "" {
					execLog.Input = input.Content
					// 尝试将输入内容作为 toolInput 参数传入
					var paramsMap map[string]interface{}
					if err := json.Unmarshal([]byte(boundParamsJSON), &paramsMap); err == nil {
						// 如果 params 中没有 toolInput，自动添加输入内容
						if _, exists := paramsMap["toolInput"]; !exists {
							paramsMap["toolInput"] = input.Content
						}
						if pBytes, err := json.Marshal(paramsMap); err == nil {
							actualParamsJSON = string(pBytes)
						}
					}
				}

				// 发送工具调用详情
				if callback != nil {
					callback(&global.Chunk{
						NodeId:     nodeId,
						NodeType:   "tool",
						NodeLabel:  node.Data.Label,
						ToolName:   boundToolName,
						ToolParams: actualParamsJSON,
						ToolStatus: "running",
						ShowMsg:    fmt.Sprintf("[%s] 调用工具: %s", node.Data.Label, boundToolName),
					})
				}

				// 直接执行工具（不经过 LLM）
				var result string
				execErr := executeWithRetry(toolRetryConfig, func() error {
					var invokeErr error
					result, invokeErr = invokableTool.InvokableRun(ctx, actualParamsJSON)
					return invokeErr
				})

				if execErr != nil {
					log.Error("工具直接执行失败", zap.String("nodeId", nodeId), zap.String("toolName", boundToolName), zap.Error(execErr))
					execLog.Status = "failed"
					execLog.Error = execErr.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbTool("failed")
					return nil, execErr
				}

				// 发送工具执行结果
				if callback != nil {
					callback(&global.Chunk{
						NodeId:     nodeId,
						NodeType:   "tool",
						NodeLabel:  node.Data.Label,
						ToolName:   boundToolName,
						ToolResult: result,
						ToolStatus: "completed",
						ShowMsg:    fmt.Sprintf("[%s] 工具 %s 执行完成", node.Data.Label, boundToolName),
					})
				}

				execLog.Status = "completed"
				execLog.Output = result
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbTool("completed")
				return schema.AssistantMessage(result, nil), nil
			}))

		case "condition":
			// 条件节点：执行条件判断并添加分支
			conditionConfig := DefaultConditionConfig()
			if node.Data.ConditionConf != nil {
				oldConfig := node.Data.ConditionConf
				conditionConfig.Type = ConditionType(oldConfig.Type)
				conditionConfig.Expression = oldConfig.Expression
				conditionConfig.LLMPrompt = oldConfig.LLMPrompt
				conditionConfig.ToolName = oldConfig.ToolName
				conditionConfig.ToolResultKey = oldConfig.ToolResultKey
				conditionConfig.ExpectedValue = oldConfig.ExpectedValue
				if oldConfig.FailureAction != "" {
					conditionConfig.FailureAction = FailureAction(oldConfig.FailureAction)
				}
				conditionConfig.FailureBranch = oldConfig.FailureBranch
			}

			// 添加条件判断节点
			statusCbCond := nodeStatusCallback(callback, nodeId, "condition", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "condition",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbCond("running")
				log.Info("Executing Condition node",
					zap.String("nodeId", nodeId),
					zap.String("type", string(conditionConfig.Type)))

				if input == nil {
					log.Warn("Condition node received nil input")
					// 返回带错误标记的消息
					return schema.UserMessage("ERROR:EMPTY_INPUT"), nil
				}

				var result *ConditionResult

				switch conditionConfig.Type {
				case ConditionTypeExpression:
					// 表达式判断 - 使用变量替换
					eval := NewExpressionEvaluator()
					vars := map[string]string{
						"output": input.Content,
					}
					result = eval.EvaluateWithVars(conditionConfig.Expression, vars)

				case ConditionTypeLLM:
					// AI判断 - 使用 LLM 判断
					llmResult, err := a.executeLLMCondition(ctx, endpoint, apiKey, model, conditionConfig.LLMPrompt, input.Content, callback)
					if err != nil {
						result = NewConditionError(err, conditionConfig)
					} else {
						result = llmResult
					}

				case ConditionTypeToolResult:
					// 工具结果判断 - 需要结合前面的工具执行结果
					toolResult := a.executeToolResultCondition(ctx, conditionConfig, input)
					result = toolResult

				default:
					result = NewConditionError(fmt.Errorf("未知的条件类型: %s", conditionConfig.Type), conditionConfig)
				}

				log.Info("Condition result",
					zap.Bool("result", result.Result),
					zap.String("branch", result.Branch),
					zap.String("reason", result.Reason))

				if result.Error != nil {
					log.Error("Condition evaluation error", zap.Error(result.Error))
				}

				// 根据结果设置分支
				branchId := conditionConfig.GetFalseHandle()
				if result.Result {
					branchId = conditionConfig.GetTrueHandle()
				}
				if result.Error != nil && result.Branch != "" {
					branchId = result.Branch
				}

				// 在消息的 Extra 中存储分支信息
				outputMsg := schema.AssistantMessage(input.Content, nil)
				outputMsg.Extra = map[string]interface{}{
					"condition_branch": branchId,
					"condition_result": result.Result,
					"condition_reason": result.Reason,
				}

				execLog.Input = input.Content
				execLog.Output = fmt.Sprintf("result=%v, branch=%s, reason=%s", result.Result, branchId, result.Reason)
				execLog.Status = "completed"
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbCond("completed")
				return outputMsg, nil
			}))

		case "code":
			// 代码节点：执行自定义代码
			codeConfig := node.Data.CodeConfig
			if codeConfig == nil {
				log.Warn("代码节点配置为空，跳过", zap.String("nodeId", nodeId))
				continue
			}

			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				// executeCodeNode 内部已处理 running/completed/failed 状态和执行日志
				result, err := executeCodeNode(ctx, nodeId, node.Data.Label, codeConfig, input, callback)
				if err != nil {
					return nil, err
				}
				return result, nil
			}))

		case "http":
			// HTTP 节点：发送 HTTP 请求
			httpConfig := node.Data.HTTPConfig
			if httpConfig == nil {
				log.Warn("HTTP 节点配置为空，跳过", zap.String("nodeId", nodeId))
				continue
			}

			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				// executeHTTPNode 内部已处理 running/completed/failed 状态和执行日志
				result, err := executeHTTPNode(ctx, nodeId, node.Data.Label, httpConfig, input, callback)
				if err != nil {
					return nil, err
				}
				return result, nil
			}))

		case "subworkflow":
			// 子工作流节点：调用其他工作流
			subWorkflowConfig := node.Data.SubWorkflowConfig
			if subWorkflowConfig == nil {
				log.Warn("子工作流节点配置为空，跳过", zap.String("nodeId", nodeId))
				continue
			}

			// 暂时跳过子工作流执行，需要注入 SubWorkflowExecutor
			log.Warn("子工作流节点暂未实现完整执行逻辑", zap.String("nodeId", nodeId))
			statusCbSub := nodeStatusCallback(callback, nodeId, "subworkflow", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "subworkflow",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbSub("running")
				// TODO: 实现子工作流执行
				if input != nil {
					execLog.Input = input.Content
				}
				execLog.Status = "completed"
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbSub("completed")
				return input, nil
			}))

		case "agent":
			// Agent 节点：支持多轮工具调用循环（参照阿里百炼 Agent 节点设计）
			agentConfig := node.Data.AgentConfig
			if agentConfig == nil {
				log.Warn("Agent 节点配置为空，跳过", zap.String("nodeId", nodeId))
				continue
			}

			// 解析节点模型信息
			agentModelName := ""
			if node.Data.ModelConfig != nil && node.Data.ModelConfig.Model != "" {
				agentModelName = node.Data.ModelConfig.Model
			}
			agentEndpoint, agentAPIKey, agentModel := a.resolveModelInfo(agentModelName, endpoint, apiKey, model)

			// 字段合并：ModelConfig 优先，缺失字段回退 AgentConfig / Merge fields: ModelConfig wins, fall back to AgentConfig
			systemPrompt := agentConfig.SystemPrompt
			var agentTools []string
			agentMaxRunSteps := 30
			if agentConfig.MaxRunSteps > 0 {
				agentMaxRunSteps = agentConfig.MaxRunSteps
			}
			// 仅 ModelConfig 提供 / ModelConfig-only fields
			maxToolRounds := 0
			temperature := float32(0.7)
			maxTokens := 4096
			var retryConfig *RetryConfig

			if node.Data.ModelConfig != nil {
				mc := node.Data.ModelConfig
				if mc.SystemPrompt != "" {
					systemPrompt = mc.SystemPrompt
				}
				if len(mc.Tools) > 0 {
					agentTools = mc.Tools
				}
				if mc.MaxToolRounds > 0 {
					maxToolRounds = mc.MaxToolRounds
				}
				if mc.Temperature > 0 {
					temperature = float32(mc.Temperature)
				}
				if mc.MaxTokens > 0 {
					maxTokens = mc.MaxTokens
				}
				retryConfig = mc.Retry
			}
			// agentConfig.Tools 仅在 modelConfig 未指定时回退 / Fallback tools to agentConfig only when modelConfig didn't provide any
			if len(agentTools) == 0 && len(agentConfig.Tools) > 0 {
				agentTools = agentConfig.Tools
			}

			// 第一版：MaxToolRounds / Retry 暂未接入 ToolCallAgent 内部循环，仅记录日志
			// First version: MaxToolRounds / Retry are not yet wired into ToolCallAgent's internal loop, only logged
			if maxToolRounds > 0 || retryConfig != nil {
				log.Info("Agent 节点高级配置已读取（第一版未接入 ToolCallAgent 内部循环）",
					zap.String("nodeId", nodeId),
					zap.Int("maxToolRounds", maxToolRounds),
					zap.Any("retry", retryConfig))
			}

			// 创建 ToolCallAgent 实例
			toolCallAgent := NewToolCallAgent(a.resProvider)
			toolCallAgent.SetSystemPrompt(systemPrompt)
			toolCallAgent.SetMaxRunSteps(agentMaxRunSteps)

			// 如果指定了工具列表，按名称过滤
			if len(agentTools) > 0 {
				toolCallAgent.tools = a.getToolsByNames(agentTools)
			}

			// 捕获变量供闭包使用 / Capture variables for closure
			agentTemperature := temperature
			agentMaxTokens := maxTokens

			statusCbAgent := nodeStatusCallback(callback, nodeId, "agent", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "agent",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbAgent("running")
				log.Info("Executing Agent node", zap.String("nodeId", nodeId), zap.String("model", agentModel))

				inputContent := ""
				if input != nil {
					inputContent = input.Content
					execLog.Input = inputContent
				}

				// 包装 callback，为 ToolCallAgent 发送的所有 Chunk 注入 NodeId
				agentNodeCallback := func(chunk *global.Chunk) error {
					chunk.NodeId = nodeId
					chunk.NodeType = "agent"
					chunk.NodeLabel = node.Data.Label
					return callback(chunk)
				}

				// 使用 ToolCallAgent 执行多轮工具调用循环
				response, err := toolCallAgent.ExecuteStreamWithConfig(ctx, agentEndpoint, agentAPIKey, agentModel, inputContent, "", agentNodeCallback, &AgentExecConfig{
					Temperature: &agentTemperature,
					MaxTokens:   &agentMaxTokens,
				})
				if err != nil {
					log.Error("Agent node execution failed", zap.Error(err))
					execLog.Status = "failed"
					execLog.Error = err.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbAgent("failed")
					return nil, err
				}

				execLog.Status = "completed"
				execLog.Output = response
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbAgent("completed")
				return schema.AssistantMessage(response, nil), nil
			}))

		case "parallel":
			// 并行组入口节点：创建并行执行节点
			parallelConfig := &ParallelConfig{
				MaxConcurrency: 0,
				WaitStrategy:   "all",
				Timeout:        0,
			}
			// 从节点配置读取
			if node.Data.ParallelConfig != nil {
				parallelConfig = node.Data.ParallelConfig
			}

			// 创建并行执行器
			parallelExecutor := NewParallelExecutor(a, parallelConfig.MaxConcurrency, endpoint, apiKey, model)

			statusCbParallel := nodeStatusCallback(callback, nodeId, "parallel", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "parallel",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbParallel("running")

				inputContent := ""
				if input != nil {
					inputContent = input.Content
					execLog.Input = inputContent
				}

				// 识别并行组
				parallelGroups, err := parallelExecutor.IdentifyParallelGroups(&topo)
				if err != nil {
					log.Error("识别并行组失败", zap.String("nodeId", nodeId), zap.Error(err))
					execLog.Status = "failed"
					execLog.Error = err.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbParallel("failed")
					return nil, err
				}

				// 找到当前 parallel 节点对应的并行组
				var currentGroup *ParallelGroup
				for _, group := range parallelGroups {
					if group.ParallelID == nodeId {
						currentGroup = group
						break
					}
				}

				if currentGroup == nil {
					log.Warn("未找到对应的并行组", zap.String("nodeId", nodeId))
					execLog.Status = "completed"
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbParallel("completed")
					return input, nil
				}

				// 执行并行组
				results, err := parallelExecutor.ExecuteParallelGroup(ctx, currentGroup, inputContent, callback)
				if err != nil {
					log.Error("并行组执行失败", zap.String("nodeId", nodeId), zap.Error(err))
					execLog.Status = "failed"
					execLog.Error = err.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbParallel("failed")
					return nil, err
				}

				// 合并结果
				mergedOutput := parallelExecutor.MergeResults(results)

				execLog.Status = "completed"
				execLog.Output = mergedOutput
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbParallel("completed")

				// 并行节点完成：结果已经写入 executionLog.output，
				// 这里不再通过 callback 推送 Content（避免被前端当作最终输出渲染）
				// Only emit a lightweight progress hint so the node panel updates.
				log.Info("并行节点执行完成",
					zap.String("nodeId", nodeId),
					zap.Int("outputLen", len(mergedOutput)),
					zap.Bool("hasCallback", callback != nil))
				if callback != nil {
					if err := callback(&global.Chunk{
						NodeId:     nodeId,
						NodeType:   "parallel",
						NodeStatus: "completed",
						ShowMsg:    "并行执行完成",
					}); err != nil {
						log.Error("并行结果 callback 失败", zap.Error(err))
					}
				}

				return schema.AssistantMessage(mergedOutput, nil), nil
			}))

		case "join":
			// 汇聚节点：创建汇聚等待节点（实际汇聚逻辑在 parallel 节点完成）
			// Join node: the actual join logic is handled in the parallel node
			statusCbJoin := nodeStatusCallback(callback, nodeId, "join", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "join",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbJoin("running")

				inputContent := ""
				if input != nil {
					inputContent = input.Content
					execLog.Input = inputContent
				}

				// 汇聚节点直接返回输入（实际汇聚逻辑在 parallel 节点完成）
				execLog.Status = "completed"
				execLog.Output = inputContent
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbJoin("completed")

				return input, nil
			}))

		}
	}

	// 收集条件节点的分支映射
	conditionBranches := make(map[string]map[string]string) // nodeId -> {handle: targetId}
	for _, edge := range topo.Edges {
		// 查找源节点类型
		var sourceNodeType string
		for _, n := range topo.Nodes {
			if n.Id == edge.Source {
				sourceNodeType = n.Data.NodeType
				break
			}
		}

		// 如果是条件节点的边，记录分支映射
		if sourceNodeType == "condition" {
			if conditionBranches[edge.Source] == nil {
				conditionBranches[edge.Source] = make(map[string]string)
			}
			handle := edge.SourceHandle
			if handle == "" {
				handle = "false" // 默认走 false 分支
			}
			conditionBranches[edge.Source][handle] = edge.Target
		}
	}

	// 添加边（跳过条件节点的边，改用 Branch）
	for _, edge := range topo.Edges {
		// 查找源节点类型
		var sourceNodeType string
		for _, n := range topo.Nodes {
			if n.Id == edge.Source {
				sourceNodeType = n.Data.NodeType
				break
			}
		}

		// 条件节点使用 Branch 而不是 Edge
		if sourceNodeType == "condition" {
			continue
		}

		// 跳过并行组相关的边（分支节点由 ParallelExecutor 内部执行，不需要 graph 边连接）
		// Skip parallel group edges (branch nodes are executed internally by ParallelExecutor)
		if sourceNodeType == "parallel" {
			continue
		}
		if _, isBranchNode := parallelBranchNodes[edge.Source]; isBranchNode {
			// branch 占位节点的出边跳过（branch → join 由 parallel → join 代替）
			continue
		}
		if _, isBranchNode := parallelBranchNodes[edge.Target]; isBranchNode {
			continue
		}

		err := graph.AddEdge(edge.Source, edge.Target)
		if err != nil {
			log.Warn("添加边失败", zap.String("source", edge.Source), zap.String("target", edge.Target), zap.Error(err))
		}
	}

	// 为每个并行节点补一条直接到下游 join 节点的边（让 eino 调度能找到 join）
	buildParallelJoinEdges(graph, &topo, parallelBranchNodes)

	// 为条件节点添加 Branch
	for conditionNodeId, branches := range conditionBranches {
		// 构建分支条件函数
		branchCondition := func(ctx context.Context, in *schema.StreamReader[*schema.Message]) (string, error) {
			msg, err := in.Recv()
			if err != nil {
				return "", err
			}

			// 从消息 Extra 中获取分支信息
			branch := ""
			if msg.Extra != nil {
				if b, ok := msg.Extra["condition_branch"].(string); ok {
					branch = b
				}
			}

			// 如果没有分支信息，默认走 false 分支
			if branch == "" {
				branch = "false"
			}

			log.Info("Condition branch selected",
				zap.String("nodeId", conditionNodeId),
				zap.String("branch", branch))

			// 返回目标节点 ID
			if targetId, ok := branches[branch]; ok {
				return targetId, nil
			}

			// 如果找不到对应分支，尝试走 false 分支
			if targetId, ok := branches["false"]; ok {
				return targetId, nil
			}

			// 最后尝试走 true 分支
			if targetId, ok := branches["true"]; ok {
				return targetId, nil
			}

			return compose.END, nil
		}

		// 构建可能的目标节点映射
		possibleTargets := make(map[string]bool)
		for _, targetId := range branches {
			possibleTargets[targetId] = true
		}
		possibleTargets[compose.END] = true // 允许结束

		// 添加分支
		err := graph.AddBranch(conditionNodeId, compose.NewStreamGraphBranch[*schema.Message](branchCondition, possibleTargets))
		if err != nil {
			log.Warn("添加条件分支失败",
				zap.String("conditionNodeId", conditionNodeId),
				zap.Any("branches", branches),
				zap.Error(err))
		} else {
			log.Info("添加条件分支成功",
				zap.String("conditionNodeId", conditionNodeId),
				zap.Any("branches", branches))
		}
	}

	if startNodeId != "" {
		graph.AddEdge(compose.START, startNodeId)
	}
	if endNodeId != "" {
		graph.AddEdge(endNodeId, compose.END)
	}

	return graph, nil
}

// buildParallelJoinEdges 为每个并行节点补一条直接到下游 join 节点的边
// 因为 branch 占位节点不会被 eino 调度，join 节点需要直接连到 parallel 节点
// Add edge: parallelId → joinId, skipping branch nodes that never receive inputs.
func buildParallelJoinEdges(graph *compose.Graph[[]*schema.Message, *schema.Message], topo *Topology, parallelBranchNodes map[string]struct{}) {
	// 先一次性算出所有并行组
	groups, _ := NewParallelExecutor(nil, 0, "", "", "").IdentifyParallelGroups(topo)

	// branchId -> parallelId
	branchToParallel := make(map[string]string)
	for _, g := range groups {
		for _, branch := range g.Branches {
			for _, b := range branch {
				branchToParallel[b.Id] = g.ParallelNode.Id
			}
		}
	}

	// 用 set 去重：每个 (parallelId, joinId) 只加一次
	addedPairs := make(map[string]struct{})

	// 对每条 (branch -> X) 边，如果 branch 属于某个 parallel，则加 (parallel -> X) 边
	for _, edge := range topo.Edges {
		if _, isBranch := parallelBranchNodes[edge.Source]; !isBranch {
			continue
		}
		parallelId, ok := branchToParallel[edge.Source]
		if !ok {
			continue
		}
		pairKey := parallelId + "->" + edge.Target
		if _, already := addedPairs[pairKey]; already {
			continue
		}
		addedPairs[pairKey] = struct{}{}

		err := graph.AddEdge(parallelId, edge.Target)
		if err != nil {
			log.Warn("添加并行→汇聚边失败",
				zap.String("source", parallelId),
				zap.String("target", edge.Target),
				zap.Error(err))
			continue
		}
		log.Info("已建立并行→汇聚边", zap.String("source", parallelId), zap.String("target", edge.Target))
	}
}

// ExecuteStream 覆写流式执行方法
func (a *WorkflowAgent) ExecuteStream(ctx context.Context, endpoint string, apiKey string, model string,
	input string, filePath string, callback func(chunk *global.Chunk) error) (string, error) {

	// 保存凭证供并行执行器使用 / Save credentials for parallel executor
	a.endpoint = endpoint
	a.apiKey = apiKey
	a.model = model

	// 包装 callback，追踪图执行过程中是否已发送过 Content
	// 避免 BaseAgent.ExecuteStream 在图执行完毕后重复发送
	contentSentDuringGraph := false
	wrappedCallback := func(chunk *global.Chunk) error {
		if chunk.Content != "" {
			contentSentDuringGraph = true
		}
		return callback(chunk)
	}

	graph, err := a.BuildGraph(ctx, endpoint, apiKey, model, wrappedCallback)
	if err != nil {
		log.Error("构建执行图失败", zap.Error(err))
		return "", err
	}
	a.SetGraph(graph)

	// 执行图
	response, err := a.Execute(ctx, endpoint, apiKey, model, input)
	if err != nil {
		return "", err
	}

	// 仅当图执行过程中未发送过 Content 时，才通过 callback 发送最终结果
	if !contentSentDuringGraph && response != "" {
		return response, callback(&global.Chunk{Content: response})
	}

	return response, nil
}

// executeLLMCondition 使用 LLM 执行条件判断
func (a *WorkflowAgent) executeLLMCondition(ctx context.Context, endpoint, apiKey, model string, prompt string, input string, callback func(chunk *global.Chunk) error) (*ConditionResult, error) {
	if prompt == "" {
		return nil, fmt.Errorf("LLM 判断提示词为空")
	}

	// 创建 LLM 模型
	maxTokens := 1024 // 判断结果不需要太长
	temperature := float32(0.1) // 低温度使结果更确定

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     endpoint,
		Model:       model,
		APIKey:      apiKey,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建判断模型失败: %w", err)
	}

	// 构建判断提示
	systemPrompt := `你是一个条件判断助手。请根据用户的输入内容，判断是否满足指定条件。

你必须严格按照以下 JSON 格式返回结果，不要包含任何其他内容：
{"result": true, "reason": "判断原因"}
或
{"result": false, "reason": "判断原因"}

result 必须是布尔值 true 或 false。
reason 简要说明判断原因。`

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(fmt.Sprintf("判断条件：%s\n\n输入内容：%s\n\n请判断输入内容是否满足条件。", prompt, input)),
	}

	if callback != nil {
		callback(&global.Chunk{ShowMsg: "[条件判断] AI 正在分析..."})
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 判断失败: %w", err)
	}

	// 解析 JSON 响应
	content := strings.TrimSpace(response.Content)
	// 尝试提取 JSON（处理可能的 markdown 代码块）
	if strings.HasPrefix(content, "```") {
		// 移除 markdown 代码块
		lines := strings.Split(content, "\n")
		var jsonLines []string
		inBlock := false
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				inBlock = !inBlock
				continue
			}
			if inBlock {
				jsonLines = append(jsonLines, line)
			}
		}
		content = strings.Join(jsonLines, "\n")
	}

	var llmResponse LLMJudgmentResponse
	if err := json.Unmarshal([]byte(content), &llmResponse); err != nil {
		// 尝试从文本中提取 true/false
		lowerContent := strings.ToLower(content)
		if strings.Contains(lowerContent, "true") || strings.Contains(lowerContent, "是") || strings.Contains(lowerContent, "yes") {
			return NewConditionResult(true, "从响应文本中推断: "+content), nil
		}
		if strings.Contains(lowerContent, "false") || strings.Contains(lowerContent, "否") || strings.Contains(lowerContent, "no") {
			return NewConditionResult(false, "从响应文本中推断: "+content), nil
		}
		return nil, fmt.Errorf("解析 LLM 判断结果失败: %w, 响应内容: %s", err, content)
	}

	if callback != nil {
		callback(&global.Chunk{ShowMsg: fmt.Sprintf("[条件判断] 结果: %v, 原因: %s", llmResponse.Result, llmResponse.Reason)})
	}

	return NewConditionResult(llmResponse.Result, llmResponse.Reason), nil
}

// executeToolResultCondition 基于工具结果执行条件判断
func (a *WorkflowAgent) executeToolResultCondition(ctx context.Context, config *ConditionConfigV2, input *schema.Message) *ConditionResult {
	// 工具结果判断需要从消息的 Extra 中获取工具执行结果
	if input == nil || input.Extra == nil {
		return NewConditionError(fmt.Errorf("无法获取工具执行结果"), config)
	}

	// 尝试从 Extra 中获取工具结果
	var toolResult interface{}
	var found bool

	// 查找工具结果
	if config.ToolResultKey != "" {
		if result, ok := input.Extra[config.ToolResultKey]; ok {
			toolResult = result
			found = true
		}
	}

	// 如果没有指定 key，尝试从常见字段获取
	if !found {
		for _, key := range []string{"result", "output", "data", "tool_result"} {
			if result, ok := input.Extra[key]; ok {
				toolResult = result
				found = true
				break
			}
		}
	}

	if !found {
		// 如果 Extra 中没有工具结果，尝试使用消息内容
		toolResult = input.Content
	}

	// 将结果转为字符串进行比较
	resultStr := fmt.Sprintf("%v", toolResult)

	// 与期望值比较
	if config.ExpectedValue != "" {
		if resultStr == config.ExpectedValue {
			return NewConditionResult(true, fmt.Sprintf("工具结果 '%s' 等于期望值 '%s'", resultStr, config.ExpectedValue))
		}
		return NewConditionResult(false, fmt.Sprintf("工具结果 '%s' 不等于期望值 '%s'", resultStr, config.ExpectedValue))
	}

	// 如果没有期望值，检查是否有表达式
	if config.Expression != "" {
		eval := NewExpressionEvaluator()
		return eval.Evaluate(config.Expression, resultStr)
	}

	// 默认：非空即为 true
	if resultStr != "" && resultStr != "null" && resultStr != "nil" {
		return NewConditionResult(true, "工具结果非空: "+resultStr)
	}

	return NewConditionResult(false, "工具结果为空")
}
