package chat

import (
	"fmt"
	"strings"
	"testing"

	"txing-ai/internal/service/channel"
)

// TestFriendlyChatErrorMessage 验证渠道相关错误能转换为
// 包含原因和操作指引的中文提示，而不是笼统的 "System error"
func TestFriendlyChatErrorMessage(t *testing.T) {
	model := "deepseek-v3-250324"

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "no channel found",
			err:  fmt.Errorf("%w: %s", channel.ErrNoChannelFound, model),
			want: "暂无可用",
		},
		{
			name: "no available channel",
			err:  fmt.Errorf("%w: %s", channel.ErrNoAvailableChannel, model),
			want: "联网搜索",
		},
		{
			name: "unknown error falls back to generic message",
			err:  fmt.Errorf("some other error"),
			want: defaultErrRespMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := friendlyChatErrorMessage(tt.err, model)
			if !strings.Contains(msg, tt.want) {
				t.Fatalf("expected message to contain %q, got: %q", tt.want, msg)
			}
			if tt.name == "unknown error falls back to generic message" {
				if msg != defaultErrRespMessage {
					t.Fatalf("expected generic message %q, got: %q", defaultErrRespMessage, msg)
				}
				return
			}
			// 友好提示应包含模型名，方便用户定位问题
			if !strings.Contains(msg, model) {
				t.Fatalf("expected message to contain model %q, got: %q", model, msg)
			}
		})
	}
}
