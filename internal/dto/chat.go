package dto

type WsMessageRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Model   string `json:"model"`
	// 暂时不支持由客户端指定context
	//Context   int    `json:"context"`
	EnableWeb bool `json:"enableWeb"`

	// optional fields
	MaxTokens         *int     `json:"max_tokens,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"top_p,omitempty"`
	TopK              *int     `json:"top_k,omitempty"`
	PresencePenalty   *float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float32 `json:"frequency_penalty,omitempty"`
	RepetitionPenalty *float32 `json:"repetition_penalty,omitempty"`
}

type WsMessageResponse struct {
	ConversationId int64  `json:"conversationId"`
	Content        string `json:"content"`
	// 思考过程消息
	ReasoningContent string `json:"reasoning_content"`
	End              bool   `json:"end"`
}
