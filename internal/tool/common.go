package tool

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
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

// resolveSavedFile 解析工具保存的文件路径为绝对路径：
// 兼容 markdown_save_tool 返回的 "./xxx.md" 相对路径（相对当前用户的保存目录）
// 与绝对路径；限制在保存目录内，避免任意路径读取
func resolveSavedFile(ctx context.Context, filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("filePath 为空")
	}
	if filepath.IsAbs(filePath) {
		return filePath, nil
	}
	savePath, err := buildSaveDir(ctx)
	if err != nil {
		return "", err
	}
	rel := strings.TrimPrefix(filePath, "./")
	rel = strings.TrimPrefix(rel, `.\\`)
	abs := filepath.Join(savePath, filepath.Clean(rel))
	// 防目录穿越：解析后的路径必须仍在保存目录内
	if !strings.HasPrefix(abs, filepath.Clean(savePath)+string(filepath.Separator)) &&
		abs != filepath.Clean(savePath) {
		return "", fmt.Errorf("非法文件路径: %s", filePath)
	}
	return abs, nil
}
