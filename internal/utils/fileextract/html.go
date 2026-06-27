package fileextract

import (
	"context"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// HtmlExtractor HTML 文件提取器
type HtmlExtractor struct{}

// Extract 提取 HTML 文件中的文本内容
func (e *HtmlExtractor) Extract(ctx context.Context, filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	e.extractText(doc, &sb)
	return sb.String(), nil
}

// extractText 递归提取 HTML 节点中的文本
func (e *HtmlExtractor) extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}

	// 跳过 script 和 style 标签
	if n.Type == html.ElementNode {
		tag := strings.ToLower(n.Data)
		if tag == "script" || tag == "style" {
			return
		}
		// 在块级元素后添加换行
		if tag == "p" || tag == "div" || tag == "br" || tag == "h1" || tag == "h2" ||
			tag == "h3" || tag == "h4" || tag == "h5" || tag == "h6" || tag == "li" {
			sb.WriteString("\n")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.extractText(c, sb)
	}
}

// Supports 返回支持的 MIME 类型
func (e *HtmlExtractor) Supports() []string {
	return []string{
		"text/html",
		"application/xhtml+xml",
	}
}
