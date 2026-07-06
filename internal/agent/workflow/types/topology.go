package types

// TopoNode 拓扑节点
type TopoNode struct {
	Id       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Data     NodeData `json:"data"`
}

// Position 节点位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// TopoEdge 拓扑边
type TopoEdge struct {
	Id           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"`
	TargetHandle string `json:"targetHandle,omitempty"`
}

// SchemaField Schema 字段定义
type SchemaField struct {
	Name        string `json:"name"`                    // 字段名称（英文标识，用作表单字段名）
	Type        string `json:"type"`                    // 字段类型: file, text, textarea
	Label       string `json:"label,omitempty"`         // 显示标签（如"上传简历"）
	Placeholder string `json:"placeholder,omitempty"`   // 占位提示文字
	Required    bool   `json:"required,omitempty"`      // 是否必填
	Accept      string `json:"accept,omitempty"`        // 文件类型限制（type=file 时），如 ".pdf,.doc,.docx"
	Default     string `json:"default,omitempty"`       // 默认值
	Description string `json:"description,omitempty"`   // 字段描述说明
}

// WorkflowConfig 工作流级别配置
type WorkflowConfig struct {
	DefaultModel string        `json:"defaultModel,omitempty"` // 默认模型名称
	MaxRunSteps  int           `json:"maxRunSteps,omitempty"`  // 最大执行步数
	InputSchema  []SchemaField `json:"inputSchema,omitempty"`  // 输入 Schema
	OutputSchema []SchemaField `json:"outputSchema,omitempty"` // 输出 Schema
}

// Topology 工作流拓扑图结构
type Topology struct {
	Nodes  []TopoNode     `json:"nodes"`
	Edges  []TopoEdge     `json:"edges"`
	Config *WorkflowConfig `json:"config,omitempty"`
}

// NodeData 节点数据（配置直接放在 data 层级，与前端 JSON 结构一致）
type NodeData struct {
	NodeType          string              `json:"nodeType"`
	Label             string              `json:"label"`
	ModelConfig       *ModelConfig        `json:"modelConfig,omitempty"`
	ToolConfig        *ToolConfig         `json:"toolConfig,omitempty"`
	ConditionConf     *ConditionConfig    `json:"conditionConfig,omitempty"`
	CodeConfig        *CodeConfig         `json:"codeConfig,omitempty"`
	HTTPConfig        *HTTPConfig         `json:"httpConfig,omitempty"`
	SubWorkflowConfig *SubWorkflowConfig  `json:"subWorkflowConfig,omitempty"`
	AgentConfig       *AgentConfig        `json:"agentConfig,omitempty"`
	ParallelConfig    *ParallelConfig     `json:"parallelConfig,omitempty"`  // 并行组配置 / Parallel group config
	JoinConfig        *JoinConfig         `json:"joinConfig,omitempty"`      // 汇聚节点配置 / Join node config
	Extra             map[string]interface{} `json:"extra,omitempty"`        // 扩展字段，用于存储 parallelId 等 / Extra fields for parallelId etc.
}
