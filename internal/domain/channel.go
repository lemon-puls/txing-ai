package domain

type Channel struct {
	BaseModel
	Name     string   `gorm:"type:varchar(255);not null;comment:渠道名称" json:"name"`
	Type     string   `gorm:"type:varchar(50);not null;comment:渠道类型" json:"type"`
	Priority int      `gorm:"type:int;default:0;comment:优先级" json:"priority"`
	Weight   int      `gorm:"type:int;default:0;comment:权重" json:"weight"`
	Models   []string `gorm:"type:json;comment:支持的模型列表" json:"models"`
	Retry    int      `gorm:"type:int;default:3;comment:重试次数" json:"retry"`
	Secret   string   `gorm:"type:varchar(255);comment:密钥" json:"secret"`
	Endpoint string   `gorm:"type:varchar(255);not null;comment:服务地址" json:"endpoint"`
	Status   bool     `gorm:"type:int;default:0;comment:启用状态(0: 正常 1: 停用)" json:"status"`
}
