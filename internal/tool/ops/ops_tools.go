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

// OpsToolDeps 运营助手工具的依赖（按请求构造，非全局单例）
// OpsToolDeps holds the dependencies for ops tools. Tools are built per
// request (not via a global registry) so each request captures its own
// DB handle, COS client and the requesting admin's user id.
type OpsToolDeps struct {
	DB     *gorm.DB
	COS    *utils.COSClient
	UserID int64 // COS 对象路径前缀 / COS object path prefix
}

// ProvideOpsTools 构建运营助手专用工具集（只读，不含任何写库/删除操作）
// ProvideOpsTools builds the dedicated ops tool set. Read-only by design:
// the agent proposes, the admin confirms and the existing admin API writes.
func ProvideOpsTools(deps OpsToolDeps) []tool.BaseTool {
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

	// 复用统一的容错包装：工具报错时返回错误信息给 LLM 而非中断流程
	return mytool.WrapSafeTools([]tool.BaseTool{fetchTool, previewTool})
}
