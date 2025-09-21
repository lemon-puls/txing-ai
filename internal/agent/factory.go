package agent

import (
	"fmt"
)

// AgentType represents the type of agent to create
type AgentType string

const (
	// GeneralAgentType is a general-purpose agent
	GeneralAgentType AgentType = "general"
)

// AgentFactory creates agents of different types
type AgentFactory interface {
	// CreateAgent creates an agent of the specified type
	CreateAgent(agentType AgentType) (Agent, error)
}

// SimpleAgentFactory is a basic implementation of AgentFactory
type SimpleAgentFactory struct {
	// Registry of agent constructors
	constructors map[AgentType]func() Agent
}

// 接口实现校验
var _ AgentFactory = (*SimpleAgentFactory)(nil)

// NewSimpleAgentFactory creates a new simple agent factory
func NewSimpleAgentFactory() AgentFactory {
	factory := SimpleAgentFactory{
		constructors: make(map[AgentType]func() Agent),
	}
	// 注册一个通用 agent 类型
	factory.RegisterAgentType(GeneralAgentType, func() Agent {
		return NewGeneralAgent()
	})
	return &factory
}

// RegisterAgentType registers a constructor for an agent type
func (f *SimpleAgentFactory) RegisterAgentType(agentType AgentType, constructor func() Agent) {
	f.constructors[agentType] = constructor
}

// CreateAgent creates an agent of the specified type
func (f *SimpleAgentFactory) CreateAgent(agentType AgentType) (Agent, error) {
	// TODO 加上缓存机制，避免重复创建
	constructor, exists := f.constructors[agentType]
	if !exists {
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}

	return constructor(), nil
}
