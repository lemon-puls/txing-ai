package vo

import (
	"time"
	"txing-ai/internal/domain"
)

type AgentFlowVO struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Topology    string    `json:"topology"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

func ToAgentFlowVO(flow domain.AgentFlow) AgentFlowVO {
	return AgentFlowVO{
		Id:          flow.Id,
		Name:        flow.Name,
		Description: flow.Description,
		Topology:    flow.Topology,
		CreateTime:  flow.CreateTime,
		UpdateTime:  flow.UpdateTime,
	}
}

func ToAgentFlowVOs(flows []domain.AgentFlow) []AgentFlowVO {
	var vos []AgentFlowVO
	for _, flow := range flows {
		vos = append(vos, ToAgentFlowVO(flow))
	}
	return vos
}
