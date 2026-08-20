package chat

import (
	"testing"
	"time"

	"txing-ai/internal/dto"
	"txing-ai/internal/global"
)

// TestProgressFromChunk 验证执行 chunk 与前端工作流进度消息的转换
func TestProgressFromChunk(t *testing.T) {
	chunk := &global.Chunk{
		NodeId:     "agent_travel",
		NodeType:   "agent",
		NodeLabel:  "旅游攻略生成",
		NodeStatus: "running",
		ToolName:   "web_search_tool",
		ToolCallId: "call_1",
		ToolParams: `{"q":"湛江"}`,
		ToolStatus: "running",
	}
	p := progressFromChunk(chunk)
	if p.NodeID != "agent_travel" || p.ToolName != "web_search_tool" || p.Status != "running" {
		t.Fatalf("unexpected progress: %+v", p)
	}
	if p.ToolArgs != `{"q":"湛江"}` {
		t.Fatalf("expected ToolArgs, got %q", p.ToolArgs)
	}
}

// TestWorkflowSessionAttachFreshNoSnapshot 验证全新工作流会话 attach 时：
// 不发送 resume 快照（避免干扰普通消息流），直接建立消费者，后续增量正常送达
func TestWorkflowSessionAttachFreshNoSnapshot(t *testing.T) {
	ws := &WorkflowSession{
		convId:    1,
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		status:    "running",
	}
	fake := &fakeSender{}
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer ws.Detach(fake)

	if msgs := fake.messages(); len(msgs) != 0 {
		t.Fatalf("expected no snapshot for fresh session, got %d: %+v", len(msgs), msgs)
	}

	// 广播增量应到达消费者
	ws.broadcast(partialChunk{Chunk: &global.Chunk{NodeId: "start_1", NodeStatus: "running"}})
	if !waitUntil(2*time.Second, func() bool {
		return len(fake.messages()) >= 1
	}) {
		t.Fatal("expected live chunk to reach consumer")
	}
	msg := fake.messages()[0]
	if msg.End {
		t.Fatalf("expected non-end message, got %+v", msg)
	}
	if msg.Workflow == nil || msg.Workflow.NodeID != "start_1" {
		t.Fatalf("expected workflow progress message, got %+v", msg)
	}
}

// TestWorkflowSessionAttachReplaysProgress 验证已有累积进度时 attach：
// 先发 resume 快照（含累积内容与 active 标记），再回放进度，最后接收实时增量
func TestWorkflowSessionAttachReplaysProgress(t *testing.T) {
	ws := &WorkflowSession{
		convId:    1,
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		status:    "running",
	}
	ws.appendContent("已生成部分内容")
	ws.appendProgress(dto.WorkflowProgress{Status: "running", NodeID: "start_1", NodeStatus: "running"})
	ws.appendProgress(dto.WorkflowProgress{Status: "running", NodeID: "agent_travel", NodeStatus: "running", ToolName: "web_search_tool", ToolStatus: "running"})

	fake := &fakeSender{}
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer ws.Detach(fake)

	msgs := fake.messages()
	if len(msgs) != 3 {
		t.Fatalf("expected resume + 2 replay, got %d: %+v", len(msgs), msgs)
	}
	if msgs[0].Type != "resume" || !msgs[0].Active {
		t.Fatalf("expected active resume snapshot, got %+v", msgs[0])
	}
	if msgs[0].Content != "已生成部分内容" {
		t.Fatalf("expected snapshot content, got %q", msgs[0].Content)
	}
	if msgs[1].Workflow == nil || msgs[1].Workflow.NodeID != "start_1" {
		t.Fatalf("expected replay of start node, got %+v", msgs[1])
	}
	if msgs[2].Workflow == nil || msgs[2].Workflow.ToolName != "web_search_tool" {
		t.Fatalf("expected replay of tool progress, got %+v", msgs[2])
	}

	// 实时增量继续送达
	ws.broadcast(partialChunk{Chunk: &global.Chunk{NodeId: "agent_travel", ToolName: "web_search_tool", ToolStatus: "completed"}})
	if !waitUntil(2*time.Second, func() bool { return len(fake.messages()) >= 4 }) {
		t.Fatal("expected live chunk after replay")
	}
	live := fake.messages()[3]
	if live.Workflow == nil || live.Workflow.ToolStatus != "completed" {
		t.Fatalf("expected live tool progress, got %+v", live)
	}
}

