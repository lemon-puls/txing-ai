package adaptercommon

import (
	"context"
	"txing-ai/internal/global"
	"txing-ai/internal/iface"
)

// 聊天请求器
type ChatRequester interface {
	// 流式聊天请求
	StreamChat(ctx context.Context, conf *ChatConfig, callback global.Hook) error
}

// 聊天请求器工厂
type ChatRequesterFactory interface {
	// 创建聊天请求器
	CreateChatRequester(conf iface.ChannelConfig) (ChatRequester, error)
}
