package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"txing-ai/internal/domain"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"

	"gorm.io/gorm"
)

const defaultConversationName = "new chat"

// 查询 conversation 详情
func QueryConversationById(db *gorm.DB, id int64) (*domain.Conversation, error) {
	var conversation domain.Conversation

	result := db.Where("id = ?", id).First(&conversation)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("conversation not found: %v", id)
		}
		return nil, fmt.Errorf("failed to query conversation: %v", result.Error)
	}

	// 把 conversation.Message(json 字符串) 转换为 conversation.FormattedMessage
	if conversation.Message != "" {
		var messages []global.Message
		if err := json.Unmarshal([]byte(conversation.Message), &messages); err != nil {
			return nil, fmt.Errorf("failed to parse conversation messages: %v", err)
		}
		conversation.FormattedMessage = messages
	}

	return &conversation, nil
}

// 根据情况获取到 conversation 实例
func ExtractConversation(db *gorm.DB, id int64, userId int64, presetId string) *domain.Conversation {

	var conversation *domain.Conversation
	if userId == -1 {
		// 未登录用户，创建匿名 conversation
		conversation = NewAnonymousConversation()
	} else if id == -1 {
		// 需要新建 conversation
		conversation = NewConversation(userId)
		log.Info("presetId: ", zap.String("presetId", presetId))
		if presetId != "" {
			// 初始化 preset 结构体
			preset := &domain.Preset{}
			// 查询出预设详情，直接使用 presetId 而不是其地址
			if err := db.Where("id = ?", presetId).First(preset).Error; err != nil {
				log.Error("get preset failed", zap.Error(err))
				panic("get preset failed")
			}
			// 添加初始系统消息，用于给大模型预设一些初始话题
			conversation.AddMessageFromSystem(preset.Context)
			// 添加初始 AI 消息，用于引导用户进行对话 `你好！我是 ${assistant.name}，${assistant.description}`
			helloMsg := fmt.Sprintf("你好！我是 %s，%s", preset.Name, preset.Description)
			conversation.AddMessageFromAssistant(helloMsg, "")
		}

		if err := db.Create(conversation).Error; err != nil {
			log.Error("failed to create conversation", zap.Error(err))
			return nil
		}
	} else {
		// 查询已有的 conversation
		var err error
		conversation, err = QueryConversationById(db, id)
		if err != nil {
			conversation = NewConversation(userId)
			if err := db.Create(conversation).Error; err != nil {
				log.Error("failed to create conversation", zap.Error(err))
				return nil
			}
		}
	}

	return conversation
}

// 构建 conversation 实例
func NewConversation(userId int64) *domain.Conversation {
	conversation := domain.Conversation{
		Auth:             true,
		UserID:           userId,
		Message:          "[]",
		Name:             defaultConversationName,
		Model:            global.ModelDeepSeekV3,
		FormattedMessage: []global.Message{},
	}
	return &conversation
}

// 创建匿名 conversation
func NewAnonymousConversation() *domain.Conversation {
	return &domain.Conversation{
		Auth:             false,
		UserID:           -1,
		Message:          "[]",
		Name:             defaultConversationName,
		Model:            global.ModelDeepSeekV3,
		FormattedMessage: []global.Message{},
	}
}
