package chat

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging"
	"txing-ai/internal/service/channel"
	"txing-ai/internal/utils"
)

// TestMain 初始化日志，避免测试中调用 log.Error 时全局 Logger 为 nil 导致 panic
func TestMain(m *testing.M) {
	logging.InitLogger(&global.LogConfig{
		Level:      "error",
		FileName:   filepath.Join(os.TempDir(), "txing-ai-chat-test.log"),
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
	}, "test")
	os.Exit(m.Run())
}

// fakeSender 测试用发送器：记录消息，可模拟连接断开
type fakeSender struct {
	mu   sync.Mutex
	msgs []dto.WsMessageResponse
	fail bool
}

func (f *fakeSender) Send(msg dto.WsMessageResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return errors.New("connection closed")
	}
	f.msgs = append(f.msgs, msg)
	return nil
}

func (f *fakeSender) messages() []dto.WsMessageResponse {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]dto.WsMessageResponse(nil), f.msgs...)
}

func newTestSession() *StreamSession {
	return &StreamSession{
		convId:    1,
		buffer:    utils.NewChatRespBuffer(),
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		cancel:    func() {},
	}
}

// waitUntil 轮询等待条件满足，超时返回 false
func waitUntil(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

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

// TestSessionAttachEmptyBufferNoSnapshot 验证全新空会话 attach 时：
// 不发送 resume 快照（避免普通消息触发前端续流逻辑），
// 消费者直接建立，后续增量正常送达
func TestSessionAttachEmptyBufferNoSnapshot(t *testing.T) {
	s := newTestSession()

	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer s.Detach(fake)

	// 空会话不应发送任何快照消息
	if msgs := fake.messages(); len(msgs) != 0 {
		t.Fatalf("expected no snapshot for empty session, got %d messages: %+v", len(msgs), msgs)
	}

	// 消费者应已建立，后续增量正常送达
	s.broadcast(partialChunk{Chunk: &global.Chunk{Content: "归并"}, End: false})
	if !waitUntil(2*time.Second, func() bool { return len(fake.messages()) == 1 }) {
		t.Fatalf("expected 1 chunk message, got %d", len(fake.messages()))
	}
	if got := fake.messages()[0]; got.Content != "归并" || got.End {
		t.Fatalf("unexpected chunk message: %+v", got)
	}
}

// TestSessionAttachSnapshotActive 验证进行中的流 attach 时：
// 先收到包含累积内容的 resume 快照（active=true），再持续接收增量
func TestSessionAttachSnapshotActive(t *testing.T) {
	s := newTestSession()
	s.buffer.Write("你好")
	s.buffer.WriteReasoningContent("思考中")

	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer s.Detach(fake)

	msgs := fake.messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 snapshot message, got %d", len(msgs))
	}
	snap := msgs[0]
	if snap.Type != MsgTypeResume || !snap.Active {
		t.Fatalf("expected resume snapshot active=true, got %+v", snap)
	}
	if snap.Content != "你好" || snap.ReasoningContent != "思考中" {
		t.Fatalf("unexpected snapshot content: %+v", snap)
	}
}

// TestSessionAttachFinished 验证已结束的流 attach 时：
// 仅回放最终快照（active=false），不建立消费者
func TestSessionAttachFinished(t *testing.T) {
	s := newTestSession()
	s.buffer.Write("最终回答")
	s.finish(nil)

	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}

	msgs := fake.messages()
	if len(msgs) != 1 {
		t.Fatalf("expected only snapshot message, got %d", len(msgs))
	}
	if msgs[0].Active {
		t.Fatal("expected active=false for finished session")
	}
	if msgs[0].Content != "最终回答" {
		t.Fatalf("unexpected snapshot content: %+v", msgs[0])
	}
}

// TestSessionBroadcastAndEnd 验证广播增量与结束标志：
// 空会话 attach 不发送快照，直接收到增量 chunk，流结束时收到 End 消息
func TestSessionBroadcastAndEnd(t *testing.T) {
	s := newTestSession()
	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer s.Detach(fake)

	s.broadcast(partialChunk{Chunk: &global.Chunk{Content: "你"}, End: false})
	s.broadcast(partialChunk{Chunk: &global.Chunk{Content: "好"}, End: false})
	s.finish(nil)

	if !waitUntil(2*time.Second, func() bool { return len(fake.messages()) >= 3 }) {
		t.Fatalf("expected 2 chunks + end, got %d messages: %+v", len(fake.messages()), fake.messages())
	}

	msgs := fake.messages()
	// msgs[0]=你, msgs[1]=好, msgs[2]=End
	if msgs[0].Content != "你" || msgs[1].Content != "好" {
		t.Fatalf("unexpected chunk messages: %+v", msgs)
	}
	if !msgs[2].End {
		t.Fatalf("expected last message End=true, got %+v", msgs[2])
	}
	if !s.Finished() {
		t.Fatal("expected session finished after finish")
	}
	select {
	case <-s.Done():
	default:
		t.Fatal("expected done channel closed after finish")
	}
}

