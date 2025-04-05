package domain

type User struct {
	BaseModel
	Username string `gorm:"unique; varchar(255); not null" json:"userName"`
	Password string `gorm:"type:varchar(255); not null" json:"-"`
	Email    string `gorm:"type:varchar(255); not null" json:"email"`
	Phone    string `gorm:"type:varchar(255); not null" json:"phone"`
	Gender   int8   `gorm:"type:tinyint" json:"gender"`
	Status   int8   `gorm:"type:tinyint" json:"status"`
	Age      int8   `gorm:"type:tinyint" json:"age"`
	Avatar   string `gorm:"type:varchar(255)" json:"avatar"`
}
