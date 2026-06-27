# 工作流测试指南

本文档说明如何测试 txing-ai 的工作流功能。

## 前置条件

1. **启动 MySQL 数据库**
   ```bash
   # 确保MySQL服务运行
   # 创建数据库（如果不存在）
   mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS txing_ai CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
   ```

2. **启动 Redis 服务**
   ```bash
   # 确保Redis服务运行
   redis-cli ping  # 应返回 PONG
   ```

3. **配置应用**
   ```bash
   # 复制配置文件模板
   cp config.yaml.sample runtime/config.yaml

   # 编辑配置文件，修改数据库和Redis连接信息
   # 确保 MySQL 用户名、密码、数据库名正确
   ```

## 步骤 1: 启动后端服务

```bash
# 进入项目目录
cd D:\develop_project\txing-ai

# 启动服务（会自动创建数据库表）
go run cmd/main.go
```

启动成功后会显示：
```
🚀 Txing AI service started successfully!
📚 API Documentation: http://localhost:8080/swagger/index.html
🏠 Homepage: http://localhost:8080/dash
```

## 步骤 2: 创建测试用户

由于 API 需要认证，需要先创建测试用户。

### 方式 1: 通过前端页面注册

1. 访问 http://localhost:8080/dash
2. 点击"注册"按钮
3. 填写用户信息完成注册

### 方式 2: 直接插入数据库

```sql
-- 连接数据库
mysql -u root -p txing_ai

-- 插入测试用户（密码为 admin123）
INSERT INTO `users` (`username`, `password`, `role`, `create_time`, `update_time`)
VALUES ('admin', '$2a$10$rR8FJ3vXW4K9Y8eH3mPqZ.d6S5.iYI3yPMhIAX6oI3H5qZhPV3Fim', 0, NOW(), NOW());
```

## 步骤 3: 创建示例工作流

### 方式 1: 通过 SQL 插入

```bash
# 执行 SQL 文件
mysql -u root -p txing_ai < sql/init_data.sql
```

### 方式 2: 通过前端创建（推荐）

1. 访问 http://localhost:8080/dash
2. 登录后进入"工作流管理"
3. 点击"新增工作流"，填写名称和描述
4. 点击"设计流程"进入编辑器
5. 从左侧拖拽节点到画布
6. 连接各节点形成流程
7. 配置各节点参数
8. 点击"保存"按钮

示例工作流节点布局：
```
[开始] -> [理解问题(LLM)] -> [搜索信息(工具)] -> [生成答案(LLM)] -> [结束]
```

## 步骤 4: 测试工作流（前端测试）

### 使用前端页面测试（推荐）

1. 在工作流编辑器页面，点击顶部的 **"运行测试"** 按钮（绿色按钮）
2. 在弹出的对话框中输入测试内容
3. 点击 **"开始运行"** 按钮
4. 等待工作流执行完成
5. 查看运行结果

### 运行测试功能说明

- **运行测试按钮**：位于顶部工具栏，点击后会打开测试对话框
- **输入测试内容**：可以输入任何想要测试的内容，如"请介绍一下人工智能的发展历史"
- **实时显示**：支持 SSE 流式输出，可以实时看到模型生成的响应
- **状态提示**：运行中会显示加载动画和状态信息
- **结果展示**：执行完成后显示完整的响应结果

## 步骤 5: 测试工作流（其他方式）

### 方式 1: 通过 API 测试

```bash
# 获取 token（从浏览器开发者工具中复制）
# 运行工作流
curl -X POST http://localhost:8080/api/workflow/1/run \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "content=请介绍一下人工智能的发展历史"
```

### 方式 2: 使用测试脚本

```bash
# 先登录获取 token
# 在浏览器开发者工具中复制 token

# 运行测试脚本
go run scripts/test_workflow.go -id 1 -token YOUR_TOKEN -content "请介绍一下人工智能的发展历史"
```

## 示例工作流配置

### 智能问答助手

节点配置：
1. **开始节点** - 接收用户输入
2. **理解问题节点（LLM）**:
   - 模型: deepseek-chat
   - 提示词: "你是一个专业的问答助手。请仔细分析用户的问题，提取其中的关键词和核心意图。"
   - 温度: 0.3
   - maxTokens: 1024
3. **搜索信息节点（工具）**:
   - 工具: web_search_tool
4. **生成答案节点（LLM）**:
   - 模型: deepseek-chat
   - 提示词: "你是一个专业的问答助手。请根据用户的问题和搜索到的信息，生成一个完整、准确、有帮助的回答。"
   - 温度: 0.7
   - maxTokens: 2048
5. **结束节点** - 返回最终答案

## 节点配置说明

### LLM 节点配置项

- **模型选择**：从下拉列表选择已配置的模型
- **系统提示词**：定义模型的角色和行为
- **温度参数**：控制输出的随机性（0-2）
- **最大Token数**：限制响应长度
- **上下文记忆**：是否记住历史对话

### 工具节点配置项

- **工具选择**：勾选需要使用的工具
- **参数配置**：为工具配置自定义参数

### 条件节点配置项

- **条件类型**：
  - 表达式判断：基于表达式结果判断
  - AI判断：让大模型根据内容判断
  - 工具结果：根据工具执行结果判断

## 常见问题

### 1. 认证失败

确保请求头中包含正确的 JWT token：
```
Authorization: Bearer YOUR_TOKEN
```

### 2. 模型不存在

确保数据库中已配置相应的模型，并且至少有一个渠道支持该模型。

### 3. 工具执行失败

检查 SearchAPI 配置是否正确，确保 API key 有效。

### 4. 连线无法建立

确保从一个节点的输出 handle 拖动到另一个节点的输入 handle。

## 下一步

- 尝试创建更复杂的工作流（添加条件分支）
- 测试不同的节点配置
- 探索工具节点的高级配置
- 优化提示词以获得更好的响应
