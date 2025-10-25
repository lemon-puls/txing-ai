package agent

import (
	"context"
	"encoding/json"
	"fmt"
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

// ToolCallAgent 通用智能体实现
type ToolCallAgent struct {
	*BaseAgent
	tools []tool.BaseTool
}

// NewToolCallAgent 创建一个新的通用智能体
func NewToolCallAgent(res iface.ResourceProvider) *ToolCallAgent {
	baseAgent := NewBaseAgent("ToolCallAgent", "A general-purpose AI agent")

	baseAgent.SetSystemPrompt("You are a helpful AI assistant that can use tools to solve problems.")

	return &ToolCallAgent{
		BaseAgent: baseAgent,
		tools:     mytool.ProvideTools(res),
	}
}

// Execute 执行通用智能体任务
func (a *ToolCallAgent) Execute(ctx context.Context,
	endpoint string, apiKey string, model string, input string) (string, error) {

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: endpoint,
		Model:   model, // 使用的模型版本
		APIKey:  apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("Failed to create chat model: %w", err)
	}

	// 创建一个包含工具的执行图
	graph, err := newGraph(context.Background(), chatModel, a.tools, func(chunk *global.Chunk) error {
		// 不需要处理chunk
		return nil
	})
	if err != nil {
		log.Error("Failed to create graph", zap.Error(err))
		return "", err
	}
	a.SetGraph(graph)

	response, err := a.BaseAgent.Execute(ctx, endpoint, apiKey, model, input)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (a *ToolCallAgent) ExecuteStream(ctx context.Context, endpoint string, apiKey string, model string,
	input string, filePath string, callback func(chunk *global.Chunk) error) (string, error) {

	// 设置 LLM 响应最大 token 数量，一些模型（例如 DeepSeek v3）默认是 4k，这里上调到 8k，否则最终生成的结果可能会超长导致被截断
	maxTokens := 8192
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:   endpoint,
		Model:     model, // 使用的模型版本
		APIKey:    apiKey,
		MaxTokens: &maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("Failed to create chat model: %w", err)
	}

	// 创建一个包含工具的执行图
	graph, err := newGraph(ctx, chatModel, a.tools, callback)
	if err != nil {
		log.Error("Failed to create graph", zap.Error(err))
		return "", err
	}
	a.SetGraph(graph)

	return a.BaseAgent.ExecuteStream(ctx, endpoint, apiKey, model, input, filePath, callback)
}

