package openai

import (
	"txing-ai/internal/adapter/common"
	"txing-ai/internal/iface"
)

// https://code.poloapi.com/

type OpenaiFactory struct {
	// apiKey -> requester
	requesterMap map[string]adaptercommon.ChatRequester
}

func (v OpenaiFactory) CreateChatRequester(conf iface.ChannelConfig) (adaptercommon.ChatRequester, error) {
	// 判断 requester 是否已经存在，如果存在直接返回，否则新建一个
	// TODO: 保证并发安全
	secret := conf.GetRandomSecret()
	if requester, ok := v.requesterMap[secret]; ok {
		return requester, nil
	}
	requester := NewChatClient(conf.GetEndpoint(), secret)
	v.requesterMap[secret] = requester
	return requester, nil
}

func NewOpenaiFactory() *OpenaiFactory {
	return &OpenaiFactory{
		requesterMap: make(map[string]adaptercommon.ChatRequester),
	}
}

var _ adaptercommon.ChatRequesterFactory = (*OpenaiFactory)(nil)
