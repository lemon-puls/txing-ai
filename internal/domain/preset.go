package domain

// Preset 聊天预设表
type Preset struct {
	BaseModel
	UserID      *int64 `gorm:"column:user_id;type:bigint;index;comment:用户ID" json:"user_id"`
	Avatar      string `gorm:"type:varchar(255);comment:预设头像" json:"avatar"`
	Name        string `gorm:"type:varchar(255);comment:预设名称" json:"name"`
	Description string `gorm:"type:text;comment:预设描述" json:"description"`
	Context     string `gorm:"type:text;comment:预设上下文" json:"context"`
	// 是否官方预设
	Official bool  `gorm:"comment:是否官方预设" json:"official"`
	User     *User `gorm:"foreignKey:UserID;references:Id;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION" json:"-"`
}
