# Txing AI

<div align="center">
  <img src="docs/images/logo.png" alt="Txing AI Logo" width="200"/>

  <h1>Txing AI</h1>
  <p>🤖 智能对话平台 | 多模型支持 | 智能负载均衡</p>

  [![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://go.dev/)
  [![Vue Version](https://img.shields.io/badge/Vue-3.x-brightgreen.svg)](https://vuejs.org/)
  [![Element Plus](https://img.shields.io/badge/Element%20Plus-Latest-blue)](https://element-plus.org/)
  [![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

  [📚 文档](docs/README.md) |
  [🚀 快速开始](#快速开始) |
  [👥 贡献](#贡献指南) |
  [📝 更新日志](CHANGELOG.md)
</div>

---

Txing AI 是一个现代化的 AI 聊天平台，基于 Vue 3 和 Go 开发，支持多模型接入和智能会话管理。平台采用前后端分离架构，提供流畅的用户体验和强大的后台管理功能。

## 🌟 功能特点

### 👥 用户端功能

| 功能模块 | 子功能 | 状态 | 说明 |
|---------|--------|------|------|
| **多模型支持** | 多种大语言模型接入 | ✅ 已实现 | 支持 DeepSeek、ChatGPT 等 |
| | 模型市场自由选择 | ✅ 已实现 | 支持模型评分和评价 |
| | 实时对话流式响应 | ✅ 已实现 | 打字机效果，WebSocket 通信 |
| | 智能上下文管理 | ✅ 已实现 | 支持多轮对话 |
| | 网页搜索增强 | 🔄 开发中 | 正在开发中 |
| **预设系统** | 自定义 AI 助手角色 | ✅ 已实现 | 支持角色背景、性格特征设定 |
| | 预设市场一键使用 | ✅ 已实现 | 支持按场景分类 |
| | 个性化对话场景定制 | ✅ 已实现 | 完整的预设管理功能 |
| | 预设分享与导入功能 | 🔄 开发中 | 正在开发中 |
| | 预设评分和收藏系统 | 🔄 开发中 | 正在开发中 |
| **会话管理** | 多会话并行支持 | ✅ 已实现 | 快速切换，实时保存 |
| | 历史记录完整保存 | ✅ 已实现 | 支持查看和导出 |
| | 会话导出备份 | ✅ 已实现 | 支持多种格式 |
| | 会话标题自动生成 | ✅ 已实现 | 智能标题生成 |
| | 会话分类管理 | 🔄 开发中 | 正在开发中 |

### 👨‍💼 管理端功能

| 功能模块 | 子功能 | 状态 | 说明 |
|---------|--------|------|------|
| **渠道管理** | 多渠道统一管理 | ✅ 已实现 | 支持多个供应商 |
| | 智能负载均衡策略 | ✅ 已实现 | 优先级控制、权重分配 |
| | 自动故障转移 | ✅ 已实现 | 故障自动切换 |
| | 模型组合配置 | ✅ 已实现 | 一个渠道支持多个模型 |
| | 实时状态监控 | ✅ 已实现 | 调用量、响应时间、错误率统计 |
| | 成本统计 | 🔄 开发中 | 正在开发中 |
| **预设管理** | 预设审核发布流程 | ✅ 已实现 | 完整的审核流程 |
| | 分类标签管理系统 | ✅ 已实现 | 支持分类和标签 |
| | 热门推荐设置 | 🔄 开发中 | 正在开发中 |
| | 内容质量控制 | ✅ 已实现 | 基础内容管理 |
| | 用户反馈处理 | 🔄 开发中 | 正在开发中 |
| | 预设排行榜 | 🔄 开发中 | 正在开发中 |
| **用户管理** | 用户权限分级控制 | ✅ 已实现 | 支持角色分配 |
| | 使用量统计分析 | ✅ 已实现 | 调用次数、使用时长统计 |
| | 费用统计 | 🔄 开发中 | 正在开发中 |
| | 异常行为监控 | 🔄 开发中 | 正在开发中 |
| | 账户状态管理 | ✅ 已实现 | 启用/禁用用户 |
| | 用户反馈处理 | 🔄 开发中 | 正在开发中 |
| **模型市场** | 模型上架审核流程 | ✅ 已实现 | 完整的审核流程 |
| | 价格策略设置 | ✅ 已实现 | 按次计费、按 Token 计费 |
| | 使用量统计 | ✅ 已实现 | 基础统计功能 |
| | 评分反馈系统 | 🔄 开发中 | 正在开发中 |
| | 模型性能监控 | 🔄 开发中 | 正在开发中 |
| | 模型版本管理 | 🔄 开发中 | 正在开发中 |

### 📊 实现状态统计

- ✅ **已实现**: 65% - 核心功能已完成
- 🔄 **开发中**: 25% - 正在积极开发
- 📋 **计划中**: 10% - 后续版本规划

### 🎯 核心特性

| 特性 | 状态 | 技术实现 |
|------|------|----------|
| **智能负载均衡** | ✅ 已实现 | Channel 系统，优先级权重调度 |
| **高性能设计** | ✅ 已实现 | WebSocket、流式响应、连接池 |
| **安全性** | ✅ 已实现 | JWT 认证、权限控制、数据加密 |
| **可扩展性** | ✅ 已实现 | 模块化架构、插件化模型接入 |

## 🏗️ 系统设计

### 📐 系统架构图

```mermaid
graph TD
    %% 定义样式
    classDef frontend fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef gateway fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef business fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef data fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef adapter fill:#fce4ec,stroke:#880e4f,stroke-width:2px
    classDef external fill:#f1f8e9,stroke:#33691e,stroke-width:2px
    
    %% 1. 前端层
    subgraph Frontend
        A[Vue 3 Frontend App]
    end
    
    %% 2. 网关层
    subgraph Gateway
        E[Gin Router Gateway]
        E1[JWT Auth Middleware]
    end
    
    %% 3. 业务层
    subgraph Business
        I[User Service]
        P[Channel Service]
        T[Preset Service]
        L[Chat Service WebSocket]
    end

    %% 4. 数据层
    subgraph Data
        CC[MySQL Database]
        HH[Redis Cache]
    end
    
    %% 5. 适配器层
    subgraph Adapter
        ZZ[Adapter Factory]
        W[OpenAI Adapter]
        Y[Volcengine Adapter]
        AA[Polo Adapter]
    end
    
    %% 6. 外部服务
    subgraph External
        X[OpenAI API]
        Z[Volcengine API]
        BB[Polo API]
    end
    
    %% 连接关系
    A -->|HTTP API Request| E
    E -->|Route| E1
    E1 --> I
    E1 --> P
    E1 --> T

    A -.->|WebSocket| E
    E -.->|Upgrade| L
    A <--> L
    
    I --> CC
    I --> HH
    L --> CC
    L --> HH
    P --> CC
    T --> CC

    L --> ZZ
    ZZ --> W
    ZZ --> Y
    ZZ --> AA

    W --> X
    Y --> Z
    AA --> BB
    
    %% 应用样式
    class A frontend
    class E,E1 gateway
    class I,P,T,L business
    class CC,HH data
    class ZZ,W,Y,AA adapter
    class X,Z,BB external
```

### 🔗 核心组件关系图

```mermaid
graph LR
    subgraph "用户端 User Side"
        A[用户 User] --> B[会话 Conversation]
        A --> C[预设 Preset]
        B --> D[消息 Message]
    end
    
    subgraph "管理端 Admin Side"
        E[管理员 Admin] --> F[渠道 Channel]
        E --> G[模型 Model]
        E --> H[预设管理 Preset Management]
        E --> I[用户管理 User Management]
    end
    
    subgraph "系统核心 System Core"
        F --> J[负载均衡器 Load Balancer]
        J --> K[适配器工厂 Adapter Factory]
        K --> L[OpenAI 适配器]
        K --> M[火山引擎适配器]
        K --> N[Polo 适配器]
        K --> O[...]
        
        B --> P[聊天服务 Chat Service]
        P --> J
        P --> Q[消息限制器 Message Limiter]
        P --> R[响应缓冲区 Response Buffer]   
        
        %% 新增：Channel 和 Model 的关系
        F -.->|支持模型列表 Models| G
        F -.->|模型映射 Mappings| G
        G -.->|被渠道支持| F
    end
    
    subgraph "外部服务 External Services"
        L --> S[OpenAI API]
        M --> T[火山引擎 API]
        N --> U[Polo API]
    end
    
    A --> P
    C --> P
    F --> P
```

### 💬 聊天请求处理流程图

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant WS as WebSocket
    participant Auth as 认证中间件
    participant Chat as 聊天服务
    participant Channel as 渠道管理
    participant Adapter as 适配器层
    participant AI as AI 服务商
    participant DB as 数据库
    participant Redis as Redis 缓存
    
    Client->>WS: 建立 WebSocket 连接
    WS->>Auth: 验证用户身份
    Auth->>Redis: 检查 Token
    Redis-->>Auth: Token 有效
    Auth-->>WS: 认证通过
    WS-->>Client: 连接建立成功
    
    Client->>WS: 发送聊天消息
    WS->>Chat: 处理聊天请求
    Chat->>DB: 保存用户消息
    Chat->>Redis: 检查消息限制
    Redis-->>Chat: 限制检查通过
    
    Chat->>Channel: 获取支持模型的渠道
    Channel->>DB: 查询渠道配置
    DB-->>Channel: 返回渠道列表
    Channel->>Chat: 选择最优渠道
    
    Chat->>Adapter: 创建适配器实例
    Adapter->>AI: 发送流式请求
    AI-->>Adapter: 流式响应数据
    
    loop 流式响应处理
        Adapter->>Chat: 发送响应块
        Chat->>WS: 转发给客户端
        WS->>Client: 实时显示内容
    end
    
    Chat->>DB: 保存 AI 响应
    Chat->>Redis: 更新使用统计
    Chat->>WS: 发送结束标志
    WS->>Client: 完成响应
```

### 🎯 核心概念说明

#### 🔗 Channel（渠道）
- **定义**: 连接不同 AI 服务商的抽象层
- **功能**: 
  - 统一管理多个 AI 服务商
  - 支持优先级和权重配置
  - 实现智能负载均衡
  - 提供故障转移机制
- **配置项**:
  - `name`: 渠道名称
  - `type`: 渠道类型（OpenAI、火山引擎、Polo）
  - `priority`: 优先级（1-100）
  - `weight`: 权重分配
  - `models`: 支持的模型列表
  - `mappings`: 模型映射关系
  - `endpoint`: 服务地址
  - `secret`: API 密钥

#### 🤖 Model（模型）
- **定义**: AI 模型的具体实现
- **功能**:
  - 定义模型的基本信息
  - 支持模型标签分类
  - 配置模型特性（高上下文、默认模型等）
- **配置项**:
  - `name`: 模型名称
  - `description`: 模型描述
  - `default`: 是否为默认模型
  - `high_context`: 是否支持高上下文
  - `avatar`: 模型头像
  - `tag`: 模型标签

#### 🔄 Channel-Model 关系
- **多对多关系**: 一个渠道可以支持多个模型，一个模型可以被多个渠道支持
- **模型映射**: 渠道可以配置模型映射规则，根据条件动态选择具体的模型版本
- **负载均衡**: 系统根据渠道的优先级和权重，智能选择最优渠道处理请求
- **映射示例**:
  ```json
  {
    "sourceModel": "deepseek-r1",
    "conditions": [
      {
        "targetModel": "deepseek-r1-250120",
        "conditions": {"enableWeb": true}
      },
      {
        "targetModel": "deepseek-r1-250121", 
        "conditions": {"enableWeb": false}
      }
    ]
  }
  ```

#### 🎭 Preset（预设）
- **定义**: AI 助手的角色配置
- **功能**:
  - 定义 AI 助手的性格特征
  - 设置对话上下文
  - 提供场景化对话
- **属性**:
  - `name`: 预设名称
  - `description`: 预设描述
  - `context`: 系统提示词
  - `avatar`: 预设头像
  - `tags`: 标签分类

#### 💬 Conversation（会话）
- **定义**: 用户与 AI 的对话实例
- **功能**:
  - 管理对话历史
  - 保存消息记录
  - 维护对话上下文
- **属性**:
  - `name`: 会话标题
  - `model`: 使用的模型
  - `messages`: 消息历史
  - `presetId`: 关联的预设

### 🔄 负载均衡策略

#### 优先级策略
```mermaid
graph LR
    A[用户请求] --> B{检查优先级}
    B -->|优先级高| C[选择高优先级渠道]
    B -->|优先级相同| D[权重分配]
    D --> E[按权重随机选择]
    C --> F[发送请求]
    E --> F
    F --> G{请求成功?}
    G -->|是| H[返回响应]
    G -->|否| I[故障转移]
    I --> J[选择下一个渠道]
    J --> F
```

#### 权重分配算法
```javascript
// 权重分配示例
const channels = [
  { name: 'Channel A', weight: 60, priority: 100 },
  { name: 'Channel B', weight: 30, priority: 100 },
  { name: 'Channel C', weight: 10, priority: 100 }
];

// 60% 的请求会路由到 Channel A
// 30% 的请求会路由到 Channel B  
// 10% 的请求会路由到 Channel C
```

### 🔄 模型映射机制

#### 模型映射流程图
```mermaid
sequenceDiagram
    participant User as 用户
    participant Chat as 聊天服务
    participant Channel as 渠道管理
    participant Model as 模型映射
    participant Adapter as 适配器层
    participant AI as AI 服务商
    
    User->>Chat: 发送聊天请求 (model: deepseek-r1)
    Chat->>Channel: 查询支持 deepseek-r1 的渠道
    Channel-->>Chat: 返回渠道列表
    
    Note over Chat,Model: 模型映射处理
    Chat->>Model: 检查模型映射规则
    Model->>Model: 根据条件映射模型
    Note right of Model: 例如：enableWeb=true → deepseek-r1-250120<br/>enableWeb=false → deepseek-r1-250121
    
    Model-->>Chat: 返回映射后的模型
    Chat->>Adapter: 使用映射后的模型发送请求
    Adapter->>AI: 调用 AI 服务商 API
    AI-->>Adapter: 返回响应
    Adapter-->>Chat: 转发响应
    Chat-->>User: 返回聊天结果
```

#### 模型映射配置示例
```json
{
  "sourceModel": "deepseek-r1",
  "conditions": [
    {
      "targetModel": "deepseek-r1-250120",
      "conditions": {
        "enableWeb": true,
        "temperature": 0.7
      }
    },
    {
      "targetModel": "deepseek-r1-250121", 
      "conditions": {
        "enableWeb": false,
        "temperature": 0.5
      }
    }
  ]
}
```

#### 映射规则说明
- **源模型**: 用户请求的模型名称
- **目标模型**: 实际发送给 AI 服务商的模型名称
- **条件映射**: 根据请求参数动态选择目标模型
- **默认映射**: 如果没有匹配的条件，使用源模型作为目标模型

### 🛡️ 安全机制

#### 认证授权流程
```mermaid
graph TD
    A[用户登录] --> B[JWT Token 生成]
    B --> C[Token 存储到 Redis]
    C --> D[前端保存 Token]
    D --> E[请求携带 Token]
    E --> F[中间件验证 Token]
    F --> G{Token 有效?}
    G -->|是| H[获取用户信息]
    G -->|否| I[返回 401 错误]
    H --> J[权限检查]
    J --> K{有权限?}
    K -->|是| L[执行业务逻辑]
    K -->|否| M[返回 403 错误]
```

#### 消息限制机制
- **每日消息限制**: 防止滥用
- **用户等级限制**: 不同用户等级有不同的限制
- **实时监控**: 实时统计和告警
- **自动恢复**: 每日零点自动重置限制

## 🔧 技术架构

### 🎨 前端技术栈
- Vue 3 (Composition API)
  - 响应式数据处理
  - 组件化开发
- Element Plus
  - 美观的 UI 组件
  - 响应式布局
- Vite
  - 快速的开发构建
  - 模块热重载
- Pinia
  - 状态管理
  - 持久化存储
- Axios
  - 请求拦截器
  - 响应处理
- WebSocket
  - 实时通信
  - 断线重连

### ⚙️ 后端技术栈
- Go
  - 高性能
  - 并发处理
- Gin
  - 路由管理
  - 中间件支持
- GORM
  - ORM 映射
  - 数据库操作
- Redis
  - 缓存处理
  - 会话存储
- WebSocket
  - 长连接管理
  - 消息推送
- Zap
  - 日志记录
  - 性能监控

## 💫 系统特点

1. **智能负载均衡**
   - Channel 系统支持多模型配置
   - 基于优先级和权重的智能调度
   - 自动故障转移机制
   - 动态负载分配
   - 实时监控和调整
   - 成本优化策略

2. **高性能设计**
   - WebSocket 实时通信
   - 流式响应处理
   - 连接池优化
   - 多级缓存策略
   - 异步处理机制
   - 性能监控告警

3. **安全性**
   - 完整的权限控制系统
   - 数据加密传输
   - 敏感信息保护
   - 操作日志审计
   - 防攻击策略
   - 数据备份恢复

4. **可扩展性**
   - 模块化架构设计
   - 插件化模型接入
   - 灵活的配置系统
   - 开放的 API 设计
   - 第三方集成支持
   - 自定义扩展能力

## ✨ 特性

- 🚀 基于 Go 1.23+ 的高性能后端服务
    - Gin 框架提供高效的 HTTP 服务
    - GORM 实现优雅的数据库操作
    - Redis 支持高性能缓存
    - Zap 处理日志记录
    - Swagger 自动生成 API 文档

- 🎯 基于 Vue 3 的现代化前端框架
    - Vite 构建工具确保快速的开发体验
    - Element Plus 组件库提供美观的 UI
    - Pinia 状态管理
    - Vue Router 路由管理

## 🚀 快速开始

### 环境要求

- Go 1.23+
- MySQL 5.7+
- Redis 6.0+
- Node.js 18+
- pnpm 8.0+

### 获取代码

```bash
git clone git@github.com:lemon-puls/Vine.git
cd Vine
```

### 后端服务

1. 配置数据库
```bash
# 创建配置文件
cp config.toml.sample runtime/config.toml
# 根据实际情况修改配置
```

2. 生成 API 文档
```bash
swag init -g cmd/main.go
```

3. 前端项目初始化
```bash
# 进入前端目录下
cd static/frontend
# 安装依赖
npm install
# 项目打包
npm run build
```
> 注意：前端部分的打包生成目录 dist 会在 GO 项目中嵌入作为静态资源对外提供服务，所以不需单独启动前端项目。

4. 启动服务
   启动时会自动执行数据库迁移，生成表，无需手动创建。
```bash
go run cmd/main.go
```

## 📚 项目结构

```
├── cmd                 # 主程序入口
├── internal           # 内部包
│   ├── app           # 应用生命周期
│   ├── controller    # 控制器
│   ├── middleware    # 中间件
│   ├── domain        # 数据库实体
│   ├── dto           # 数据传输对象
│   ├── vo            # 视图对象
│   ├── enum          # 枚举
│   ├── global        # 公共包
│   ├── iface         # 接口定义
│   ├── service      # 业务服务
│   ├── route         # 路由
│   └── utils        # 工具函数
├── runtime           # 运行时文件（日志、配置文件等）
├── scripts          # 脚本工具
└── static/frontend              # 前端项目
    ├── src
    │   ├── api     # API 接口
    │   ├── assets  # 静态资源
    │   ├── components # 组件
    │   ├── layouts   # 布局
    │   ├── router    # 路由
    │   ├── stores    # 状态管理
    │   └── views     # 页面
    └── public      # 公共资源
```

## 开发说明
### SVG 图标使用
项目封装了 SvgIcon 组件，可以直接使用，主要支持配置项：
- `icon`：图标名称，必填
- `size`：图标大小，默认 16px
- `color`：图标颜色
- `click`：是否开启点击动画，默认 false
- `class`：自定义类名
- `spin`：是否开启旋转动画，默认 false
- `rotate`：旋转角度（顺时针），默认 0
- `hover`: 是否开启鼠标悬停动画，默认 false

使用示例：
```vue
<SvgIcon icon="theme" size="20" color="#1890ff" click/>
```

### 前端权限指令使用

```js

// 权限指令
// Permission directive

// 使用方式 Usage:
v-permission:role="['admin', 'editor']"  // 角色权限 Role permission

v-permission:perm="['create', 'edit']"   // 操作权限 Permission-based

v-permission:role.hide="['admin']"       // 无权限时隐藏元素 Hide when no permission

v-permission:perm.hide="['create']"      // 无权限时隐藏元素 Hide when no permission
```

在需要进行前端权限的元素上添加 v-permission 指令即可，实例如下：

```js
<el-button v-permission:role.hide="['admin']" type="primary" circle class="new-chat-button" @click="createNewChat('')">
    <el-icon>
         <Plus class="icon-bounce"/>
    </el-icon>
</el-button>
```

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 鸣谢

- [Gin](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Vue.js](https://vuejs.org/)
- [Element Plus](https://element-plus.org/)
- [Vite](https://vitejs.dev/)

## 📊 统计

![Visitors](https://visitor-badge.laobi.icu/badge?page_id=txing-ai.readme)
[![Stars](https://img.shields.io/github/stars/txing-ai/txing-ai)](https://github.com/txing-ai/txing-ai/stargazers)
[![Forks](https://img.shields.io/github/forks/txing-ai/txing-ai)](https://github.com/txing-ai/txing-ai/network/members)
[![Issues](https://img.shields.io/github/issues/txing-ai/txing-ai)](https://github.com/txing-ai/txing-ai/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/txing-ai/txing-ai)](https://github.com/txing-ai/txing-ai/pulls)

## 🎉 展示

### 🏠 首页
<div align="center">
  <img src="static/imgs/首页.png" alt="首页"/>
  <p><em>现代化的首页设计，展示平台核心功能和特色</em></p>
</div>

### 💬 AI 聊天功能
<div align="center">
  <img src="static/imgs/AI聊天页.png" alt="AI聊天页" width="65%"/>
  <img src="static/imgs/AI聊天参数设置.png" alt="AI聊天参数设置" width="30%"/>
  <p><em>智能对话界面，支持多模型切换和参数自定义</em></p>
</div>

### 🤖 AI 助手预设
<div align="center">
  <img src="static/imgs/AI助手页（预设）.png" alt="AI助手预设" />
  <p><em>丰富的 AI 助手预设市场，支持角色定制和场景应用</em></p>
</div>

### 👨‍💼 后台管理系统

#### 📊 控制台
<div align="center">
  <img src="static/imgs/后台管理-控制台.png" alt="后台管理控制台"/>
  <p><em>数据统计仪表板，实时监控平台运营状态</em></p>
</div>

#### 👥 用户管理
<div align="center">
  <img src="static/imgs/后台管理-用户管理.png" alt="用户管理"/>
  <p><em>用户权限管理，支持角色分配和状态控制</em></p>
</div>

#### 🔗 渠道管理
<div align="center">
  <img src="static/imgs/后台管理-渠道管理.png" alt="渠道管理"/>
  <p><em>多渠道配置管理，支持负载均衡和故障转移</em></p>
</div>

#### 🤖 模型管理
<div align="center">
  <img src="static/imgs/后台管理-模型管理.png" alt="模型管理"/>
  <p><em>模型市场管理，支持模型上架和价格策略配置</em></p>
</div>

#### 🎯 AI 助手管理
<div align="center">
  <img src="static/imgs/后台管理-AI助手管理.png" alt="AI助手管理"/>
  <p><em>预设内容管理，支持审核发布和分类管理</em></p>
</div>

---

<div align="center">
  <sub>Built with ❤️ by The Txing AI Team</sub>
</div>
