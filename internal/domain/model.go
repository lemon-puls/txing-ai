package domain

// ModelTag 模型标签类型
type ModelTag []string

// Model 模型表
type Model struct {
	BaseModel
	Name        string `gorm:"type:varchar(100);not null;comment:模型名称" json:"name"`
	Description string `gorm:"type:varchar(1024);comment:模型描述" json:"description"`
	Default     bool   `gorm:"type:boolean;not null;default:false;comment:是否为默认模型" json:"default"`
	HighContext bool   `gorm:"type:boolean;not null;default:false;comment:是否支持高上下文" json:"high_context"`
	Avatar      string `gorm:"type:varchar(255);comment:模型头像" json:"avatar"`
	Tag         string `gorm:"type:varchar(255);comment:模型标签(多个标签以英文逗号分隔)" json:"tag"`
}
