package utils

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"time"
	"txing-ai/internal/dto"
	"txing-ai/internal/global/logging/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 用于管理底层 WebSocket 连接的读写等操作
type WebSocket struct {
	Conn       *websocket.Conn
	Ctx        *gin.Context
	Closed     bool
	MaxTimeout time.Duration
}

func NewWebSocket(ctx *gin.Context) *WebSocket {

	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有源
		},
	}

	var conn *websocket.Conn
	var err error

	if conn, err = upgrader.Upgrade(ctx.Writer, ctx.Request, nil); err != nil {
		ErrorWithMsg(ctx, err.Error(), err)
		return nil
	}

	instance := &WebSocket{
		Conn: conn,
		Ctx:  ctx,
	}

	return instance
}

// websocket 初始化
func (w *WebSocket) Init() {
	w.Closed = false
	// 设置关闭连接的回调函数，当客户端关闭连接时，将 Closed 标记为 true
	w.Conn.SetCloseHandler(func(code int, text string) error {
		w.Closed = true
		return nil
	})
	// 设置心跳检测，如果超过 MaxTimeout 时间没有收到客户端的 Pong 消息，则关闭连接
	w.Conn.SetPongHandler(func(appData string) error {
		return w.Conn.SetReadDeadline(time.Now().Add(w.MaxTimeout))
	})
}

// 发送消息（写）
func (w *WebSocket) Write(messageType int, message []byte) error {
	return w.Conn.WriteMessage(messageType, message)
}

// 接收消息（读）
func (w *WebSocket) Read() (messageType int, message []byte, err error) {
	return w.Conn.ReadMessage()
}

func (w *WebSocket) IsClosed() bool {
	return w.Closed
}

func (w *WebSocket) Send(v interface{}) error {
	log.Info("发送消息：", zap.Any("message", v))
	return w.SendJson(v)
}

func (w *WebSocket) SendJson(v interface{}) error {
	return w.Conn.WriteJSON(v)
}

// 读取消息 并且转换为指定类型
func ReadForType[T any](w *WebSocket) (*T, error) {
	_, message, err := w.Read()
	if err != nil {
		return nil, err
	}

	var msg dto.WsMessageRequest
	if err := json.Unmarshal(message, &msg); err == nil {
		// 判断是否为 Ping 消息
		if msg.Type == "ping" {
			log.Debug("收到 Ping 消息")
			return ReadForType[T](w)
		}
	}

	var result T
	if err := json.Unmarshal(message, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
