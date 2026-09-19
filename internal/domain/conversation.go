package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"unicode/utf8"

	"github.com/samber/lo"

	"go.uber.org/zap"
	"gorm.io/gorm"
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

	// 预设 id
	PresetID *int64 `gorm:"type:bigint;comment:预设 id" json:"presetId"`

	// 非数据库字段
	FormattedMessage []global.Message `gorm:"-" json:"formattedMessage"`
}

const (
	defaultContextLength = 8
	// imageKeepRounds 发送给 LLM 时保留原图的最大用户轮数（从最新消息往前数）。
	// 更早轮次 user 消息中的图片会被剥离并替换为文本占位，控制视觉 token 随对话累积；
	// 只影响发送副本，持久化的 FormattedMessage 与前端历史回显不受影响。
	imageKeepRounds = 2
)

// 处理消息
func (c *Conversation) HandleMessage(msg *dto.WsMessageRequest, db *gorm.DB) error {
	// 如果是该会话的第一条用户发的消息，则更新会话名称
	c.setNameFromFirstUserMessage(msg.Content)
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

// setNameFromFirstUserMessage 若会话中还没有任何用户消息，则将会话名称设置为消息内容
// （最多 35 个字符，超出按 rune 截断，避免截断多字节字符）
func (c *Conversation) setNameFromFirstUserMessage(content string) {
	count := lo.CountBy(c.FormattedMessage, func(m global.Message) bool {
		return m.Role == global.User
	})
	if count > 0 {
		return
	}

	if utf8.RuneCountInString(content) > 35 {
		// 找到第 35 个字符的位置
		pos := 0
		for i := 0; i < 35; i++ {
			_, size := utf8.DecodeRuneInString(content[pos:])
			pos += size
		}
		c.Name = content[:pos]
	} else {
		c.Name = content
	}
}

// 将 WsMessageRequest 消息添加到会话消息记录中
func (c *Conversation) addMessageFromWsMessageRequest(msg *dto.WsMessageRequest) error {
	// 如果消息内容为空，则不添加到消息记录中
	if len(msg.Content) == 0 && len(msg.Images) == 0 && len(msg.Attachments) == 0 {
		return errors.New("message content is empty")
	}

	// 转换附件列表
	var attachments []global.Attachment
	for _, a := range msg.Attachments {
		attachments = append(attachments, global.Attachment{
			FileName: a.FileName,
			FileURL:  a.FileURL,
			FileType: a.FileType,
			FileSize: a.FileSize,
		})
	}

	// 调试日志：打印接收到的多模态信息
	log.Info("收到消息",
		zap.String("content", msg.Content),
		zap.Int("images_count", len(msg.Images)),
		zap.Int("attachments_count", len(msg.Attachments)),
		zap.Any("images", msg.Images),
	)

	// 将消息添加到会话消息记录中
	c.addMessage(global.Message{
		Role:        global.User,
		Content:     msg.Content,
		Images:      msg.Images,
		Attachments: attachments,
	})
	// 应用调用参数
	c.applyCallParams(msg)
	return nil
}

// HandleWorkflowMessage 处理工作流用户消息（包含应用名称和文件信息）
func (c *Conversation) HandleWorkflowMessage(msg *dto.WsMessageRequest, appName string, db *gorm.DB) error {
	if len(msg.Content) == 0 {
		return errors.New("message content is empty")
	}

	// 如果是该会话的第一条用户发的消息，则更新会话名称
	// （前端发送时已剥离 @应用名 前缀，此处直接用消息内容作为名称）
	c.setNameFromFirstUserMessage(msg.Content)

	// 提取文件名列表
	var fileNames []string
	for _, f := range msg.Files {
		if f.FileName != "" {
			fileNames = append(fileNames, f.FileName)
		}
	}

	// 将消息添加到会话消息记录中（包含应用名称和文件信息）
	c.addMessage(global.Message{
		Role:    global.User,
		Content: msg.Content,
		AppName: appName,
		Files:   fileNames,
	})
	c.applyCallParams(msg)

	// 更新会话信息到数据库
	return c.updateOrCreate(db)
}

func (c *Conversation) SaveResponse(db *gorm.DB, content string, reasoningContent string) {
	// 添加消息到会话消息记录中
	c.AddMessageFromAssistant(content, reasoningContent)

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
	c.Temperature = msg.Temperature
	c.Model = msg.Model
	//c.setContextLength(msg.Context)

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
	// 消息转换为 json 字符串
	jsonBytes, err := json.Marshal(c.FormattedMessage)
	if err != nil {
		log.Error("json.Marshal failed", zap.Error(err))
		return err
	}
	c.Message = string(jsonBytes)

	// 执行新增或更新操作
	tx := db.Save(c)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// 获取到用于发送给大模型的消息切片（context 的长度）
// 返回的是副本：下游会剥离旧轮次图片，绝不能改动持久化的 FormattedMessage
// （前端历史回显依赖它保留完整图片）。
func (c *Conversation) GetChatMessages() []global.Message {
	length := c.getContextLength()

	start := 0
	if len(c.FormattedMessage) > length {
		start = len(c.FormattedMessage) - length
	}
	window := c.FormattedMessage[start:]

	messages := make([]global.Message, len(window))
	copy(messages, window)

	stripOldImages(messages)
	return messages
}

// stripOldImages 剥离早于最近 imageKeepRounds 轮的 user 消息中的图片（含图片类附件），
// 替换为文本占位提示。图片是高成本的视觉 token，旧轮次的图片通常不再是追问焦点，
// 剥离后可显著降低请求体积与费用。
// 原地修改入参，调用方必须传入副本（见 GetChatMessages）。
func stripOldImages(messages []global.Message) {
	// 从后往前定位第 imageKeepRounds 条 user 消息，只剥离它之前的 user 消息
	keepFrom := -1
	userSeen := 0
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != global.User {
			continue
		}
		userSeen++
		if userSeen == imageKeepRounds {
			keepFrom = i
			break
		}
	}
	// 窗口内 user 消息不足 imageKeepRounds 条，或第 K 条就是首条：无可剥离
	if keepFrom <= 0 {
		return
	}

	for i := 0; i < keepFrom; i++ {
		msg := &messages[i]
		if msg.Role != global.User {
			continue
		}

		imageCount := len(msg.Images)
		var keptAttachments []global.Attachment
		for _, att := range msg.Attachments {
			if isImageFileType(att.FileType) {
				imageCount++
			} else {
				keptAttachments = append(keptAttachments, att)
			}
		}
		if imageCount == 0 {
			continue
		}

		// 只改副本字段，不 mutate 底层切片数据
		hint := fmt.Sprintf("（此消息原有 %d 张图片，因对话较长已省略）", imageCount)
		if msg.Content != "" {
			msg.Content += "\n" + hint
		} else {
			msg.Content = hint
		}
		msg.Images = nil
		msg.Attachments = keptAttachments
	}
}

// isImageFileType 判断文件类型是否为图片（与 adapter 层 isImageType 保持一致）
func isImageFileType(fileType string) bool {
	switch fileType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml":
		return true
	}
	return false
}

func (c *Conversation) AddMessageFromAssistant(content, reasoningContent string) {
	// 如果内容与思考过程都为空，则不添加到消息记录中。
	// 注意：只产出思考过程（reasoning 非空、content 为空）时也必须保存，
	// 否则推理模型的思考内容会被整个丢弃
	if len(content) == 0 && len(reasoningContent) == 0 {
		log.Error("response is empty, skip AddMessageFromAssistant")
		return
	}

	// 将消息添加到会话消息记录中
	c.addMessage(global.Message{
		Role:             global.Assistant,
		Content:          content,
		ReasoningContent: reasoningContent,
	})
}

func (c *Conversation) AddMessageFromSystem(content string) {
	if len(content) == 0 {
		log.Error("response is empty, skip AddMessageFromSystem")
		return
	}

	c.addMessage(global.Message{
		Role:    global.System,
		Content: content,
	})
}
