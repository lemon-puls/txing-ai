# internal/agent 包结构重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将扁平的 `internal/agent` 单包拆分为职责清晰的多个子包（agent / workflow / workflow/{types,condition,node,validator,parallel,resolver}），通过接口注入打破 core↔parallel 循环依赖。

**Architecture:** 自底向上分层迁移。先建 `types` 基础包（共享数据类型 + 执行日志辅助），再迁移仅依赖 types 的叶子包（condition/resolver/node/validator），然后迁移 parallel（定义 `NodeExecutor` 接口回调 core），再迁移完全独立的 agent 子包，最后迁移 workflow 根包（WorkflowAgent 实现 NodeExecutor），末尾改造 6 个外部调用点。每步以 `go build ./...` + `go test ./internal/agent/...` 通过为验收标准。

**Tech Stack:** Go 1.x, GORM, CloudWeGo Eino, Gin。

## Global Constraints

- 纯结构重构：不改变任何运行时行为、函数逻辑、JSON tag，仅调整包路径、包名、import、以及为跨包调用而导出（首字母大写）原本未导出的符号。
- 每完成一个 Task 必须 `go build ./...` 通过；涉及测试的 Task 还须 `go test ./internal/agent/...` 通过。
- 频繁提交：每个 Task 末尾提交一次。
- 包路径前缀：`txing-ai/internal/agent`。子包如 `txing-ai/internal/agent/workflow/types`。
- 中英双语注释保持原样，不改动注释内容。
- 不保留旧 `agent` 包的兼容转发层。
- git 移动文件用 `git mv` 保留历史。

## 关键跨包导出清单（贯穿全程，务必一致）

以下符号原为包内私有，迁移后需**导出**以供跨包调用，全程使用这些新名字：

| 原符号（private） | 新符号（exported） | 归属包 |
|---|---|---|
| `sendExecutionLog` | `types.SendExecutionLog` | types |
| `executeCodeNode` | `node.ExecuteCodeNode` | node |
| `executeHTTPNode` | `node.ExecuteHTTPNode` | node |
| `executeSubWorkflowNode` | `node.ExecuteSubWorkflowNode` | node |

以下符号原已导出，仅随迁移换包路径（引用处加包名前缀）：
`ModelInfo, ModelResolver, RetryConfig, ModelConfig, ToolConfig, ConditionConfig, CodeConfig, HTTPConfig, SubWorkflowConfig, AgentConfig, NodeData, NodeExecutionLog, TopoNode, Position, TopoEdge, SchemaField, WorkflowConfig, Topology, ParallelConfig, JoinConfig, ParallelResult, JoinResult` → `types.*`；`ConditionType, FailureAction, ConditionConfigV2, ConditionResult, ExpressionEvaluator, LLMJudgmentResponse, NewExpressionEvaluator, NewConditionResult, NewConditionError, DefaultConditionConfig` → `condition.*`；`ValidateTopology, ValidateTopologyWithLLM, ValidationResult, ValidationError, ValidationLevel` → `validator.*`；`ChannelModelResolver, NewChannelModelResolver` → `resolver.*`；`ParallelExecutor, NewParallelExecutor, ParallelContext, ...` → `parallel.*`；agent 相关 → `agent.*`（新子包，包名仍为 `agent`）。

## 目录目标结构

```
internal/agent/
├── agent/          base.go factory.go state.go errors.go general.go resume.go travel.go toolcall.go
└── workflow/
    ├── types/      topology.go config.go model.go result.go execlog.go
    ├── condition/  condition.go expression.go expression_test.go
    ├── node/       code.go http.go subworkflow.go
    ├── validator/  structural.go llm.go
    ├── parallel/   executor.go context.go context_test.go group.go vars.go
    ├── resolver/   channel.go
    └── engine.go   (WorkflowAgent 核心 + retry + condition_exec + join_edges，可再拆多文件)
```

---

## Task 1: 建立 workflow/types 基础包

**Files:**
- Create: `internal/agent/workflow/types/topology.go`
- Create: `internal/agent/workflow/types/config.go`
- Create: `internal/agent/workflow/types/model.go`
- Create: `internal/agent/workflow/types/result.go`
- Create: `internal/agent/workflow/types/execlog.go`
- Modify: `internal/agent/workflow_agent.go`（删除已迁出的类型/函数定义，暂不改包名——本 Task 结束时该文件会因缺失类型而无法编译，属预期，最终验证放在整体阶段。见下方步骤说明）

> **重要**：Go 不允许项目中长期存在无法编译的包。为保证每步可编译，本 Task 采用「**复制到新包 + 在旧包用类型别名桥接**」策略：新包定义真身，旧 `agent` 包临时用 `type X = types.X` 别名和 `var SendExecutionLog = ...` 转发，使旧包仍编译通过。桥接别名在 Task 9（迁移 core 后）统一删除。

**Interfaces:**
- Produces（types 包对外提供）：
  - `types.Topology{Nodes []TopoNode; Edges []TopoEdge; Config *WorkflowConfig}`
  - `types.TopoNode{Id string; Type string; Position Position; Data NodeData}`
  - `types.Position{X,Y float64}`，`types.TopoEdge{Id,Source,Target,SourceHandle,TargetHandle string}`
  - `types.NodeData{...}`（含 `ParallelConfig *ParallelConfig`、`JoinConfig *JoinConfig`、`ConditionConf *ConditionConfig` 等字段，全部指向 types 包内类型）
  - `types.WorkflowConfig`, `types.SchemaField`, `types.ModelConfig`, `types.ToolConfig`, `types.ConditionConfig`, `types.CodeConfig`, `types.HTTPConfig`, `types.SubWorkflowConfig`, `types.AgentConfig`, `types.RetryConfig`
  - `types.ModelInfo{Endpoint,APIKey,Model string}`，`types.ModelResolver interface{ Resolve(string)(*ModelInfo,error) }`
  - `types.ParallelConfig{MaxConcurrency int; WaitStrategy string; Timeout int; BranchRetry *RetryConfig}`
  - `types.JoinConfig{Strategy string; Timeout int}`
  - `types.ParallelResult{...}`，`types.JoinResult{...}`（见 parallel_context.go 原定义）
  - `types.NodeExecutionLog{...}`
  - `types.SendExecutionLog(callback func(*global.Chunk) error, execLog *NodeExecutionLog)`

