package fileextract

import (
	"context"
	"fmt"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// DocxExtractor Word 文档提取器
type DocxExtractor struct{}

// Extract 提取 Word 文档中的文本内容
func (e *DocxExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read docx: %w", err)
	}
	defer r.Close()

	doc := r.Editable()
	content := doc.GetContent()

	// 清理 XML 标签，提取纯文本
	text := e.cleanXmlTags(content)
	return text, nil
}

// cleanXmlTags 清理 XML 标签，保留文本内容
func (e *DocxExtractor) cleanXmlTags(content string) string {
	var sb strings.Builder
	inTag := false

	for _, ch := range content {
		switch ch {
		case '<':
			inTag = true
		case '>':
			inTag = false
			sb.WriteString(" ")
		default:
			if !inTag {
				sb.WriteRune(ch)
			}
		}
	}

	// 清理多余空白
	text := sb.String()
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// Supports 返回支持的 MIME 类型
func (e *DocxExtractor) Supports() []string {
	return []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
}
