package global

// 钩子函数 底层收到消息后会调用此函数将消息分块传递给业务层处理
type Hook func(chunk *Chunk) error

// 聊天消息
type Message struct {
	Role             string  `json:"role"`
	Content          string  `json:"content"`
	ReasoningContent string  `json:"reasoning_content"`
	Name             *string `json:"name,omitempty"`
	// 工作流相关字段（用于持久化工作流执行结果）
	WorkflowStatus string   `json:"workflow_status,omitempty"` // completed/failed
	Artifacts      string   `json:"artifacts,omitempty"`       // 产物 JSON
	AppName        string   `json:"app_name,omitempty"`        // 应用名称
	Files          []string `json:"files,omitempty"`           // 用户上传的文件名列表
	ExecutionLogs  string   `json:"execution_logs,omitempty"`  // 节点执行日志 JSON
	// 多模态相关字段
	Images      []string     `json:"images,omitempty"`      // 图片 URL 列表
	Attachments []Attachment `json:"attachments,omitempty"` // 文件附件列表
}

// Attachment 文件附件信息
type Attachment struct {
	FileName string `json:"fileName"` // 文件名
	FileURL  string `json:"fileUrl"`  // 文件 URL
	FileType string `json:"fileType"` // 文件类型 (image/pdf/doc/etc)
	FileSize int64  `json:"fileSize"` // 文件大小
}

// 流式聊天响应消息块
type Chunk struct {
	Content string `json:"content"`
	// 思考过程消息
	ReasoningContent string `json:"reasoning_content"`
	// toolCall Id
	ToolCallId string `json:"tool_call_id"`
	// 工具名称
	ToolName string `json:"tool_name"`
	// 工具调用参数
	ToolParams string `json:"tool_params"`
	// 工具返回结果
	ToolResult string `json:"tool_result"`
	// 工具调用状态 running/completed/failed
	ToolStatus string `json:"tool_status,omitempty"`
	// 显示信息（用于前端显示）
	ShowMsg string `json:"show_msg"`
	// 工作流节点状态信息
	NodeId     string `json:"node_id,omitempty"`
	NodeType   string `json:"node_type,omitempty"`
	NodeLabel  string `json:"node_label,omitempty"`
	NodeStatus string `json:"node_status,omitempty"` // "running" | "completed" | "failed"
	// 执行日志信息
	ExecutionLog *ExecutionLogInfo `json:"execution_log,omitempty"`
}

// ExecutionLogInfo 执行日志信息
type ExecutionLogInfo struct {
	StartTime int64  `json:"startTime"` // 开始时间戳（毫秒）
	EndTime   int64  `json:"endTime"`   // 结束时间戳（毫秒）
	Duration  int64  `json:"duration"`  // 执行耗时（毫秒）
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	Retry     int    `json:"retry,omitempty"` // 重试次数
}

// ModelMapping 模型映射规则
type ModelMapping struct {
	SourceModel string                  `json:"sourceModel"` // 源模型
	Conditions  []ModelMappingCondition `json:"conditions"`  // 条件列表
}

// ModelMappingCondition 模型映射条件
type ModelMappingCondition struct {
	TargetModel string                 `json:"targetModel"` // 目标模型
	Conditions  map[string]interface{} `json:"conditions"`  // 条件映射，key 为条件名称，value 为条件值
}
