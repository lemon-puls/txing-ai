# internal/agent 包结构重构设计

日期：2026-07-05
状态：设计已确认，待评审

## 背景与目标

`internal/agent/` 当前是一个扁平的单包，含 21 个文件、约 6164 行代码，把多个不同职责混在一起：

- Agent 抽象与实现（general/resume/travel/toolcall）
- 工作流引擎核心（`workflow_agent.go`，1668 行）
- 并行执行（`workflow_parallel.go`，1306 行）
- 节点执行器（code/http/subworkflow）
- 条件与表达式求值
- 拓扑校验（结构 + LLM）
- 模型解析

**目标**：拆分为多个职责清晰、可独立测试的子包，通过接口打破 core↔parallel 循环依赖，提升可读性。外部调用点直接改造，不保留转发层。

## 依赖关系分析（重构可行性的核心约束）

### 循环耦合

```
WorkflowAgent (workflow_agent.go)
  ├─ 创建 → NewParallelExecutor(a, ...)                        // core → parallel
  └─ 定义 → func (e *WorkflowAgent) ExecuteLLMNodeInParallel() // 方法却在 parallel 文件里

ParallelExecutor (workflow_parallel.go)
  └─ 持有 e.agent *WorkflowAgent，回调 e.agent.ExecuteLLMNodeInParallel()  // parallel → core
```

`ParallelExecutor` 从 `WorkflowAgent` 只需要 **两个方法**：

- `ExecuteLLMNodeInParallel(ctx, node, input, callback) (string, error)`
- `ExecuteToolNodeInParallel(ctx, node, input, callback) (string, error)`

core 反向需要 parallel 的 `ExecuteParallelGroup / IdentifyParallelGroups / MergeResults`。

### 打破循环的方案：接口注入

在 `parallel` 包定义窄接口：

```go
// workflow/parallel/executor.go
type NodeExecutor interface {
    ExecuteLLMNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(*global.Chunk) error) (string, error)
    ExecuteToolNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(*global.Chunk) error) (string, error)
}
```

`ParallelExecutor` 持有 `NodeExecutor` 接口而非具体 `*WorkflowAgent`。`WorkflowAgent`（workflow 根包）实现该接口。依赖方向变为单向：`workflow → parallel → types`。

### 共享类型/辅助下沉

被多个子包引用、必须下沉到底层 `types` 包的项：

- 数据类型：`Topology`、`TopoNode`、`TopoEdge`、`Position`、`NodeData`、`WorkflowConfig` 及各 `*Config`（Model/Tool/Condition/Code/HTTP/SubWorkflow/Agent/Retry）、`SchemaField`
- 模型类型：`ModelInfo`、`ModelResolver` 接口
- 结果类型：`ParallelResult`、`JoinResult`、`NodeExecutionLog`
- 执行日志辅助：`sendExecutionLog`（被 node 包的 code/http/subworkflow 引用）

只在 core 使用、随 core 留在 `workflow/` 根包的项：

- retry 辅助：`executeWithRetry`、`calculateRetryDelay`、`getMaxRetries`
- 回调辅助：`nodeStatusCallback`
- 条件执行：`executeLLMCondition`、`executeToolResultCondition`
- `buildParallelJoinEdges`

### agents 独立性

`general/resume/travel/toolcall/base/factory/state/errors` 完全不引用任何 workflow 类型，可干净独立成 `agent/` 子包。

## 目标包结构

依赖方向：下层不依赖上层。

