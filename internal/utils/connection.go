package utils

import (
	"go.uber.org/zap"
	"txing-ai/internal/dto"
	"txing-ai/internal/global/logging/log"
)

type Stack chan *dto.WsMessageRequest

// 偏上层（业务层）的连接对象，通过 WebSocket 连接拿到客户端的数据，并丢给业务层处理
// 或者把业务层的数据通过 WebSocket 发送给客户端
type Connection struct {
	conn  *WebSocket
	auth  bool
	stack Stack
}

func NewConnection(conn *WebSocket, auth bool, bufferSize int) *Connection {
	return &Connection{
		conn:  conn,
		auth:  auth,
		stack: make(Stack, bufferSize),
	}
}

// 处理消息循环，从消息栈中读取消息，并处理
func (c *Connection) HandleMessageLoop(handler func(*dto.WsMessageRequest) error) {
	for {
		if msg := c.Read(); msg != nil {
			if err := handler(msg); err != nil {
				log.Error("handle message error in HandleMessageLoop, exit", zap.Error(err))
				return
			}
		} else {
			return
		}
	}
}

// 从 WebSocket 读取消息，并放入消息栈
func (c *Connection) ReadFromWsLoop() {
	for {
		if c.IsClosed() {
			// 连接已断开，退出循环
			break
		}

		if message, err := ReadForType[dto.WsMessageRequest](c.conn); err != nil {
			// 读取消息失败，退出循环
			break
		} else {
			// 读取到消息，放入消息栈
			c.Write(message)
		}
	}

	// 跳出循环，往消息栈中写入 nil 值，表示连接已断开，通知 HandleMessageLoop 协程退出
	c.Write(nil)
}

// 开启消息处理
func (c *Connection) Handle(handler func(*dto.WsMessageRequest) error) {
	go c.HandleMessageLoop(handler)
	c.ReadFromWsLoop()
}

// 从消息栈中读取消息
func (c *Connection) Read() *dto.WsMessageRequest {
	return <-c.stack
}

// 往消息栈中写入消息
func (c *Connection) Write(msg *dto.WsMessageRequest) {
	if len(c.stack) == cap(c.stack) {
		// 消息栈已满，丢弃消息
		log.Warn("message stack is full, discard message")
		c.skipMessage()
	}
	c.stack <- msg
}

// 连接是否已断开
func (c *Connection) IsClosed() bool {
	return c.conn.IsClosed()
}

// 跳过消息(丢弃)
func (c *Connection) skipMessage() {
	<-c.stack
}

func (c *Connection) Send(msg dto.WsMessageResponse) error {
	return c.conn.Send(msg)
}
