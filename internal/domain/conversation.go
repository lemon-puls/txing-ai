package domain

import "txing-ai/internal/global"

type Conversation struct {
	BaseModel
	Auth      bool   `gorm:"type:boolean;not null;default:false;comment:是否已认证" json:"auth"`
	UserID    int64  `gorm:"type:bigint;not null;index;comment:用户ID" json:"userId"`
	Name      string `gorm:"type:varchar(255);comment:会话名称" json:"name"`
	Message   string `gorm:"type:mediumtext;comment:会话消息记录" json:"message"`
	Model     string `gorm:"type:varchar(50);not null;comment:使用的模型" json:"model"`
	EnableWeb bool   `gorm:"type:boolean;not null;default:false;comment:是否启用网页搜索" json:"enableWeb"`
	Context   int    `gorm:"type:int;not null;default:0;comment:上下文长度" json:"context"`

	// 可选的模型参数
	MaxTokens         *int     `gorm:"type:int;comment:最大token数" json:"maxTokens,omitempty"`
	Temperature       *float32 `gorm:"type:float;comment:温度参数" json:"temperature,omitempty"`
	TopP              *float32 `gorm:"type:float;comment:Top-P采样参数" json:"topP,omitempty"`
	TopK              *int     `gorm:"type:int;comment:Top-K采样参数" json:"topK,omitempty"`
	PresencePenalty   *float32 `gorm:"type:float;comment:存在惩罚参数" json:"presencePenalty,omitempty"`
	FrequencyPenalty  *float32 `gorm:"type:float;comment:频率惩罚参数" json:"frequencyPenalty,omitempty"`
	RepetitionPenalty *float32 `gorm:"type:float;comment:重复惩罚参数" json:"repetitionPenalty,omitempty"`

	// 非数据库字段
	FormattedMessage []global.Message `gorm:"-" json:"formattedMessage"`
}
