package config

import (
	"fmt"
	"time"
	model "txing-ai/internal/domain"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// 分页
func Paginate(page int, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func NewMysqlDB(conf *global.MysqlConfig) *gorm.DB {

	username := conf.Username
	password := conf.Password
	host := conf.Host
	port := conf.Port
	database := conf.Database
	maxIdleConns := conf.MaxIdleConns
	maxOpenConns := conf.MaxOpenConns
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username,
		password,
		host,
		port,
		database,
	)
	// gorm 日志配置
	newLogger := logging.NewGormLogger(zap.L())

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true, // 使用单数
		},
	})

	if err != nil {
		log.Error("mysql connect failed, panic", zap.Error(err))
		panic(err)
	}
	// 连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("get sqlDB failed, panic", zap.Error(err))
		panic(err)
	}
	sqlDB.SetMaxIdleConns(maxIdleConns) // 设置连接池中空闲连接的最大数量
	sqlDB.SetMaxOpenConns(maxOpenConns) //设置打开数据库连接的最大数量
	sqlDB.SetConnMaxLifetime(time.Hour) //设置了连接可复用的最大时间

	//数据库迁移
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Model{})
	db.AutoMigrate(&model.Channel{})
	db.AutoMigrate(&model.Preset{})
	db.AutoMigrate(&model.Conversation{})
	db.AutoMigrate(&model.Website{})

	// 设置 GORM 的 JSON 序列化器
	db.Config.PrepareStmt = true
	db.Config.DisableForeignKeyConstraintWhenMigrating = true

	return db
}
