package chat

import (
	"context"
	"sync"
	"time"

	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"

	"go.uber.org/zap"
)

// WorkflowSession 一次应用（工作流）执行的流式会话，与 WS 连接解耦：
// 客户端断开后工作流继续执行，输出与进度在服务端累积；
// 新连接可通过 resume 恢复（快照 + 进度回放 + 实时增量）。
type WorkflowSession struct {
	mu        sync.Mutex
	convId    int64
	consumers map[chunkSender]*streamConsumer
	done      chan struct{}
	cancel    context.CancelFunc

	// 累积状态（供 resume 快照与最终落库）
	currentContent string                 // 当前已累积的输出内容（resume 快照）
	progress       []dto.WorkflowProgress // 按序累积的工作流进度（节点/工具状态），resume 时回放
	content        string                 // 最终输出内容（结束后写入，写协程发送给客户端）
	artifacts      []dto.ArtifactInfo     // 最终产物
	rawErr         string                 // 执行失败时的原始错误（前端"查看详情"）
	status         string                 // running/completed/failed
	err            error
	finished       bool
	finishedAt     time.Time
	doneOnce       sync.Once
}

// WorkflowStreamManager 管理进行中的工作流执行会话，支持客户端刷新后恢复
type WorkflowStreamManager struct {
	mu       sync.Mutex
	sessions map[int64]*WorkflowSession
}

// 包级单例：所有 WS 连接共享，保证同一会话只能有一个进行中的工作流
var workflowStreamManager = &WorkflowStreamManager{sessions: make(map[int64]*WorkflowSession)}

func init() {
	go func() {
		ticker := time.NewTicker(sessionCleanupInterval)
		for range ticker.C {
			workflowStreamManager.cleanup()
		}
	}()
}

// Start 注册（或替换）会话的工作流执行任务；返回会话与解耦后的执行上下文。
// 若该会话已有进行中的工作流，会先取消旧流（同一会话不允许并行执行）。
// 执行上下文的生命周期不依赖 WS 连接：连接断开后工作流继续，直到完成或被取消。
func (m *WorkflowStreamManager) Start(convId int64) (*WorkflowSession, context.Context) {
	m.Cancel(convId)

	ctx, cancel := context.WithCancel(context.Background())
	ws := &WorkflowSession{
		convId:    convId,
		consumers: make(map[chunkSender]*streamConsumer),
		done:      make(chan struct{}),
		cancel:    cancel,
		status:    "running",
	}

	m.mu.Lock()
	m.sessions[convId] = ws
	m.mu.Unlock()

	return ws, ctx
}

// Get 获取会话的工作流（进行中或刚结束的）
func (m *WorkflowStreamManager) Get(convId int64) *WorkflowSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[convId]
}

// Cancel 取消会话的工作流执行（用户点击停止生成）
func (m *WorkflowStreamManager) Cancel(convId int64) {
	m.mu.Lock()
	ws := m.sessions[convId]
	m.mu.Unlock()
	if ws != nil {
		ws.cancel()
		ws.detachAll()
	}
}

// cleanup 清理超过保留期的已结束会话
func (m *WorkflowStreamManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, ws := range m.sessions {
		ws.mu.Lock()
		expired := ws.finished && now.Sub(ws.finishedAt) > sessionRetention
		ws.mu.Unlock()
		if expired {
			delete(m.sessions, id)
		}
	}
}

// Done 返回工作流结束通知通道
func (s *WorkflowSession) Done() <-chan struct{} { return s.done }

// Err 返回工作流结束时的错误（无错误为 nil）
func (s *WorkflowSession) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// appendContent 累积输出内容（resume 快照用）
func (s *WorkflowSession) appendContent(content string) {
	s.mu.Lock()
	s.currentContent += content
	s.mu.Unlock()
}

// appendProgress 累积一个工作流进度（resume 回放用）
func (s *WorkflowSession) appendProgress(p dto.WorkflowProgress) {
	s.mu.Lock()
	s.progress = append(s.progress, p)
	s.mu.Unlock()
}