- [ ] **Step 1: 创建 topology.go**

写入 `internal/agent/workflow/types/topology.go`，`package types`，把这些类型从 `workflow_agent.go` 复制过来（内容逐字不变，仅改包名）：`TopoNode`、`Position`、`TopoEdge`、`SchemaField`、`WorkflowConfig`、`Topology`、`NodeData`。

注意 `NodeData` 引用 `ModelConfig/ToolConfig/ConditionConfig/CodeConfig/HTTPConfig/SubWorkflowConfig/AgentConfig/ParallelConfig/JoinConfig`——这些也都在 types 包内（config.go），所以同包直接引用，无需前缀。

- [ ] **Step 2: 创建 config.go**

写入 `internal/agent/workflow/types/config.go`，`package types`，复制：`RetryConfig`、`ModelConfig`、`ToolConfig`、`ConditionConfig`、`CodeConfig`、`HTTPConfig`、`SubWorkflowConfig`、`AgentConfig`、`ParallelConfig`、`JoinConfig`（后两个原在 `workflow_parallel.go` 第 24-36 行，一并搬来）。逐字复制，含全部 JSON tag 与注释。

- [ ] **Step 3: 创建 model.go**

写入 `internal/agent/workflow/types/model.go`，`package types`，复制 `ModelInfo`、`ModelResolver`（原 workflow_agent.go 第 23-33 行）。

- [ ] **Step 4: 创建 result.go**

写入 `internal/agent/workflow/types/result.go`，`package types`，复制 `ParallelResult`、`JoinResult`（原 `parallel_context.go` 第 120-末尾，找到这两个 struct 定义逐字复制）。

- [ ] **Step 5: 创建 execlog.go**

写入 `internal/agent/workflow/types/execlog.go`，`package types`。复制 `NodeExecutionLog` 类型（原 workflow_agent.go 133-145）与 `sendExecutionLog` 函数体（原 workflow_agent.go 293-314），**函数改名为导出的 `SendExecutionLog`**。import `"txing-ai/internal/global"`。函数签名：

```go
func SendExecutionLog(callback func(chunk *global.Chunk) error, execLog *NodeExecutionLog) {
    // 函数体与原 sendExecutionLog 完全一致
}
```

- [ ] **Step 6: 在旧 agent 包建立桥接别名**

创建 `internal/agent/types_bridge.go`，`package agent`：

```go
package agent

import (
	"txing-ai/internal/global"
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

func sendExecutionLog(callback func(chunk *global.Chunk) error, execLog *NodeExecutionLog) {
	types.SendExecutionLog(callback, execLog)
}
```

- [ ] **Step 7: 从旧文件删除已迁移的定义**

- 从 `workflow_agent.go` 删除这些类型定义：`ModelInfo`、`ModelResolver`、`RetryConfig`、`ModelConfig`、`ToolConfig`、`ConditionConfig`、`CodeConfig`、`HTTPConfig`、`SubWorkflowConfig`、`AgentConfig`、`NodeData`、`NodeExecutionLog`、`TopoNode`、`Position`、`TopoEdge`、`SchemaField`、`WorkflowConfig`、`Topology`（第 23-33、47-200 行区间内的这些 struct/interface），以及 `sendExecutionLog` 函数（293-314）。保留 `WorkflowAgent`、`WorkflowAgentState`、retry 辅助函数等。
- 从 `workflow_parallel.go` 删除 `ParallelConfig`、`JoinConfig` 定义（24-36）。
- 从 `parallel_context.go` 删除 `ParallelResult`、`JoinResult` 定义（保留 `ParallelContext`）。
- 检查 workflow_agent.go 顶部 import：若删除 `sendExecutionLog` 后 `"math"`/`"time"` 等仍被 retry 辅助用到则保留，`go build` 会提示未使用的 import，据此清理。

- [ ] **Step 8: 构建验证**

Run: `go build ./...`
Expected: 通过。旧 agent 包通过桥接别名引用 types，其余代码不受影响。

- [ ] **Step 9: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS（现有 expression / parallel_context 测试仍在旧包，别名保证类型一致）。

- [ ] **Step 10: Commit**

```bash
git add internal/agent/workflow/types/ internal/agent/types_bridge.go internal/agent/workflow_agent.go internal/agent/workflow_parallel.go internal/agent/parallel_context.go
git commit -m "refactor(agent): 抽取共享数据类型到 workflow/types 基础包"
```

---

## Task 2: 迁移 condition 子包（条件与表达式）

**Files:**
- Create: `internal/agent/workflow/condition/condition.go`（from `condition_types.go`）
- Create: `internal/agent/workflow/condition/expression.go`（from `expression_eval.go`）
- Create: `internal/agent/workflow/condition/expression_test.go`（from `expression_eval_test.go`）
- Delete: `internal/agent/condition_types.go`, `expression_eval.go`, `expression_eval_test.go`
- Modify: `internal/agent/types_bridge.go`（追加 condition 桥接别名）

