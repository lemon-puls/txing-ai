package fileextract

import (
	"context"
	"os"
)

// MarkdownExtractor Markdown 文件提取器
type MarkdownExtractor struct{}

// Extract 提取 Markdown 文件内容（直接返回原始文本）
func (e *MarkdownExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Supports 返回支持的 MIME 类型
func (e *MarkdownExtractor) Supports() []string {
	return []string{
		"text/markdown",
		"text/x-markdown",
	}
}
