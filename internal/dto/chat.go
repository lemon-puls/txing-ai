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

	// 多模态相关字段（可选）
	Images      []string     `json:"images,omitempty"`      // 图片 URL 列表
	Attachments []Attachment `json:"attachments,omitempty"` // 文件附件列表

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

// Attachment 文件附件
type Attachment struct {
	FileName string `json:"fileName"` // 文件名
	FileURL  string `json:"fileUrl"`  // 文件 URL
	FileType string `json:"fileType"` // 文件类型
	FileSize int64  `json:"fileSize"` // 文件大小
}

type WsMessageResponse struct {
	ConversationId int64  `json:"conversationId"`
	// 消息类型：空表示普通流式内容；"resume" 表示续流应答
	Type string `json:"type,omitempty"`
	// 续流应答：当前是否还有进行中的流（true 表示已续上，后续仍有增量推送）
	Active bool `json:"active,omitempty"`
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
	ToolCallId string `json:"toolCallId,omitempty"` // 工具调用 ID（用于精确匹配同轮多次调用）
	ToolArgs   string `json:"toolArgs,omitempty"`   // 工具调用参数（截断后，仅用于展示）
	ToolStatus string `json:"toolStatus,omitempty"` // 工具调用状态 running/completed/failed
	ToolResult string `json:"toolResult,omitempty"` // 工具结果（截断后，仅用于展示）
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

// UpdateConversationParamsReq 更新会话高级参数请求（指针字段：仅更新非空字段）
type UpdateConversationParamsReq struct {
	MaxTokens         *int     `json:"max_tokens,omitempty"`
	Temperature       *float32 `json:"temperature,omitempty"`
	TopP              *float32 `json:"top_p,omitempty"`
	TopK              *int     `json:"top_k,omitempty"`
	PresencePenalty   *float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float32 `json:"frequency_penalty,omitempty"`
	RepetitionPenalty *float32 `json:"repetition_penalty,omitempty"`
	EnableWeb         *bool    `json:"enableWeb,omitempty"`
}
