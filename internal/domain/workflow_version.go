package domain

// AgentFlowVersion 工作流版本模型
type AgentFlowVersion struct {
	BaseModel
	FlowID      int64  `gorm:"column:flow_id;not null;index;comment:工作流ID" json:"flowId"`
	Version     int    `gorm:"column:version;not null;comment:版本号" json:"version"`
	Name        string `gorm:"column:name;type:varchar(100);not null;comment:版本名称" json:"name"`
	Description string `gorm:"column:description;type:varchar(500);comment:版本描述" json:"description"`
	Topology    string `gorm:"column:topology;type:text;comment:工作流拓扑图数据" json:"topology"`
	IsPublished bool   `gorm:"column:is_published;default:false;comment:是否已发布" json:"isPublished"`
	ChangeLog   string `gorm:"column:change_log;type:text;comment:变更日志" json:"changeLog"`
}

func (AgentFlowVersion) TableName() string {
	return "agent_flow_versions"
}
