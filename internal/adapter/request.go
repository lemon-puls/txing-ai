package adapter

import (
	"context"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/global"
	"txing-ai/internal/iface"
)

func NewChatRequest(ctx context.Context, channelConfig iface.ChannelConfig, chatConfig *adaptercommon.ChatConfig, hook global.Hook) error {

	// TODO 实现限流和重试等机制

	return createChatRequest(ctx, channelConfig, chatConfig, hook)

}
