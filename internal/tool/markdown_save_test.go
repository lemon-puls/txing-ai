package tool

import (
	"context"
	"strings"
	"testing"
)

// TestSaveMarkdownValidation 校验 saveMarkdown 的参数防护：
// - 空文件名 / 空内容必须被拒绝（避免生成 0 字节空文件浪费模型轮次）
// - 文件名含路径分隔符或 .. 必须被拒绝（防目录穿越）
// 校验发生在 buildSaveDir 之前，因此不依赖 runtime/config.yaml，可在测试环境运行
func TestSaveMarkdownValidation(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		params     *markdownSaveParams
		wantErrMsg string // 命中即认为被拒绝
	}{
		{
			name:       "空文件名被拒绝",
			params:     &markdownSaveParams{Filename: "", Content: "正文"},
			wantErrMsg: "文件名为空",
		},
		{
			name:       "空内容被拒绝",
			params:     &markdownSaveParams{Filename: "guide", Content: ""},
			wantErrMsg: "内容为空",
		},
		{
			name:       "空白内容被拒绝",
			params:     &markdownSaveParams{Filename: "guide", Content: "  \n "},
			wantErrMsg: "内容为空",
		},
		{
			name:       "文件名含路径分隔符被拒绝",
			params:     &markdownSaveParams{Filename: "../evil", Content: "正文"},
			wantErrMsg: "文件名不合法",
		},
		{
			name:       "文件名含反斜杠被拒绝",
			params:     &markdownSaveParams{Filename: `a\b`, Content: "正文"},
			wantErrMsg: "文件名不合法",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := saveMarkdown(ctx, tt.params)
			if err != nil {
				t.Fatalf("saveMarkdown() 不应返回 error，实际: %v", err)
			}
			if !strings.Contains(got, tt.wantErrMsg) {
				t.Errorf("saveMarkdown() = %q, 期望包含拒绝提示 %q", got, tt.wantErrMsg)
			}
		})
	}
}
