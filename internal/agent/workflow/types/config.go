package types

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int    `json:"maxRetries,omitempty"`  // 最大重试次数，默认 0（不重试）
	RetryDelay  int    `json:"retryDelay,omitempty"`  // 重试间隔（毫秒），默认 1000
	BackoffType string `json:"backoffType,omitempty"` // 退避策略: "fixed" | "exponential"，默认 "fixed"
}

// ModelConfig 模型配置
type ModelConfig struct {
	Model          string       `json:"model"`
	SystemPrompt   string       `json:"systemPrompt"`
	Temperature    float64      `json:"temperature"`
	MaxTokens      int          `json:"maxTokens"`
	ContextEnabled bool         `json:"contextEnabled"`
	Tools          []string     `json:"tools,omitempty"`  // 绑定的工具列表（LLM 通过 Function Calling 自主调用）
	MaxToolRounds  int          `json:"maxToolRounds,omitempty"` // 最大工具调用轮次，默认 5
	Retry          *RetryConfig `json:"retry,omitempty"` // 重试配置
}

// ToolConfig 工具配置
type ToolConfig struct {
	ToolName string                 `json:"toolName,omitempty"` // 单个工具名称（直接执行模式）
	Params   map[string]interface{} `json:"params,omitempty"`   // 工具参数（直接执行模式）
	Tools    []string               `json:"tools,omitempty"`    // 工具名称列表（兼容旧配置）
	Retry    *RetryConfig           `json:"retry,omitempty"`    // 重试配置
}

// ConditionConfig 条件配置（旧版本，保持兼容）
type ConditionConfig struct {
	Type          string `json:"type"` // expression | llm | tool_result
	Expression    string `json:"expression,omitempty"`
	LLMPrompt     string `json:"llmPrompt,omitempty"`
	ToolName      string `json:"toolName,omitempty"`
	ToolResultKey string `json:"toolResultKey,omitempty"`
	ExpectedValue string `json:"expectedValue,omitempty"` // 新增：期望值
	FailureAction string `json:"failureAction,omitempty"` // 新增：错误处理策略
	FailureBranch string `json:"failureBranch,omitempty"` // 新增：错误时的默认分支
}

// CodeConfig 代码节点配置
type CodeConfig struct {
	Language string `json:"language"`           // 语言: "javascript" | "python" | "go"
	Code     string `json:"code"`               // 代码内容
	Timeout  int    `json:"timeout,omitempty"`   // 超时时间（秒），默认 30
}

// HTTPConfig HTTP 节点配置
type HTTPConfig struct {
	Method  string            `json:"method"`            // HTTP 方法: "GET" | "POST" | "PUT" | "DELETE"
	URL     string            `json:"url"`               // 请求 URL
	Headers map[string]string `json:"headers,omitempty"` // 请求头
	Body    string            `json:"body,omitempty"`    // 请求体（支持 {{output}} 变量替换）
	Timeout int               `json:"timeout,omitempty"` // 超时时间（秒），默认 30
}

// SubWorkflowConfig 子工作流节点配置
type SubWorkflowConfig struct {
	WorkflowID int64  `json:"workflowId"`          // 子工作流 ID
	Input      string `json:"input,omitempty"`     // 输入模板（支持 {{output}} 变量替换）
	Timeout    int    `json:"timeout,omitempty"`   // 超时时间（秒），默认 60
}

// AgentConfig Agent 节点配置（支持多轮工具调用循环）
type AgentConfig struct {
	SystemPrompt string   `json:"systemPrompt"`         // 系统提示词
	Tools        []string `json:"tools,omitempty"`       // 工具名称列表（为空则使用全部工具）
	MaxRunSteps  int      `json:"maxRunSteps,omitempty"` // 最大执行步数，默认 30
}

// ParallelConfig 并行组节点配置
// ParallelConfig defines the configuration for parallel execution groups
type ParallelConfig struct {
	MaxConcurrency int          `json:"maxConcurrency"` // 最大并发数，0=无限制 / Max concurrency, 0=unlimited
	WaitStrategy   string      `json:"waitStrategy"`   // 等待策略：all / Wait strategy: all
	Timeout        int         `json:"timeout"`        // 超时时间（秒）/ Timeout in seconds
	BranchRetry    *RetryConfig `json:"branchRetry,omitempty"`
}

// JoinConfig 汇聚节点配置
// JoinConfig defines the configuration for join nodes
type JoinConfig struct {
	Strategy string `json:"strategy"` // 汇聚策略：all/any / Join strategy: all/any
	Timeout  int    `json:"timeout"`  // 超时时间（秒）/ Timeout in seconds
}
