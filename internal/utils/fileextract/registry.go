package fileextract

import (
	"fmt"
	"strings"
)

// Registry 提取器注册表
type Registry struct {
	extractors map[string]TextExtractor // key: MIME type (小写)
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		extractors: make(map[string]TextExtractor),
	}
}

// Register 注册提取器
func (r *Registry) Register(extractor TextExtractor) {
	for _, mimeType := range extractor.Supports() {
		r.extractors[strings.ToLower(mimeType)] = extractor
	}
}

// GetExtractor 根据文件类型获取提取器
func (r *Registry) GetExtractor(fileType string) (TextExtractor, error) {
	ext, ok := r.extractors[strings.ToLower(fileType)]
	if !ok {
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
	return ext, nil
}

// IsSupported 检查文件类型是否支持
func (r *Registry) IsSupported(fileType string) bool {
	_, ok := r.extractors[strings.ToLower(fileType)]
	return ok
}

// Default 默认注册表实例
var Default = NewRegistry()

func init() {
	// 注册所有内置提取器
	Default.Register(&TxtExtractor{})
	Default.Register(&MarkdownExtractor{})
	Default.Register(&HtmlExtractor{})
	Default.Register(&PdfExtractor{})
	Default.Register(&DocxExtractor{})
}
