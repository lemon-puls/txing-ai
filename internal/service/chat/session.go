package chat

import (
	"context"
	"fmt"
	"sync"
	"time"

	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 续流相关常量
const (
	// MsgTypeResume 续流应答消息类型
	MsgTypeResume = "resume"

	// streamConsumerBuffer 单个消费者的增量缓冲大小
	streamConsumerBuffer = 512

	// sessionRetention 会话结束后的保留时长，期间仍可被 resume 查询到最终内容
	sessionRetention = 60 * time.Second

	// sessionCleanupInterval 清理协程的执行间隔
	sessionCleanupInterval = 30 * time.Second
)

// chunkSender 流消息发送接口，便于测试解耦；*utils.Connection 天然满足
type chunkSender interface {
	Send(msg dto.WsMessageResponse) error
}

// StreamSession 一次会话的流式生成任务，与 WS 连接解耦：
// 客户端断开后生成继续，内容在服务端累积；新连接可通过 resume 恢复接收
type StreamSession struct {
	mu           sync.Mutex
	convId       int64
	db           *gorm.DB
	conversation *domain.Conversation
	buffer       *utils.ChatRespBuffer
	consumers    map[chunkSender]*streamConsumer
	done         chan struct{}
	cancel       context.CancelFunc
	err          error
	finished     bool
	finishedAt   time.Time
	doneOnce     sync.Once
}

// streamConsumer 单个客户端的写入协程
type streamConsumer struct {
	conn chunkSender
	ch   chan partialChunk
	quit chan struct{}
}

// StreamManager 管理进行中的流式会话，支持客户端刷新后恢复
type StreamManager struct {
	mu       sync.Mutex
	sessions map[int64]*StreamSession
}

// 包级单例：所有 WS 连接共享，保证同一会话只能有一个进行中的流
var streamManager = &StreamManager{sessions: make(map[int64]*StreamSession)}

func init() {
	go func() {
		ticker := time.NewTicker(sessionCleanupInterval)
		for range ticker.C {
			streamManager.cleanup()
		}
	}()
}

// Start 启动（或替换）会话的流式生成任务。
// 若该会话已有进行中的流，会先取消旧流（同一会话不允许并行生成）。
// 流的生命周期不依赖 WS 连接：连接断开后生成继续，直到完成或被取消
func (m *StreamManager) Start(db *gorm.DB, conversation *domain.Conversation, config *adaptercommon.ChatConfig) *StreamSession {
	// 取消并摘除旧会话
	m.Cancel(conversation.Id)

	streamCtx, cancel := context.WithCancel(context.Background())
	s := &StreamSession{
		convId:       conversation.Id,
		db:           db,
		conversation: conversation,
		buffer:       utils.NewChatRespBuffer(),
		consumers:    make(map[chunkSender]*streamConsumer),
		done:         make(chan struct{}),
		cancel:       cancel,
	}

	m.mu.Lock()
	m.sessions[conversation.Id] = s
	m.mu.Unlock()

	// 启动生产协程：调用大模型流式生成，写入累积缓冲并广播增量
	go s.produce(streamCtx, config)

	return s
}

// Get 获取会话的流（进行中或刚结束的）
func (m *StreamManager) Get(convId int64) *StreamSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[convId]
}

// Cancel 取消会话的流式生成（用户点击停止生成）
func (m *StreamManager) Cancel(convId int64) {
	m.mu.Lock()
	s := m.sessions[convId]
	m.mu.Unlock()
	if s != nil {
		s.cancel()
		s.detachAll()
	}
}

// cleanup 清理超过保留期的已结束会话
func (m *StreamManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, s := range m.sessions {
		s.mu.Lock()
		expired := s.finished && now.Sub(s.finishedAt) > sessionRetention
		s.mu.Unlock()
		if expired {
			delete(m.sessions, id)
		}
	}
}

// Done 返回流结束通知通道
func (s *StreamSession) Done() <-chan struct{} { return s.done }

// Err 返回流结束时的错误（无错误为 nil）
func (s *StreamSession) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Finished 流是否已结束
func (s *StreamSession) Finished() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.finished
}

// Content 返回当前累积的完整内容（快照/保存用）
func (s *StreamSession) Content() (content, reasoningContent string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffer.Content, s.buffer.ReasoningContent
}

