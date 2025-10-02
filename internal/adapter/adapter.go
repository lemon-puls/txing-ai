package adapter

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/adapter/eino_openai"
	"txing-ai/internal/adapter/openai"
	"txing-ai/internal/adapter/polo"
	myVolcengine "txing-ai/internal/adapter/volcengine"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
)

// 存储所有 ChatRequester 工厂，发送请求时根据 channelType 选择对应的工厂创建 ChatRequester
var chatRequesterFactories = map[string]adaptercommon.ChatRequesterFactory{
	global.ChannelTypeVolcengine: myVolcengine.NewVoclEngineFactory(),
	global.ChannelTypePolo:       polo.NewPoloFactory(),
	global.ChannelOpenai:         openai.NewOpenaiFactory(),
	global.ChannelEinoOpenai:     eino_openai.NewEinoOpenaiFactory(),
}

func createChatRequest(ctx context.Context, channelConfig iface.ChannelConfig, chatConfig *adaptercommon.ChatConfig, hook global.Hook) error {
	channelType := channelConfig.GetType()
	if factory, ok := chatRequesterFactories[channelType]; ok {
		requester, err := factory.CreateChatRequester(channelConfig)
		if err != nil {
			log.Error("failed to create chat requester for channel", zap.String("channel_type", channelType), zap.Error(err))
			return err
		}
		err = requester.StreamChat(ctx, chatConfig, hook)
		if err != nil {
			log.Error("failed to stream chat for channel", zap.String("channel_type", channelType), zap.Error(err))
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown channel type %s (channel #%d)", channelType, channelConfig.GetId())
}