// TestSessionConsumerDisconnectKeepsStream 验证客户端断开场景：
// 消费者写入失败后自动摘除，流本身继续生成（不 panic、不阻塞）
func TestSessionConsumerDisconnectKeepsStream(t *testing.T) {
	s := newTestSession()
	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}

	// 模拟连接断开：快照已发出，后续写入失败
	fake.mu.Lock()
	fake.fail = true
	fake.mu.Unlock()

	// 发送一个 chunk，写协程发送失败后应摘除消费者
	s.broadcast(partialChunk{Chunk: &global.Chunk{Content: "hi"}, End: false})

	if !waitUntil(2*time.Second, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return len(s.consumers) == 0
	}) {
		t.Fatal("consumer not detached after send failure")
	}

	// 流仍可继续广播（无消费者也不 panic、不阻塞）
	s.broadcast(partialChunk{Chunk: &global.Chunk{Content: "still streaming"}, End: false})
	s.finish(nil)

	if !waitUntil(2*time.Second, func() bool { return s.Finished() }) {
		t.Fatal("session should finish normally")
	}
}

// TestSessionAttachTwiceNoDuplicate 验证同一连接重复 attach 是幂等的：
// 不会建立第二个消费者，也不会重复发送快照（避免同一连接多写协程并发写）
func TestSessionAttachTwiceNoDuplicate(t *testing.T) {
	s := newTestSession()
	s.buffer.Write("你好")

	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}

	if !waitUntil(2*time.Second, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return len(s.consumers) == 1
	}) {
		t.Fatal("expected exactly one consumer after duplicate attach")
	}

	msgs := fake.messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 snapshot message, got %d", len(msgs))
	}
	s.Detach(fake)
}

// TestStreamManagerGetAndCancel 验证会话注册表的查询与取消
func TestStreamManagerGetAndCancel(t *testing.T) {
	m := &StreamManager{sessions: make(map[int64]*StreamSession)}
	s := newTestSession()
	m.sessions[1] = s

	if got := m.Get(1); got != s {
		t.Fatal("Get returned wrong session")
	}
	if got := m.Get(2); got != nil {
		t.Fatal("Get should return nil for missing conversation")
	}

	cancelled := false
	s.cancel = func() { cancelled = true }
	m.Cancel(1)
	if !cancelled {
		t.Fatal("Cancel did not call session cancel")
	}
}

// TestStreamManagerCancelDeliversEnd 验证用户停止生成（普通聊天）时：
// 消费者不被提前摘除，生产协程收尾（finish）后 End 消息送达前端收尾，
// 前端据此用已累积的部分内容收尾，而不是永远停留在"生成中"
func TestStreamManagerCancelDeliversEnd(t *testing.T) {
	s := newTestSession()
	// 累积部分内容，模拟停止前已生成的内容
	s.buffer.Write("已生成的部分回答")
	fake := &fakeSender{}
	if err := s.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer s.Detach(fake)

	m := &StreamManager{sessions: map[int64]*StreamSession{1: s}}
	m.Cancel(1)

	s.finish(context.Canceled)

	<-s.Done()

	// 第一条是 Attach 发送的 resume 快照，随后应收到 End 终态消息
	if !waitUntil(2*time.Second, func() bool {
		for _, msg := range fake.messages() {
			if msg.End {
				return true
			}
		}
		return false
	}) {
		t.Fatalf("expected end message after cancel, got %d messages", len(fake.messages()))
	}
}

// TestSessionContent 验证累积内容可读取（含思考过程）
func TestSessionContent(t *testing.T) {
	s := newTestSession()
	s.buffer.Write("回答内容")
	s.buffer.WriteReasoningContent("思考内容")

	content, reasoning := s.Content()
	if content != "回答内容" || reasoning != "思考内容" {
		t.Fatalf("unexpected content: %q / %q", content, reasoning)
	}
}
