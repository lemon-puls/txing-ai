OUT_DIR = output
# 必要的变量注入，如 commit_hash 以跟踪代码版本
# 提交 hash
GIT_HASH := $(shell git show -s --format=%h --abbrev=8)
BIN_NAME := txing-ai
PKG_NAME := ${BIN_NAME}_${GIT_HASH}

# 构建标志
BUILD_FLAGS := -trimpath -ldflags="-s -w"

.PHONY: win build docker gen

win: build-win

docker: gen build-docker

gen:
	which swag || go install github.com/swaggo/swag/cmd/swag@latest
	go generate ./...

# 快速构建，跳过生成代码步骤；仅供调试时确定没有改动生成代码用，供生产的构建务必用 build
build:
	go build ${BUILD_FLAGS} -o ${OUT_DIR}/${PKG_NAME} cmd/main.go

build-docker:
	go build ${BUILD_FLAGS} -o ${BIN_NAME} cmd/main.go

build-win:
	go build ${BUILD_FLAGS} -o ${OUT_DIR}/${PKG_NAME}.exe cmd/main.go