**Interfaces:**
- Consumes: 无（condition/expression 自包含，不依赖其它子包）
- Produces:
  - `condition.ConditionType`, `condition.FailureAction`, `condition.ConditionConfigV2`, `condition.ConditionResult`
  - `condition.NewConditionResult(result bool, reason string) *ConditionResult`
  - `condition.NewConditionError(err error, config *ConditionConfigV2) *ConditionResult`
  - `condition.DefaultConditionConfig() *ConditionConfigV2`
  - `condition.LLMJudgmentResponse`
  - `condition.ExpressionEvaluator`，`condition.NewExpressionEvaluator() *ExpressionEvaluator`
  - `(*ExpressionEvaluator).Evaluate(expression, input string) *ConditionResult` 等方法（签名不变）

- [ ] **Step 1: git mv 三个文件到 condition 目录并改名**

```bash
git mv internal/agent/condition_types.go internal/agent/workflow/condition/condition.go
git mv internal/agent/expression_eval.go internal/agent/workflow/condition/expression.go
git mv internal/agent/expression_eval_test.go internal/agent/workflow/condition/expression_test.go
```

- [ ] **Step 2: 改包名**

将三个文件顶部 `package agent` 改为 `package condition`。

- [ ] **Step 3: 追加桥接别名**

在 `internal/agent/types_bridge.go` 追加：

```go
import "txing-ai/internal/agent/workflow/condition"

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
```

- [ ] **Step 4: 构建验证**

Run: `go build ./...`
Expected: 通过。core（workflow_agent.go）通过别名调用 `NewExpressionEvaluator()`/`ConditionResult` 不变。

- [ ] **Step 5: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS（expression_test.go 现属 condition 包，测试逻辑不变）。

- [ ] **Step 6: Commit**

```bash
git add -A internal/agent/workflow/condition internal/agent/types_bridge.go
git commit -m "refactor(agent): 迁移条件与表达式求值到 workflow/condition 子包"
```

---

## Task 3: 迁移 resolver 子包（模型解析）

**Files:**
- Create: `internal/agent/workflow/resolver/channel.go`（from `channel_model_resolver.go`）
- Delete: `internal/agent/channel_model_resolver.go`
- Modify: `internal/agent/types_bridge.go`（追加 resolver 桥接）

**Interfaces:**
- Consumes: `types.ModelInfo`（返回值类型）
- Produces:
  - `resolver.ChannelModelResolver`
  - `resolver.NewChannelModelResolver(db *gorm.DB) *ChannelModelResolver`
  - `(*ChannelModelResolver).Resolve(modelName string) (*types.ModelInfo, error)`

- [ ] **Step 1: git mv**

```bash
git mv internal/agent/channel_model_resolver.go internal/agent/workflow/resolver/channel.go
```

- [ ] **Step 2: 改包名并加 types 前缀**

- 顶部 `package agent` → `package resolver`。
- 增加 import `"txing-ai/internal/agent/workflow/types"`。
- 函数返回类型 `*ModelInfo` → `*types.ModelInfo`（Resolve 方法签名与内部 `return &ModelInfo{...}` → `&types.ModelInfo{...}`）。

- [ ] **Step 3: 追加桥接别名**

在 `types_bridge.go` 追加：

```go
import "txing-ai/internal/agent/workflow/resolver"

type ChannelModelResolver = resolver.ChannelModelResolver

var NewChannelModelResolver = resolver.NewChannelModelResolver
```

- [ ] **Step 4: 构建验证**

Run: `go build ./...`
Expected: 通过。

