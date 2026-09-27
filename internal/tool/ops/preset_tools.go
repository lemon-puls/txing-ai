package ops

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"txing-ai/internal/domain"
)

// PresetPresetTags AI 助手（预设）标签枚举，提案标签必须从中选择。
// 前端镜像见 static/frontend/src/views/admin/preset/PresetList.vue 的 tagOptions，两处需同步维护。
var PresetPresetTags = []string{
	"popular",
	"tools",
	"writing",
	"coding",
	"learning",
	"life",
	"other",
}

// presetContextMaxRunes 上下文设定（系统提示词）长度上限
const presetContextMaxRunes = 8000

// presetListContextMaxRunes 列表条目返回给 LLM 的上下文设定截断长度（控制 token 消耗）
const presetListContextMaxRunes = 500

// presetListRequest preset_list_tool 请求参数
type presetListRequest struct {
	// 按预设名称模糊检索的关键字，可为空
	Keyword string `json:"keyword,omitempty"`
	// 返回条数上限，缺省 20，最大 50
	Limit int `json:"limit,omitempty"`
}

// presetItem 预设列表条目（Context 截断返回，控制 token 消耗）
type presetItem struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// 上下文设定（系统提示词，截断至前 500 字）
	Context string `json:"context,omitempty"`
	Tags    string `json:"tags,omitempty"`
	// 是否官方预设
	Official bool `json:"official"`
}

// presetListResult preset_list_tool 返回结果
type presetListResult struct {
	Total   int64        `json:"total"`
	Presets []presetItem `json:"presets"`
}

