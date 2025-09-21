// Package agent provides a layered architecture for AI agents
package agent

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	_ "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"txing-ai/internal/iface"
)

// Agent 定义智能体接口
type Agent interface {
	// Execute 执行智能体任务
	Execute(ctx context.Context, channelConfig iface.ChannelConfig, model string, input string) (string, error)
	// GetName 获取智能体名称
	GetName() string
	// GetDescription 获取智能体描述
	GetDescription() string
}

// 校验接口实现
var _ Agent = (*BaseAgent)(nil)

// BaseAgent 提供Agent接口的基本实现
type BaseAgent struct {
	name         string
	description  string
	model        *openai.ChatModel
	graph        *compose.Graph[[]*schema.Message, *schema.Message]
	systemPrompt string
}

// NewBaseAgent 创建一个新的基础智能体
func NewBaseAgent(name, description string) *BaseAgent {
	return &BaseAgent{
		name:        name,
		description: description,
	}
}

// GetName 获取智能体名称
func (a *BaseAgent) GetName() string {
	return a.name
}

// GetDescription 获取智能体描述
func (a *BaseAgent) GetDescription() string {
	return a.description
}

// SetGraph 设置智能体的执行图
func (a *BaseAgent) SetGraph(graph *compose.Graph[[]*schema.Message, *schema.Message]) {
	a.graph = graph
}

// SetSystemPrompt 设置系统提示
func (a *BaseAgent) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

// Execute 执行智能体任务的默认实现
func (a *BaseAgent) Execute(ctx context.Context,
	channelConfig iface.ChannelConfig, model string, input string) (string, error) {
	if a.graph == nil {
		// 如果没有设置执行图，则直接使用模型生成回复
		messages := []*schema.Message{
			schema.SystemMessage("You are a helpful AI assistant."),
			schema.UserMessage(input),
		}

		response, err := a.model.Generate(ctx, messages)
		if err != nil {
			return "", err
		}

		return response.Content, nil
	}

	// 使用执行图处理输入
	messages := []*schema.Message{
		schema.SystemMessage(a.systemPrompt),
		schema.UserMessage(input),
	}

	// 4. 编译Graph，并设置最大步数防止无限循环
	agent, err := a.graph.Compile(ctx, compose.WithMaxRunSteps(500))

	response, err := agent.Invoke(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}
