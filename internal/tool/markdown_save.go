package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

// markdownSaveParams 保存Markdown文件参数
type markdownSaveParams struct {
	Content  string `json:"content" jsonschema:"description=要保存的Markdown内容"`
	Filename string `json:"filename" jsonschema:"description=文件名(不含扩展名)"`
}

// saveMarkdown 将Markdown内容保存到本地文件
func saveMarkdown(ctx context.Context, params *markdownSaveParams) (string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("获取当前工作目录失败", zap.Error(err))
		return "", fmt.Errorf("获取当前工作目录失败: %v", err)
	}
	savePath := currentDir
	savePath = filepath.Join(savePath, "runtime", "temp")

	// 确保文件名有.md扩展名
	filename := params.Filename
	if filepath.Ext(filename) != ".md" {
		filename = filename + ".md"
	}

	// 构建完整的文件路径
	fullPath := filepath.Join(savePath, filename)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error("创建目录失败", zap.String("dir", dir), zap.Error(err))
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	err = os.WriteFile(fullPath, []byte(params.Content), 0644)
	if err != nil {
		log.Error("保存Markdown文件失败", zap.String("path", fullPath), zap.Error(err))
		return "", fmt.Errorf("保存Markdown文件失败: %v", err)
	}

	// 记录成功日志
	log.Info("Markdown文件保存成功",
		zap.String("path", fullPath),
		zap.Int("contentLength", len(params.Content)),
		zap.Time("timestamp", time.Now()))

	return fmt.Sprintf("Markdown文件已成功保存到: %s", fullPath), nil
}
