package fileextract

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTxtExtractor(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!\nThis is a test file."
	os.WriteFile(tmpFile, []byte(content), 0644)

	extractor := &TxtExtractor{}
	result, err := extractor.Extract(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if result != content {
		t.Errorf("Expected %q, got %q", content, result)
	}
}

func TestMarkdownExtractor(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	content := "# Title\n\nThis is **bold** text."
	os.WriteFile(tmpFile, []byte(content), 0644)

	extractor := &MarkdownExtractor{}
	result, err := extractor.Extract(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if result != content {
		t.Errorf("Expected %q, got %q", content, result)
	}
}

func TestHtmlExtractor(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `<html><body><h1>Title</h1><p>Hello <b>World</b></p></body></html>`
	os.WriteFile(tmpFile, []byte(htmlContent), 0644)

	extractor := &HtmlExtractor{}
	result, err := extractor.Extract(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	t.Logf("Extracted HTML text: %q", result)
	// HTML 提取应该包含文本内容
	if result == "" {
		t.Error("Expected non-empty result from HTML extraction")
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&TxtExtractor{})
	registry.Register(&PdfExtractor{})

	// 测试支持的类型
	if !registry.IsSupported("text/plain") {
		t.Error("Expected text/plain to be supported")
	}
	if !registry.IsSupported("application/pdf") {
		t.Error("Expected application/pdf to be supported")
	}
	if registry.IsSupported("application/msword") {
		t.Error("Expected application/msword to NOT be supported")
	}

	// 测试获取提取器
	ext, err := registry.GetExtractor("text/plain")
	if err != nil {
		t.Fatalf("GetExtractor failed: %v", err)
	}
	if ext == nil {
		t.Error("Expected non-nil extractor")
	}

	// 测试不支持的类型
	_, err = registry.GetExtractor("application/msword")
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// 测试默认注册表是否包含所有内置提取器
	supportedTypes := []string{
		"text/plain",
		"text/markdown",
		"text/html",
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}

	for _, mimeType := range supportedTypes {
		if !Default.IsSupported(mimeType) {
			t.Errorf("Expected %s to be supported in default registry", mimeType)
		}
	}
}
