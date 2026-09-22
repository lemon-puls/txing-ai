package domain

// OpsAgentLog 后台运营助手运行记录（审计用）
// 记录每次运营助手对话的输入、工具调用、提案与输出，便于追溯与排查
type OpsAgentLog struct {
	BaseModel
	UserID int64  `gorm:"column:user_id;type:bigint;not null;index;comment:操作用户ID" json:"userId"`
	Page   string `gorm:"column:page;type:varchar(50);comment:来源页面标识" json:"page"`
	// 页面上下文 JSON（page/draft 等）
	Context string `gorm:"column:context;type:text;comment:页面上下文JSON" json:"context"`
	// 对话消息 JSON（[{role,content}]）
	Input string `gorm:"column:input;type:text;comment:对话消息JSON" json:"input"`
	// 工具调用记录 JSON（含参数与结果）
	ToolCalls string `gorm:"column:tool_calls;type:mediumtext;comment:工具调用记录JSON" json:"toolCalls"`
	// 最终结构化提案 JSON（如网站录入提案）
	Proposal string `gorm:"column:proposal;type:text;comment:结构化提案JSON" json:"proposal"`
	Output   string `gorm:"column:output;type:mediumtext;comment:模型最终输出" json:"output"`
	Model    string `gorm:"column:model;type:varchar(100);comment:使用的模型名" json:"model"`
	// running/completed/failed/interrupted
	Status string `gorm:"column:status;type:varchar(20);default:'running';comment:运行状态" json:"status"`
	Error  string `gorm:"column:error;type:text;comment:错误详情" json:"error"`
	// 耗时（毫秒）
	DurationMs int64 `gorm:"column:duration_ms;type:bigint;comment:耗时(毫秒)" json:"durationMs"`
}

func (OpsAgentLog) TableName() string {
	return "ops_agent_logs"
}
