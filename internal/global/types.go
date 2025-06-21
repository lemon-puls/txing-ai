package global

// 钩子函数 底层收到消息后会调用此函数将消息分块传递给业务层处理
type Hook func(chunk *Chunk) error

// 聊天消息
type Message struct {
	Role             string  `json:"role"`
	Content          string  `json:"content"`
	ReasoningContent string  `json:"reasoning_content"`
	Name             *string `json:"name,omitempty"`
}

// 流式聊天响应消息块
type Chunk struct {
	Content string `json:"content"`
	// 思考过程消息
	ReasoningContent string `json:"reasoning_content"`
}

// ModelMapping 模型映射规则
type ModelMapping struct {
	SourceModel string                  `json:"sourceModel"` // 源模型
	Conditions  []ModelMappingCondition `json:"conditions"`  // 条件列表
}

// ModelMappingCondition 模型映射条件
type ModelMappingCondition struct {
	TargetModel string                 `json:"targetModel"` // 目标模型
	Conditions  map[string]interface{} `json:"conditions"`  // 条件映射，key 为条件名称，value 为条件值
}