```
internal/agent/
│
├── agent/                    # AI Agent 抽象与实现（独立于 workflow 引擎）
│   ├── base.go               # Agent 接口 + BaseAgent
│   ├── factory.go            # AgentFactory / SimpleAgentFactory
│   ├── state.go              # AgentState
│   ├── errors.go
│   ├── general.go            # GeneralAgent
│   ├── resume.go             # ResumeAgent
│   ├── travel.go             # TravelAgent
│   └── toolcall.go           # ToolCallAgent
│
└── workflow/                 # 工作流引擎
    ├── types/                # 底层：共享数据类型（无同域依赖）
    │   ├── topology.go       # Topology, TopoNode, TopoEdge, Position, NodeData
    │   ├── config.go         # WorkflowConfig, ModelConfig, ToolConfig, RetryConfig, ...
    │   ├── model.go          # ModelInfo, ModelResolver 接口
    │   ├── result.go         # ParallelResult, JoinResult
    │   └── execlog.go        # NodeExecutionLog + sendExecutionLog
    │
    ├── condition/            # 条件与表达式求值（依赖 types）
    │   ├── condition.go      # ConditionType, ConditionConfigV2, ConditionResult
    │   ├── expression.go     # ExpressionEvaluator
    │   └── expression_test.go
    │
    ├── node/                 # 节点执行器（依赖 types）
    │   ├── code.go           # executeCodeNode + JS/Python 运行时
    │   ├── http.go           # executeHTTPNode
    │   └── subworkflow.go    # executeSubWorkflowNode + SubWorkflowExecutor 接口
    │
    ├── validator/            # 拓扑校验（依赖 types）
    │   ├── structural.go     # ValidateTopology, detectCycle, validate*Node
    │   └── llm.go            # ValidateTopologyWithLLM, buildTopologySummary
    │
    ├── parallel/             # 并行执行器（依赖 types + condition，接口回调 core）
    │   ├── executor.go       # ParallelExecutor + NodeExecutor 接口（打破循环）
    │   ├── context.go        # ParallelContext
    │   ├── context_test.go
    │   ├── group.go          # 并行组识别 / 分支追踪 / 拓扑排序
    │   └── vars.go           # replaceVars* / toString / escapeJSON 工具
    │
    ├── resolver/             # 模型解析（依赖 types）
    │   └── channel.go        # ChannelModelResolver
    │
    └── engine.go             # WorkflowAgent 核心（实现 parallel.NodeExecutor）
                              # + retry 辅助 + condition 执行 + BuildGraph/ExecuteStream
```

> 注：`engine.go` 可按需再拆为 `engine.go`（WorkflowAgent + BuildGraph/ExecuteStream）、`retry.go`（retry 辅助）、`condition_exec.go`（executeLLMCondition/executeToolResultCondition）、`join_edges.go`（buildParallelJoinEdges）等多个文件，均在 workflow 根包内，不影响外部。

### 包命名注意

`ExecuteLLMNodeInParallel / ExecuteToolNodeInParallel` 这两个方法当前定义在 parallel 文件但挂在 `WorkflowAgent` 上，重构后随 core 移入 `workflow/` 根包（作为 `NodeExecutor` 接口的实现）。

## 对外调用点改造（6 处，直接改）

| 文件 | 原符号 | 新引用 |
|------|--------|--------|
| `internal/app/app.go` | `agent.NewSimpleAgentFactory` | `agent.NewSimpleAgentFactory`（新 agent 子包，import 路径变 `.../agent/agent`）|
| `internal/middleware/middleware.go` | `agent.AgentFactory` | 同上 |
| `internal/middleware/builtins.go` | `agent.AgentFactory` | 同上 |
| `internal/controller/agent/controller.go` | `agent.{AgentFactory,AgentType,Execute,ExecuteStream}` | agent 子包 |
| `internal/controller/workflow/controller.go` | `agent.{NewChannelModelResolver,NewWorkflowAgent,Topology,ValidateTopology,ValidateTopologyWithLLM,ValidationError,ValidationResult}` | 分别指向 `resolver`、`workflow`、`types`、`validator` |
| `internal/service/chat/workflow_chat.go` | `agent.{NewChannelModelResolver,NewWorkflowAgent,Topology,WorkflowConfig}` | `resolver`、`workflow`、`types` |

> import 别名建议：为避免 `agent/agent` 双名歧义，agent 子包 import 时用别名 `agentpkg` 或保持默认 `agent`（包名仍为 `agent`）。workflow 相关子包按目录名 `workflow` / `types` / `validator` / `resolver` 引入。

## 迁移策略

采用**分层自底向上、每层可编译**的方式，降低单步风险：

1. `types` 子包（无依赖，先建）
2. `condition`、`resolver`、`node`、`validator`（只依赖 types，可并行迁移）
3. `parallel`（依赖 types + condition + 定义 NodeExecutor 接口）
4. `agent` 子包（完全独立，可任意时机迁移）
5. `workflow` 根包 `engine.go`（依赖以上所有，实现 NodeExecutor）
6. 改造 6 个外部调用点的 import

每完成一步执行 `go build ./...` 与 `go test ./internal/agent/...` 验证。

## 测试与验证

- 现有测试：`expression_eval_test.go` → `condition/expression_test.go`；`parallel_context_test.go` → `parallel/context_test.go`。测试逻辑不改，仅调整包名与 import。
- 全程以 `go build ./...`、`go test ./...` 作为每步验收标准。
- 本次为**纯结构重构**：不改变任何运行时行为、函数签名语义（仅包路径/接口注入方式变化），不新增功能。

## 非目标（YAGNI）

- 不重写任何业务逻辑
- 不优化算法或性能
- 不新增 agent 类型或 node 类型
- 不动前端代码
- 不保留旧 `agent` 包的兼容转发层