// Attach 绑定一个客户端：先回放当前累积内容快照，再持续推送实时增量。
// 若流已结束，仅回放最终快照（active=false），不建立消费者；
// 若该连接已是消费者（重复 resume），直接返回，避免同一连接多个写协程并发写
func (s *StreamSession) Attach(conn chunkSender) error {
	s.mu.Lock()
	if _, ok := s.consumers[conn]; ok {
		s.mu.Unlock()
		return nil
	}
	content := s.buffer.Content
	reasoning := s.buffer.ReasoningContent
	finished := s.finished
	s.mu.Unlock()

	// 先发送快照（保证 wire 顺序：快照 → 实时增量）
	if err := conn.Send(dto.WsMessageResponse{
		Type:             MsgTypeResume,
		Active:           !finished,
		Content:          content,
		ReasoningContent: reasoning,
		ConversationId:   s.convId,
	}); err != nil {
		return err
	}

	if finished {
		return nil
	}

	c := &streamConsumer{
		conn: conn,
		ch:   make(chan partialChunk, streamConsumerBuffer),
		quit: make(chan struct{}),
	}
	s.mu.Lock()
	s.consumers[conn] = c
	s.mu.Unlock()

	go c.writeLoop(s)
	return nil
}

// Detach 摘除单个客户端（连接断开或写入失败时调用）
func (s *StreamSession) Detach(conn chunkSender) {
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

// detachAll 摘除所有客户端（会话被新消息替换或取消时调用）
func (s *StreamSession) detachAll() {
	s.mu.Lock()
	consumers := s.consumers
	s.consumers = make(map[chunkSender]*streamConsumer)
	s.mu.Unlock()
	for _, c := range consumers {
		close(c.quit)
	}
}

// produce 生产协程：调用大模型流式生成，写入累积缓冲并广播增量
func (s *StreamSession) produce(ctx context.Context, config *adaptercommon.ChatConfig) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("chat stream produce panic", zap.Any("err", r))
			s.finish(fmt.Errorf("chat stream panic: %v", r))
		}
	}()

	err := NewChatRequest(ctx, s.db, config, func(chunk *global.Chunk) error {
		s.mu.Lock()
		s.buffer.WriteChunk(chunk)
		s.mu.Unlock()
		s.broadcast(partialChunk{Chunk: chunk, End: false, Err: nil})
		return nil
	})

	s.finish(err)
}

// finish 流结束（正常/错误/取消）统一收尾
func (s *StreamSession) finish(err error) {
	s.doneOnce.Do(func() {
		s.mu.Lock()
		s.err = err
		s.finished = true
		s.finishedAt = time.Now()
		s.mu.Unlock()

		// 通知所有仍在线的消费者发送结束标志
		s.broadcast(partialChunk{End: true, Err: err})
		close(s.done)
	})
}

// broadcast 向所有消费者广播一个消息块（非阻塞，慢消费者不拖慢生产）
func (s *StreamSession) broadcast(p partialChunk) {
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
			// 消费者积压过慢（大概率已断开且尚未摘除），跳过本次增量。
			// 断开的消费者会在写入失败后自行摘除，不阻塞流本身
			log.Warn("stream consumer backlog, drop chunk",
				zap.Int64("conversationId", s.convId))
		}
	}
}

// writeLoop 消费者写入协程：把消息块转发给客户端
func (c *streamConsumer) writeLoop(s *StreamSession) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("stream consumer panic", zap.Any("err", r))
		}
		s.Detach(c.conn)
	}()

	for {
		select {
		case p := <-c.ch:
			if p.End {
				if p.Err != nil {
					// 流中途出错：发送友好提示（不含模型名时回退通用提示）
					model := ""
					if s.conversation != nil {
						model = s.conversation.Model
					}
					_ = c.conn.Send(dto.WsMessageResponse{
						Content:        friendlyChatErrorMessage(p.Err, model),
						End:            true,
						ConversationId: s.convId,
					})
				} else {
					_ = c.conn.Send(dto.WsMessageResponse{
						End:            true,
						ConversationId: s.convId,
					})
				}
				return
			}

			if p.Chunk != nil {
				// 发送前确认自己仍是该会话的消费者：
				// 会话被新消息替换（supersede）后，残留增量不应再发往连接
				s.mu.Lock()
				_, stillAttached := s.consumers[c.conn]
				s.mu.Unlock()
				if !stillAttached {
					return
				}

				if err := c.conn.Send(dto.WsMessageResponse{
					Content:          p.Chunk.Content,
					ReasoningContent: p.Chunk.ReasoningContent,
					End:              false,
					ConversationId:   s.convId,
				}); err != nil {
					// 客户端已断开，摘除消费者（流本身继续生成，供后续 resume）
					log.Error("failed to send message to client", zap.Error(err))
					return
				}
			}
		case <-c.quit:
			return
		}
	}
}
