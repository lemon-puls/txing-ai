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

// ModelInfo 模型信息（包含端点和密钥）
type ModelInfo struct {
	Endpoint string
	APIKey   string
	Model    string // 映射后的模型名称
}

// ModelResolver 模型解析器接口，用于根据模型名称获取对应的端点和密钥
type ModelResolver interface {
	Resolve(modelName string) (*ModelInfo, error)
}

// WorkflowAgent 工作流智能体
type WorkflowAgent struct {
	*BaseAgent
	tools         []tool.BaseTool
	topology      string
	modelResolver ModelResolver
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int    `json:"maxRetries,omitempty"`  // 最大重试次数，默认 0（不重试）
	RetryDelay  int    `json:"retryDelay,omitempty"`  // 重试间隔（毫秒），默认 1000
	BackoffType string `json:"backoffType,omitempty"` // 退避策略: "fixed" | "exponential"，默认 "fixed"
}

// ModelConfig 模型配置
type ModelConfig struct {
	Model          string       `json:"model"`
	SystemPrompt   string       `json:"systemPrompt"`
	Temperature    float64      `json:"temperature"`
	MaxTokens      int          `json:"maxTokens"`
	ContextEnabled bool         `json:"contextEnabled"`
	Retry          *RetryConfig `json:"retry,omitempty"` // 重试配置
}

// ToolConfig 工具配置
type ToolConfig struct {
	Tools  []string               `json:"tools"`
	Params map[string]interface{} `json:"params"`
	Retry  *RetryConfig           `json:"retry,omitempty"` // 重试配置
}

// ConditionConfig 条件配置（旧版本，保持兼容）
type ConditionConfig struct {
	Type          string `json:"type"` // expression | llm | tool_result
	Expression    string `json:"expression,omitempty"`
	LLMPrompt     string `json:"llmPrompt,omitempty"`
	ToolName      string `json:"toolName,omitempty"`
	ToolResultKey string `json:"toolResultKey,omitempty"`
	ExpectedValue string `json:"expectedValue,omitempty"` // 新增：期望值
	FailureAction string `json:"failureAction,omitempty"` // 新增：错误处理策略
	FailureBranch string `json:"failureBranch,omitempty"` // 新增：错误时的默认分支
}

// NodeData 节点数据（配置直接放在 data 层级，与前端 JSON 结构一致）
type NodeData struct {
	NodeType      string           `json:"nodeType"`
	Label         string           `json:"label"`
	ModelConfig   *ModelConfig     `json:"modelConfig,omitempty"`
	ToolConfig    *ToolConfig      `json:"toolConfig,omitempty"`
	ConditionConf *ConditionConfig `json:"conditionConfig,omitempty"`
}

// NodeExecutionLog 节点执行日志
type NodeExecutionLog struct {
	NodeID    string `json:"nodeId"`
	NodeType  string `json:"nodeType"`
	NodeLabel string `json:"nodeLabel"`
	Status    string `json:"status"` // running, completed, failed
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Duration  int64  `json:"duration"` // 毫秒
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	Retry     int    `json:"retry,omitempty"` // 当前重试次数
}

// TopoNode 拓扑节点
type TopoNode struct {
	Id       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
}

// Position 节点位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// TopoEdge 拓扑边
type TopoEdge struct {
	Id           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"`
	TargetHandle string `json:"targetHandle,omitempty"`
}

// WorkflowConfig 工作流级别配置
type WorkflowConfig struct {
	DefaultModel string `json:"defaultModel,omitempty"` // 默认模型名称
	MaxRunSteps  int    `json:"maxRunSteps,omitempty"`  // 最大执行步数
}