// presetPreviewRequest preset_preview_tool 请求参数（LLM 拟提交的预设资料）。
// Id > 0 表示对已有预设的优化提案，其余字段为优化后的完整目标值。
type presetPreviewRequest struct {
	Id          int64    `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Context     string   `json:"context"`
	Tags        []string `json:"tags"`
	Official    bool     `json:"official"`
}

// presetPreviewResult preset_preview_tool 返回结果（信封与 websitePreviewResult 同形）
type presetPreviewResult struct {
	// "ok" | "duplicate" | "invalid"
	Status string `json:"status"`
	// 人类可读的校验说明（中文）
	Message string `json:"message,omitempty"`
	// 校验通过后的规范提案（invalid 时为 nil；duplicate 时附带便于用户决策）
	Proposal *PresetProposal `json:"proposal,omitempty"`
	// 优化提案时的原值（Id>0 时返回，Context 同样截断）
	Original *presetItem `json:"original,omitempty"`
	// 重复记录的 ID（duplicate 时返回）
	ExistingID int64 `json:"existingId,omitempty"`
}

// listPresets 查询预设列表（只读），供 LLM 了解已有助手、保持命名/标签风格一致
func (d OpsToolDeps) listPresets(ctx context.Context, req *presetListRequest) (presetListResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	query := d.DB.Model(&domain.Preset{})
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return presetListResult{}, fmt.Errorf("查询预设总数失败: %w", err)
	}

	var presets []domain.Preset
	if err := query.Order("id DESC").Limit(limit).Find(&presets).Error; err != nil {
		return presetListResult{}, fmt.Errorf("查询预设列表失败: %w", err)
	}

	items := make([]presetItem, 0, len(presets))
	for _, p := range presets {
		items = append(items, presetItemFromDomain(p))
	}
	return presetListResult{Total: total, Presets: items}, nil
}

// previewPreset 校验并规范化预设提案（只读，不写库）
func (d OpsToolDeps) previewPreset(ctx context.Context, req *presetPreviewRequest) (presetPreviewResult, error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	contextText := strings.TrimSpace(req.Context)

	// 字段长度校验（与前端表单规则一致）
	if n := len([]rune(name)); n < 2 || n > 50 {
		return presetPreviewResult{Status: "invalid", Message: fmt.Sprintf("助手名称长度需在 2-50 个字之间，当前 %d 个字", n)}, nil
	}
	if n := len([]rune(description)); n > 200 {
		return presetPreviewResult{Status: "invalid", Message: fmt.Sprintf("助手描述不能超过 200 个字，当前 %d 个字", n)}, nil
	}
	if contextText == "" {
		return presetPreviewResult{Status: "invalid", Message: "上下文设定（系统提示词）不能为空"}, nil
	}
	if n := len([]rune(contextText)); n > presetContextMaxRunes {
		return presetPreviewResult{Status: "invalid", Message: fmt.Sprintf("上下文设定不能超过 %d 个字，当前 %d 个字", presetContextMaxRunes, n)}, nil
	}

	// 标签：限定枚举值且最多 3 个
	tags := normalizeTags(req.Tags)
	if len(tags) == 0 {
		return presetPreviewResult{Status: "invalid", Message: "至少需要 1 个标签"}, nil
	}
	if len(tags) > 3 {
		return presetPreviewResult{Status: "invalid", Message: fmt.Sprintf("标签最多 3 个，当前 %d 个", len(tags))}, nil
	}
	for _, t := range tags {
		if !knownPresetTag(t) {
			return presetPreviewResult{Status: "invalid", Message: fmt.Sprintf("标签 %q 不在允许范围（%v）内", t, PresetPresetTags)}, nil
		}
	}

	// 优化提案：先加载原值（同时验证目标存在）
	var original *presetItem
	if req.Id > 0 {
		var existing domain.Preset
		err := d.DB.First(&existing, req.Id).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return presetPreviewResult{Status: "invalid", Message: fmt.Sprintf("要优化的助手不存在（id=%d）", req.Id)}, nil
			}
			return presetPreviewResult{}, fmt.Errorf("查询预设失败: %w", err)
		}
		item := presetItemFromDomain(existing)
		original = &item
	}

	// 名称查重（只读查询；优化时排除自身）
	if d.DB != nil {
		query := d.DB.Model(&domain.Preset{}).Where("name = ?", name)
		if req.Id > 0 {
			query = query.Where("id <> ?", req.Id)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return presetPreviewResult{}, fmt.Errorf("查询重复预设失败: %w", err)
		}
		if count > 0 {
			var dup domain.Preset
			_ = d.DB.Where("name = ?", name).First(&dup).Error
			return presetPreviewResult{
				Status:     "duplicate",
				Message:    "已存在同名助手，请勿重复录入",
				Proposal:   buildPresetProposal(name, description, contextText, req, tags),
				Original:   original,
				ExistingID: dup.Id,
			}, nil
		}
	}

	return presetPreviewResult{
		Status:   "ok",
		Message:  "校验通过",
		Proposal: buildPresetProposal(name, description, contextText, req, tags),
		Original: original,
	}, nil
}

// buildPresetProposal 构建规范预设提案
func buildPresetProposal(name, description, contextText string, req *presetPreviewRequest, tags []string) *PresetProposal {
	return &PresetProposal{
		Type:        ProposalTypePreset,
		Id:          req.Id,
		Name:        name,
		Description: description,
		Context:     contextText,
		Tags:        strings.Join(tags, ","),
		Official:    req.Official,
	}
}

// presetItemFromDomain 领域预设转列表条目（Context 截断）
func presetItemFromDomain(p domain.Preset) presetItem {
	return presetItem{
		Id:          p.Id,
		Name:        p.Name,
		Description: p.Description,
		Context:     truncateRunes(p.Context, presetListContextMaxRunes),
		Tags:        p.Tags,
		Official:    p.Official,
	}
}

// knownPresetTag 是否为允许的预设标签
func knownPresetTag(t string) bool {
	for _, known := range PresetPresetTags {
		if known == t {
			return true
		}
	}
	return false
}

// truncateRunes 按 rune 截断字符串，超出部分以省略号结尾
func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…（已截断）"
}
