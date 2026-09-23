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

// OpsChatAppendReq 运营助手会话追加请求（持久化前端本地产生的消息，如提案确认提示）
type OpsChatAppendReq struct {
	Role    string `json:"role" binding:"required,oneof=assistant" example:"assistant"`
	Content string `json:"content" binding:"required,max=8000" example:"✅ 提案已确认，网站录入完成。"`
	// 是否将最后一条未确认的提案标记为已确认
	MarkProposalConfirmed bool `json:"markProposalConfirmed"`
}
