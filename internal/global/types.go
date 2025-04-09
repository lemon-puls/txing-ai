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