- [ ] **Step 5: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add -A internal/agent/workflow/resolver internal/agent/types_bridge.go
git commit -m "refactor(agent): 迁移模型解析到 workflow/resolver 子包"
```

---

## Task 4: 迁移 node 子包（节点执行器）

**Files:**
- Create: `internal/agent/workflow/node/code.go`（from `node_code.go`）
- Create: `internal/agent/workflow/node/http.go`（from `node_http.go`）
- Create: `internal/agent/workflow/node/subworkflow.go`（from `node_subworkflow.go`）
- Delete: `internal/agent/node_code.go`, `node_http.go`, `node_subworkflow.go`
- Modify: `internal/agent/workflow_agent.go`（core 对 node 执行器的调用改为 `node.ExecuteXxx` + import）

**Interfaces:**
- Consumes: `types.CodeConfig/HTTPConfig/SubWorkflowConfig`, `types.NodeExecutionLog`, `types.SendExecutionLog`
- Produces:
  - `node.ExecuteCodeNode(ctx, nodeId, nodeLabel string, config *types.CodeConfig, input *schema.Message, callback func(*global.Chunk) error) (*schema.Message, error)`
  - `node.ExecuteHTTPNode(ctx, nodeId, nodeLabel string, config *types.HTTPConfig, input *schema.Message, callback func(*global.Chunk) error) (*schema.Message, error)`
  - `node.SubWorkflowExecutor interface { ExecuteSubWorkflow(ctx, workflowID int64, input string, callback func(*global.Chunk) error) (string, error) }`
  - `node.ExecuteSubWorkflowNode(ctx, nodeId, nodeLabel string, config *types.SubWorkflowConfig, input *schema.Message, executor node.SubWorkflowExecutor, callback func(*global.Chunk) error) (*schema.Message, error)`

- [ ] **Step 1: git mv 三文件**

```bash
git mv internal/agent/node_code.go internal/agent/workflow/node/code.go
git mv internal/agent/node_http.go internal/agent/workflow/node/http.go
git mv internal/agent/node_subworkflow.go internal/agent/workflow/node/subworkflow.go
```

- [ ] **Step 2: 改包名、导出函数、加 types 前缀**

对三个文件：
- `package agent` → `package node`。
- 增加 import `"txing-ai/internal/agent/workflow/types"`。
- 导出执行函数：`executeCodeNode` → `ExecuteCodeNode`，`executeHTTPNode` → `ExecuteHTTPNode`，`executeSubWorkflowNode` → `ExecuteSubWorkflowNode`。
- config 参数类型加前缀：`*CodeConfig` → `*types.CodeConfig`，`*HTTPConfig` → `*types.HTTPConfig`，`*SubWorkflowConfig` → `*types.SubWorkflowConfig`。
- `NodeExecutionLog{...}` → `types.NodeExecutionLog{...}`（三文件内 execLog 构造处）。
- 若函数体内调用了 `sendExecutionLog(...)`，改为 `types.SendExecutionLog(...)`。
- `SubWorkflowExecutor` 接口保持在 subworkflow.go（现为 `node.SubWorkflowExecutor`），签名不变。

- [ ] **Step 3: 修改 core 调用点**

在 `workflow_agent.go`：
- 增加 import `"txing-ai/internal/agent/workflow/node"`。
- 第 979 行 `executeCodeNode(ctx, nodeId, node.Data.Label, codeConfig, input, callback)` → `node.ExecuteCodeNode(...)`。
- 第 996 行 `executeHTTPNode(...)` → `node.ExecuteHTTPNode(...)`。
- 搜索 `executeSubWorkflowNode` 调用点，改为 `node.ExecuteSubWorkflowNode(...)`。
- 若 core 有实现/传入 `SubWorkflowExecutor` 的地方，类型引用改为 `node.SubWorkflowExecutor`。

> 注意变量名冲突：core 里循环变量常叫 `node`（`for _, node := range ...`）。引入 `node` 包会与之冲突。**用 import 别名规避**：`nodeexec "txing-ai/internal/agent/workflow/node"`，调用写 `nodeexec.ExecuteCodeNode(...)`。全 Task 统一用 `nodeexec` 别名。

- [ ] **Step 4: 构建验证**

Run: `go build ./...`
Expected: 通过。

- [ ] **Step 5: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add -A internal/agent/workflow/node internal/agent/workflow_agent.go
git commit -m "refactor(agent): 迁移节点执行器到 workflow/node 子包"
```

---

## Task 5: 迁移 validator 子包（拓扑校验）

**Files:**
- Create: `internal/agent/workflow/validator/structural.go`（from `workflow_validator.go`）
- Create: `internal/agent/workflow/validator/llm.go`（from `workflow_llm_validator.go`）
- Delete: `internal/agent/workflow_validator.go`, `workflow_llm_validator.go`
- Modify: `internal/agent/types_bridge.go`（追加 validator 桥接，供外部调用点过渡期使用）

**Interfaces:**
- Consumes: `types.Topology/TopoNode/TopoEdge/NodeData`
- Produces:
  - `validator.ValidationLevel`, `validator.LevelError/LevelWarning`（const）
  - `validator.ValidationError{Level,...}`，`validator.ValidationResult{...}`
  - `validator.ValidateTopology(topologyJSON string) *ValidationResult`
  - `validator.ValidateTopologyWithLLM(ctx, endpoint, apiKey, model, topologyJSON string) (*ValidationResult, error)`

- [ ] **Step 1: git mv**

```bash
git mv internal/agent/workflow_validator.go internal/agent/workflow/validator/structural.go
git mv internal/agent/workflow_llm_validator.go internal/agent/workflow/validator/llm.go
```

- [ ] **Step 2: 改包名、加 types 前缀**

两文件 `package agent` → `package validator`，增加 import `"txing-ai/internal/agent/workflow/types"`。将文件内所有 `Topology`→`types.Topology`、`TopoNode`→`types.TopoNode`、`TopoEdge`→`types.TopoEdge`、`NodeData`→`types.NodeData`（含函数参数、局部变量声明 `var topo Topology` 等）。`ValidationResult/ValidationError/ValidationLevel` 是本包定义，不加前缀。

- [ ] **Step 3: 追加桥接别名**

在 `types_bridge.go` 追加：

```go
import "txing-ai/internal/agent/workflow/validator"

type (
	ValidationResult = validator.ValidationResult
	ValidationError  = validator.ValidationError
)

var (
	ValidateTopology        = validator.ValidateTopology
	ValidateTopologyWithLLM = validator.ValidateTopologyWithLLM
)
```

- [ ] **Step 4: 构建验证**

Run: `go build ./...`
Expected: 通过。

