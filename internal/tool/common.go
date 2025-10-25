package tool

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"
)

func buildSaveDir(ctx context.Context) (string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("获取当前工作目录失败", zap.Error(err))
		return "", err
	}
	localUploadConfig := global.LoadConfig().LocalUploadConfig

	savePath := currentDir
	savePath = filepath.Join(savePath, localUploadConfig.Dir)
	userId, exist := utils.GetUIDFromContextAllowEmpty(ctx)
	if exist {
		savePath = filepath.Join(savePath, fmt.Sprintf("%d", userId))
	}
	currentDate := time.Now().Format("2006-01-02")
	savePath = filepath.Join(savePath, currentDate)
	// 确保目录存在
	if err := os.MkdirAll(savePath, 0755); err != nil {
		log.Error("创建保存目录失败", zap.String("dir", savePath), zap.Error(err))
		return "", err
	}
	return savePath, nil
}
