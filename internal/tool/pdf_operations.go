package tool

import (
	"context"
	"fmt"
	"os"
	"strings"
	"txing-ai/internal/global/logging/log"

	"github.com/ledongthuc/pdf"
	"go.uber.org/zap"
)

// PDF文件读取参数
type pdfReadParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要读取的PDF文件路径"`
}

// 读取PDF文件内容
func readPdfText(ctx context.Context, params *pdfReadParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(params.FilePath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.FilePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(params.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", params.FilePath)
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(params.FilePath), ".pdf") {
		return "", fmt.Errorf("文件不是PDF格式: %s", params.FilePath)
	}

	// 打开PDF文件
	f, r, err := pdf.Open(params.FilePath)
	if err != nil {
		log.Error("打开PDF文件失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("打开PDF文件失败: %v", err)
	}
	defer f.Close()

	// 提取文本内容
	var textBuilder strings.Builder
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			log.Error("提取PDF页面文本失败",
				zap.String("path", params.FilePath),
				zap.Int("page", pageIndex),
				zap.Error(err))
			continue
		}

		textBuilder.WriteString(fmt.Sprintf("--- 第 %d 页 ---\n", pageIndex))
		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n")
	}

	return textBuilder.String(), nil
}

// 验证PDF文件参数
type pdfValidateParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要验证的PDF文件路径"`
}

// 验证PDF文件
func validatePdf(ctx context.Context, params *pdfValidateParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(params.FilePath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.FilePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(params.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", params.FilePath)
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(params.FilePath), ".pdf") {
		return "", fmt.Errorf("文件不是PDF格式: %s", params.FilePath)
	}

	// 尝试打开PDF文件
	f, r, err := pdf.Open(params.FilePath)
	if err != nil {
		log.Error("PDF文件验证失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("无效的PDF文件: %v", err)
	}
	defer f.Close()

	// 检查页数
	pageCount := r.NumPage()
	if pageCount <= 0 {
		return "", fmt.Errorf("PDF文件无有效页面")
	}

	return fmt.Sprintf("PDF文件有效，共 %d 页", pageCount), nil
}