- [ ] **Step 5: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add -A internal/agent/workflow/validator internal/agent/types_bridge.go
git commit -m "refactor(agent): 迁移拓扑校验到 workflow/validator 子包"
```

---

## Task 6: 迁移 parallel 子包并定义 NodeExecutor 接口（打破循环）

**Files:**
- Create: `internal/agent/workflow/parallel/executor.go`（ParallelExecutor 主体 + NodeExecutor 接口）
- Create: `internal/agent/workflow/parallel/context.go`（from `parallel_context.go` 的 ParallelContext）
- Create: `internal/agent/workflow/parallel/context_test.go`（from `parallel_context_test.go`）
- Create: `internal/agent/workflow/parallel/group.go`（分组识别/分支追踪/拓扑排序）
- Create: `internal/agent/workflow/parallel/vars.go`（replaceVars* / toString / escapeJSON）
- Delete: `internal/agent/workflow_parallel.go`, `parallel_context.go`, `parallel_context_test.go`
- Modify: `internal/agent/workflow_agent.go`（core 改用 `parallel.NewParallelExecutor`，并把 `ExecuteLLMNodeInParallel`/`ExecuteToolNodeInParallel` 两方法留在 core 作为接口实现）

**Interfaces:**
- Consumes: `types.*`（TopoNode/Topology/ParallelConfig/JoinConfig/ParallelResult/JoinResult/NodeExecutionLog）, `condition.NewExpressionEvaluator`
- Produces:
  - `parallel.NodeExecutor interface { ExecuteLLMNodeInParallel(ctx, node *types.TopoNode, input string, callback func(*global.Chunk) error) (string, error); ExecuteToolNodeInParallel(ctx, node *types.TopoNode, input string, callback func(*global.Chunk) error) (string, error) }`
  - `parallel.ParallelExecutor`，`parallel.ParallelGroup`
  - `parallel.NewParallelExecutor(executor NodeExecutor, maxWorkers int, endpoint, apiKey, model string) *ParallelExecutor`
  - `parallel.ParallelContext`，`parallel.NewParallelContext(total int) *ParallelContext`
  - `(*ParallelExecutor).IdentifyParallelGroups(topo *types.Topology) ([]*ParallelGroup, error)`
  - `(*ParallelExecutor).ExecuteParallelGroup(...)`, `.MergeResults(...)`, `.WaitForJoin(...)`, `.TopologicalSort(...)`, `.ValidateParallelGroup(...)`, `.BuildParallelGraph(...)`（签名中的 `*TopoNode`/`*Topology` 加 `types.` 前缀）

- [ ] **Step 1: git mv 主体文件与拆分**

```bash
git mv internal/agent/workflow_parallel.go internal/agent/workflow/parallel/executor.go
git mv internal/agent/parallel_context.go internal/agent/workflow/parallel/context.go
git mv internal/agent/parallel_context_test.go internal/agent/workflow/parallel/context_test.go
```

（group.go / vars.go 的拆分在 Step 5 做，先让主体跑通。）

- [ ] **Step 2: 定义 NodeExecutor 接口，改造 ParallelExecutor 持有接口**

在 `executor.go`：
- `package agent` → `package parallel`。
- 增加 import：`"txing-ai/internal/agent/workflow/types"`、`"txing-ai/internal/agent/workflow/condition"`。
- 在文件顶部新增接口：

```go
// NodeExecutor 由 workflow 核心实现，供并行分支回调执行 LLM/Tool 节点
// NodeExecutor is implemented by the workflow core to execute LLM/Tool nodes within parallel branches
type NodeExecutor interface {
	ExecuteLLMNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error)
	ExecuteToolNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error)
}
```

- 改 `ParallelExecutor` 结构体字段：`agent *WorkflowAgent` → `executor NodeExecutor`。
- 改构造函数签名：`NewParallelExecutor(agent *WorkflowAgent, ...)` → `NewParallelExecutor(executor NodeExecutor, ...)`，函数体 `agent: agent` → `executor: executor`。
- 把回调点 `e.agent.ExecuteLLMNodeInParallel(...)` → `e.executor.ExecuteLLMNodeInParallel(...)`，`e.agent.ExecuteToolNodeInParallel(...)` → `e.executor.ExecuteToolNodeInParallel(...)`。

- [ ] **Step 3: 移除挂在 WorkflowAgent 上的两个方法**

`executor.go` 中原有 `func (e *WorkflowAgent) ExecuteLLMNodeInParallel(...)`（516 行起）和 `func (e *WorkflowAgent) ExecuteToolNodeInParallel(...)`（718 行起）**从本文件删除**——它们将在 Task 9 移入 core（workflow 根包）作为 `NodeExecutor` 实现。本 Task 先删除，Step 8 会临时在旧 agent 包补桩，保证编译。

- [ ] **Step 4: 全文件类型加前缀**

`executor.go`、`context.go` 内所有：`*TopoNode`→`*types.TopoNode`、`*Topology`→`*types.Topology`、`ParallelConfig`→`types.ParallelConfig`、`JoinConfig`→`types.JoinConfig`、`ParallelResult`→`types.ParallelResult`、`JoinResult`→`types.JoinResult`、`RetryConfig`→`types.RetryConfig`、`NodeExecutionLog`→`types.NodeExecutionLog`。表达式求值 `NewExpressionEvaluator()`（789 行）→ `condition.NewExpressionEvaluator()`。`sendExecutionLog` → `types.SendExecutionLog`。`context_test.go` 同步改 `package parallel` 及类型前缀。

- [ ] **Step 5: 拆分 group.go 与 vars.go**

从 `executor.go` 剪切以下函数到新文件（同 `package parallel`）：
- `group.go`：`IdentifyParallelGroups`、`extractParallelGroup`、`traceBranchPath`、`findBranchesByParallelId`、`getNodeParallelId`、`TopologicalSort`、`ValidateParallelGroup`、`BuildParallelGraph`。
- `vars.go`：`replaceVarsInParams`、`replaceVarValue`、`replaceNestedVars`、`toString`、`escapeJSON`。

新文件顶部按需 import（`types`、`regexp`、`strings`、`sort`、`fmt`、`encoding/json`）。

- [ ] **Step 6: core 改用 parallel 包**

在 `workflow_agent.go`：
- 增加 import `"txing-ai/internal/agent/workflow/parallel"`。
- `NewParallelExecutor(a, 10, "", "", "")`（429 行）→ `parallel.NewParallelExecutor(a, 10, "", "", "")`（`a` 是 `*WorkflowAgent`，实现了 NodeExecutor 接口，可直接传）。
- `NewParallelExecutor(a, parallelConfig.MaxConcurrency, ...)`（1172 行）→ `parallel.NewParallelExecutor(a, ...)`。
- `NewParallelExecutor(nil, 0, "", "", "")`（1442 行 buildParallelJoinEdges 内）→ `parallel.NewParallelExecutor(nil, 0, "", "", "")`。
- 变量类型 `parallelExecutor`/`parallelExecutorForScan` 现为 `*parallel.ParallelExecutor`，方法调用 `.IdentifyParallelGroups/.ExecuteParallelGroup/.MergeResults` 不变。

- [ ] **Step 7: 追加 parallel 桥接（供旧包内其它引用）**

在 `types_bridge.go` 追加：

```go
import "txing-ai/internal/agent/workflow/parallel"

