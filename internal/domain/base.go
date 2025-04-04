package domain

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	Id         int64          `gorm:"column:id;primary_key" json:"id"`
	CreateTime time.Time      `gorm:"column:createTime" json:"createTime"`
	UpdateTime time.Time      `gorm:"column:updateTime" json:"updateTime"`
	DeleteTime gorm.DeletedAt `gorm:"column:deleteTime" json:"-"`
}

// gorm 钩子函数，用于创建时间和更新时间的自动赋值
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	b.CreateTime = time.Now()
	b.UpdateTime = time.Now()
	return nil
}

func (b *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	b.UpdateTime = time.Now()
	return nil
}
