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
type PdfReadParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要读取的PDF文件路径"`
}

// 读取PDF文件内容
func ReadPdfText(ctx context.Context, params *PdfReadParams) (string, error) {
	// 先解析为绝对路径（兼容 ./xxx.pdf 相对保存目录 / runtime/xxx.pdf 相对工作目录）
	absPath, resolveErr := resolveSavedFile(ctx, params.FilePath)
	if resolveErr != nil {
		log.Error("解析PDF路径失败", zap.String("path", params.FilePath), zap.Error(resolveErr))
		return fmt.Sprintf("无法解析PDF路径: %s（%v）。仅允许访问 runtime 目录下的文件", params.FilePath, resolveErr), nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Error("文件不存在", zap.String("path", absPath))
		return fmt.Sprintf("文件不存在: %s。若 PDF 尚未生成，请先调用 markdown_to_pdf_file_tool 将 Markdown 转换为 PDF，再读取其返回的路径", params.FilePath), nil
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(absPath), ".pdf") {
		log.Error("文件不是PDF格式", zap.String("path", absPath))
		return fmt.Sprintf("文件不是PDF格式: %s", params.FilePath), nil
	}

	// 打开PDF文件
	f, r, err := pdf.Open(absPath)
	if err != nil {
		log.Error("打开PDF文件失败", zap.String("path", absPath), zap.Error(err))
		return fmt.Sprintf("打开PDF文件失败: %v", err), nil
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
	absPath, resolveErr := resolveSavedFile(ctx, params.FilePath)
	if resolveErr != nil {
		return "", fmt.Errorf("无法解析PDF路径: %s（%v）", params.FilePath, resolveErr)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", params.FilePath)
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(absPath), ".pdf") {
		return "", fmt.Errorf("文件不是PDF格式: %s", params.FilePath)
	}

	// 尝试打开PDF文件
	f, r, err := pdf.Open(absPath)
	if err != nil {
		log.Error("PDF文件验证失败", zap.String("path", absPath), zap.Error(err))
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