type (
	ParallelExecutor = parallel.ParallelExecutor
	ParallelContext  = parallel.ParallelContext
	ParallelGroup    = parallel.ParallelGroup
)

var (
	NewParallelExecutor = parallel.NewParallelExecutor
	NewParallelContext  = parallel.NewParallelContext
)
```

- [ ] **Step 8: 补 core 侧 NodeExecutor 实现桩（临时）**

因 Step 3 删除了两方法、而 Task 9 才正式迁移，先在旧 `workflow_agent.go` 保留 `ExecuteLLMNodeInParallel`/`ExecuteToolNodeInParallel` 的**完整实现**（原 executor.go 516-674、718-775 行的函数体整体搬到 workflow_agent.go，方法接收者仍是 `*WorkflowAgent`），并把其中 `*TopoNode`→`*types.TopoNode` 等类型加前缀。这样 `a *WorkflowAgent` 满足 `parallel.NodeExecutor` 接口。

> 该实现在 Task 9 会随 core 一起进入 workflow 根包，无需再动。

- [ ] **Step 9: 构建验证**

Run: `go build ./...`
Expected: 通过。此时 core↔parallel 循环已由接口打破：parallel 只依赖 types+condition，core 依赖 parallel 并实现其接口。

- [ ] **Step 10: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS（context_test.go 属 parallel 包）。

- [ ] **Step 11: Commit**

```bash
git add -A internal/agent/workflow/parallel internal/agent/workflow_agent.go internal/agent/types_bridge.go
git commit -m "refactor(agent): 迁移并行执行器到 workflow/parallel 子包并以 NodeExecutor 接口打破循环"
```

---

## Task 7: 迁移 agent 子包（Agent 抽象与实现）

**Files:**
- Create: `internal/agent/agent/base.go`（from `base_agent.go`）
- Create: `internal/agent/agent/factory.go`（from `factory.go`）
- Create: `internal/agent/agent/state.go`（from `state.go`）
- Create: `internal/agent/agent/errors.go`（from `errors.go`）
- Create: `internal/agent/agent/general.go`（from `general_agent.go`）
- Create: `internal/agent/agent/resume.go`（from `resume_agent.go`）
- Create: `internal/agent/agent/travel.go`（from `travel_agent.go`）
- Create: `internal/agent/agent/toolcall.go`（from `toolcall_agent.go`）
- Delete: 上述 8 个旧文件

**Interfaces:**
- Consumes: 无 workflow 依赖（这些文件仅用 iface/tool/eino，Task 前置检查已确认独立）
- Produces:
  - `agent.Agent`（接口）, `agent.BaseAgent`, `agent.NewBaseAgent(name, description string) *BaseAgent`
  - `agent.AgentType`, `agent.AgentFactory`, `agent.SimpleAgentFactory`, `agent.NewSimpleAgentFactory(res iface.ResourceProvider) AgentFactory`
  - `agent.AgentState`, `agent.GeneralAgent`, `agent.ResumeAgent`, `agent.TravelAgent`, `agent.ToolCallAgent`, `agent.AgentExecConfig`
  - 错误变量 `agent.ErrNoHandler` 等

> **注意**：新子包目录名 `agent`，包名保持 `agent`。它与父目录 `internal/agent`（旧包，最终将只剩 engine 相关文件并改名，见 Task 9）路径不同：`txing-ai/internal/agent/agent`。

- [ ] **Step 1: git mv 8 个文件**

```bash
git mv internal/agent/base_agent.go     internal/agent/agent/base.go
git mv internal/agent/factory.go        internal/agent/agent/factory.go
git mv internal/agent/state.go          internal/agent/agent/state.go
git mv internal/agent/errors.go         internal/agent/agent/errors.go
git mv internal/agent/general_agent.go  internal/agent/agent/general.go
git mv internal/agent/resume_agent.go   internal/agent/agent/resume.go
git mv internal/agent/travel_agent.go   internal/agent/agent/travel.go
git mv internal/agent/toolcall_agent.go internal/agent/agent/toolcall.go
```

- [ ] **Step 2: 包名保持 agent**

这些文件顶部已是 `package agent`，目录变了但包名不变，无需改包名声明。确认 8 文件均 `package agent`。

- [ ] **Step 3: 检查 WorkflowAgent 引用**

`WorkflowAgent`（在旧父包）内嵌 `*BaseAgent`、调用 `NewBaseAgent`。现 `BaseAgent` 移入子包 `agent`。core 需在 Task 9 import `agentpkg "txing-ai/internal/agent/agent"` 并改 `*BaseAgent`→`*agentpkg.BaseAgent`、`NewBaseAgent`→`agentpkg.NewBaseAgent`。**本 Task 先在 types_bridge.go 加桥接**，让父包 workflow_agent.go 暂时仍能用 `BaseAgent`/`NewBaseAgent`：

```go
import agentpkg "txing-ai/internal/agent/agent"

