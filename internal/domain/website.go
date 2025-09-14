package domain

// Website 网站实体
type Website struct {
	BaseModel
	Name        string `gorm:"column:name;type:varchar(100);not null;comment:网站名称" json:"name"`
	Description string `gorm:"column:description;type:varchar(500);not null;comment:网站描述" json:"description"`
	Url         string `gorm:"column:url;type:varchar(500);not null;comment:网站地址" json:"url"`
	Avatar      string `gorm:"column:avatar;type:varchar(500);comment:网站头像" json:"avatar"`
	Tags        string `gorm:"column:tags;type:varchar(500);comment:标签,逗号分隔" json:"tags"`
	Sort        int    `gorm:"column:sort;type:int;default:0;comment:排序" json:"sort"`
	Status      int    `gorm:"column:status;type:tinyint;default:1;comment:状态 1启用 0禁用" json:"status"`
}

func (Website) TableName() string {
	return "websites"
}