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

// saveDirQuiet 获取当前保存目录；配置不可用（如单元测试环境没有 runtime/config.yaml）时
// 返回空串而不是 panic，避免 isPathAllowed 等路径校验在测试中崩溃
func saveDirQuiet(ctx context.Context) (dir string) {
	defer func() {
		if r := recover(); r != nil {
			dir = ""
		}
	}()
	d, err := buildSaveDir(ctx)
	if err != nil {
		return ""
	}
	return d
}

// resolveSavedFile 解析工具保存的文件路径为绝对路径：
// 兼容 markdown_save_tool 返回的 "./xxx.md" 相对路径（相对当前用户的保存目录）、
// 相对工作目录的路径（如 runtime/temp_files/2026-08-20/xxx.md）与绝对路径；
// 限制在 runtime 目录内，避免任意路径读取
func resolveSavedFile(ctx context.Context, filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("filePath 为空")
	}
	if filepath.IsAbs(filePath) {
		if !isPathUnderRuntime(filePath) {
			return "", fmt.Errorf("非法文件路径: %s（仅允许 runtime 目录下的文件）", filePath)
		}
		return filepath.Clean(filePath), nil
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// 候选路径：相对保存目录（./xxx.md）与相对工作目录（runtime/xxx.md）各解析一份
	candidates := []string{filepath.Join(currentDir, filepath.Clean(filePath))}
	if savePath := saveDirQuiet(ctx); savePath != "" {
		rel := strings.TrimPrefix(strings.TrimPrefix(filePath, "./"), `.\\`)
		candidates = append(candidates, filepath.Join(savePath, filepath.Clean(rel)))
	}

	// 优先返回已存在的文件；都存在/都不存在时按顺序取第一个合法候选
	for _, c := range candidates {
		if isPathUnderRuntime(c) {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}
	for _, c := range candidates {
		if isPathUnderRuntime(c) {
			return c, nil
		}
	}
	return "", fmt.Errorf("非法文件路径: %s（仅允许 runtime 目录下的文件）", filePath)
}

// isPathUnderRuntime 判断绝对路径是否位于当前工作目录的 runtime 目录内
func isPathUnderRuntime(absPath string) bool {
	currentDir, err := os.Getwd()
	if err != nil {
		return false
	}
	abs := filepath.Clean(absPath)
	root := filepath.Clean(filepath.Join(currentDir, "runtime"))
	return abs == root || strings.HasPrefix(abs, root+string(filepath.Separator))
}
