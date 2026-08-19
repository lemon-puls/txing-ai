package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"

	"txing-ai/internal/global"
	"txing-ai/internal/global/logging"
)

// TestMain 初始化日志，避免测试中调用 log.Error/Warn 时全局 Logger 为 nil 导致 panic
func TestMain(m *testing.M) {
	logging.InitLogger(&global.LogConfig{
		Level:      "error",
		FileName:   filepath.Join(os.TempDir(), "txing-ai-workflow-test.log"),
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	}, "test")
	os.Exit(m.Run())
}

// TestValidateToolCallArgs 验证非法工具调用参数会被标记并回喂可操作指引：
// - 空参数不再静默跳过
// - 大段正文导致 JSON 被截断时，PDF 工具的错误提示引导改用 filePath（落盘后转换）
// - 其他工具提示重新生成合法 JSON
func TestValidateToolCallArgs(t *testing.T) {
	tests := []struct {
		name        string
		toolCalls   []schema.ToolCall
		wantInvalid bool
		wantHint    string
	}{
		{
			name: "合法参数不标记",
			toolCalls: []schema.ToolCall{
				{ID: "c1", Function: schema.FunctionCall{Name: "web_search_tool", Arguments: `{"q":"湛江"}`}},
			},
			wantInvalid: false,
		},
		{
			name: "空参数被标记并提示",
			toolCalls: []schema.ToolCall{
				{ID: "c2", Function: schema.FunctionCall{Name: "web_search_tool", Arguments: ""}},
			},
			wantInvalid: true,
			wantHint:    "参数为空",
		},
		{
			name: "PDF 工具截断 JSON 提示改用 filePath",
			toolCalls: []schema.ToolCall{
				{ID: "c3", Function: schema.FunctionCall{Name: "markdown_to_pdf_file_tool", Arguments: `{"content": "未闭合`}},
			},
			wantInvalid: true,
			wantHint:    "filePath",
		},
		{
			name: "其他工具截断 JSON 提示重新生成",
			toolCalls: []schema.ToolCall{
				{ID: "c4", Function: schema.FunctionCall{Name: "web_search_tool", Arguments: `{"q": `}},
			},
			wantInvalid: true,
			wantHint:    "重新生成",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs, invalid := validateToolCallArgs(tt.toolCalls)
			if invalid != tt.wantInvalid {
				t.Fatalf("expected invalid=%v, got %v", tt.wantInvalid, invalid)
			}
			if tt.wantInvalid && len(msgs) == 0 {
				t.Fatal("expected error messages")
			}
			if tt.wantHint != "" {
				var joined strings.Builder
				for _, m := range msgs {
					joined.WriteString(m.Content)
				}
				if !strings.Contains(joined.String(), tt.wantHint) {
					t.Fatalf("expected hint %q in %q", tt.wantHint, joined.String())
				}
			}
		})
	}
}