type BaseAgent = agentpkg.BaseAgent
var NewBaseAgent = agentpkg.NewBaseAgent
```

（`Agent` 接口、`WorkflowAgentState` 等按 workflow_agent.go 实际引用补齐别名。）

- [ ] **Step 4: 构建验证**

Run: `go build ./...`
Expected: 通过。外部 6 调用点仍引用旧 `agent` 包的桥接别名（`agent.NewSimpleAgentFactory` 等），编译无碍。

- [ ] **Step 5: 测试验证**

Run: `go test ./internal/agent/...`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add -A internal/agent/agent internal/agent/types_bridge.go
git commit -m "refactor(agent): 迁移 Agent 抽象与实现到 agent 子包"
```

---

## Task 8: 迁移 workflow 核心到 workflow 根包（engine.go）

**Files:**
- Create: `internal/agent/workflow/engine.go`（WorkflowAgent 主体 + BuildGraph + ExecuteStream）
- Create: `internal/agent/workflow/retry.go`（retry 辅助 + nodeStatusCallback）
- Create: `internal/agent/workflow/condition_exec.go`（executeLLMCondition/executeToolResultCondition）
- Create: `internal/agent/workflow/join_edges.go`（buildParallelJoinEdges）
- Create: `internal/agent/workflow/node_exec.go`（ExecuteLLMNodeInParallel/ExecuteToolNodeInParallel —— NodeExecutor 实现）
- Delete: `internal/agent/workflow_agent.go`
- Modify: `internal/agent/types_bridge.go`（删除 core 相关别名，仅保留外部调用点 Task 9 需要的过渡别名——或本 Task 直接清空，见步骤）

**Interfaces:**
- Consumes: `types.*`, `condition.*`, `nodeexec.ExecuteCodeNode/ExecuteHTTPNode/ExecuteSubWorkflowNode`, `parallel.NewParallelExecutor/NodeExecutor`, `validator.*`（若 core 内部用到）, `agentpkg.BaseAgent/NewBaseAgent`
- Produces:
  - `workflow.WorkflowAgent`（实现 `parallel.NodeExecutor`）
  - `workflow.NewWorkflowAgent(res iface.ResourceProvider, topology string, modelResolver types.ModelResolver) *WorkflowAgent`
  - `(*WorkflowAgent).BuildGraph(...)`, `.ExecuteStream(...)`
  - `(*WorkflowAgent).ExecuteLLMNodeInParallel(...)`, `.ExecuteToolNodeInParallel(...)`

- [ ] **Step 1: git mv 主体，按职责拆文件**

```bash
git mv internal/agent/workflow_agent.go internal/agent/workflow/engine.go
```

然后从 `engine.go` 剪切到新文件（均 `package workflow`）：
- `retry.go`：`calculateRetryDelay`、`getMaxRetries`、`executeWithRetry`、`nodeStatusCallback`。
- `condition_exec.go`：`executeLLMCondition`、`executeToolResultCondition`。
- `join_edges.go`：`buildParallelJoinEdges`。
- `node_exec.go`：`ExecuteLLMNodeInParallel`、`ExecuteToolNodeInParallel`（Task 6 Step 8 已搬进来的实现）。

- [ ] **Step 2: 改包名**

所有上述文件 `package agent` → `package workflow`。

- [ ] **Step 3: 加各子包 import 与前缀**

在需要的文件顶部 import 并替换引用：
- `"txing-ai/internal/agent/workflow/types"`：所有 `TopoNode/Topology/NodeData/ModelInfo/ModelResolver/*Config/NodeExecutionLog/ParallelConfig/JoinConfig/ParallelResult/JoinResult` 加 `types.` 前缀。`sendExecutionLog(...)` → `types.SendExecutionLog(...)`。
- `"txing-ai/internal/agent/workflow/condition"`：`NewExpressionEvaluator/ConditionResult/ConditionConfigV2/DefaultConditionConfig/NewConditionResult/NewConditionError/LLMJudgmentResponse` 加 `condition.` 前缀。
- `nodeexec "txing-ai/internal/agent/workflow/node"`：`node.ExecuteCodeNode` 等（Task 4 已改为别名 `nodeexec`，此处沿用）。
- `"txing-ai/internal/agent/workflow/parallel"`：`parallel.NewParallelExecutor`。
- `agentpkg "txing-ai/internal/agent/agent"`：`*BaseAgent`→`*agentpkg.BaseAgent`，`NewBaseAgent`→`agentpkg.NewBaseAgent`。
- `NewWorkflowAgent` 第三参数类型 `modelResolver ModelResolver` → `types.ModelResolver`。

- [ ] **Step 4: 删除桥接文件**

删除 `internal/agent/types_bridge.go`（其使命完成——core 已迁出旧包，旧父目录 `internal/agent` 下不再有 `.go` 文件）。

```bash
git rm internal/agent/types_bridge.go
```

确认 `internal/agent/*.go`（父目录直属）已无剩余文件：`ls internal/agent/*.go 2>/dev/null` 应为空。

- [ ] **Step 5: 构建验证（预期外部调用点报错）**

Run: `go build ./internal/agent/...`
Expected: `internal/agent/...` 各子包编译通过。
Run: `go build ./...`
Expected: **失败**——6 个外部调用点仍写 `agent.NewWorkflowAgent` 等，旧包已空。这是预期，Task 9 修复。

- [ ] **Step 6: 测试验证（子包）**

Run: `go test ./internal/agent/...`
Expected: PASS（各子包内测试）。

- [ ] **Step 7: Commit**

