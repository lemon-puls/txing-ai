package utils

import (
	"time"
	"txing-ai/internal/global"
)

// 聊天响应消息缓冲区：用于存储大模型的输出（单条消息），每次回答都创建新的
type ChatRespBuffer struct {
	// 消息内容
	Content string `json:"content"`
	// 思考过程消息
	ReasoningContent string `json:"reasoning_content"`
	// 最后的消息块
	Last string `json:"last"`
	// 消息块数
	Count int `json:"count"`
	// 消息开始时间
	StartTime *time.Time `json:"-"`
}

func NewChatRespBuffer() *ChatRespBuffer {
	now := time.Now()
	return &ChatRespBuffer{
		StartTime: &now,
	}
}

func (b *ChatRespBuffer) WriteChunk(chunk *global.Chunk) (content, reasoningContent string) {
	if chunk == nil {
		return
	}

	if chunk.Content != "" {
		b.Write(chunk.Content)
		content = chunk.Content
	}
	if chunk.ReasoningContent != "" {
		b.WriteReasoningContent(chunk.ReasoningContent)
		reasoningContent = chunk.ReasoningContent
	}

	return content, reasoningContent
}

// 写入消息
func (b *ChatRespBuffer) Write(content string) {
	b.Last = content
	b.Count++
	b.Content += content
}

// 写入思考过程消息
func (b *ChatRespBuffer) WriteReasoningContent(reasoningContent string) {
	b.ReasoningContent += reasoningContent
}

func (b *ChatRespBuffer) IsEmpty() bool {
	return len(b.Content) == 0
}

// 获取消息 （如果消息为空，则返回默认消息）
func (b *ChatRespBuffer) GetOrDefault(defaultMessage string) (content, reasoningContent string) {
	if b.IsEmpty() {
		return defaultMessage, ""
	}
	return b.Content, b.ReasoningContent
}
