# Txing AI



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