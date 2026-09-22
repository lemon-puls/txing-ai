package dto

// OpsChatStreamReq 运营助手对话请求
type OpsChatStreamReq struct {
	Messages []OpsChatMessage `json:"messages" binding:"required,min=1,max=40,dive"`
	// 页面上下文（可选），由前端各管理页注入
	Context *OpsChatContext `json:"context,omitempty"`
}

// OpsChatMessage 对话消息
type OpsChatMessage struct {
	Role    string `json:"role" binding:"required,oneof=user assistant" example:"user"`
	Content string `json:"content" binding:"required,max=8000" example:"收录 https://github.com/cloudwego/eino"`
}

// OpsChatContext 页面上下文
type OpsChatContext struct {
	// 来源页面标识，如 "websites"、"ops-assistant"
	Page string `json:"page,omitempty" example:"websites"`
	// 页面草稿数据，如 {"url":"https://github.com/cloudwego/eino"}
	Draft map[string]interface{} `json:"draft,omitempty"`
}
