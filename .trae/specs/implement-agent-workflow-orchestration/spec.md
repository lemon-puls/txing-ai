# Agent Workflow Orchestration Spec

## Why
随着大模型能力的提升，单一 Agent 难以处理复杂的综合性任务。为了实现更高级的业务场景（如自动化开发、复杂数据分析、多角色协作等），我们需要引入多智能体协作（Multi-Agent Collaboration）功能，并提供直观的拖拽式工作流可视化编排工具，使用户能够轻松定义 AI 角色及其交互流程。

## What Changes
- 引入前端可视化编排库（Vue Flow），实现拖拽式的工作流图编辑界面。
- 新增 `AgentFlow` 数据库模型及对应的 CRUD 接口，用于保存和管理用户编排的工作流拓扑结构（JSON格式）。
- 后端基于项目中已集成的字节跳动 `Eino` 框架，实现根据 JSON 拓扑动态构建和执行 `compose.Graph` 的能力。
- 创新性结合点：支持主流 Multi-Agent 编排模式，例如“规划-执行（Plan-and-Execute）”或“主管-员工（Supervisor-Worker）”模式的节点组件化，支持大模型节点、工具节点以及条件分支。

## Impact
- Affected specs: 现有的系统将扩展支持复杂编排的工作流执行能力。
- Affected code:
  - Frontend: `static/frontend/src/views/admin/workflow/*` (新增), `static/frontend/package.json`, 路由及菜单配置。
  - Backend: `internal/domain/workflow.go`, `internal/controller/workflow/*`, `internal/service/workflow/*`, `internal/agent/workflow_agent.go`

## ADDED Requirements
### Requirement: Workflow Visual Editor
The system SHALL provide a drag-and-drop visual editor for multi-agent workflows.

#### Scenario: User creates a new multi-agent workflow
- **WHEN** user opens the workflow editor and drags different agent/tool nodes onto the canvas
- **THEN** the system renders the nodes, allows connecting them with edges to define the execution flow, and saves the topology graph as JSON data.

### Requirement: Dynamic Multi-Agent Execution
The system SHALL support executing a saved workflow graph dynamically.

#### Scenario: User triggers a deployed workflow
- **WHEN** user inputs a task to a deployed workflow
- **THEN** the backend dynamically constructs an Eino `compose.Graph` from the saved JSON definition, executes the multi-agent nodes collaboratively, and returns the final result.
