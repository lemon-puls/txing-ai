package dto

import "txing-ai/internal/utils/page"

// OpsChatStreamReq 运营助手对话请求（服务端持久化会话，前端只传本次输入与会话ID）
type OpsChatStreamReq struct {
	// 会话ID，0 表示新建会话
	SessionId int64 `json:"sessionId" example:"0"`
	// 本次用户输入
	Content string `json:"content" binding:"required,max=8000" example:"收录 https://github.com/cloudwego/eino"`
	// 页面上下文（可选），由前端各管理页注入
	Context *OpsChatContext `json:"context,omitempty"`
}

// OpsChatContext 页面上下文
type OpsChatContext struct {
	// 来源页面标识，如 "websites"、"ops-assistant"
	Page string `json:"page,omitempty" example:"websites"`
	// 页面草稿数据，如 {"url":"https://github.com/cloudwego/eino"}
	Draft map[string]interface{} `json:"draft,omitempty"`
}

// OpsChatSessionListReq 运营助手会话列表请求（游标分页）
type OpsChatSessionListReq struct {
	page.CursorPageBaseRequest
}

// OpsChatAppendReq 运营助手会话追加请求（持久化前端本地产生的消息，如提案确认提示；
// rewindLastRound 为 true 时表示回退最后一轮问答，用于重新生成/失败重试）
type OpsChatAppendReq struct {
	Role    string `json:"role" binding:"required,oneof=assistant" example:"assistant"`
	Content string `json:"content" binding:"max=8000" example:"✅ 提案已确认，网站录入完成。"`
	// 是否将未确认的提案标记为已确认
	MarkProposalConfirmed bool `json:"markProposalConfirmed"`
	// 被确认提案的 name（批量录入时精确匹配；为空时回退为"最后一条未确认提案"）
	ProposalName string `json:"proposalName,omitempty" example:"deepseek-v3"`
	// 回退最后一轮问答（重新生成/失败重试前调用）：移除末尾的助手消息与其对应用户消息；
	// 为 true 时 Content 可为空（不追加任何消息）
	RewindLastRound bool `json:"rewindLastRound,omitempty"`
}
