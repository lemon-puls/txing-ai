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
			"List existing AI models (id, name, description, tags, high_context/multimodal/default flags). "+
				"ALWAYS call this tool first to learn what models already exist, avoid duplicates, "+
				"and keep naming/description/tag style consistent before proposing anything.",
			deps.listModels)
		if err != nil {
			panic(err)
		}
		previewTool, err := toolutils.InferTool(
			ModelPreviewToolName,
			"Validate and canonicalize a model-entry proposal before presenting it to the admin. "+
				"Pass the name/description/flags/tags you intend to submit, plus id when optimizing an existing model. "+
				"It checks field constraints, checks the database for duplicates, and returns the canonical proposal JSON. "+
				"ALWAYS call this tool right before presenting a proposal, "+
				"and present its proposal JSON as-is without modifying any field.",
			deps.previewModel)
		if err != nil {
			panic(err)
		}
		return mytool.WrapSafeTools([]tool.BaseTool{listTool, previewTool})

	case OpsPageChannels:
		listTool, err := toolutils.InferTool(
			ChannelListToolName,
			"List existing channels (id, name, type, endpoint, models, priority/weight/retry, status, model mappings). "+
				"Secrets are redacted: you only see secretConfigured/secretKeyCount and never receive API keys. "+
				"ALWAYS call this tool first to learn what channels already exist and avoid duplicates "+
				"before proposing anything.",
			deps.listChannels)
		if err != nil {
			panic(err)
		}
		previewTool, err := toolutils.InferTool(
			ChannelPreviewToolName,
			"Validate and canonicalize a channel-entry proposal before presenting it to the admin. "+
				"Pass the name/type/endpoint/models/priority/weight/retry/status/mappings you intend to submit, "+
				"plus id when optimizing an existing channel. There is NO secret parameter: the admin types the "+
				"secret manually on the confirmation card. It checks field constraints, mapping structure, "+
				"duplicates, and returns the canonical proposal JSON. ALWAYS call this tool right before "+
				"presenting a proposal, and present its proposal JSON as-is without modifying any field.",
			deps.previewChannel)
		if err != nil {
			panic(err)
		}
		return mytool.WrapSafeTools([]tool.BaseTool{listTool, previewTool})

	case OpsPagePresets:
		listTool, err := toolutils.InferTool(
			PresetListToolName,
			"List existing AI assistant presets (id, name, description, truncated system prompt, tags, official flag). "+
				"ALWAYS call this tool first to learn what assistants already exist, avoid duplicates, "+
				"and keep naming/description/tag style consistent before proposing anything.",
			deps.listPresets)
		if err != nil {
			panic(err)
		}
		previewTool, err := toolutils.InferTool(
			PresetPreviewToolName,
			"Validate and canonicalize an AI-assistant-preset proposal before presenting it to the admin. "+
				"Pass the name/description/context (the system prompt)/tags/official you intend to submit, "+
				"plus id when optimizing an existing preset. It checks field constraints, the tag enum, "+
				"duplicates, and returns the canonical proposal JSON. ALWAYS call this tool right before "+
				"presenting a proposal, and present its proposal JSON as-is without modifying any field.",
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
		"Fetch raw facts about a website or a GitHub repository. For github.com/{owner}/{repo} URLs "+
			"it calls the GitHub API (name, description, topics, homepage, avatar); otherwise it scrapes "+
			"the page (title, description, text excerpt). It also extracts the site icon, uploads it to "+
			"COS and returns a presigned URL as avatarCandidate. ALWAYS call this tool first, before "+
			"proposing anything, and base your proposal only on the data it returns.",
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
		"Validate and canonicalize a website-entry proposal before presenting it to the admin. "+
			"Pass the name/description/url/avatar/tags you intend to submit. It checks field constraints, "+
			"normalizes the URL, checks the database for duplicates, and returns the canonical proposal JSON "+
			"that will be rendered as a preview card (tags matching the preset category set are reordered first). "+
			"ALWAYS call this tool right before presenting a proposal, "+
			"and present its proposal JSON as-is without modifying any field.",
		deps.previewWebsite)
	if err != nil {
		panic(err)
	}
	return previewTool
}
