package adaptercommon

import "txing-ai/internal/global"

type ChatConfig struct {
	Model     string           `json:"model"`
	Message   []global.Message `json:"message,omitempty"`
	EnableWeb bool             `json:"enable_web,omitempty"`
	// 可选的模型参数
	MaxTokens         *int     `json:"max_tokens,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"top_p,omitempty"`
	TopK              *int     `json:"top_k,omitempty"`
	PresencePenalty   *float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float32 `json:"frequency_penalty,omitempty"`
	RepetitionPenalty *float32 `json:"repetition_penalty,omitempty"`
}