// Topology 工作流拓扑图结构
type Topology struct {
	Nodes  []TopoNode     `json:"nodes"`
	Edges  []TopoEdge     `json:"edges"`
	Config *WorkflowConfig `json:"config,omitempty"`
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

// sendExecutionLog 发送执行日志到回调
func sendExecutionLog(callback func(chunk *global.Chunk) error, execLog *NodeExecutionLog) {
	if callback == nil || execLog == nil {
		return
	}
	callback(&global.Chunk{
		NodeId:     execLog.NodeID,
		NodeType:   execLog.NodeType,
		NodeLabel:  execLog.NodeLabel,
		NodeStatus: execLog.Status,
		ShowMsg:    fmt.Sprintf("[%s] %s (耗时: %dms)", execLog.NodeLabel, execLog.Status, execLog.Duration),
		ExecutionLog: &global.ExecutionLogInfo{
			StartTime: execLog.StartTime,
			EndTime:   execLog.EndTime,
			Duration:  execLog.Duration,
			Input:     execLog.Input,
			Output:    execLog.Output,
			Error:     execLog.Error,
			Retry:     execLog.Retry,
		},
	})
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
	for _, node := range topo.Nodes {
		nodeId := node.Id

		switch node.Data.NodeType {
		case "start":
			startNodeId = nodeId
			statusCb := nodeStatusCallback(callback, nodeId, "start", node.Data.Label)
			// 开始节点：将输入消息转换为消息列表
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
				statusCb("running")
				log.Info("Executing start node", zap.String("nodeId", nodeId))
				var result *schema.Message
				if len(input) > 0 {
					result = input[len(input)-1]
				} else {
					result = schema.UserMessage("")
				}
				statusCb("completed")
				return result, nil
			}))

		case "end":
			endNodeId = nodeId
			statusCb := nodeStatusCallback(callback, nodeId, "end", node.Data.Label)
			// 结束节点：直接返回输入
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				statusCb("running")
				log.Info("Executing end node", zap.String("nodeId", nodeId))
				statusCb("completed")
				return input, nil
			}))

		case "llm":
			// LLM 节点：根据配置创建模型
			nodeModelName := ""
			nodeMaxTokens := defaultMaxTokens
			var nodeTemperature float32 = 0.7
			systemPrompt := ""
			var retryConfig *RetryConfig

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

			// 如果有系统提示词，创建包含系统提示的 Lambda
			if systemPrompt != "" {
				statusCb := nodeStatusCallback(callback, nodeId, "llm", node.Data.Label)
				llmRetryConfig := retryConfig
				graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
					execLog := &NodeExecutionLog{
						NodeID:    nodeId,
						NodeType:  "llm",
						NodeLabel: node.Data.Label,
						StartTime: time.Now().UnixMilli(),
					}
					statusCb("running")
					log.Info("Executing LLM node", zap.String("nodeId", nodeId), zap.String("model", nodeModel))

					// 构建消息列表
					messages := []*schema.Message{
						schema.SystemMessage(systemPrompt),
					}
					if input != nil && input.Content != "" {
						messages = append(messages, input)
						execLog.Input = input.Content
					}

					var response *schema.Message
					// 带重试的执行
					execErr := executeWithRetry(llmRetryConfig, func() error {
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
						return nil, execErr
					}

					// 回调
					if callback != nil && response.Content != "" {
						callback(&global.Chunk{Content: response.Content, ShowMsg: fmt.Sprintf("[%s] 思考中...", node.Data.Label)})
					}

					execLog.Status = "completed"
					execLog.Output = response.Content
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCb("completed")
					return response, nil
				}))
			} else {
				statusCb2 := nodeStatusCallback(callback, nodeId, "llm", node.Data.Label)
				llmRetryConfig2 := retryConfig
				graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
					execLog := &NodeExecutionLog{
						NodeID:    nodeId,
						NodeType:  "llm",
						NodeLabel: node.Data.Label,
						StartTime: time.Now().UnixMilli(),
					}
					statusCb2("running")
					log.Info("Executing LLM node", zap.String("nodeId", nodeId), zap.String("model", nodeModel))

					var messages []*schema.Message
					if input != nil && input.Content != "" {
						messages = []*schema.Message{input}
						execLog.Input = input.Content
					}

					var response *schema.Message
					execErr := executeWithRetry(llmRetryConfig2, func() error {
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
						return nil, execErr
					}

					if callback != nil && response.Content != "" {
						callback(&global.Chunk{Content: response.Content, ShowMsg: fmt.Sprintf("[%s] 思考中...", node.Data.Label)})
					}

					execLog.Status = "completed"
					execLog.Output = response.Content
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCb2("completed")
					return response, nil
				}))
			}

		case "tool":
			// 工具节点：使用 LLM 调用工具
			var nodeTools []tool.BaseTool
			var toolRetryConfig *RetryConfig
			if node.Data.ToolConfig != nil {
				nodeTools = a.getToolsByNames(node.Data.ToolConfig.Tools)
				toolRetryConfig = node.Data.ToolConfig.Retry
			} else {
				nodeTools = a.tools
			}

			// 解析工具节点的模型信息（支持节点级别覆盖）
			toolModelName := ""
			if node.Data.ModelConfig != nil && node.Data.ModelConfig.Model != "" {
				toolModelName = node.Data.ModelConfig.Model
			}
			toolEndpoint, toolAPIKey, toolModel := a.resolveModelInfo(toolModelName, endpoint, apiKey, model)

			// 为工具创建一个带工具绑定的模型
			toolChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
				BaseURL:     toolEndpoint,
				Model:       toolModel,
				APIKey:      toolAPIKey,
				MaxTokens:   &defaultMaxTokens,
				Temperature: &defaultTemperature,
			})
			if err != nil {
				log.Error("创建工具模型失败", zap.String("nodeId", nodeId), zap.Error(err))
				continue
			}

			// 绑定工具
			nodeToolInfos := make([]*schema.ToolInfo, 0, len(nodeTools))
			for _, t := range nodeTools {
				info, err := t.Info(ctx)
				if err != nil {
					continue
				}
				nodeToolInfos = append(nodeToolInfos, info)
			}
			if err := toolChatModel.BindTools(nodeToolInfos); err != nil {
				log.Warn("绑定工具失败", zap.String("nodeId", nodeId), zap.Error(err))
			}

			// 创建工具节点执行器
			toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
				Tools:               nodeTools,
				ExecuteSequentially: true,
			})
			if err != nil {
				log.Error("创建工具节点失败", zap.String("nodeId", nodeId), zap.Error(err))
				continue
			}

			// 使用 Lambda 包装工具调用
			statusCbTool := nodeStatusCallback(callback, nodeId, "tool", node.Data.Label)
			nodeToolRetryConfig := toolRetryConfig
			graph.AddLambdaNode(nodeId, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
				execLog := &NodeExecutionLog{
					NodeID:    nodeId,
					NodeType:  "tool",
					NodeLabel: node.Data.Label,
					StartTime: time.Now().UnixMilli(),
				}
				statusCbTool("running")
				log.Info("Executing Tool node", zap.String("nodeId", nodeId), zap.Strings("tools", func() []string {
					names := make([]string, len(nodeTools))
					for i, t := range nodeTools {
						info, _ := t.Info(ctx)
						if info != nil {
							names[i] = info.Name
						}
					}
					return names
				}()))

				if input != nil {
					execLog.Input = input.Content
				}

				var response *schema.Message
				// 带重试的执行
				execErr := executeWithRetry(nodeToolRetryConfig, func() error {
					var genErr error
					response, genErr = toolChatModel.Generate(ctx, []*schema.Message{input})
					return genErr
				})

				if execErr != nil {
					log.Error("工具模型调用失败", zap.Error(execErr))
					execLog.Status = "failed"
					execLog.Error = execErr.Error()
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					statusCbTool("failed")
					return nil, execErr
				}

				// 检查是否有工具调用
				if len(response.ToolCalls) > 0 {
					if callback != nil {
						callback(&global.Chunk{ShowMsg: fmt.Sprintf("[%s] 正在调用工具...", node.Data.Label)})
					}

					// 执行工具调用
					results, err := toolNode.Invoke(ctx, response)
					if err != nil {
						log.Error("工具执行失败", zap.Error(err))
						execLog.Status = "failed"
						execLog.Error = err.Error()
						execLog.EndTime = time.Now().UnixMilli()
						execLog.Duration = execLog.EndTime - execLog.StartTime
						sendExecutionLog(callback, execLog)
						statusCbTool("failed")
						return response, nil
					}

					if callback != nil {
						callback(&global.Chunk{ShowMsg: fmt.Sprintf("[%s] 工具执行完成", node.Data.Label)})
					}

					// 返回最后一个结果
					execLog.Status = "completed"
					execLog.EndTime = time.Now().UnixMilli()
					execLog.Duration = execLog.EndTime - execLog.StartTime
					sendExecutionLog(callback, execLog)
					if len(results) > 0 {
						execLog.Output = results[len(results)-1].Content
						statusCbTool("completed")
						return results[len(results)-1], nil
					}
					statusCbTool("completed")
					return response, nil
				}

				execLog.Status = "completed"
				execLog.Output = response.Content
				execLog.EndTime = time.Now().UnixMilli()
				execLog.Duration = execLog.EndTime - execLog.StartTime
				sendExecutionLog(callback, execLog)
				statusCbTool("completed")
				return response, nil
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

				statusCbCond("completed")
				return outputMsg, nil
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

		err := graph.AddEdge(edge.Source, edge.Target)
		if err != nil {
			log.Warn("添加边失败", zap.String("source", edge.Source), zap.String("target", edge.Target), zap.Error(err))
		}
	}

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

// ExecuteStream 覆写流式执行方法
func (a *WorkflowAgent) ExecuteStream(ctx context.Context, endpoint string, apiKey string, model string,
	input string, filePath string, callback func(chunk *global.Chunk) error) (string, error) {

	graph, err := a.BuildGraph(ctx, endpoint, apiKey, model, callback)
	if err != nil {
		log.Error("构建执行图失败", zap.Error(err))
		return "", err
	}
	a.SetGraph(graph)

	return a.BaseAgent.ExecuteStream(ctx, endpoint, apiKey, model, input, filePath, callback)
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
