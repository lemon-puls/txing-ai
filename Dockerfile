# 前端构建阶段
FROM node:18-alpine as frontend-builder

# 设置 npm 镜像源
RUN npm config set registry https://registry.npmmirror.com

# 设置工作目录
WORKDIR /app/static/frontend

# 复制所有源代码
COPY static/frontend .

# 安装主项目依赖
RUN npm install

# 安装生成的 API 客户端依赖
WORKDIR /app/static/frontend/src/api/generated
RUN npm install

# 返回主工作目录并构建
WORKDIR /app/static/frontend
RUN npm run build && \
    ls -la dist && \
    touch /tmp/frontend_build_complete

# Go 构建阶段
FROM golang:1.23 as backend-builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 设置 GOPROXY
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖项
RUN go mod download

# 复制后端源代码
COPY . .

# 从前端构建阶段复制构建产物，并验证
COPY --from=frontend-builder /tmp/frontend_build_complete /tmp/frontend_build_complete
COPY --from=frontend-builder /app/static/frontend/dist /app/static/frontend/dist

# 验证前端构建产物存在
RUN ls -la /app/static/frontend/dist && \
    test -f /tmp/frontend_build_complete

# 构建应用程序
RUN make docker && \
    chmod +x txing-oj && \
    ls -hail txing-oj && \
    touch build_complete

# 最终运行阶段
FROM debian:bookworm-slim as final

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=backend-builder /app/build_complete /app/build_complete
COPY --from=backend-builder /app/txing-oj /app/txing-oj

# 清理构建标记文件
RUN rm build_complete && \
    ls -hail txing-oj

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 启动命令
ENTRYPOINT ["./txing-oj"]