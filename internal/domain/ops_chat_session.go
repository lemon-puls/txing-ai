package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"txing-ai/internal/dto"
	"txing-ai/internal/global/logging/log"
	"unicode/utf8"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// maxStoredToolResultChars 持久化的单条工具结果最大字符数（超长截断，避免 mediumtext 膨胀）
const maxStoredToolResultChars = 4000

// OpsChatToolCall 运营助手会话中的工具调用记录（字段名与前端 ToolCallItem 对齐）
type OpsChatToolCall struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Params string `json:"params"`
	Result string `json:"result"`
	Status string `json:"status"` // running/completed/failed/interrupted
}

// OpsChatProposalItem 单个预览提案（批量录入时一条回复可携带多张提案卡片）
// Status 取 ok|duplicate|confirmed，confirmed 表示提案已确认入库，前端据此禁用确认按钮
type OpsChatProposalItem struct {
	// Proposal 结构化提案原始 JSON（website/model/channel/preset 各自的扁平结构，
	// 前端按 type 字段分发渲染卡片。
	// internal/tool/ops 反向依赖了 domain，这里不能 import tool 包，由 controller 层透传原始 JSON）
	// swaggertype:object 让 swag 将 json.RawMessage 识别为任意 JSON 对象
	Proposal json.RawMessage `json:"proposal" swaggertype:"object"`
	Status   string          `json:"status,omitempty"`
	Message  string          `json:"message,omitempty"`
}

// OpsChatMessage 运营助手会话消息（富结构：含工具调用与提案卡片，需在刷新后完整回放）
type OpsChatMessage struct {
	Role             string             `json:"role"` // user/assistant
	Content          string             `json:"content"`
	Reasoning        string             `json:"reasoning,omitempty"`
	ToolCalls        []OpsChatToolCall  `json:"toolCalls,omitempty"`
	// Proposal/ProposalStatus/ProposalMessage 为旧版单提案字段，仅用于反序列化历史数据；
	// QueryOpsChatSessionById 加载时经 NormalizeProposals 统一并入 Proposals，保存后即落为新结构
	Proposal        json.RawMessage `json:"proposal,omitempty" swaggertype:"object"`
	ProposalStatus  string          `json:"proposalStatus,omitempty"`
	ProposalMessage string          `json:"proposalMessage,omitempty"`
	// Proposals 本条消息携带的全部提案卡片
	Proposals []OpsChatProposalItem `json:"proposals,omitempty"`
	Error     string                `json:"error,omitempty"`
	// Interrupted 表示本次回复被用户主动中断（内容为已生成的部分）
	Interrupted bool `json:"interrupted,omitempty"`
	// Confirmed 标记提案确认完成的提示消息（前端本地产生，经 append 接口持久化）
	Confirmed bool `json:"confirmed,omitempty"`
}

// NormalizeProposals 将旧版单提案字段并入 Proposals 数组（历史数据渐进迁移，
// 加载即归一化，保存后统一为新结构）
func (m *OpsChatMessage) NormalizeProposals() {
	if len(m.Proposal) == 0 {
		return
	}
	m.Proposals = append([]OpsChatProposalItem{{
		Proposal: m.Proposal,
		Status:   m.ProposalStatus,
		Message:  m.ProposalMessage,
	}}, m.Proposals...)
	m.Proposal = nil
	m.ProposalStatus = ""
	m.ProposalMessage = ""
}

// OpsChatSession 运营助手会话（对齐用户端 Conversation 的持久化模式：
// 一条会话一行，消息以 JSON blob 存于 Messages 列）
type OpsChatSession struct {
	BaseModel
	UserID int64  `gorm:"column:user_id;type:bigint;not null;index;comment:用户ID" json:"userId"`
	Title  string `gorm:"column:title;type:varchar(255);comment:会话标题" json:"title"`
	Model  string `gorm:"column:model;type:varchar(100);comment:使用的模型名" json:"model"`
	// 最近一次请求的页面上下文 JSON（page/draft 等）
	Context string `gorm:"column:context;type:text;comment:页面上下文JSON" json:"-"`
	// 会话消息 JSON blob
	Messages string `gorm:"column:messages;type:mediumtext;comment:会话消息JSON" json:"-"`

	// 非数据库字段
	FormattedMessages []OpsChatMessage `gorm:"-" json:"-"`
}

