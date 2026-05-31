package domain

// AgentFlow 工作流模型
type AgentFlow struct {
	BaseModel
	Name            string `gorm:"column:name;type:varchar(100);not null;comment:工作流名称" json:"name"`
	Description     string `gorm:"column:description;type:varchar(500);comment:工作流描述" json:"description"`
	Topology        string `gorm:"column:topology;type:text;comment:工作流拓扑图数据" json:"topology"`
	CurrentVersion  int    `gorm:"column:current_version;default:1;comment:当前版本号" json:"currentVersion"`
	PublishedVersion int   `gorm:"column:published_version;default:0;comment:已发布版本号" json:"publishedVersion"`
	IsTemplate      bool   `gorm:"column:is_template;default:false;comment:是否为模板" json:"isTemplate"`
	TemplateCategory string `gorm:"column:template_category;type:varchar(50);comment:模板分类" json:"templateCategory"`
	Status          string `gorm:"column:status;type:varchar(20);default:'draft';comment:状态: draft/published" json:"status"`
}

func (AgentFlow) TableName() string {
	return "agent_flows"
}
