package agent

// GeneralAgent is a general-purpose AI agent
type GeneralAgent struct {
	*ToolCallAgent
}

// NewGeneralAgent creates a new general-purpose agent
func NewGeneralAgent() *GeneralAgent {
	a := &GeneralAgent{
		ToolCallAgent: NewToolCallAgent(),
	}
	a.SetSystemPrompt("You are a helpful AI assistant. Use your capabilities to assist the user.")
	return a
}
