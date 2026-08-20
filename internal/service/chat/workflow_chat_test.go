package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"txing-ai/internal/dto"
	"txing-ai/internal/global"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
)

// einoNodeError 模拟 eino compose.internalError 的包装结构：
// Error() 带 node path 尾巴，Unwrap() 返回原始错误（与真实错误链一致）
type einoNodeError struct {
	orig error
	path string
}

func (e *einoNodeError) Error() string {
	return fmt.Sprintf("[NodeRunError] %s\n------------------------\nnode path: [%s]", e.orig.Error(), e.path)
}

func (e *einoNodeError) Unwrap() error { return e.orig }

// TestFilterFinalArtifacts 验证产物过滤：生成 PDF 后剔除中间 Markdown 分片，
// 未生成 PDF 时保留全部产物（纯 Markdown 交付场景不受影响）
func TestFilterFinalArtifacts(t *testing.T) {
	md := func(name string) dto.ArtifactInfo {
		return dto.ArtifactInfo{Name: name, URL: "/api/file/download?filePath=" + name, Category: "markdown"}
	}
	pdf := func(name string) dto.ArtifactInfo {
		return dto.ArtifactInfo{Name: name, URL: "/api/file/download?filePath=" + name, Category: "pdf"}
	}
	img := func(name string) dto.ArtifactInfo {
		return dto.ArtifactInfo{Name: name, URL: "/api/file/download?filePath=" + name, Category: "image"}
	}

	tests := []struct {
		name      string
		artifacts []dto.ArtifactInfo
		wantLen   int
		wantPDF   bool
	}{
		{
			name: "有 PDF 时剔除 Markdown 分片",
			artifacts: []dto.ArtifactInfo{
				md("guide_part1.md"), md("guide_part2.md"), md("guide_part3.md"),
				pdf("guide.pdf"),
			},
			wantLen: 1,
			wantPDF: true,
		},
		{
			name: "有 PDF 时保留图片产物",
			artifacts: []dto.ArtifactInfo{
				md("guide_part1.md"), pdf("guide.pdf"), img("photo1.jpg"),
			},
			wantLen: 2,
			wantPDF: true,
		},
		{
			name: "无 PDF 时全部保留（纯 Markdown 交付）",
			artifacts: []dto.ArtifactInfo{
				md("guide.md"),
			},
			wantLen: 1,
		},
		{
			name:      "空产物保持为空",
			artifacts: nil,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterFinalArtifacts(tt.artifacts)
			if len(got) != tt.wantLen {
				t.Fatalf("expected %d artifacts, got %d: %+v", tt.wantLen, len(got), got)
			}
			if tt.wantPDF {
				found := false
				for _, a := range got {
					if a.Category == "pdf" {
						found = true
					}
					if a.Category == "markdown" {
						t.Fatalf("markdown artifact should be filtered out when pdf exists: %+v", a)
					}
				}
				if !found {
					t.Fatal("expected pdf artifact to remain")
				}
			}
		})
	}
}

