package domain

// AgentFlow 工作流模型
type AgentFlow struct {
	BaseModel
	Name        string `gorm:"column:name;type:varchar(100);not null;comment:工作流名称" json:"name"`
	Description string `gorm:"column:description;type:varchar(500);comment:工作流描述" json:"description"`
	Topology    string `gorm:"column:topology;type:text;comment:工作流拓扑图数据" json:"topology"`
}

func (AgentFlow) TableName() string {
	return "agent_flows"
}
