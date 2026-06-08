package fileextract

import (
	"context"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PdfExtractor PDF 文件提取器
type PdfExtractor struct{}

// Extract 提取 PDF 文件中的文本内容
func (e *PdfExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}

		sb.WriteString(text)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// Supports 返回支持的 MIME 类型
func (e *PdfExtractor) Supports() []string {
	return []string{
		"application/pdf",
	}
}