// TestFriendlyWorkflowErrorMessage 验证工作流执行失败时，
// 用户看到的是提取根因后的友好提示（含原因与操作指引），
// 而不是 [NodeRunError]/node path 等内部包装噪音
func TestFriendlyWorkflowErrorMessage(t *testing.T) {
	// wrapLLM 模拟 llm_exec + retry 的逐层包装
	wrapLLM := func(err error) error {
		return fmt.Errorf("LLM 调用失败: %w", fmt.Errorf("执行失败（已重试 0 次）: %w", err))
	}
	// wrapEino 模拟 eino 节点错误包装（带 node path 尾巴）
	wrapEino := func(err error) error {
		return &einoNodeError{orig: err, path: "agent_travel"}
	}

	tests := []struct {
		name string
		err  error
		want string // 期望消息包含的子串
	}{
		{
			name: "401 invalid api key 给出配置指引",
			err: wrapEino(wrapLLM(&einoopenai.APIError{
				Message:        "Invalid API Key",
				HTTPStatus:     "401 Unauthorized",
				HTTPStatusCode: 401,
			})),
			want: "API Key 无效或已过期",
		},
		{
			name: "429 限流给出重试指引",
			err: wrapEino(wrapLLM(&einoopenai.APIError{
				Message:        "Rate limit reached",
				HTTPStatus:     "429 Too Many Requests",
				HTTPStatusCode: 429,
			})),
			want: "请求过于频繁",
		},
		{
			name: "404 模型不存在给出配置指引",
			err: wrapEino(wrapLLM(&einoopenai.APIError{
				Message:        "Model Not Found",
				HTTPStatus:     "404 Not Found",
				HTTPStatusCode: 404,
			})),
			want: "模型不存在",
		},
		{
			name: "5xx 服务不可用",
			err: wrapEino(wrapLLM(&einoopenai.APIError{
				Message:        "Internal Server Error",
				HTTPStatus:     "500 Internal Server Error",
				HTTPStatusCode: 500,
			})),
			want: "暂时不可用",
		},
		{
			name: "其他 4xx 保留根因消息与错误码",
			err: wrapEino(wrapLLM(&einoopenai.APIError{
				Message:        "context length exceeded",
				HTTPStatus:     "400 Bad Request",
				HTTPStatusCode: 400,
			})),
			want: "context length exceeded",
		},
		{
			name: "超时给出重试指引",
			err:  wrapEino(wrapLLM(context.DeadlineExceeded)),
			want: "响应超时",
		},
		{
			name: "网络错误给出检查指引",
			err:  wrapEino(wrapLLM(errors.New("dial tcp: connection refused"))),
			want: "无法连接模型服务",
		},
		{
			name: "未知错误剥离内部噪音提取根因",
			err: wrapEino(fmt.Errorf("LLM 调用失败: %w",
				fmt.Errorf("执行失败（已重试 2 次）: %w", errors.New("工具执行异常: boom")))),
			want: "工具执行异常: boom",
		},
		{
			name: "nil 错误返回通用提示且不 panic",
			err:  nil,
			want: "未知错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := friendlyWorkflowErrorMessage(tt.err)
			if !strings.Contains(msg, tt.want) {
				t.Fatalf("expected message to contain %q, got: %q", tt.want, msg)
			}
			// 友好提示不应泄漏内部包装噪音
			if strings.Contains(msg, "[NodeRunError]") || strings.Contains(msg, "node path") {
				t.Fatalf("friendly message should not contain internal noise, got: %q", msg)
			}
		})
	}
}

// TestExtractArtifactFromChunkWithRelPath 验证产物 URL 携带实际保存目录相对路径：
// 下载端按 2/2026-08-20/xxx.pdf 定位文件，避免仅凭文件名跨天猜日期目录 404
func TestExtractArtifactFromChunkWithRelPath(t *testing.T) {
	chunk := &global.Chunk{
		NodeId:     "agent_travel",
		ToolName:   "markdown_to_pdf_file_tool",
		ToolResult: "PDF已成功保存: ./广西北海一日游攻略.pdf",
	}
	a := extractArtifactFromChunk(chunk, "2/2026-08-20")
	if a == nil {
		t.Fatal("expected artifact")
	}
	if a.Name != "广西北海一日游攻略.pdf" {
		t.Fatalf("unexpected name: %q", a.Name)
	}
	// URL 应携带 相对目录/文件名 且经 URL 编码（文件名含中文）
	if !strings.Contains(a.URL, "filePath=2%2F2026-08-20%2F") &&
		!strings.Contains(a.URL, "filePath=2/2026-08-20/") {
		t.Fatalf("expected URL to carry rel path dir, got: %q", a.URL)
	}

	// 无相对目录（旧路径）：URL 退化为仅文件名
	b := extractArtifactFromChunk(chunk, "")
	if b == nil || b.URL != "/api/file/download?filePath=%E5%B9%BF%E8%A5%BF%E5%8C%97%E6%B5%B7%E4%B8%80%E6%97%A5%E6%B8%B8%E6%94%BB%E7%95%A5.pdf" {
		t.Fatalf("unexpected bare URL: %+v", b)
	}
}
