package global

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"time"

	"github.com/spf13/viper"
)

const (
	// 运行时数据包括配置、日志都统一放这里，方便后续迁移
	RuntimeDir = "runtime"
	// 默认配置文件名
	defaultConfigName = "config.yaml"
)

type AppConfig struct {
	*ServerConfig    `mapstructure:"server"`
	*LogConfig       `mapstructure:"log"`
	*MysqlConfig     `mapstructure:"mysql"`
	*RedisConfig     `mapstructure:"redis"`
	*SnowflakeConfig `mapstructure:"snowflake"`
	*AuthConfig      `mapstructure:"auth"`
	*CosConfig       `mapstructure:"cos"`
	*AmapConfig      `mapstructure:"amap"`
}

type ServerConfig struct {
	Name    string `mapstructure:"name"`
	Port    int    `mapstructure:"port"`
	Profile string `mapstructure:"profile"`
	Version string `mapstructure:"version"`
}
type CosConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	// 自定义域名 baseurl
	BaseURL string `mapstructure:"base_url"`
	// 预签名URL有效期 单位秒
	SignExpire time.Duration `mapstructure:"sign_expire"`
}

type AuthConfig struct {
	JwtSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenExpire  time.Duration `mapstructure:"access_token_expire"`
	RefreshTokenExpire time.Duration `mapstructure:"refresh_token_expire"`
}

type SnowflakeConfig struct {
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	FileName   string `mapstructure:"file_name"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MysqlConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type AmapConfig struct {
	Key      string `mapstructure:"key"`
	RegeoURL string `mapstructure:"regeo_url"`
}

func LoadConfig() *AppConfig {

	var conf = new(AppConfig)

	var configPath string
	flag.StringVar(&configPath, "cfg", "./"+RuntimeDir+"/"+defaultConfigName, "配置文件路径")
	flag.Parse()

	// 设置配置文件路径
	viper.SetConfigFile(configPath)
	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("viper.ReadInConfig failed, err:%v\n", err)
		panic(err)
	}

	// 将读取的配置绑定到结构体变量
	if err := viper.Unmarshal(conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
		panic(err)
	}

	// 监听配置文件变化
	viper.WatchConfig()
	// 注册回调函数 当配置文件变化时 更新配置
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s", e.Name)
		if err := viper.Unmarshal(conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
			panic(err)
		}
	})

	return conf
}
