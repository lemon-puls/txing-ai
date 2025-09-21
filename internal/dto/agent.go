package dto

// AgentExecReq Agent执行请求
type AgentExecReq struct {
	AgentType string `json:"agentType" binding:"required" example:"general"` // 智能体类型
	Content   string `json:"content" example:"生成一份深圳旅游攻略"`                   // 请求内容
}
