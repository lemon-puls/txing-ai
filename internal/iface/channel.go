package iface

type ChannelConfig interface {
	GetId() int64
	GetType() string
	GetRetry() int
	GetRandomSecret() string
	GetEndpoint() string
}
