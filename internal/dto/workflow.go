package dto

import "txing-ai/internal/utils/page"

type CreateAgentFlowReq struct {
	Name        string `json:"name" binding:"required" comment:"工作流名称"`
	Description string `json:"description" comment:"工作流描述"`
	Topology    string `json:"topology" comment:"工作流拓扑图数据"`
}

type UpdateAgentFlowReq struct {
	Name        string `json:"name" binding:"required" comment:"工作流名称"`
	Description string `json:"description" comment:"工作流描述"`
	Topology    string `json:"topology" comment:"工作流拓扑图数据"`
}

type ListAgentFlowReq struct {
	page.PageRequest
	Name string `form:"name" comment:"工作流名称"`
}

// ValidateWorkflowReq 工作流校验请求
type ValidateWorkflowReq struct {
	Topology string `json:"topology" comment:"工作流拓扑图数据"`
	UseLLM   bool   `json:"useLLM" comment:"是否启用LLM语义校验"`
}

// CreateVersionReq 创建版本请求
type CreateVersionReq struct {
	Name        string `json:"name" binding:"required" comment:"版本名称"`
	Description string `json:"description" comment:"版本描述"`
	ChangeLog   string `json:"changeLog" comment:"变更日志"`
}

// PublishVersionReq 发布版本请求
type PublishVersionReq struct {
	Version int `json:"version" binding:"required" comment:"版本号"`
}

// ListVersionReq 版本列表请求
type ListVersionReq struct {
	page.PageRequest
	FlowID int64 `form:"flowId" binding:"required" comment:"工作流ID"`
}

// CreateTemplateReq 创建模板请求
type CreateTemplateReq struct {
	FlowID      int64  `json:"flowId" binding:"required" comment:"工作流ID"`
	Name        string `json:"name" binding:"required" comment:"模板名称"`
	Description string `json:"description" comment:"模板描述"`
	Category    string `json:"category" comment:"模板分类"`
}

// ListTemplateReq 模板列表请求
type ListTemplateReq struct {
	page.PageRequest
	Category string `form:"category" comment:"模板分类"`
	Name     string `form:"name" comment:"模板名称"`
}

// CloneTemplateReq 克隆模板请求
type CloneTemplateReq struct {
	TemplateID int64  `json:"templateId" binding:"required" comment:"模板ID"`
	Name       string `json:"name" comment:"新工作流名称"`
}
