package agent

// ConditionType 条件类型
type ConditionType string

const (
	ConditionTypeExpression ConditionType = "expression"  // 表达式判断
	ConditionTypeLLM        ConditionType = "llm"         // AI判断
	ConditionTypeToolResult ConditionType = "tool_result" // 工具结果判断
)

// FailureAction 错误处理策略
type FailureAction string

const (
	FailureActionDefaultFalse  FailureAction = "default_false" // 出错时走 false 分支（默认）
	FailureActionTerminate     FailureAction = "terminate"      // 出错时终止工作流
	FailureActionConfigurable  FailureAction = "configurable"   // 出错时走用户配置的默认分支
)

// ConditionConfigV2 扩展的条件配置
type ConditionConfigV2 struct {
	Type           ConditionType `json:"type"`                     // 条件类型
	Expression     string        `json:"expression,omitempty"`      // 表达式（expression 类型）
	LLMPrompt      string        `json:"llmPrompt,omitempty"`      // AI判断提示词（llm 类型）
	ToolName       string        `json:"toolName,omitempty"`        // 工具名称（tool_result 类型）
	ToolResultKey  string        `json:"toolResultKey,omitempty"`   // 结果字段（tool_result 类型）
	ExpectedValue  string        `json:"expectedValue,omitempty"`   // 期望值（tool_result 类型）
	TrueHandle     string        `json:"trueHandle,omitempty"`      // true 分支的 Handle ID，默认 "true"
	FalseHandle    string        `json:"falseHandle,omitempty"`     // false 分支的 Handle ID，默认 "false"
	FailureAction  FailureAction `json:"failureAction,omitempty"`   // 错误处理策略
	FailureBranch  string        `json:"failureBranch,omitempty"`  // 错误时的默认分支（configurable 类型）
}

// ConditionResult 条件判断结果
type ConditionResult struct {
	Result bool     `json:"result"`          // 判断结果
	Reason string   `json:"reason"`          // 判断原因
	Error  error    `json:"error,omitempty"` // 错误信息
	Branch string   `json:"branch"`          // 应该走的分支 Handle ID
}

// NewConditionResult 创建条件结果
func NewConditionResult(result bool, reason string) *ConditionResult {
	handle := "false"
	if result {
		handle = "true"
	}
	return &ConditionResult{
		Result: result,
		Reason: reason,
		Branch: handle,
	}
}

// NewConditionError 创建错误结果
func NewConditionError(err error, config *ConditionConfigV2) *ConditionResult {
	branch := "false" // 默认走 false 分支

	// 根据错误处理策略决定分支
	switch config.FailureAction {
	case FailureActionTerminate:
		// 终止工作流，Branch 留空表示终止
		return &ConditionResult{
			Result: false,
			Reason: err.Error(),
			Error:  err,
			Branch: "",
		}
	case FailureActionConfigurable:
		// 使用用户配置的默认分支
		branch = config.FailureBranch
		if branch == "" {
			branch = "false"
		}
	case FailureActionDefaultFalse:
		fallthrough
	default:
		branch = "false"
	}

	return &ConditionResult{
		Result: false,
		Reason: err.Error(),
		Error:  err,
		Branch: branch,
	}
}

// GetTrueHandle 获取 true 分支 Handle ID
func (c *ConditionConfigV2) GetTrueHandle() string {
	if c.TrueHandle == "" {
		return "true"
	}
	return c.TrueHandle
}

// GetFalseHandle 获取 false 分支 Handle ID
func (c *ConditionConfigV2) GetFalseHandle() string {
	if c.FalseHandle == "" {
		return "false"
	}
	return c.FalseHandle
}

// LLMJudgmentResponse LLM 判断响应结构
type LLMJudgmentResponse struct {
	Result bool   `json:"result"`
	Reason string `json:"reason"`
}

// DefaultConditionConfig 创建默认条件配置
func DefaultConditionConfig() *ConditionConfigV2 {
	return &ConditionConfigV2{
		Type:          ConditionTypeExpression,
		Expression:    "",
		FailureAction: FailureActionDefaultFalse,
		TrueHandle:    "true",
		FalseHandle:   "false",
	}
}
