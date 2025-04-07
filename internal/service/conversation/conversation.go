package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"txing-ai/internal/domain"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

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
func ExtractConversation(db *gorm.DB, id int64, userId int64) *domain.Conversation {

	var conversation *domain.Conversation
	if userId == -1 {
		// 未登录用户，创建匿名 conversation
		conversation = NewAnonymousConversation()
	} else if id == -1 {
		// 需要新建 conversation
		conversation = NewConversation(userId)
		if err := db.Create(conversation).Error; err != nil {
			log.Error("failed to create conversation", zap.Error(err))
			return nil
		}
	} else {
		// 查询已有的 conversation
		conversation, err := QueryConversationById(db, id)
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
