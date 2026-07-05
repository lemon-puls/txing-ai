package agent

import (
	"txing-ai/internal/global"
	"txing-ai/internal/agent/workflow/condition"
	"txing-ai/internal/agent/workflow/parallel"
	"txing-ai/internal/agent/workflow/resolver"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/agent/workflow/validator"
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

type (
	ParallelExecutor = parallel.ParallelExecutor
	ParallelContext  = parallel.ParallelContext
	ParallelGroup    = parallel.ParallelGroup
)

type ChannelModelResolver = resolver.ChannelModelResolver

type (
	ValidationResult = validator.ValidationResult
	ValidationError  = validator.ValidationError
)

var (
	NewExpressionEvaluator  = condition.NewExpressionEvaluator
	NewConditionResult      = condition.NewConditionResult
	NewConditionError       = condition.NewConditionError
	DefaultConditionConfig  = condition.DefaultConditionConfig
)

var (
	NewParallelExecutor = parallel.NewParallelExecutor
	NewParallelContext  = parallel.NewParallelContext
	ReplaceVarsInParams = parallel.ReplaceVarsInParams
)

var NewChannelModelResolver = resolver.NewChannelModelResolver

var (
	ValidateTopology        = validator.ValidateTopology
	ValidateTopologyWithLLM = validator.ValidateTopologyWithLLM
)

const (
	ConditionTypeExpression = condition.ConditionTypeExpression
	ConditionTypeLLM        = condition.ConditionTypeLLM
	ConditionTypeToolResult = condition.ConditionTypeToolResult
)

func sendExecutionLog(callback func(chunk *global.Chunk) error, execLog *NodeExecutionLog) {
	types.SendExecutionLog(callback, execLog)
}
