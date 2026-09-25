package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"

	"txing-ai/internal/agent/workflow/types"
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

// TestSplitToolCalls 验证工具调用按参数合法性分组，非法参数回喂可操作指引：
// - 空参数不再静默跳过
// - 大段正文导致 JSON 被截断时，PDF 工具的错误提示引导改用 filePath（落盘后转换）
// - 其他工具提示重新生成合法 JSON
func TestSplitToolCalls(t *testing.T) {
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
			name: "Markdown 保存工具截断 JSON 提示分段追加",
			toolCalls: []schema.ToolCall{
				{ID: "c5", Function: schema.FunctionCall{Name: "markdown_save_tool", Arguments: `{"content": "超长内容未闭合`}},
			},
			wantInvalid: true,
			wantHint:    "is_append",
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
			valid, msgs := splitToolCalls(tt.toolCalls)
			invalid := len(msgs) > 0
			if invalid != tt.wantInvalid {
				t.Fatalf("expected invalid=%v, got %v", tt.wantInvalid, invalid)
			}
			if tt.wantInvalid && len(msgs) == 0 {
				t.Fatal("expected error messages")
			}
			if !tt.wantInvalid && len(valid) != len(tt.toolCalls) {
				t.Fatalf("expected all %d calls valid, got %d", len(tt.toolCalls), len(valid))
			}
			if tt.wantInvalid && len(valid) != 0 {
				t.Fatalf("expected no valid calls, got %d", len(valid))
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

// TestExecuteWithRetryBailsOnCanceledCtx 验证上下文取消/超时后重试立即返回：
// 不再按重试策略 sleep 空转，让取消信号尽快传播到上层
// （用户点击停止生成时的响应速度，避免停止后仍在等待重试）
func TestExecuteWithRetryBailsOnCanceledCtx(t *testing.T) {
	calls := 0
	err := executeWithRetry(&types.RetryConfig{MaxRetries: 3, RetryDelay: 100}, func() error {
		calls++
		return context.Canceled
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call (bail on cancel), got %d", calls)
	}

	// 超时同样立即返回
	calls = 0
	err = executeWithRetry(&types.RetryConfig{MaxRetries: 3, RetryDelay: 100}, func() error {
		calls++
		return context.DeadlineExceeded
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call (bail on deadline), got %d", calls)
	}
}

// TestSplitToolCallsMixed 验证同一轮里合法与非法参数混合时的分组：
// 合法子集照常执行，非法子集逐个回喂错误（每个 tool_call_id 一条消息，
// 满足 provider 对 assistant.tool_calls 的完整 tool 响应校验）
func TestSplitToolCallsMixed(t *testing.T) {
	toolCalls := []schema.ToolCall{
		{ID: "a1", Function: schema.FunctionCall{Name: "t_valid", Arguments: `{"x":1}`}},
		{ID: "a2", Function: schema.FunctionCall{Name: "t_truncated", Arguments: `{"q": `}},
		{ID: "a3", Function: schema.FunctionCall{Name: "t_valid2", Arguments: `{"y":"ok"}`}},
		{ID: "a4", Function: schema.FunctionCall{Name: "t_empty", Arguments: ""}},
	}
	valid, errMsgs := splitToolCalls(toolCalls)
	if len(valid) != 2 {
		t.Fatalf("expected 2 valid calls, got %d", len(valid))
	}
	if valid[0].ID != "a1" || valid[1].ID != "a3" {
		t.Fatalf("unexpected valid ids: %s, %s", valid[0].ID, valid[1].ID)
	}
	if len(errMsgs) != 2 {
		t.Fatalf("expected 2 error messages, got %d", len(errMsgs))
	}
	ids := map[string]bool{}
	for _, m := range errMsgs {
		if m.Role != schema.Tool || m.ToolCallID == "" || m.Content == "" {
			t.Fatalf("bad error message: %+v", m)
		}
		ids[m.ToolCallID] = true
	}
	if !ids["a2"] || !ids["a4"] {
		t.Fatalf("expected error messages for a2/a4, got %v", ids)
	}
}

// TestToolErrorMessages 验证失败消息逐个 tool_call_id 生成（不能只喂首条）
func TestToolErrorMessages(t *testing.T) {
	toolCalls := []schema.ToolCall{
		{ID: "b1", Function: schema.FunctionCall{Name: "t1"}},
		{ID: "b2", Function: schema.FunctionCall{Name: "t2"}},
	}
	msgs := toolErrorMessages("工具执行失败: boom", toolCalls)
	if len(msgs) != len(toolCalls) {
		t.Fatalf("expected %d messages, got %d", len(toolCalls), len(msgs))
	}
	for i, m := range msgs {
		if m.ToolCallID != toolCalls[i].ID || m.ToolName != toolCalls[i].Function.Name {
			t.Fatalf("message %d mismatches call: %+v", i, m)
		}
	}
}
