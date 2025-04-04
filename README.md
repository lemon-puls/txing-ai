# Txing-AI


## 快速开始

### 环境要求
- Go 1.22+
- MySQL
- Node.js 18+ (用于前端开发)

### 配置说明
1. 在 `runtime/config.toml` 进行配置

### 生成接口文档
```bash
swag init -g cmd/main.go
```


### 运行服务
```bash
# 启动后端服务
go run cmd/main.go