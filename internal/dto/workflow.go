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
