package dto

import "txing-ai/internal/iface/wikiagent"

// WikiMDSourceReq 新建 markdown 源（前端读文件后以文本提交，服务端上传 COS）
// 上限 1M 字符：ingest 侧按 ~12k 字符/片分片编译，该限制只为防御失控上传，
// 非单次 LLM 调用的硬约束（见 builtin.splitMarkdownChunks）
type WikiMDSourceReq struct {
	Title   string `json:"title" binding:"required,max=255" example:"个人简历 2026"`
	Content string `json:"content" binding:"required,max=1000000" example:"# 简历正文..."`
}

// WikiURLSourceReq 新建网页 URL 源（ingest 时服务端抓取正文）
type WikiURLSourceReq struct {
	Title string `json:"title" binding:"required,max=255" example:"我的技术博客"`
	Url   string `json:"url" binding:"required,url,max=1024" example:"https://blog.example.com/about"`
}

// WikiUpdateDraftReq 编辑草稿/已发布页（指针字段：只更新显式传入的字段）
type WikiUpdateDraftReq struct {
	Title    *string   `json:"title,omitempty" binding:"omitempty,max=255"`
	PageType *string   `json:"pageType,omitempty" binding:"omitempty,max=32"`
	Aliases  []string  `json:"aliases,omitempty"`
	Summary  *string   `json:"summary,omitempty" binding:"omitempty,max=512"`
	Content  *string   `json:"content,omitempty" binding:"omitempty,max=200000"`
}

// WikiAskReq 面试官问答请求（多轮上下文由前端携带，服务端无状态）
type WikiAskReq struct {
	// 前端生成的会话标识（用于服务端问答留存聚合，D10）
	SessionId string `json:"sessionId" binding:"required,max=64" example:"b7c8d9e0-..."`
	// 本次提问
	Question string `json:"question" binding:"required,max=4000" example:"他最熟悉的技术栈是什么？"`
	// 近期对话历史（role: user/assistant）
	History []wikiagent.QAHistoryItem `json:"history,omitempty"`
}
