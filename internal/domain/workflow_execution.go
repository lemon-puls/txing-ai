package domain

// WorkflowExecution 工作流执行记录（用于 AI 对话中 @ 应用的上下文注入）
type WorkflowExecution struct {
	BaseModel
	ConversationID int64  `gorm:"column:conversation_id;type:bigint;not null;index;comment:会话ID" json:"conversationId"`
	WorkflowID     int64  `gorm:"column:workflow_id;type:bigint;not null;index;comment:工作流ID" json:"workflowId"`
	Inputs         string `gorm:"column:inputs;type:text;comment:各字段输入值JSON" json:"inputs"`
	FileRefs       string `gorm:"column:file_refs;type:text;comment:文件字段引用JSON" json:"fileRefs"`
	Output         string `gorm:"column:output;type:mediumtext;comment:完整输出内容" json:"output"`
	Artifacts      string `gorm:"column:artifacts;type:text;comment:生成的文件产物JSON" json:"artifacts"`
	Status         string `gorm:"column:status;type:varchar(20);default:'completed';comment:执行状态: completed/failed" json:"status"`
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}
