package global

// 消息来源类型
const (
	System    = "system"
	User      = "user"
	Assistant = "assistant"
)

// 渠道类型
const (
	ChannelTypeVolcengine = "volcengine"
	ChannelTypePolo       = "polo"
	ChannelOpenai         = "OpenAI"
	ChannelEinoOpenai     = "Eino OpenAI"
)

// 所有大模型
const (
	ModelDeepSeekV3 = "deepseek-v3-250324"
	ModelDeepSeekR1 = "deepseek-r1-250120"
)

// 消息类型
const (
	MessageTypeChat = "chat"
	MessageTypeStop = "stop"
)

// 目标模型类型（用于模型映射条件）
const (
	// 直连 LLM
	LLMTypeModel = "model"
	// 三方提供的应用 （支持内置联网搜索等功能）
	LLMTypeAPP = "app"
)

// channel 模型映射条件的默认值
var ChannelModelMappingConditionDefaultVal = map[string]interface{}{
	"enableWeb": false,
	"type":      LLMTypeModel,
}