func newGraph(ctx context.Context, model *openai.ChatModel, tools []tool.BaseTool,
	callback func(chunk *global.Chunk) error) (*compose.Graph[[]*schema.Message, *schema.Message], error) {

	// 创建工具节点时确保正确配置
	todoToolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools:               tools,
		ExecuteSequentially: true,
	})
	if err != nil {
		log.Error("Failed to create tool node", zap.Error(err))
		return nil, err
	}

	// 设置模型
	// 获取工具信息并绑定到 ChatModel
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Error("Failed to get tool info", zap.Error(err))
		}
		toolInfos = append(toolInfos, info)
	}
	err = model.BindTools(toolInfos)
	if err != nil {
		log.Error("Failed to bind tools to model", zap.Error(err))
		return nil, err
	}

	// 1. 创建Graph，定义输入输出类型
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *AgentState {
			return &AgentState{Messages: make([]*schema.Message, 0)}
		}))

	// 2. 添加节点
	// 在节点预处理中维护消息序列
	modelPreHandle := func(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
		for _, msg := range input {
			log.Debug("model input", zap.String("role", string(msg.Role)), zap.String("content", msg.Content))
			if msg.ToolCallID != "" {
				var showMsg string
				showMsg, err = mytool.BuildResponseShowMsg(msg.ToolName, msg.Content)
				if err != nil {
					showMsg = "完成工具调用：" + msg.Content
				}
				callback(&global.Chunk{
					ToolCallId: msg.ToolCallID,
					ToolName:   msg.ToolName,
					ToolResult: msg.Content,
					ShowMsg:    showMsg,
				})
			}
		}
		state.Messages = append(state.Messages, input...)
		return state.Messages, nil
	}

	// 工具节点的预处理器，接收单个消息并添加到状态中
	toolsPreHandle := func(ctx context.Context, input *schema.Message, state *AgentState) (*schema.Message, error) {
		// 打印工具调用信息
		for _, call := range input.ToolCalls {
			log.Debug("tool call", zap.String("name", call.Function.Name), zap.Any("args", call.Function.Arguments))
			var showMsg string
			showMsg, err = mytool.BuildRequestShowMsg(call.Function.Name, call.Function.Arguments)
			if err != nil {
				showMsg = "发起工具调用：" + call.Function.Name
			}
			callback(&global.Chunk{
				ToolCallId: call.ID,
				ToolName:   call.Function.Name,
				ToolParams: call.Function.Arguments,
				ShowMsg:    showMsg,
			})
		}
		state.Messages = append(state.Messages, input)
		// 返回输入消息，保持类型一致性
		return input, nil
	}

	toolInvokeParamErrorHandlePre := func(ctx context.Context, in *schema.Message, state *AgentState) (*schema.Message, error) {
		state.Messages = append(state.Messages, in)
		return in, nil
	}

	toolInvokeParamErrorHandle := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) ([]*schema.Message, error) {
		if input.ToolCalls != nil && len(input.ToolCalls) > 0 {

			// 检查工具调用参数是否是合法的JSON
			for _, toolCall := range input.ToolCalls {
				if toolCall.Function.Arguments != "" {
					var jsonObj map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &jsonObj); err != nil {
						// 参数不是合法的JSON，构造错误信息
						errorMsg := fmt.Sprintf("工具调用参数不是合法的JSON: %s, ToolCallID: %s, 错误: %v", toolCall.Function.Name, toolCall.ID, err)
						log.Error(errorMsg)

						// 创建一个新的消息，包含错误信息
						errorResponse := &schema.Message{
							Role:       schema.Tool,
							Content:    errorMsg,
							ToolName:   toolCall.Function.Name,
							ToolCallID: toolCall.ID,
						}
						return []*schema.Message{errorResponse}, nil
					}
				}
			}
		}
		return []*schema.Message{
			{
				Role:    schema.Tool,
				Content: "工具调用参数不合法，请检查参数是否是合法的JSON",
			},
		}, nil
	})

	_ = graph.AddChatModelNode("model", model, compose.WithStatePreHandler(modelPreHandle))
	_ = graph.AddToolsNode("tools", todoToolsNode, compose.WithStatePreHandler(toolsPreHandle))
	_ = graph.AddLambdaNode("tool_invoke_param_error_handle", toolInvokeParamErrorHandle,
		compose.WithStatePreHandler(toolInvokeParamErrorHandlePre))

	//countLimit := 0

	// 定义分支条件函数 - 判断模型输出是否包含工具调用
	modelPostBranchCondition := func(ctx context.Context, in *schema.StreamReader[*schema.Message]) (string, error) {
		msg, err := in.Recv()
		if err != nil {
			return "", err
		}
		if msg.ToolCalls != nil && len(msg.ToolCalls) > 0 {
			//countLimit += 1
			//if countLimit >= 20 {
			//	return compose.END, nil // 超过10个工具调用，结束
			//}

			// 检查工具调用参数是否是合法的JSON
			for _, toolCall := range msg.ToolCalls {
				if toolCall.Function.Arguments != "" {
					var jsonObj map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &jsonObj); err != nil {
						// 参数不是合法的JSON，构造错误信息
						errorMsg := fmt.Sprintf("检测到不合法的工具调用参数: %s, 错误: %v", toolCall.Function.Name, err)
						log.Error(errorMsg)
						// 返回到模型节点，让模型处理错误
						return "tool_invoke_param_error_handle", nil
					}
				}
			}

			log.Info("GO TO tools")
			return "tools", nil // 如果包含工具调用，转到tools节点
		}
		log.Info("GO TO END")
		return compose.END, nil // 否则结束
	}

	// 3. 构建边和循环 - 这是实现多轮的关键
	_ = graph.AddEdge(compose.START, "model") // 初始输入到模型
	// 添加一个分支：判断模型输出是否包含工具调用
	_ = graph.AddBranch("model", compose.NewStreamGraphBranch[*schema.Message](modelPostBranchCondition, map[string]bool{
		"tool_invoke_param_error_handle": true,
		"tools":                          true,
		compose.END:                      true,
	}))
	_ = graph.AddEdge("tools", "model") // 工具执行结果直接反馈给模型，形成循环
	_ = graph.AddEdge("tool_invoke_param_error_handle", "model")
	return graph, nil
}
