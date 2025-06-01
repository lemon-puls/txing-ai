#! /bin/bash

IMAGE_REPO=ghcr.nju.edu.cn
IMAGE_NAME=lemon-puls/txing-ai
CONTAINER_NAME=txing-ai
NET_MODE=bridge
RUNTIME_DIR=/home/lemon/web/txing-ai/runtime
PORT=8083

# 检查是否提供了 tag 参数
if [ -n "$1" ]; then
    TAG=$1
else
    TAG=latest
fi

echo "Using image tag: ${TAG}"

# 停止并删除已存在的同名容器
if [ "$(docker ps -aq -f name=${CONTAINER_NAME})" ]; then
    echo "Stopping and removing existing container..."
    docker stop ${CONTAINER_NAME}
    docker rm ${CONTAINER_NAME}
fi

# 启动新容器
docker run \
    --net ${NET_MODE} \
    -v ${RUNTIME_DIR}:/app/runtime \
    -p ${PORT}:8080 \
    -d \
    --name ${CONTAINER_NAME} \
    --restart unless-stopped \
    ${IMAGE_REPO}/${IMAGE_NAME}:sha-${TAG}

# 清理旧镜像，保留最近的两个版本
echo "Cleaning up old images..."
# 获取所有 txing-ai 镜像，按创建时间排序（最新的在前）
IMAGES=$(docker images ${IMAGE_REPO}/${IMAGE_NAME} --format "{{.CreatedAt}}\t{{.Repository}}:{{.Tag}}" --filter "label=org.opencontainers.image.version" | sort -r | cut -f2)

# 计算镜像数量
IMAGE_COUNT=$(echo "$IMAGES" | wc -l)

if [ "$IMAGE_COUNT" -gt 2 ]; then
    # 获取需要删除的镜像（除了最新的两个）
    OLD_IMAGES=$(echo "$IMAGES" | tail -n +3)
    
    # 删除旧镜像
    for IMAGE in $OLD_IMAGES; do
        echo "Removing old image: $IMAGE"
        docker rmi $IMAGE
    done
else
    echo "No old images to clean up"
fi

echo "Container started and old images cleaned up successfully"