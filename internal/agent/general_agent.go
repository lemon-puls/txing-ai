package agent

import "txing-ai/internal/iface"

// GeneralAgent is a general-purpose AI agent
type GeneralAgent struct {
	*ToolCallAgent
}

// NewGeneralAgent creates a new general-purpose agent
func NewGeneralAgent(res iface.ResourceProvider) *GeneralAgent {
	a := &GeneralAgent{
		ToolCallAgent: NewToolCallAgent(res),
	}
	a.SetSystemPrompt("You are a helpful AI assistant. Use your capabilities to assist the user.")
	return a
}
