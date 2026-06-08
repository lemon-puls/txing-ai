package fileextract

import (
	"context"
	"os"
)

// TxtExtractor 纯文本文件提取器
type TxtExtractor struct{}

// Extract 提取文本文件内容
func (e *TxtExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Supports 返回支持的 MIME 类型
func (e *TxtExtractor) Supports() []string {
	return []string{
		"text/plain",
	}
}