// Attach 绑定一个客户端：
// 先发送 resume 快照（含已累积内容与 active 标记），再回放已累积的工作流进度
// （让前端重建节点/工具展示），最后启动写协程接收实时增量。
// 全新空会话不发送快照，直接建立消费者；重复 Attach 同一连接幂等返回。
func (s *WorkflowSession) Attach(conn chunkSender) error {
	s.mu.Lock()
	if c, ok := s.consumers[conn]; ok {
		// 连接已附加（如切换会话切回后同一连接再次 resume）：
		// 通过写协程重发 快照+进度回放，让前端重建展示；不重复注册消费者
		content := s.currentContent
		progress := append([]dto.WorkflowProgress(nil), s.progress...)
		s.mu.Unlock()
		select {
		case c.ch <- partialChunk{resync: &workflowResyncPayload{content: content, progress: progress}}:
		default:
			log.Warn("workflow resync backlog, skip",
				zap.Int64("conversationId", s.convId))
		}
		return nil
	}
	content := s.currentContent
	progress := append([]dto.WorkflowProgress(nil), s.progress...)
	finished := s.finished
	s.mu.Unlock()

	if content != "" || len(progress) > 0 || finished {
		if err := conn.Send(dto.WsMessageResponse{
			Type:           MsgTypeResume,
			Active:         !finished,
			Content:        content,
			ConversationId: s.convId,
		}); err != nil {
			return err
		}
	}

	if finished {
		return nil
	}

	// 先注册消费者：此后的实时增量进入通道缓冲，尚未启动写协程，不会直写连接
	c := &streamConsumer{
		conn: conn,
		ch:   make(chan partialChunk, streamConsumerBuffer),
		quit: make(chan struct{}),
	}
	s.mu.Lock()
	s.consumers[conn] = c
	s.mu.Unlock()

	// 回放已累积的工作流进度（写协程未启动，连接仍由本协程独占写入，
	// 保证 wire 顺序：快照 → 回放 → 实时增量）
	for i := range progress {
		resp := dto.WsMessageResponse{
			End:            false,
			ConversationId: s.convId,
			Workflow:       &progress[i],
		}
		if err := conn.Send(resp); err != nil {
			s.Detach(conn)
			return err
		}
	}

	// 启动写协程：此后实时增量由它下发
	go c.workflowWriteLoop(s)
	return nil
}

// Detach 摘除单个客户端（连接断开或写入失败时调用）
func (s *WorkflowSession) Detach(conn chunkSender) {
	s.mu.Lock()
	c, ok := s.consumers[conn]
	if ok {
		delete(s.consumers, conn)
	}
	s.mu.Unlock()
	if ok {
		close(c.quit)
	}
}

// detachAll 摘除所有客户端（会话被替换或取消时调用）
func (s *WorkflowSession) detachAll() {
	s.mu.Lock()
	consumers := s.consumers
	s.consumers = make(map[chunkSender]*streamConsumer)
	s.mu.Unlock()
	for _, c := range consumers {
		close(c.quit)
	}
}

// broadcast 向所有消费者广播一个消息块（非阻塞，慢消费者不拖慢执行）
func (s *WorkflowSession) broadcast(p partialChunk) {
	s.mu.Lock()
	consumers := make([]*streamConsumer, 0, len(s.consumers))
	for _, c := range s.consumers {
		consumers = append(consumers, c)
	}
	s.mu.Unlock()

	for _, c := range consumers {
		select {
		case c.ch <- p:
		default:
			log.Warn("workflow stream consumer backlog, drop chunk",
				zap.Int64("conversationId", s.convId))
		}
	}
}

// finish 工作流结束（正常/错误/取消）统一收尾：记录终态并通知消费者发送最终消息
func (s *WorkflowSession) finish(execErr error, content string, artifacts []dto.ArtifactInfo, rawErr string) {
	s.doneOnce.Do(func() {
		s.mu.Lock()
		s.err = execErr
		s.content = content
		s.artifacts = artifacts
		s.rawErr = rawErr
		if execErr != nil {
			s.status = "failed"
		} else {
			s.status = "completed"
		}
		s.finished = true
		s.finishedAt = time.Now()
		s.mu.Unlock()

		s.broadcast(partialChunk{End: true, Err: execErr})
		close(s.done)
	})
}

