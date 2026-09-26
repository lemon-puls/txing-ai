package global

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/spf13/viper"
)

const (
	// 运行时数据包括配置、日志都统一放这里，方便后续迁移
	RuntimeDir = "runtime"
	// 默认配置文件名
	defaultConfigName = "config.yaml"
)

var (
	appConfig  = new(AppConfig)
	configOnce sync.Once
)

type AppConfig struct {
	*ServerConfig      `mapstructure:"server"`
	*LogConfig         `mapstructure:"log"`
	*MysqlConfig       `mapstructure:"mysql"`
	*RedisConfig       `mapstructure:"redis"`
	*SnowflakeConfig   `mapstructure:"snowflake"`
	*AuthConfig        `mapstructure:"auth"`
	*CosConfig         `mapstructure:"cos"`
	*AmapConfig        `mapstructure:"amap"`
	*AWSConfig         `mapstructure:"aws"`
	*SearchAPIConfig   `mapstructure:"searchapi"`
	*ImageSearchConfig `mapstructure:"image_search"`
	*LocalUploadConfig `mapstructure:"local_upload"`
	*OpsAgentConfig    `mapstructure:"ops_agent"`
	*WikiConfig        `mapstructure:"wiki"`
	*ObservabilityConfig `mapstructure:"observability"`
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
	// CDN 加速地址
	CDNURL string `mapstructure:"cdn_url"`
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

type AWSConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	// 预签名URL有效期 单位秒
	SignExpire time.Duration `mapstructure:"sign_expire"`
}

type SearchAPIConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	ApiKey   string `mapstructure:"api_key"`
	Engine   string `mapstructure:"Engine"`
}

type ImageSearchConfig struct {
	Sougou *ImageSearchSougou `mapstructure:"sougou"`
}

type ImageSearchSougou struct {
	ID       string `mapstructure:"id"`
	Key      string `mapstructure:"key"`
	Endpoint string `mapstructure:"endpoint"`
}

type LocalUploadConfig struct {
	Dir     string `mapstructure:"dir"`
	MaxSize int    `mapstructure:"max_size"`
}

// OpsAgentConfig 后台运营助手配置
type OpsAgentConfig struct {
	// 模型名，需在渠道管理中已配置；经 ChannelModelResolver 解析
	Model string `mapstructure:"model"`
	// 最大工具调用轮次，<=0 时使用默认值
	MaxToolRounds int `mapstructure:"max_tool_rounds"`
	// GitHub API Token（可选），提升 API 限额
	GithubToken string `mapstructure:"github_token"`
}

// WikiConfig LLM Wiki 知识库配置（docs/llmwiki_design.md §9）
type WikiConfig struct {
	// 总开关（false 时公开问答端点直接下线）
	Enabled bool `mapstructure:"enabled"`
	// agent 引擎实现选择（D14）：builtin | hermes | ...，当前仅 builtin
	AgentImpl string `mapstructure:"agent_impl"`
	// ingest 编译用模型名，留空回退 qa_model
	IngestModel string `mapstructure:"ingest_model"`
	// 问答用模型名，留空走默认解析
	QAModel string `mapstructure:"qa_model"`
	// 最大工具调用轮次，<=0 时使用默认值
	MaxToolRounds int `mapstructure:"max_tool_rounds"`
	// 单 IP 每小时问答上限，<=0 时使用默认值
	RateLimitPerIPPerHour int `mapstructure:"rate_limit_per_ip_per_hour"`
	// 全站每日问答总量上限，<=0 时使用默认值
	DailyQuota int `mapstructure:"daily_quota"`
}

// ObservabilityConfig 可观测性配置（docs/observability_design.md）
type ObservabilityConfig struct {
	// 总开关：false 时所有埋点即时 no-op（热生效）
	Enabled bool `mapstructure:"enabled"`
	// 采样频率（registry.Gather 间隔）
	SampleInterval time.Duration `mapstructure:"sample_interval"`
	// 环形缓冲槽粒度（重启生效）
	SlotInterval time.Duration `mapstructure:"slot_interval"`
	// 内存保留时长（<=48h，重启清零；长期统计走 MySQL 业务表聚合）
	Retention time.Duration `mapstructure:"retention"`
	// 序列数上限（基数防御）
	MaxSeries int `mapstructure:"max_series"`
	// external 外部抓取模式
	External *ObservabilityExternalConfig `mapstructure:"external"`
}

// ObservabilityExternalConfig 外部 Prometheus/VictoriaMetrics 抓取配置
type ObservabilityExternalConfig struct {
	// 暴露 /metrics 端点（路由注册在启动时，改后需重启）
	Enabled bool `mapstructure:"enabled"`
	// 非空时 /metrics 校验 Authorization: Bearer <token>
	BearerToken string `mapstructure:"bearer_token"`
}

func LoadConfig() *AppConfig {
	configOnce.Do(func() {
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
		if err := viper.Unmarshal(appConfig); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
			panic(err)
		}

		// 设置默认值
		if appConfig.LocalUploadConfig.MaxSize == 0 {
			appConfig.LocalUploadConfig.MaxSize = 10 * 1024 * 1024 // 默认 10MB
		}

		// 监听配置文件变化
		viper.WatchConfig()
		// 注册回调函数 当配置文件变化时 更新配置
		viper.OnConfigChange(func(e fsnotify.Event) {
			fmt.Printf("Config file changed: %s", e.Name)
			if err := viper.Unmarshal(appConfig); err != nil {
				fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
				panic(err)
			}
		})
	})

	return appConfig
}