func (OpsChatSession) TableName() string {
	return "ops_chat_sessions"
}

// AppendUserMessage 追加用户消息；若会话中还没有任何用户消息，则以消息内容作为会话标题
// （最多 35 个字符，超出按 rune 截断，与用户端 Conversation.setNameFromFirstUserMessage 一致）
func (s *OpsChatSession) AppendUserMessage(content string) {
	count := 0
	for _, m := range s.FormattedMessages {
		if m.Role == "user" {
			count++
		}
	}
	if count == 0 {
		if utf8.RuneCountInString(content) > 35 {
			// 找到第 35 个字符的位置
			pos := 0
			for i := 0; i < 35; i++ {
				_, size := utf8.DecodeRuneInString(content[pos:])
				pos += size
			}
			s.Title = content[:pos]
		} else {
			s.Title = content
		}
	}

	s.FormattedMessages = append(s.FormattedMessages, OpsChatMessage{
		Role:    "user",
		Content: content,
	})
}

// AppendAssistantMessage 追加助手消息并截断超长工具结果；
// 内容、思考、工具调用、提案、错误全空时跳过（与用户端 AddMessageFromAssistant 一致）
func (s *OpsChatSession) AppendAssistantMessage(msg OpsChatMessage) {
	if msg.Content == "" && msg.Reasoning == "" && len(msg.ToolCalls) == 0 &&
		len(msg.Proposal) == 0 && len(msg.Proposals) == 0 && msg.Error == "" {
		log.Error("ops assistant response is empty, skip AppendAssistantMessage")
		return
	}

	// 工具结果可能很大（网页正文），持久化副本统一截断
	for i := range msg.ToolCalls {
		tc := &msg.ToolCalls[i]
		if utf8.RuneCountInString(tc.Result) > maxStoredToolResultChars {
			pos := 0
			for j := 0; j < maxStoredToolResultChars; j++ {
				_, size := utf8.DecodeRuneInString(tc.Result[pos:])
				pos += size
			}
			tc.Result = tc.Result[:pos] + "…（已截断）"
		}
	}

	s.FormattedMessages = append(s.FormattedMessages, msg)
}

// SetContext 记录最近一次请求的页面上下文 JSON
func (s *OpsChatSession) SetContext(ctx *dto.OpsChatContext) {
	if ctx == nil {
		s.Context = ""
		return
	}
	if b, err := json.Marshal(ctx); err == nil {
		s.Context = string(b)
	}
}

// Save 将 FormattedMessages 序列化后落库（新增或更新）
func (s *OpsChatSession) Save(db *gorm.DB) error {
	jsonBytes, err := json.Marshal(s.FormattedMessages)
	if err != nil {
		log.Error("ops chat session marshal failed", zap.Error(err))
		return err
	}
	s.Messages = string(jsonBytes)

	if tx := db.Save(s); tx.Error != nil {
		return tx.Error
	}
	return nil
}

// QueryOpsChatSessionById 查询会话详情并解析消息 JSON
func QueryOpsChatSessionById(db *gorm.DB, id int64) (*OpsChatSession, error) {
	var session OpsChatSession
	result := db.Where("id = ?", id).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ops chat session not found: %v", id)
		}
		return nil, fmt.Errorf("failed to query ops chat session: %v", result.Error)
	}

	if len(session.Messages) > 0 {
		if err := json.Unmarshal([]byte(session.Messages), &session.FormattedMessages); err != nil {
			log.Error("ops chat session unmarshal failed", zap.Int64("id", id), zap.Error(err))
			return nil, err
		}
		// 旧版单提案字段统一并入 Proposals（保存后即落为新结构，下游只需处理数组）
		for i := range session.FormattedMessages {
			session.FormattedMessages[i].NormalizeProposals()
		}
	}
	return &session, nil
}
