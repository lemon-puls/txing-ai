package ops

import "txing-ai/internal/global"

// 工具名称（与 ProvideOpsTools 注册名一致）
// Tool names, must match the names registered in ProvideOpsTools.
const (
	ModelListToolName    = "model_list_tool"
	ModelPreviewToolName = "model_preview_tool"

	ChannelListToolName    = "channel_list_tool"
	ChannelPreviewToolName = "channel_preview_tool"

	PresetListToolName    = "preset_list_tool"
	PresetPreviewToolName = "preset_preview_tool"
)

// 提案类型（前端 OpsChatPanel 按 proposal.type 分发到对应提案卡片）
// Proposal type discriminators; the frontend dispatches preview cards by proposal.type.
const (
	ProposalTypeWebsite = "website"
	ProposalTypeModel   = "model"
	ProposalTypeChannel = "channel"
	ProposalTypePreset  = "preset"
)

// PreviewToolProposalTypes 预览工具名 → 提案类型 的注册表。
// controller 据此拦截工具完成事件并统一下发 proposal SSE 帧，新增提案类型只需在此注册。
// PreviewToolProposalTypes maps preview tool names to proposal types. The controller
// intercepts completed tool calls by this registry and emits the proposal SSE frame.
var PreviewToolProposalTypes = map[string]string{
	WebsitePreviewToolName: ProposalTypeWebsite,
	ModelPreviewToolName:   ProposalTypeModel,
	ChannelPreviewToolName: ProposalTypeChannel,
	PresetPreviewToolName:  ProposalTypePreset,
}

// ModelProposal 模型录入/优化提案
// 字段与 dto.CreateModelReq/UpdateModelReq 对应，前端确认后直接提交现有管理接口。
// ModelProposal is the canonical model-entry proposal. Id > 0 marks an
// optimization proposal for an existing model (confirm then calls the update API).
type ModelProposal struct {
	Type        string `json:"type"`                   // 固定 "model"，固定值 "model"
	Id          int64  `json:"id,omitempty"`           // 目标模型ID，>0 表示优化提案 / target model id, >0 means update proposal
	Name        string `json:"name"`                   // 模型名称 / model name
	Description string `json:"description,omitempty"`  // 模型描述 / model description
	Default     bool   `json:"default,omitempty"`      // 是否默认模型 / default model flag
	HighContext bool   `json:"high_context,omitempty"` // 是否高上下文 / high-context flag（与接口字段名一致）
	Multimodal  bool   `json:"multimodal,omitempty"`   // 是否多模态 / multimodal flag
	Avatar      string `json:"avatar,omitempty"`       // 头像（可空，更新时空值不覆盖原头像）/ avatar, omitted on update keeps original
	Tags        string `json:"tags,omitempty"`         // 标签，逗号分隔 / comma-separated tags
}

// ChannelProposal 渠道录入/优化提案。
// 安全约束：绝不包含 Secret 字段——密钥永远不由 AI 读写，创建时由管理员在卡片中手动填写，
// 优化时前端确认阶段自行取回原密钥提交。
// ChannelProposal is the channel proposal. SECURITY: it deliberately has no secret
// field — the agent never reads or writes API keys; the admin types the secret on
// the card for creates, and the confirm flow re-fetches the stored secret for updates.
type ChannelProposal struct {
	Type        string                `json:"type"`               // 固定 "channel"，固定值 "channel"
	Id          int64                 `json:"id,omitempty"`       // 目标渠道ID，>0 表示优化提案 / target channel id
	Name        string                `json:"name"`               // 渠道名称 / channel name
	ChannelType string                `json:"channelType"`        // 渠道类型 / channel type（与提案类型字段名错开）
	Priority    int                   `json:"priority"`           // 优先级 / priority（0 有意义，不用 omitempty）
	Weight      int                   `json:"weight"`             // 权重 / weight
	Retry       int                   `json:"retry"`              // 重试次数 / retry times
	Models      []string              `json:"models"`             // 支持的模型列表 / supported models
	Endpoint    string                `json:"endpoint"`           // 服务地址 / service endpoint
	Status      bool                  `json:"status"`             // 启用状态 / enabled
	Mappings    []global.ModelMapping `json:"mappings,omitempty"` // 模型映射 / model mappings
}

// PresetProposal AI 助手（预设）录入/优化提案
// 字段与 dto.CreatePresetReq/UpdatePresetReq 对应；Context 为面向终端用户的系统提示词。
// PresetProposal is the preset (AI assistant) proposal; Context is the system prompt.
type PresetProposal struct {
	Type        string `json:"type"`                  // 固定 "preset"，固定值 "preset"
	Id          int64  `json:"id,omitempty"`          // 目标预设ID，>0 表示优化提案 / target preset id
	Name        string `json:"name"`                  // 预设名称 / preset name
	Description string `json:"description,omitempty"` // 预设描述 / preset description
	Context     string `json:"context,omitempty"`     // 上下文设定（系统提示词）/ system prompt
	Tags        string `json:"tags,omitempty"`        // 标签，逗号分隔 / comma-separated tags（限定枚举值）
	Official    bool   `json:"official,omitempty"`    // 是否官方预设 / official flag
}