// TestWorkflowSessionFinishBroadcastsFinal 验证 finish 后消费者收到
// 携带最终正文、工作流状态与产物的结束消息
func TestWorkflowSessionFinishBroadcastsFinal(t *testing.T) {
	ws := &WorkflowSession{
		convId:    1,
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		status:    "running",
	}
	fake := &fakeSender{}
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer ws.Detach(fake)

	artifacts := []dto.ArtifactInfo{{Name: "guide.pdf", URL: "/api/file/download?filePath=guide.pdf", Category: "pdf"}}
	ws.finish(nil, "攻略已生成", artifacts, "")

	<-ws.Done()

	if !waitUntil(2*time.Second, func() bool { return len(fake.messages()) >= 1 }) {
		t.Fatal("expected final message")
	}
	msgs := fake.messages()
	last := msgs[len(msgs)-1]
	if !last.End {
		t.Fatalf("expected end message, got %+v", last)
	}
	if last.Content != "攻略已生成" {
		t.Fatalf("expected final content, got %q", last.Content)
	}
	if last.Workflow == nil || last.Workflow.Status != "completed" {
		t.Fatalf("expected completed workflow status, got %+v", last.Workflow)
	}
	if len(last.Artifacts) != 1 || last.Artifacts[0].Name != "guide.pdf" {
		t.Fatalf("expected artifacts, got %+v", last.Artifacts)
	}
}

// TestWorkflowSessionAttachResync 验证已附加连接再次 Attach（切换会话切回后 resume）：
// 通过写协程重发 resume 快照 + 进度回放，让前端重建展示，且不重复注册消费者
func TestWorkflowSessionAttachResync(t *testing.T) {
	ws := &WorkflowSession{
		convId:    1,
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		status:    "running",
	}
	fake := &fakeSender{}
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer ws.Detach(fake)

	// 执行一段时间后累积了内容与进度
	ws.appendContent("部分内容")
	ws.appendProgress(dto.WorkflowProgress{Status: "running", NodeID: "start_1", NodeStatus: "running"})

	// 同一连接再次 Attach：应重发快照 + 回放
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}

	if !waitUntil(2*time.Second, func() bool { return len(fake.messages()) >= 2 }) {
		t.Fatalf("expected resync snapshot + replay, got %d: %+v", len(fake.messages()), fake.messages())
	}
	msgs := fake.messages()
	if msgs[0].Type != "resume" || !msgs[0].Active || msgs[0].Content != "部分内容" {
		t.Fatalf("expected resume snapshot with content, got %+v", msgs[0])
	}
	if msgs[1].Workflow == nil || msgs[1].Workflow.NodeID != "start_1" {
		t.Fatalf("expected progress replay, got %+v", msgs[1])
	}

	// 重同步后实时增量仍正常送达
	ws.broadcast(partialChunk{Chunk: &global.Chunk{NodeId: "agent_travel", ToolName: "web_search_tool", ToolStatus: "completed"}})
	if !waitUntil(2*time.Second, func() bool { return len(fake.messages()) >= 3 }) {
		t.Fatal("expected live chunk after resync")
	}
}

// TestWorkflowSessionAttachResyncNotDropped 验证消费者通道已满时再次 Attach：
// 重同步必须阻塞等待写协程腾出空间并最终送达，而不是被静默丢弃。
// 回归场景：应用执行中切走会话再切回，通道里堆满实时增量时 resume 的
// 快照+进度回放一旦丢失，前端节点/工具调用列表将无法重建
func TestWorkflowSessionAttachResyncNotDropped(t *testing.T) {
	ws := &WorkflowSession{
		convId:    1,
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		status:    "running",
	}
	fake := &fakeSender{}
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}
	defer ws.Detach(fake)

	// 填满消费者通道，模拟流式执行中增量堆积（写协程尚未消费）
	c := ws.consumers[fake]
	ws.appendContent("部分内容")
	ws.appendProgress(dto.WorkflowProgress{Status: "running", NodeID: "start_1", NodeStatus: "running"})
	for i := 0; i < streamConsumerBuffer; i++ {
		c.ch <- partialChunk{Chunk: &global.Chunk{NodeId: "n", NodeStatus: "running"}}
	}

	// 同一连接再次 Attach：即使通道已满，阻塞发送也必须在写协程
	// 消费掉若干增量后把重同步送达，而不是走 default 分支静默丢弃
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}

	// 等待重同步送达（快照 + 回放）
	if !waitUntil(3*time.Second, func() bool {
		for _, m := range fake.messages() {
			if m.Type == "resume" && m.Content == "部分内容" {
				return true
			}
		}
		return false
	}) {
		t.Fatalf("expected resync to be delivered despite full channel, got %d messages", len(fake.messages()))
	}
}

// TestWorkflowStreamManagerCancel 验证取消会结束执行上下文并摘除消费者
func TestWorkflowStreamManagerCancel(t *testing.T) {
	ws, ctx := workflowStreamManager.Start(999001)
	fake := &fakeSender{}
	if err := ws.Attach(fake); err != nil {
		t.Fatal(err)
	}

	workflowStreamManager.Cancel(999001)

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("expected context cancelled")
	}

	// 消费者已摘除：后续广播不应再送达
	ws.broadcast(partialChunk{Chunk: &global.Chunk{NodeId: "start_1"}})
	time.Sleep(50 * time.Millisecond)
	if len(fake.messages()) != 0 {
		t.Fatalf("expected no messages after cancel, got %d", len(fake.messages()))
	}
}
