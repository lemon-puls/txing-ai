package agent

import "github.com/cloudwego/eino/schema"

// 在状态中维护消息历史
type AgentState struct {
	Messages []*schema.Message
}
