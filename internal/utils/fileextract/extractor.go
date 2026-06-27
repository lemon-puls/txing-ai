package fileextract

import "context"

// TextExtractor 文件文本提取器接口
type TextExtractor interface {
	// Extract 从文件中提取文本内容
	// filePath: 文件路径
	// 返回提取的文本内容和错误
	Extract(ctx context.Context, filePath string) (string, error)
	// Supports 返回支持的文件 MIME 类型列表
	Supports() []string
}
