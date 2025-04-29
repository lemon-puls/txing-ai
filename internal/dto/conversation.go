package dto

import "txing-ai/internal/utils/page"

// ConversationListRequest 会话列表查询请求
type ConversationListRequest struct {
	page.CursorPageBaseRequest
}