```bash
git add -A internal/agent
git commit -m "refactor(agent): 迁移工作流核心到 workflow 根包并删除桥接层"
```

---

## Task 9: 改造 6 个外部调用点

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/middleware/middleware.go`
- Modify: `internal/middleware/builtins.go`
- Modify: `internal/controller/agent/controller.go`
- Modify: `internal/controller/workflow/controller.go`
- Modify: `internal/service/chat/workflow_chat.go`

**Interfaces:**
- Consumes: `agent.*`（新 agent 子包）, `workflow.*`, `types.*`, `validator.*`, `resolver.*`
- Produces: 无（调用方，末端）

- [ ] **Step 1: app.go**

- import `"txing-ai/internal/agent"` → `agent "txing-ai/internal/agent/agent"`（包名仍 `agent`，路径改子包）。
- `agent.NewSimpleAgentFactory` 保持写法（符号名不变）。

- [ ] **Step 2: middleware.go 与 builtins.go**

两文件均：import 改为 `agent "txing-ai/internal/agent/agent"`；`agent.AgentFactory` 写法不变。

- [ ] **Step 3: controller/agent/controller.go**

- import 改为 `agent "txing-ai/internal/agent/agent"`。
- `agent.AgentFactory`、`agent.AgentType`、`agent.Execute`、`agent.ExecuteStream` 符号不变（均属新 agent 子包）。

- [ ] **Step 4: controller/workflow/controller.go**

替换 import 与符号：
- 增加 import：
  ```go
  "txing-ai/internal/agent/workflow"
  "txing-ai/internal/agent/workflow/types"
  "txing-ai/internal/agent/workflow/validator"
  "txing-ai/internal/agent/workflow/resolver"
  ```
  删除旧 `"txing-ai/internal/agent"`。
- 符号替换：`agent.NewChannelModelResolver`→`resolver.NewChannelModelResolver`；`agent.NewWorkflowAgent`→`workflow.NewWorkflowAgent`；`agent.Topology`→`types.Topology`；`agent.ValidateTopology`→`validator.ValidateTopology`；`agent.ValidateTopologyWithLLM`→`validator.ValidateTopologyWithLLM`；`agent.ValidationError`→`validator.ValidationError`；`agent.ValidationResult`→`validator.ValidationResult`。

- [ ] **Step 5: service/chat/workflow_chat.go**

- import 增加 `"txing-ai/internal/agent/workflow"`、`"txing-ai/internal/agent/workflow/types"`、`"txing-ai/internal/agent/workflow/resolver"`，删除旧 `"txing-ai/internal/agent"`。
- 符号替换：`agent.NewChannelModelResolver`→`resolver.NewChannelModelResolver`；`agent.NewWorkflowAgent`→`workflow.NewWorkflowAgent`；`agent.Topology`→`types.Topology`；`agent.WorkflowConfig`→`types.WorkflowConfig`。

- [ ] **Step 6: 全量构建验证**

Run: `go build ./...`
Expected: 通过。

- [ ] **Step 7: 全量测试验证**

Run: `go test ./...`
Expected: PASS（或与重构前相同的既有失败/跳过集合——不得新增失败）。

- [ ] **Step 8: 冒烟验证（可选但推荐）**

Run: `go vet ./internal/agent/... ./internal/controller/... ./internal/service/...`
Expected: 无新增告警。

- [ ] **Step 9: Commit**

```bash
git add internal/app internal/middleware internal/controller internal/service
git commit -m "refactor(agent): 改造外部调用点适配新包结构"
```

---

## Task 10: 收尾验证与文档

**Files:**
- Modify: `CLAUDE.md` 与 `.claude/CLAUDE.md`（更新「Agent System」段落的目录说明，可选）

- [ ] **Step 1: 确认旧扁平文件已全部清除**

Run: `ls internal/agent/*.go 2>/dev/null; echo "---"; find internal/agent -name '*.go' | sort`
Expected: 父目录 `internal/agent/*.go` 为空；文件全部落在 `agent/`、`workflow/`、`workflow/{types,condition,node,validator,parallel,resolver}` 下。

- [ ] **Step 2: 确认无循环依赖**

Run: `go build ./... && echo "no import cycle"`
Expected: 通过（Go 编译器会在有循环依赖时报错）。

- [ ] **Step 3: 全量测试**

Run: `go test ./...`
Expected: PASS。

- [ ] **Step 4: 更新架构文档（可选）**

若更新 CLAUDE.md 的「Agent System」段，把目录结构改为新的子包布局。

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "docs: 更新 Agent System 目录说明为子包结构"
```

---

## Self-Review 记录

- **Spec coverage**：spec 的目标包结构（agent + workflow/{types,condition,node,validator,parallel,resolver} + engine）→ Task 1-9 逐一覆盖；接口打破循环 → Task 6；共享类型/execlog 下沉 → Task 1；外部调用点改造 → Task 9。✅
- **Placeholder scan**：无 TBD/TODO；跨包导出清单表格给出确切新旧符号名；每个 import 别名（`nodeexec`、`agentpkg`）已显式定义并全程一致。✅
- **Type consistency**：`NodeExecutor` 接口签名（Task 6）与 core 实现（Task 8 node_exec.go）一致；`SendExecutionLog`/`ExecuteCodeNode` 等导出名在清单与各 Task 一致；`types.ModelResolver` 在 resolver（Task 3）、core（Task 8）引用一致。✅
- **循环依赖处理**：桥接文件 `types_bridge.go` 使每步旧包可编译，Task 8 删除；变量名 `node` 与 `node` 包冲突用 `nodeexec` 别名规避；子包目录 `agent` 与父目录同名用路径区分 + `agentpkg` 别名。✅
