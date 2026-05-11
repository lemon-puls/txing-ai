# Tasks
- [x] Task 1: Setup Frontend Workflow Dependencies
  - [x] SubTask 1.1: Install Vue Flow (`@vue-flow/core`, `@vue-flow/background`, `@vue-flow/controls`, `@vue-flow/node-resizer`)
- [x] Task 2: Backend Database and API Implementation
  - [x] SubTask 2.1: Create `AgentFlow` model in `internal/domain` or `internal/model`
  - [x] SubTask 2.2: Implement Workflow CRUD service in `internal/service/workflow`
  - [x] SubTask 2.3: Implement Workflow Controller and Routes in `internal/controller/workflow`
- [x] Task 3: Frontend Workflow Editor Implementation
  - [x] SubTask 3.1: Create Workflow List View (`WorkflowList.vue`)
  - [x] SubTask 3.2: Create Workflow Visual Editor View using Vue Flow (`WorkflowEditor.vue`) with custom nodes (Agent, Tool, Condition)
  - [x] SubTask 3.3: Add Workflow Editor to Admin Menu and Router
- [x] Task 4: Backend Dynamic Eino Graph Execution
  - [x] SubTask 4.1: Implement dynamic graph builder parsing JSON to `compose.Graph` in `internal/agent/workflow_agent.go`
  - [x] SubTask 4.2: Implement execution endpoint handling user input and streaming/returning results

# Task Dependencies
- [Task 2] depends on None
- [Task 3] depends on [Task 1, Task 2]
- [Task 4] depends on [Task 2]