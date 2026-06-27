# SPEC: Agent 工作流条件节点完善

## 1. Objective

完善 txing-ai 项目中的 Agent 工作流拖拽功能，重点实现条件节点的完整功能。

**目标用户**: 使用 txing-ai 平台构建 AI 工作流的开发者和业务人员

**核心问题**: 当前条件节点（ConditionNode）只是简单透传，无法根据条件选择不同的执行分支，导致工作流无法实现真正的条件路由。

**验收标准**:
1. 条件节点能够根据配置的条件类型（表达式/AI判断/工具结果）执行判断逻辑
2. 支持基于 Handle ID 的分支路由（`true`/`false`）
3. 前端配置界面完整支持所有条件类型的配置
4. 条件判断出错时支持可配置的错误处理策略

---

## 2. Commands

### 构建命令
```bash
# 完整构建（包含 Swagger 文档生成）
make all

# 快速构建（跳过代码生成，仅用于调试）
make build

# 生成 Swagger API 文档
make gen

# 构建前端
make frontend
```

### 运行命令
```bash
# 开发模式运行
go run cmd/main.go

# 指定端口运行
go run cmd/main.go -port 8081
```

### 测试命令
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test -v ./internal/agent/...

# 运行前端开发服务器
cd static/frontend && pnpm dev
```

---

## 3. Project Structure

### 涉及修改的文件

#### 后端
```
internal/agent/workflow_agent.go      # 主要修改：条件节点执行逻辑
internal/agent/expression_eval.go     # 新增：表达式解析引擎
internal/agent/condition_types.go     # 新增：条件判断相关类型定义
internal/domain/workflow.go           # 可能修改：扩展条件配置结构
```

#### 前端
```
static/frontend/src/components/workflow/ConditionNode.vue    # UI 展示优化
static/frontend/src/components/workflow/PropertyPanel.vue    # 扩展条件配置项
static/frontend/src/views/admin/workflow/WorkflowEditor.vue # 连线逻辑适配
```

### 关键数据结构

```go
// ConditionConfig 条件配置
type ConditionConfig struct {
    Type           string `json:"type"`           // expression | llm | tool_result
    Expression     string `json:"expression,omitempty"`
    LLMPrompt      string `json:"llmPrompt,omitempty"`
    ToolName       string `json:"toolName,omitempty"`
    ToolResultKey  string `json:"toolResultKey,omitempty"`
    TrueHandle     string `json:"trueHandle,omitempty"`     // 默认 "true"
    FalseHandle    string `json:"falseHandle,omitempty"`    // 默认 "false"
    FailureAction  string `json:"failureAction,omitempty"`  // default_false | terminate | configurable
    FailureBranch  string `json:"failureBranch,omitempty"`  // 当 failureAction 为 configurable 时使用
}
```

---

## 4. Code Style

### Go 后端
- 使用 idiomatic Go 风格
- 错误处理必须显式，不要忽略错误
- 使用 `log` 包记录关键操作和错误
- 条件判断逻辑抽取为独立函数，便于测试
- 新增类型定义放在独立文件中

### Vue 前端
- 使用 Composition API + `<script setup>`
- Element Plus UI 组件
- 保持与现有组件风格一致
- 使用 computed 处理派生状态

### 命名约定
- Go: camelCase（私有）、PascalCase（公开）
- Vue: camelCase（变量/函数）、kebab-case（模板）

---

## 5. Testing Strategy

### 单元测试
```go
// internal/agent/expression_eval_test.go
func TestExpressionEvaluator(t *testing.T) {
    tests := []struct {
        expr     string
        input    string
        expected bool
    }{
        {`{{output}} contains "成功"`, "操作成功", true},
        {`{{output}} equals "hello"`, "hello", true},
        {`{{output}} starts_with "prefix"`, "prefix_value", true},
        {`{{output}} matches "^\\d+$"`, "12345", true},
    }
    // ...
}
```

### 集成测试
- 构建包含条件节点的工作流
- 验证 true/false 分支正确路由
- 测试错误处理策略

### 前端测试
- 手动测试条件节点配置界面
- 验证连线到不同分支的视觉效果
- 测试完整工作流执行流程

---

## 6. Boundaries

### 必须做 (Always)
- 保持与现有工作流系统的兼容性
- 所有 API 变更必须添加 Swagger 注释
- 新增配置字段必须向后兼容（使用 omitempty）
- 错误必须记录到日志

### 先询问 (Ask First)
- 修改现有数据结构（ConditionConfig）
- 新增外部依赖（表达式引擎库）
- 修改前端组件的 props 接口

### 永不做 (Never)
- 不破坏现有工作流的数据格式
- 不引入破坏性变更（breaking changes）
- 不在 hot path 中使用反射
- 不在未授权情况下执行用户提供的代码/脚本

---

## 7. Implementation Details

### 7.1 表达式判断实现

支持以下运算符：
| 运算符 | 语法 | 示例 |
|--------|------|------|
| contains | `{{output}} contains "keyword"` | 判断是否包含子串 |
| equals | `{{output}} equals "value"` | 精确匹配 |
| starts_with | `{{output}} starts_with "prefix"` | 前缀匹配 |
| ends_with | `{{output}} ends_with "suffix"` | 后缀匹配 |
| matches | `{{output}} matches "^regex$"` | 正则匹配 |
| greater_than | `{{output}} greater_than 10` | 数值大于 |
| less_than | `{{output}} less_than 10` | 数值小于 |

### 7.2 AI判断实现

LLM 返回结构化 JSON：
```json
{
  "result": true,
  "reason": "内容包含积极情绪关键词"
}
```

### 7.3 分支路由策略

基于 Handle ID：
- 边的 `sourceHandle` 字段标识分支：`"true"` 或 `"false"`
- 条件节点执行时，根据判断结果选择匹配的边
- 查找 `edges` 中 `source == nodeId && sourceHandle == result` 的目标节点

### 7.4 错误处理策略

| 策略 | 说明 |
|------|------|
| `default_false` | 出错时自动走 false 分支（默认） |
| `terminate` | 出错时终止工作流，返回错误 |
| `configurable` | 出错时走用户配置的默认分支 |

---

## 8. Implementation Tasks

### Phase 1: 后端条件判断逻辑
1. [ ] 创建 `expression_eval.go` - 表达式解析引擎
2. [ ] 创建 `condition_types.go` - 条件相关类型定义
3. [ ] 修改 `workflow_agent.go` - 实现条件节点分支路由
4. [ ] 添加单元测试

### Phase 2: 前端配置完善
5. [ ] 扩展 `PropertyPanel.vue` - 添加错误处理策略配置
6. [ ] 优化 `ConditionNode.vue` - 显示配置状态
7. [ ] 更新 `WorkflowEditor.vue` - 连线逻辑适配 Handle ID

### Phase 3: 集成测试
8. [ ] 构建测试工作流
9. [ ] 端到端测试
10. [ ] 更新文档

---

## 9. References

- n8n Condition Node: https://docs.n8n.io/integrations/builtin/core-nodes/n8n-base-node/condition/
- Flowise Condition Node: https://docs.flowiseai.com/components/logic-nodes
- Vue Flow Handle: https://vueflow.dev/guide/node.html#handles
- CloudWeGo Eino Graph: https://github.com/cloudwego/eino
