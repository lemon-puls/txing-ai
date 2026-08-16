package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"txing-ai/internal/global"
	"txing-ai/internal/global/logging"
)

// TestMain 初始化日志，避免测试中调用 log.Error 时全局 Logger 为 nil 导致 panic
func TestMain(m *testing.M) {
	logging.InitLogger(&global.LogConfig{
		Level:      "error",
		FileName:   filepath.Join(os.TempDir(), "txing-ai-domain-test.log"),
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	}, "test")
	os.Exit(m.Run())
}

// 构造 4 轮 user 消息（每轮都带一张图），中间穿插 assistant 回复
func conversationWithImageRounds() *Conversation {
	return &Conversation{
		Context: 10, // 保证窗口覆盖全部消息
		FormattedMessage: []global.Message{
			{Role: global.User, Content: "第一轮", Images: []string{"img1.png"}},
			{Role: global.Assistant, Content: "回复1"},
			{Role: global.User, Content: "第二轮", Images: []string{"img2.png"}},
			{Role: global.Assistant, Content: "回复2"},
			{Role: global.User, Content: "第三轮", Images: []string{"img3.png"}},
			{Role: global.Assistant, Content: "回复3"},
			{Role: global.User, Content: "第四轮", Images: []string{"img4.png"}},
		},
	}
}

// 只保留最近 2 轮 user 消息的图片，更早的剥离并加占位提示
func TestGetChatMessagesStripsOldImages(t *testing.T) {
	c := conversationWithImageRounds()
	msgs := c.GetChatMessages()

	if len(msgs) != 7 {
		t.Fatalf("expected 7 messages, got %d", len(msgs))
	}

	// 前两轮 user 消息：图片被剥离，内容追加占位提示
	for _, idx := range []int{0, 2} {
		m := msgs[idx]
		if len(m.Images) != 0 {
			t.Errorf("message %d: images should be stripped, got %v", idx, m.Images)
		}
		if !strings.Contains(m.Content, "已省略") {
			t.Errorf("message %d: content should contain omission hint, got %q", idx, m.Content)
		}
	}
	if msgs[0].Content != "第一轮\n（此消息原有 1 张图片，因对话较长已省略）" {
		t.Errorf("message 0: unexpected content %q", msgs[0].Content)
	}

	// 最近两轮 user 消息：图片原样保留
	for _, idx := range []int{4, 6} {
		m := msgs[idx]
		if len(m.Images) != 1 {
			t.Errorf("message %d: images should be kept, got %v", idx, m.Images)
		}
		if strings.Contains(m.Content, "已省略") {
			t.Errorf("message %d: content should not contain omission hint, got %q", idx, m.Content)
		}
	}

	// assistant 消息不受影响
	if msgs[1].Content != "回复1" || len(msgs[1].Images) != 0 {
		t.Errorf("assistant message altered: %+v", msgs[1])
	}
}

// 关键约束：剥离只作用于发送副本，持久化的 FormattedMessage 必须原样保留
func TestGetChatMessagesDoesNotTouchPersistedHistory(t *testing.T) {
	c := conversationWithImageRounds()
	_ = c.GetChatMessages()

	for i, m := range c.FormattedMessage {
		if m.Role != global.User {
			continue
		}
		if len(m.Images) != 1 {
			t.Errorf("persisted message %d: images lost after GetChatMessages, got %v", i, m.Images)
		}
		if strings.Contains(m.Content, "已省略") {
			t.Errorf("persisted message %d: content altered: %q", i, m.Content)
		}
	}
}

// 图片类附件一并剥离且计入数量，非图片附件保留
func TestStripOldImagesHandlesAttachments(t *testing.T) {
	c := &Conversation{
		Context: 10,
		FormattedMessage: []global.Message{
			{
				Role:    global.User,
				Content: "旧消息",
				Images:  []string{"img1.png"},
				Attachments: []global.Attachment{
					{FileName: "a.jpg", FileType: "image/jpeg"},
					{FileName: "b.pdf", FileType: "application/pdf"},
				},
			},
			{Role: global.User, Content: "新消息1"},
			{Role: global.User, Content: "新消息2"},
		},
	}
	msgs := c.GetChatMessages()

	m := msgs[0]
	if len(m.Images) != 0 {
		t.Errorf("images should be stripped, got %v", m.Images)
	}
	if len(m.Attachments) != 1 || m.Attachments[0].FileName != "b.pdf" {
		t.Errorf("non-image attachments should be kept, got %+v", m.Attachments)
	}
	if !strings.Contains(m.Content, "2 张图片") {
		t.Errorf("hint should count image + image attachment, got %q", m.Content)
	}
}

// 窗口内 user 消息不足 2 条时不做任何剥离
func TestStripOldImagesSkipsWhenFewUserMessages(t *testing.T) {
	c := &Conversation{
		Context: 10,
		FormattedMessage: []global.Message{
			{Role: global.User, Content: "唯一一轮", Images: []string{"img1.png"}},
			{Role: global.Assistant, Content: "回复"},
		},
	}
	msgs := c.GetChatMessages()

	if len(msgs[0].Images) != 1 {
		t.Errorf("images should be kept when fewer than %d user messages, got %v", imageKeepRounds, msgs[0].Images)
	}
	if strings.Contains(msgs[0].Content, "已省略") {
		t.Errorf("content should not be altered, got %q", msgs[0].Content)
	}
}

// 仅发图未发文字的消息：剥离后占位提示成为正文，避免产生空消息
func TestStripOldImagesFillsEmptyContent(t *testing.T) {
	c := &Conversation{
		Context: 10,
		FormattedMessage: []global.Message{
			{Role: global.User, Content: "", Images: []string{"img1.png", "img2.png"}},
			{Role: global.User, Content: "新消息1"},
			{Role: global.User, Content: "新消息2"},
		},
	}
	msgs := c.GetChatMessages()

	if msgs[0].Content != "（此消息原有 2 张图片，因对话较长已省略）" {
		t.Errorf("placeholder should become the content, got %q", msgs[0].Content)
	}
}

// 只有思考过程（reasoning 非空、content 为空）的响应也必须保存，
// 否则推理模型只输出思考时整条消息会被丢弃
func TestAddMessageFromAssistantKeepsReasoningOnly(t *testing.T) {
	c := &Conversation{}
	c.AddMessageFromAssistant("", "我先分析一下归并排序的思路")

	msgs := c.FormattedMessage
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != global.Assistant {
		t.Fatalf("unexpected role: %v", msgs[0].Role)
	}
	if msgs[0].Content != "" || msgs[0].ReasoningContent != "我先分析一下归并排序的思路" {
		t.Fatalf("unexpected message: %+v", msgs[0])
	}
}

// 正文与思考过程都为空时才跳过
func TestAddMessageFromAssistantSkipsBothEmpty(t *testing.T) {
	c := &Conversation{}
	c.AddMessageFromAssistant("", "")
	if len(c.FormattedMessage) != 0 {
		t.Fatalf("expected no message saved, got %d", len(c.FormattedMessage))
	}
}

// 正常响应不受影响
func TestAddMessageFromAssistantNormal(t *testing.T) {
	c := &Conversation{}
	c.AddMessageFromAssistant("归并排序是一种分治算法", "先分解再合并")
	if len(c.FormattedMessage) != 1 {
		t.Fatalf("expected 1 message, got %d", len(c.FormattedMessage))
	}
	m := c.FormattedMessage[0]
	if m.Content != "归并排序是一种分治算法" || m.ReasoningContent != "先分解再合并" {
		t.Fatalf("unexpected message: %+v", m)
	}
}
