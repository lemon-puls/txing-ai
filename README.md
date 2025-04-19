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

- **多模型支持**
  - 支持接入多种大语言模型（如 DeepSeek、ChatGPT 等）
  - 模型市场自由选择，支持模型评分和评价
  - 实时对话流式响应，打字机效果
  - 智能上下文管理，支持多轮对话
  - 支持网页搜索增强

- **预设系统**
  - 自定义 AI 助手角色（支持设定角色背景、性格特征等）
  - 预设市场一键使用，支持按场景分类
  - 个性化对话场景定制
  - 预设分享与导入功能
  - 预设评分和收藏系统

- **会话管理**
  - 多会话并行支持，快速切换
  - 历史记录完整保存与查看
  - 会话导出备份（支持多种格式）
  - 实时保存同步
  - 会话标题自动生成
  - 会话分类管理

### 👨‍ 管理端功能

- **渠道管理**
  - 多渠道统一管理（支持多个供应商）
  - 智能负载均衡策略
    - 优先级控制（1-100，数字越大优先级越高）
    - 权重分配（按百分比分配流量）
    - 自动故障转移
  - 模型组合配置（一个渠道支持多个模型）
  - 实时状态监控
    - 调用量统计
    - 响应时间监控
    - 错误率统计
    - 成本统计

- **预设管理**
  - 预设审核发布流程
  - 分类标签管理系统
  - 热门推荐设置
  - 内容质量控制
  - 用户反馈处理
  - 预设排行榜

- **用户管理**
  - 用户权限分级控制
  - 使用量统计分析
    - 调用次数统计
    - 使用时长统计
    - 费用统计
  - 异常行为监控
  - 账户状态管理
  - 用户反馈处理

- **模型市场**
  - 模型上架审核流程
  - 价格策略设置
    - 按次计费
    - 按 Token 计费
  - 使用量统计
  - 评分反馈系统
  - 模型性能监控
  - 模型版本管理

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

<div align="center">
  <img src="docs/images/screenshot1.png" alt="Screenshot 1" width="45%"/>
  <img src="docs/images/screenshot2.png" alt="Screenshot 2" width="45%"/>
</div>

<div align="center">
  <img src="docs/images/screenshot3.png" alt="Screenshot 3" width="45%"/>
  <img src="docs/images/screenshot4.png" alt="Screenshot 4" width="45%"/>
</div>

---

<div align="center">
  <sub>Built with ❤️ by The Txing AI Team</sub>
</div>