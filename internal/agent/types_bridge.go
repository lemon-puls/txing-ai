package agent

import (
	"txing-ai/internal/global"
	"txing-ai/internal/agent/workflow/condition"
	"txing-ai/internal/agent/workflow/types"
)

// 临时桥接：Task 9 迁移 core 后删除本文件
type (
	ModelInfo          = types.ModelInfo
	ModelResolver      = types.ModelResolver
	RetryConfig        = types.RetryConfig
	ModelConfig        = types.ModelConfig
	ToolConfig         = types.ToolConfig
	ConditionConfig    = types.ConditionConfig
	CodeConfig         = types.CodeConfig
	HTTPConfig         = types.HTTPConfig
	SubWorkflowConfig  = types.SubWorkflowConfig
	AgentConfig        = types.AgentConfig
	NodeData           = types.NodeData
	NodeExecutionLog   = types.NodeExecutionLog
	TopoNode           = types.TopoNode
	Position           = types.Position
	TopoEdge           = types.TopoEdge
	SchemaField        = types.SchemaField
	WorkflowConfig     = types.WorkflowConfig
	Topology           = types.Topology
	ParallelConfig     = types.ParallelConfig
	JoinConfig         = types.JoinConfig
	ParallelResult     = types.ParallelResult
	JoinResult         = types.JoinResult
)

type (
	ConditionType       = condition.ConditionType
	FailureAction       = condition.FailureAction
	ConditionConfigV2   = condition.ConditionConfigV2
	ConditionResult     = condition.ConditionResult
	ExpressionEvaluator = condition.ExpressionEvaluator
	LLMJudgmentResponse = condition.LLMJudgmentResponse
)

var (
	NewExpressionEvaluator  = condition.NewExpressionEvaluator
	NewConditionResult      = condition.NewConditionResult
	NewConditionError       = condition.NewConditionError
	DefaultConditionConfig  = condition.DefaultConditionConfig
)

const (
	ConditionTypeExpression = condition.ConditionTypeExpression
	ConditionTypeLLM        = condition.ConditionTypeLLM
	ConditionTypeToolResult = condition.ConditionTypeToolResult
)

func sendExecutionLog(callback func(chunk *global.Chunk) error, execLog *NodeExecutionLog) {
	types.SendExecutionLog(callback, execLog)
}
