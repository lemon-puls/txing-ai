package dto

type WsMessageRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Model   string `json:"model"`
	// 暂时不支持由客户端指定context
	//Context   int    `json:"context"`
	EnableWeb bool `json:"enableWeb"`

	// 工作流集成（可选）
	WorkflowID *int64         `json:"workflowId,omitempty"` // 指定应用 ID
	Files      []WorkflowFile `json:"files,omitempty"`      // 附件列表

	// optional fields
	MaxTokens         *int     `json:"max_tokens,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"top_p,omitempty"`
	TopK              *int     `json:"top_k,omitempty"`
	PresencePenalty   *float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float32 `json:"frequency_penalty,omitempty"`
	RepetitionPenalty *float32 `json:"repetition_penalty,omitempty"`
}

// WorkflowFile 工作流附件
type WorkflowFile struct {
	FieldName string `json:"fieldName"` // inputSchema 中的字段名
	FileURL   string `json:"fileUrl"`   // 文件 URL
	FileName  string `json:"fileName"`  // 原始文件名
}

type WsMessageResponse struct {
	ConversationId int64  `json:"conversationId"`
	Content        string `json:"content"`
	// 思考过程消息
	ReasoningContent string `json:"reasoning_content"`
	End              bool   `json:"end"`

	// 工作流集成（可选）
	Workflow  *WorkflowProgress `json:"workflow,omitempty"`  // 工作流执行进度
	Artifacts []ArtifactInfo    `json:"artifacts,omitempty"` // 生成的文件产物
}

// WorkflowProgress 工作流执行进度
type WorkflowProgress struct {
	Status     string `json:"status"`               // running/completed/failed
	NodeID     string `json:"nodeId,omitempty"`     // 当前节点 ID
	NodeType   string `json:"nodeType,omitempty"`   // 节点类型
	NodeLabel  string `json:"nodeLabel,omitempty"`  // 节点标签
	NodeStatus string `json:"nodeStatus,omitempty"` // 节点状态
	ShowMsg    string `json:"showMsg,omitempty"`    // 显示消息
	ToolName   string `json:"toolName,omitempty"`   // 工具名称
	ToolResult string `json:"toolResult,omitempty"` // 工具结果
}

// ArtifactInfo 文件产物信息
type ArtifactInfo struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	Ids []int64 `json:"ids" binding:"required"` // 会话ID列表
}
