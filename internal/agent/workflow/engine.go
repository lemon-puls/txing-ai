package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	agentpkg "txing-ai/internal/agent/agent"
	"txing-ai/internal/agent/workflow/condition"
	nodeexec "txing-ai/internal/agent/workflow/node"
	"txing-ai/internal/agent/workflow/parallel"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
	mytool "txing-ai/internal/tool"
)

// WorkflowAgent 工作流智能体
type WorkflowAgent struct {
	*agentpkg.BaseAgent
	tools         []tool.BaseTool
	topology      string
	modelResolver types.ModelResolver
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
func NewWorkflowAgent(res iface.ResourceProvider, topology string, modelResolver types.ModelResolver) *WorkflowAgent {
	baseAgent := agentpkg.NewBaseAgent("WorkflowAgent", "A dynamic workflow agent based on JSON topology")
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


// BuildGraph 构建执行图（简化版本，使用 DAG 模式）
func (a *WorkflowAgent) BuildGraph(ctx context.Context, endpoint, apiKey, model string, callback func(chunk *global.Chunk) error) (*compose.Graph[[]*schema.Message, *schema.Message], error) {
	var topo types.Topology
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
	parallelExecutorForScan := parallel.NewParallelExecutor(a, 10, "", "", "")
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
				execLog := &types.NodeExecutionLog{
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
				types.SendExecutionLog(callback, execLog)
				statusCb("completed")
				return result, nil
			}))

		case "end":
			endNodeId = nodeId
			statusCb := nodeStatusCallback(callback, nodeId, "end", node.Data.Label)
			// 结束节点：直接返回输入
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &types.NodeExecutionLog{
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
				types.SendExecutionLog(callback, execLog)
				statusCb("completed")
				return input, nil
			}))

		case "llm":
			// LLM 节点：根据配置创建模型，支持绑定工具（Function Calling 多轮调用）
			nodeModelName := ""
			nodeMaxTokens := defaultMaxTokens
			var nodeTemperature float32 = 0.7
			systemPrompt := ""
			var retryConfig *types.RetryConfig
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

			statusCbLLM := nodeStatusCallback(callback, nodeId, "llm", node.Data.Label)

			// 构建共享 LLM 执行配置，实际执行委托给 ExecuteLLM
			// Build the shared LLM exec config; execution is delegated to ExecuteLLM.
			llmNodeCfg := &LLMExecConfig{
				NodeID:            nodeId,
				NodeLabel:         node.Data.Label,
				NodeType:          "llm",
				ModelResolver:     a.modelResolver,
				ModelName:         nodeModelName,
				DefaultEndpoint:   endpoint,
				DefaultAPIKey:     apiKey,
				DefaultModel:      model,
				FallbackChatModel: defaultChatModel,
				MaxTokens:         nodeMaxTokens,
				Temperature:       nodeTemperature,
				SystemPrompt:      systemPrompt,
				AllTools:          a.tools,
				ToolNames:         llmToolNames,
				MaxToolRounds:     llmMaxToolRounds,
				Retry:             retryConfig,
				EmitFinalContent:  true,
			}

			// 创建 LLM 节点 Lambda
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &types.NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "llm",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbLLM("running")
				log.Info("Executing LLM node", zap.String("nodeId", nodeId))

				inputContent := ""
				if input != nil {
					inputContent = input.Content
					execLog.Input = inputContent
				}

				// 委托共享执行核心（模型解析/工具绑定/多轮工具循环均在 ExecuteLLM 内完成）
				// Delegate to the shared execution core.
				result, execErr := ExecuteLLM(ctx, llmNodeCfg, inputContent, callback)
				if execErr != nil {
					execLog.Status = "failed"
					execLog.Error = execErr.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					types.SendExecutionLog(callback, execLog)
					statusCbLLM("failed")
					return nil, execErr
				}

				execLog.Status = "completed"
				execLog.Output = result
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				types.SendExecutionLog(callback, execLog)
				statusCbLLM("completed")
				return schema.AssistantMessage(result, nil), nil
			}))

		case "tool":
			// 工具节点：直接执行工具（不经过 LLM，不消耗 Token）
			var toolName string
			var toolParams map[string]interface{}
			var toolRetryConfig *types.RetryConfig
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
				execLog := &types.NodeExecutionLog{
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
					types.SendExecutionLog(callback, execLog)
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
				types.SendExecutionLog(callback, execLog)
				statusCbTool("completed")
				return schema.AssistantMessage(result, nil), nil
			}))

		case "condition":
			// 条件节点：执行条件判断并添加分支
			conditionConfig := condition.DefaultConditionConfig()
			if node.Data.ConditionConf != nil {
				oldConfig := node.Data.ConditionConf
				conditionConfig.Type = condition.ConditionType(oldConfig.Type)
				conditionConfig.Expression = oldConfig.Expression
				conditionConfig.LLMPrompt = oldConfig.LLMPrompt
				conditionConfig.ToolName = oldConfig.ToolName
				conditionConfig.ToolResultKey = oldConfig.ToolResultKey
				conditionConfig.ExpectedValue = oldConfig.ExpectedValue
				if oldConfig.FailureAction != "" {
					conditionConfig.FailureAction = condition.FailureAction(oldConfig.FailureAction)
				}
				conditionConfig.FailureBranch = oldConfig.FailureBranch
			}

			// 添加条件判断节点
			statusCbCond := nodeStatusCallback(callback, nodeId, "condition", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &types.NodeExecutionLog{
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

				var result *condition.ConditionResult

				switch conditionConfig.Type {
				case condition.ConditionTypeExpression:
					// 表达式判断 - 使用变量替换
					eval := condition.NewExpressionEvaluator()
					vars := map[string]string{
						"output": input.Content,
					}
					result = eval.EvaluateWithVars(conditionConfig.Expression, vars)

				case condition.ConditionTypeLLM:
					// AI判断 - 使用 LLM 判断
					llmResult, err := a.executeLLMCondition(ctx, endpoint, apiKey, model, conditionConfig.LLMPrompt, input.Content, callback)
					if err != nil {
						result = condition.NewConditionError(err, conditionConfig)
					} else {
						result = llmResult
					}

				case condition.ConditionTypeToolResult:
					// 工具结果判断 - 需要结合前面的工具执行结果
					toolResult := a.executeToolResultCondition(ctx, conditionConfig, input)
					result = toolResult

				default:
					result = condition.NewConditionError(fmt.Errorf("未知的条件类型: %s", conditionConfig.Type), conditionConfig)
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
				types.SendExecutionLog(callback, execLog)
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
				result, err := nodeexec.ExecuteCodeNode(ctx, nodeId, node.Data.Label, codeConfig, input, callback)
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
				result, err := nodeexec.ExecuteHTTPNode(ctx, nodeId, node.Data.Label, httpConfig, input, callback)
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
				execLog := &types.NodeExecutionLog{
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
				types.SendExecutionLog(callback, execLog)
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
			var retryConfig *types.RetryConfig

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

			// 工具调用轮次上限：优先 ModelConfig.MaxToolRounds，否则回退 AgentConfig.MaxRunSteps
			// Tool-call round limit: ModelConfig.MaxToolRounds wins, falling back to AgentConfig.MaxRunSteps
			agentMaxToolRounds := agentMaxRunSteps
			if maxToolRounds > 0 {
				agentMaxToolRounds = maxToolRounds
			}

			statusCbAgent := nodeStatusCallback(callback, nodeId, "agent", node.Data.Label)

			// 构建共享 LLM 执行配置；未配置工具时回退为全部工具（保留 Agent 节点语义）
			// Build the shared LLM exec config. When no tools are configured,
			// fall back to all tools to preserve Agent-node semantics.
			agentNodeCfg := &LLMExecConfig{
				NodeID:               nodeId,
				NodeLabel:            node.Data.Label,
				NodeType:             "agent",
				ModelResolver:        a.modelResolver,
				ModelName:            agentModelName,
				DefaultEndpoint:      endpoint,
				DefaultAPIKey:        apiKey,
				DefaultModel:         model,
				MaxTokens:            maxTokens,
				Temperature:          temperature,
				SystemPrompt:         systemPrompt,
				AllTools:             a.tools,
				ToolNames:            agentTools,
				UseAllToolsWhenEmpty: true,
				MaxToolRounds:        agentMaxToolRounds,
				Retry:                retryConfig,
				EmitFinalContent:     true,
			}
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &types.NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "agent",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbAgent("running")
				log.Info("Executing Agent node", zap.String("nodeId", nodeId))

				inputContent := ""
				if input != nil {
					inputContent = input.Content
					execLog.Input = inputContent
				}

				// 包装 callback，为执行核心发送的所有 Chunk 注入节点信息
				// Inject node info into every chunk emitted by the shared executor.
				agentNodeCallback := callback
				if callback != nil {
					agentNodeCallback = func(chunk *global.Chunk) error {
						chunk.NodeId = nodeId
						chunk.NodeType = "agent"
						chunk.NodeLabel = node.Data.Label
						return callback(chunk)
					}
				}

				// 委托共享执行核心执行多轮工具调用循环
				// Delegate the multi-round tool-calling loop to the shared executor.
				response, err := ExecuteLLM(ctx, agentNodeCfg, inputContent, agentNodeCallback)
				if err != nil {
					log.Error("Agent node execution failed", zap.Error(err))
					execLog.Status = "failed"
					execLog.Error = err.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					types.SendExecutionLog(callback, execLog)
					statusCbAgent("failed")
					return nil, err
				}

				execLog.Status = "completed"
				execLog.Output = response
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				types.SendExecutionLog(callback, execLog)
				statusCbAgent("completed")
				return schema.AssistantMessage(response, nil), nil
			}))

		case "parallel":
			// 并行组入口节点：创建并行执行节点
			parallelConfig := &types.ParallelConfig{
				MaxConcurrency: 0,
				WaitStrategy:   "all",
				Timeout:        0,
			}
			// 从节点配置读取
			if node.Data.ParallelConfig != nil {
				parallelConfig = node.Data.ParallelConfig
			}

			// 创建并行执行器
			parallelExecutor := parallel.NewParallelExecutor(a, parallelConfig.MaxConcurrency, endpoint, apiKey, model)

			statusCbParallel := nodeStatusCallback(callback, nodeId, "parallel", node.Data.Label)
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &types.NodeExecutionLog{
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
					types.SendExecutionLog(callback, execLog)
					statusCbParallel("failed")
					return nil, err
				}

				// 找到当前 parallel 节点对应的并行组
				var currentGroup *parallel.ParallelGroup
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
					types.SendExecutionLog(callback, execLog)
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
					types.SendExecutionLog(callback, execLog)
					statusCbParallel("failed")
					return nil, err
				}

				// 合并结果
				mergedOutput := parallelExecutor.MergeResults(results)

				execLog.Status = "completed"
				execLog.Output = mergedOutput
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				types.SendExecutionLog(callback, execLog)
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
				execLog := &types.NodeExecutionLog{
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
				types.SendExecutionLog(callback, execLog)
				statusCbJoin("completed")

				return input, nil
			}))

		}
	}

	// 预建节点类型索引，避免内层循环重复查找
	nodeTypeMap := make(map[string]string, len(topo.Nodes))
	for _, n := range topo.Nodes {
		nodeTypeMap[n.Id] = n.Data.NodeType
	}

	// 收集条件节点的分支映射
	conditionBranches := make(map[string]map[string]string) // nodeId -> {handle: targetId}
	for _, edge := range topo.Edges {
		if nodeTypeMap[edge.Source] == "condition" {
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
		sourceNodeType := nodeTypeMap[edge.Source]

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
	buildParallelJoinEdges(graph, &topo)

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

// ExecuteLLMNodeInParallel 在并行上下文中执行 LLM 节点（实现 parallel.NodeExecutor 接口）
// ExecuteLLMNodeInParallel executes an LLM node in parallel context with real LLM calls
