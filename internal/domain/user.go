package domain

type User struct {
	BaseModel
	Username string `gorm:"unique; varchar(255); not null" json:"userName"`
	Password string `gorm:"type:varchar(255); not null" json:"-"`
	Email    string `gorm:"type:varchar(255); not null" json:"email"`
	Phone    string `gorm:"type:varchar(255); not null" json:"phone"`
	Gender   int8   `gorm:"type:tinyint" json:"gender"`
	Status   int8   `gorm:"type:tinyint;comment:用户状态(0:正常 1:禁用)" json:"status"`
	Role     int8   `gorm:"type:tinyint;default:0;comment:用户角色(0:普通用户 1:超管)" json:"role"`
	Age      int8   `gorm:"type:tinyint" json:"age"`
	Avatar   string `gorm:"type:varchar(255)" json:"avatar"`
}
