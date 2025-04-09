package domain

import (
	"errors"
	"gorm.io/gorm"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

type Conversation struct {
	BaseModel
	Auth      bool   `gorm:"type:boolean;not null;default:false;comment:是否已认证" json:"auth"`
	UserID    int64  `gorm:"type:bigint;not null;index;comment:用户ID" json:"userId"`
	Name      string `gorm:"type:varchar(255);comment:会话名称" json:"name"`
	Message   string `gorm:"type:mediumtext;comment:会话消息记录" json:"message"`
	Model     string `gorm:"type:varchar(50);not null;comment:使用的模型" json:"model"`
	EnableWeb bool   `gorm:"type:boolean;not null;default:false;comment:是否启用网页搜索" json:"enableWeb"`
	Context   int    `gorm:"type:int;not null;default:0;comment:上下文长度" json:"context"`

	// 可选的模型参数
	MaxTokens         *int     `gorm:"type:int;comment:最大token数" json:"maxTokens,omitempty"`
	Temperature       *float32 `gorm:"type:float;comment:温度参数" json:"temperature,omitempty"`
	TopP              *float32 `gorm:"type:float;comment:Top-P采样参数" json:"topP,omitempty"`
	TopK              *int     `gorm:"type:int;comment:Top-K采样参数" json:"topK,omitempty"`
	PresencePenalty   *float32 `gorm:"type:float;comment:存在惩罚参数" json:"presencePenalty,omitempty"`
	FrequencyPenalty  *float32 `gorm:"type:float;comment:频率惩罚参数" json:"frequencyPenalty,omitempty"`
	RepetitionPenalty *float32 `gorm:"type:float;comment:重复惩罚参数" json:"repetitionPenalty,omitempty"`

	// 非数据库字段
	FormattedMessage []global.Message `gorm:"-" json:"formattedMessage"`
}

const (
	defaultContextLength = 8
)

// 处理消息
func (c *Conversation) HandleMessage(msg *dto.WsMessageRequest, db *gorm.DB) error {
	// 如果是会话的第一条消息，则更新会话名称
	if len(c.Message) == 0 {
		// 更新会话名称
		c.Name = msg.Content
	}
	// 添加消息到会话消息记录中，并应用调用参数
	if err := c.addMessageFromWsMessageRequest(msg); err != nil {
		return err
	}

	// 更新会话信息到数据库
	if err := c.updateOrCreate(db); err != nil {
		return err
	}
	return nil
}

// 将 WsMessageRequest 消息添加到会话消息记录中
func (c *Conversation) addMessageFromWsMessageRequest(msg *dto.WsMessageRequest) error {
	// 如果消息内容为空，则不添加到消息记录中
	if len(msg.Content) == 0 {
		return errors.New("message content is empty")
	}

	// 将消息添加到会话消息记录中
	c.addMessage(global.Message{
		Role:    global.User,
		Content: msg.Content,
	})
	// 应用调用参数
	c.applyCallParams(msg)
	return nil
}

func (c *Conversation) SaveResponse(db *gorm.DB, response string) {
	// 添加消息到会话消息记录中
	c.addMessageFromAssistant(response)

	// 更新会话信息到数据库
	c.updateOrCreate(db)
}

// 添加消息到会话消息记录中
func (c *Conversation) addMessage(msg global.Message) {
	c.FormattedMessage = append(c.FormattedMessage, msg)
}

// 应用本次调用参数
func (c *Conversation) applyCallParams(msg *dto.WsMessageRequest) {
	c.MaxTokens = msg.MaxTokens
	c.TopP = msg.TopP
	c.TopK = msg.TopK
	c.PresencePenalty = msg.PresencePenalty
	c.FrequencyPenalty = msg.FrequencyPenalty
	c.RepetitionPenalty = msg.RepetitionPenalty
	c.EnableWeb = msg.EnableWeb
	c.Model = msg.Model
	c.setContextLength(msg.Context)

}

// 设置上下文长度
func (c *Conversation) setContextLength(context int) {
	if context <= 0 {
		c.Context = defaultContextLength
	} else {
		c.Context = context
	}
}

func (c *Conversation) getContextLength() int {
	if c.Context <= 0 {
		return defaultContextLength
	}
	return c.Context
}

func (c *Conversation) updateOrCreate(db *gorm.DB) error {
	// 执行新增或更新操作
	tx := db.Save(c)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 获取到用于发送给大模型的消息切片（context 的长度）
func (c *Conversation) GetChatMessages() []global.Message {
	//// 深复制消息
	//var cp []global.Message
	//err := copier.Copy(&cp, &c.FormattedMessage)
	//if err != nil {
	//	log.Error("copier.Copy failed", zap.Error(err))
	//	panic(err)
	//}

	length := c.getContextLength()

	if len(c.FormattedMessage) < length {
		return c.FormattedMessage
	}

	return c.FormattedMessage[len(c.FormattedMessage)-length:]
}

func (c *Conversation) addMessageFromAssistant(response string) {
	// 如果消息内容为空，则不添加到消息记录中
	if len(response) == 0 {
		log.Error("response is empty, skip addMessageFromAssistant")
		return
	}

	// 将消息添加到会话消息记录中
	c.addMessage(global.Message{
		Role:    global.Assistant,
		Content: response,
	})
}
