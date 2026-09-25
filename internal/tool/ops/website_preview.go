package ops

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"txing-ai/internal/domain"
)

// websitePreviewRequest website_preview_tool 请求参数（LLM 依据 fetch 结果填写）
type websitePreviewRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Avatar      string   `json:"avatar,omitempty"` // 预签名 URL，来自 fetch 工具的 avatarCandidate
	Tags        []string `json:"tags"`
}

// WebsiteProposal 规范化后的网站录入提案
// 字段与 dto.CreateWebsiteReq 一一对应，前端确认后直接提交现有创建接口
type WebsiteProposal struct {
	Type        string `json:"type"` // 固定 "website"
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Avatar      string `json:"avatar,omitempty"`
	Tags        string `json:"tags"` // 逗号分隔
}

// websitePreviewResult website_preview_tool 返回结果
type websitePreviewResult struct {
	// "ok" | "duplicate" | "invalid"
	Status string `json:"status"`
	// 人类可读的校验说明（中文）
	Message string `json:"message,omitempty"`
	// 校验通过后的规范提案（invalid 时为 nil；duplicate 时附带便于用户决策）
	Proposal *WebsiteProposal `json:"proposal,omitempty"`
	// 重复记录的 ID（duplicate 时返回）
	ExistingID int64 `json:"existingId,omitempty"`
}

// previewWebsite 校验并规范化网站录入提案（只读，不写库）
func (d OpsToolDeps) previewWebsite(ctx context.Context, req *websitePreviewRequest) (websitePreviewResult, error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)

	// URL 校验与规范化
	rawURL := strings.TrimSpace(req.URL)
	if rawURL != "" && !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || !isValidHost(parsed) || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return invalidResult("URL 不合法，必须是完整的 http(s) 地址"), nil
	}
	// 去掉尾斜杠统一形态，保证 "https://a.com" 与 "https://a.com/" 查重结果一致
	normalizedURL := parsed.Scheme + "://" + parsed.Host + strings.TrimSuffix(parsed.Path, "/")

	// 字段长度校验（与 dto.CreateWebsiteReq 的 binding 规则一致）
	if n := len([]rune(name)); n < 2 || n > 50 {
		return invalidResult(fmt.Sprintf("网站名称长度需在 2-50 个字之间，当前 %d 个字", n)), nil
	}
	if description == "" {
		return invalidResult("网站描述不能为空"), nil
	}
	if n := len([]rune(description)); n > 200 {
		return invalidResult(fmt.Sprintf("网站描述不能超过 200 个字，当前 %d 个字", n)), nil
	}
	tags := reorderTagsPresetFirst(normalizeTags(req.Tags))
	if len(tags) == 0 {
		return invalidResult("至少需要 1 个标签"), nil
	}

	// URL 查重（只读查询）
	if d.DB != nil {
		var existing domain.Website
		err := d.DB.Where("url = ?", normalizedURL).First(&existing).Error
		if err == nil {
			return websitePreviewResult{
				Status:     "duplicate",
				Message:    "该网站地址已存在，请勿重复录入",
				Proposal:   buildProposal(name, description, normalizedURL, strings.TrimSpace(req.Avatar), tags),
				ExistingID: existing.Id,
			}, nil
		}
		if err != gorm.ErrRecordNotFound {
			return websitePreviewResult{}, fmt.Errorf("查询重复网站失败: %w", err)
		}

		// 同名提示（非阻断，仅附加警告）
		var nameCount int64
		if err := d.DB.Model(&domain.Website{}).Where("name = ?", name).Count(&nameCount).Error; err == nil && nameCount > 0 {
			warn := fmt.Sprintf("注意：已存在同名网站（%d 条），建议确认是否为同一站点", nameCount)
			proposal := buildProposal(name, description, normalizedURL, strings.TrimSpace(req.Avatar), tags)
			return websitePreviewResult{
				Status:   "ok",
				Message:  warn,
				Proposal: proposal,
			}, nil
		}
	}

	return websitePreviewResult{
		Status:   "ok",
		Message:  "校验通过",
		Proposal: buildProposal(name, description, normalizedURL, strings.TrimSpace(req.Avatar), tags),
	}, nil
}

// buildProposal 构建规范提案
func buildProposal(name, description, rawURL, avatar string, tags []string) *WebsiteProposal {
	return &WebsiteProposal{
		Type:        "website",
		Name:        name,
		Description: description,
		Url:         rawURL,
		Avatar:      avatar,
		Tags:        strings.Join(tags, ","),
	}
}

// normalizeTags 清洗标签：去空白、去重、限制 1-5 个
func normalizeTags(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	tags := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		tags = append(tags, t)
		if len(tags) >= 5 {
			break
		}
	}
	return tags
}

// invalidResult 构建校验失败结果
func invalidResult(msg string) websitePreviewResult {
	return websitePreviewResult{
		Status:  "invalid",
		Message: msg,
	}
}

// domainRegex 合法域名（支持多级子域、单标签如 localhost、Punycode）
var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

// isValidHost 校验 URL 的主机名：非空，且为合法域名或 IP（v4/v6）
// net/url 对畸形主机（如 "https://::::"）过于宽松，这里做二次校验
func isValidHost(u *url.URL) bool {
	host := u.Hostname()
	if host == "" {
		return false
	}
	if strings.Contains(host, ":") {
		// IPv6（Hostname() 已去掉方括号），交给 net.ParseIP 严格校验
		return net.ParseIP(host) != nil
	}
	// 域名或 IPv4
	return net.ParseIP(host) != nil || domainRegex.MatchString(host)
}
