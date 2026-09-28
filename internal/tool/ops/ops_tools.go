package ops

import (
	"gorm.io/gorm"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"

	mytool "txing-ai/internal/tool"
	"txing-ai/internal/utils"
)

// 工具名称
const (
	WebsiteFetchToolName   = "website_fetch_tool"
	WebsitePreviewToolName = "website_preview_tool"
)

// OpsPage 页面标识（与前端各管理页传入的 context.page 一致）
const (
	OpsPageWebsites = "websites"
	OpsPageModels   = "models"
	OpsPageChannels = "channels"
	OpsPagePresets  = "presets"
)

// OpsToolDeps 运营助手工具的依赖（按请求构造，非全局单例）
// OpsToolDeps holds the dependencies for ops tools. Tools are built per
// request (not via a global registry) so each request captures its own
// DB handle, COS client and the requesting admin's user id.
type OpsToolDeps struct {
	DB     *gorm.DB
	COS    *utils.COSClient
	UserID int64 // COS 对象路径前缀 / COS object path prefix
}

// ProvideOpsTools 按页面构建运营助手专用工具集（只读，不含任何写库/删除操作）。
// 每个页面只暴露该页相关的查询 + 预览校验工具：控制工具数量（MaxToolRounds 有限）、
// 收窄提示词选择面、避免跨页误用；未知页面回退为网站工具（兼容旧会话）。
// ProvideOpsTools builds the page-scoped ops tool set. Read-only by design:
// the agent proposes, the admin confirms and the existing admin API writes.
func ProvideOpsTools(deps OpsToolDeps, page string) []tool.BaseTool {
	switch page {
	case OpsPageModels:
		listTool, err := toolutils.InferTool(
			ModelListToolName,
			"查询已有的 AI 模型列表（id、名称、描述、标签、高上下文/多模态/默认标记）。"+
				"生成任何提案之前必须先调用本工具：了解已有模型、避免重名，并保持命名/描述/标签风格一致。",
			deps.listModels)
		if err != nil {
			panic(err)
		}
		previewTool, err := toolutils.InferTool(
			ModelPreviewToolName,
			"在向管理员呈现之前，校验并规范化模型录入/优化提案。"+
				"传入你拟提交的名称/描述/能力开关/标签；优化已有模型时必须带上 id。"+
				"工具会校验字段约束、对数据库查重，并返回规范化的提案 JSON。"+
				"呈现提案之前必须调用本工具；系统会把返回的提案 JSON 自动渲染为预览卡片，不要在回复中粘贴 JSON。",
			deps.previewModel)
		if err != nil {
			panic(err)
		}
		return mytool.WrapSafeTools([]tool.BaseTool{listTool, previewTool})

	case OpsPageChannels:
		listTool, err := toolutils.InferTool(
			ChannelListToolName,
			"查询已有的渠道列表（id、名称、类型、服务地址、模型列表、优先级/权重/重试、状态、模型映射）。"+
				"密钥已脱敏：只能看到 secretConfigured/secretKeyCount，永远不会拿到 API 密钥。"+
				"生成任何提案之前必须先调用本工具：了解已有渠道、避免重名。",
			deps.listChannels)
		if err != nil {
			panic(err)
		}
		previewTool, err := toolutils.InferTool(
			ChannelPreviewToolName,
			"在向管理员呈现之前，校验并规范化渠道录入/优化提案。"+
				"传入你拟提交的名称/类型/服务地址/模型列表/优先级/权重/重试/状态/映射；优化已有渠道时必须带上 id。"+
				"本工具没有密钥参数：密钥由管理员在确认卡片中手动填写。"+
				"工具会校验字段约束、映射结构与查重，并返回规范化的提案 JSON。"+
				"呈现提案之前必须调用本工具；系统会把返回的提案 JSON 自动渲染为预览卡片，不要在回复中粘贴 JSON。",
			deps.previewChannel)
		if err != nil {
			panic(err)
		}
		return mytool.WrapSafeTools([]tool.BaseTool{listTool, previewTool})

	case OpsPagePresets:
		listTool, err := toolutils.InferTool(
			PresetListToolName,
			"查询已有的 AI 助手预设列表（id、名称、描述、截断的系统提示词、标签、官方标记）。"+
				"生成任何提案之前必须先调用本工具：了解已有助手、避免重名，并保持命名/描述/标签风格一致。",
			deps.listPresets)
		if err != nil {
			panic(err)
		}
		previewTool, err := toolutils.InferTool(
			PresetPreviewToolName,
			"在向管理员呈现之前，校验并规范化 AI 助手预设录入/优化提案。"+
				"传入你拟提交的名称/描述/上下文设定（系统提示词）/标签/官方标记；优化已有助手时必须带上 id。"+
				"工具会校验字段约束、标签枚举与查重，并返回规范化的提案 JSON。"+
				"呈现提案之前必须调用本工具；系统会把返回的提案 JSON 自动渲染为预览卡片，不要在回复中粘贴 JSON。",
			deps.previewPreset)
		if err != nil {
			panic(err)
		}
		return mytool.WrapSafeTools([]tool.BaseTool{listTool, previewTool})

	default:
		// websites / 未知页面：保持原有网站工具（兼容旧会话）
		return mytool.WrapSafeTools([]tool.BaseTool{mustWebsiteFetchTool(deps), mustWebsitePreviewTool(deps)})
	}
}

// PageToolNames 返回页面对应的工具名列表（供 ExecuteLLM 的 ToolNames 使用，与 ProvideOpsTools 的注册一一对应）
func PageToolNames(page string) []string {
	switch page {
	case OpsPageModels:
		return []string{ModelListToolName, ModelPreviewToolName}
	case OpsPageChannels:
		return []string{ChannelListToolName, ChannelPreviewToolName}
	case OpsPagePresets:
		return []string{PresetListToolName, PresetPreviewToolName}
	default:
		return []string{WebsiteFetchToolName, WebsitePreviewToolName}
	}
}

// mustWebsiteFetchTool 构建网站抓取工具（原 ProvideOpsTools 的默认分支逻辑）
func mustWebsiteFetchTool(deps OpsToolDeps) tool.BaseTool {
	fetchTool, err := toolutils.InferTool(
		WebsiteFetchToolName,
		"获取网站或 GitHub 仓库的真实信息。github.com/{owner}/{repo} 形式的地址会调用 GitHub API"+
			"（名称、描述、topics、主页、头像）；其他地址会抓取页面（标题、描述、正文摘要）。"+
			"同时会提取站点图标并上传 COS，以 avatarCandidate 返回预签名地址。"+
			"生成任何提案之前必须先调用本工具，且提案只能基于它返回的数据。",
		deps.fetchWebsite)
	if err != nil {
		panic(err)
	}
	return fetchTool
}

// mustWebsitePreviewTool 构建网站提案校验工具（原 ProvideOpsTools 的默认分支逻辑）
func mustWebsitePreviewTool(deps OpsToolDeps) tool.BaseTool {
	previewTool, err := toolutils.InferTool(
		WebsitePreviewToolName,
		"在向管理员呈现之前，校验并规范化网站录入提案。"+
			"传入你拟提交的名称/描述/网址/头像/标签。工具会校验字段约束、规范化 URL、对数据库查重，"+
			"并返回规范化的提案 JSON（匹配内定分类的标签会被排到前面）。"+
			"呈现提案之前必须调用本工具；系统会把返回的提案 JSON 自动渲染为预览卡片，不要在回复中粘贴 JSON。",
		deps.previewWebsite)
	if err != nil {
		panic(err)
	}
	return previewTool
}
