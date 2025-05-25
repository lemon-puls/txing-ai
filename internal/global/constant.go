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