// workflowWriteLoop 消费者写入协程：把工作流进度增量/最终状态转发给客户端。
// 进度消息只带 Workflow 字段（与历史行为一致，正文最终一次性下发），
// 保证同一连接只有本协程在写。
func (c *streamConsumer) workflowWriteLoop(s *WorkflowSession) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("workflow stream consumer panic", zap.Any("err", r))
		}
		s.Detach(c.conn)
	}()

	for {
		select {
		case p := <-c.ch:
			if p.resync != nil {
				// resume 重同步：快照 + 进度回放（切换会话切回场景）。
				// 在写协程内发送，保持同一连接单写者，保证 快照→回放→后续增量 顺序
				resumeResp := dto.WsMessageResponse{
					Type:           MsgTypeResume,
					Active:         true,
					Content:        p.resync.content,
					ConversationId: s.convId,
				}
				if err := c.conn.Send(resumeResp); err != nil {
					return
				}
				for i := range p.resync.progress {
					if err := c.conn.Send(dto.WsMessageResponse{
						End:            false,
						ConversationId: s.convId,
						Workflow:       &p.resync.progress[i],
					}); err != nil {
						return
					}
				}
				continue
			}
			if p.End {
				s.mu.Lock()
				content := s.content
				status := s.status
				rawErr := s.rawErr
				artifacts := append([]dto.ArtifactInfo(nil), s.artifacts...)
				s.mu.Unlock()

				resp := dto.WsMessageResponse{
					Content:        content,
					End:            true,
					ConversationId: s.convId,
				}
				if status == "failed" {
					resp.Workflow = &dto.WorkflowProgress{Status: "failed", Error: rawErr}
				} else {
					resp.Workflow = &dto.WorkflowProgress{Status: "completed"}
				}
				if len(artifacts) > 0 {
					resp.Artifacts = artifacts
				}
				_ = c.conn.Send(resp)
				return
			}

			if p.Chunk != nil {
				// 发送前确认自己仍是该会话的消费者：
				// 会话被新工作流替换后，残留增量不应再发往连接
				s.mu.Lock()
				_, stillAttached := s.consumers[c.conn]
				s.mu.Unlock()
				if !stillAttached {
					return
				}

				resp := dto.WsMessageResponse{
					End:            false,
					ConversationId: s.convId,
				}
				if w := progressFromChunk(p.Chunk); w.NodeID != "" || w.NodeType != "" || w.ToolName != "" || w.ShowMsg != "" {
					resp.Workflow = w
				}
				if err := c.conn.Send(resp); err != nil {
					// 客户端已断开，摘除消费者（工作流本身继续执行，供后续 resume）
					log.Error("failed to send workflow message to client", zap.Error(err))
					return
				}
			}
		case <-c.quit:
			return
		}
	}
}

// HandleWorkflowResume 恢复进行中的工作流执行（客户端刷新重连后调用）。
// 工作流不存在或已结束时返回 active=false 的应答，前端据此更新界面
func HandleWorkflowResume(conn *utils.Connection, conversation *domain.Conversation) {
	ws := workflowStreamManager.Get(conversation.Id)
	if ws == nil {
		// 没有进行中的工作流：告知前端无需恢复
		_ = conn.Send(dto.WsMessageResponse{
			Type:           MsgTypeResume,
			Active:         false,
			ConversationId: conversation.Id,
		})
		return
	}
	if err := ws.Attach(conn); err != nil {
		log.Error("attach workflow stream consumer failed", zap.Error(err))
	}
}

// HandleResumeOrWorkflow 恢复进行中的流式输出：
// 优先恢复工作流会话（应用调用），否则回退普通聊天会话
func HandleResumeOrWorkflow(conn *utils.Connection, conversation *domain.Conversation) {
	if workflowStreamManager.Get(conversation.Id) != nil {
		HandleWorkflowResume(conn, conversation)
		return
	}
	HandleResume(conn, conversation)
}

// CancelWorkflowStream 取消会话进行中的工作流执行（用户点击停止生成）
func CancelWorkflowStream(convId int64) {
	workflowStreamManager.Cancel(convId)
}
