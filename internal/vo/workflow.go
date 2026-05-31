package vo

import (
	"time"
	"txing-ai/internal/domain"
)

type AgentFlowVO struct {
	Id               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Topology         string    `json:"topology"`
	CurrentVersion   int       `json:"currentVersion"`
	PublishedVersion int       `json:"publishedVersion"`
	IsTemplate       bool      `json:"isTemplate"`
	Status           string    `json:"status"`
	CreateTime       time.Time `json:"createTime"`
	UpdateTime       time.Time `json:"updateTime"`
}

func ToAgentFlowVO(flow domain.AgentFlow) AgentFlowVO {
	return AgentFlowVO{
		Id:               flow.Id,
		Name:             flow.Name,
		Description:      flow.Description,
		Topology:         flow.Topology,
		CurrentVersion:   flow.CurrentVersion,
		PublishedVersion: flow.PublishedVersion,
		IsTemplate:       flow.IsTemplate,
		Status:           flow.Status,
		CreateTime:       flow.CreateTime,
		UpdateTime:       flow.UpdateTime,
	}
}

// PublishedWorkflowVO 已发布工作流 VO（客户端用，不含拓扑数据）
type PublishedWorkflowVO struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	CreateTime  time.Time `json:"createTime"`
}

func ToPublishedWorkflowVO(flow domain.AgentFlow) PublishedWorkflowVO {
	return PublishedWorkflowVO{
		Id:          flow.Id,
		Name:        flow.Name,
		Description: flow.Description,
		Category:    flow.TemplateCategory,
		CreateTime:  flow.CreateTime,
	}
}

func ToPublishedWorkflowVOs(flows []domain.AgentFlow) []PublishedWorkflowVO {
	var vos []PublishedWorkflowVO
	for _, flow := range flows {
		vos = append(vos, ToPublishedWorkflowVO(flow))
	}
	return vos
}

func ToAgentFlowVOs(flows []domain.AgentFlow) []AgentFlowVO {
	var vos []AgentFlowVO
	for _, flow := range flows {
		vos = append(vos, ToAgentFlowVO(flow))
	}
	return vos
}

// AgentFlowVersionVO 工作流版本 VO
type AgentFlowVersionVO struct {
	Id          int64     `json:"id"`
	FlowID      int64     `json:"flowId"`
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Topology    string    `json:"topology"`
	IsPublished bool      `json:"isPublished"`
	ChangeLog   string    `json:"changeLog"`
	CreateTime  time.Time `json:"createTime"`
}

func ToAgentFlowVersionVO(version domain.AgentFlowVersion) AgentFlowVersionVO {
	return AgentFlowVersionVO{
		Id:          version.Id,
		FlowID:      version.FlowID,
		Version:     version.Version,
		Name:        version.Name,
		Description: version.Description,
		Topology:    version.Topology,
		IsPublished: version.IsPublished,
		ChangeLog:   version.ChangeLog,
		CreateTime:  version.CreateTime,
	}
}

func ToAgentFlowVersionVOs(versions []domain.AgentFlowVersion) []AgentFlowVersionVO {
	var vos []AgentFlowVersionVO
	for _, v := range versions {
		vos = append(vos, ToAgentFlowVersionVO(v))
	}
	return vos
}

// TemplateVO 模板 VO
type TemplateVO struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Topology    string    `json:"topology"`
	CreateTime  time.Time `json:"createTime"`
}

func ToTemplateVO(flow domain.AgentFlow) TemplateVO {
	return TemplateVO{
		Id:          flow.Id,
		Name:        flow.Name,
		Description: flow.Description,
		Category:    flow.TemplateCategory,
		Topology:    flow.Topology,
		CreateTime:  flow.CreateTime,
	}
}

func ToTemplateVOs(flows []domain.AgentFlow) []TemplateVO {
	var vos []TemplateVO
	for _, flow := range flows {
		vos = append(vos, ToTemplateVO(flow))
	}
	return vos
}

// ValidationResultVO 工作流校验结果
type ValidationResultVO struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationErrorVO `json:"errors,omitempty"`
	Warnings []ValidationErrorVO `json:"warnings,omitempty"`
}

// ValidationErrorVO 校验错误详情
type ValidationErrorVO struct {
	Level   string `json:"level"`           // "error" | "warning"
	NodeID  string `json:"nodeId,omitempty"` // 关联节点 ID
	Code    string `json:"code"`             // 错误码
	Message string `json:"message"`          // 人类可读描述
}
