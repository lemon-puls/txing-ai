package ops

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"txing-ai/internal/domain"
	"txing-ai/internal/global"
)

// AdminChannelTypes 渠道管理页可选的渠道类型（与前端 ChannelList.vue 的 channelTypes 保持一致）
var AdminChannelTypes = []string{"火星引擎", "polo", "OpenAI", "Eino OpenAI"}

// channelMappingConditionKeys 模型映射条件允许的 key（对齐 global.ChannelModelMappingConditionDefaultVal）
var channelMappingConditionKeys = map[string]bool{
	"enableWeb": true,
	"type":      true,
}

// channelListRequest channel_list_tool 请求参数
type channelListRequest struct {
	// 按渠道名称模糊检索的关键字，可为空
	Keyword string `json:"keyword,omitempty"`
	// 返回条数上限，缺省 20，最大 50
	Limit int `json:"limit,omitempty"`
}

// channelItem 渠道列表条目。
// 安全约束：结构上不含 Secret 字段——密钥永不返回给 LLM，仅暴露脱敏统计：
// secretConfigured（是否已配置密钥）与 secretKeyCount（已配置的密钥条数）。
type channelItem struct {
	Id          int64                 `json:"id"`
	Name        string                `json:"name"`
	ChannelType string                `json:"type"` // JSON 字段名与 domain.Channel 对齐
	Endpoint    string                `json:"endpoint"`
	Priority    int                   `json:"priority"`
	Weight      int                   `json:"weight"`
	Retry       int                   `json:"retry"`
	Models      []string              `json:"models"`
	Status      bool                  `json:"status"`
	Mappings    []global.ModelMapping `json:"mappings"`
	// 密钥脱敏信息（永不返回密钥本身）
	SecretConfigured bool `json:"secretConfigured"`
	SecretKeyCount   int  `json:"secretKeyCount"`
}

// channelListResult channel_list_tool 返回结果
type channelListResult struct {
	Total    int64         `json:"total"`
	Channels []channelItem `json:"channels"`
	// 平台已录入的模型名列表：渠道提案的 models 与映射 sourceModel 只能从中选择
	AvailableModels []string `json:"availableModels"`
}

// channelPreviewRequest channel_preview_tool 请求参数（LLM 拟提交的渠道配置）。
// 安全约束：没有密钥参数——密钥由管理员在提案卡片中手动填写，AI 不参与。
type channelPreviewRequest struct {
	Id          int64                 `json:"id,omitempty"`
	Name        string                `json:"name"`
	ChannelType string                `json:"channelType"`
	Endpoint    string                `json:"endpoint"`
	Priority    int                   `json:"priority"`
	Weight      int                   `json:"weight"`
	Retry       int                   `json:"retry"`
	Models      []string              `json:"models"`
	Status      bool                  `json:"status"`
	Mappings    []global.ModelMapping `json:"mappings,omitempty"`
}

// channelPreviewResult channel_preview_tool 返回结果（信封与 websitePreviewResult 同形）
type channelPreviewResult struct {
	// "ok" | "duplicate" | "invalid"
	Status string `json:"status"`
	// 人类可读的校验说明（中文）
	Message string `json:"message,omitempty"`
	// 校验通过后的规范提案（invalid 时为 nil；duplicate 时附带便于用户决策）
	Proposal *ChannelProposal `json:"proposal,omitempty"`
	// 优化提案时的原值（Id>0 时返回，密钥字段已脱敏）
	Original *channelItem `json:"original,omitempty"`
	// 重复记录的 ID（duplicate 时返回）
	ExistingID int64 `json:"existingId,omitempty"`
	// 平台已录入的模型名（invalid 且因模型未录入时返回，便于 LLM 同轮自我纠正）
	AvailableModels []string `json:"availableModels,omitempty"`
}

// listChannels 查询渠道列表（只读，密钥脱敏），供 LLM 了解已有渠道、避免重复录入
func (d OpsToolDeps) listChannels(ctx context.Context, req *channelListRequest) (channelListResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	query := d.DB.Model(&domain.Channel{})
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return channelListResult{}, fmt.Errorf("查询渠道总数失败: %w", err)
	}

	var channels []domain.Channel
	if err := query.Order("id DESC").Limit(limit).Find(&channels).Error; err != nil {
		return channelListResult{}, fmt.Errorf("查询渠道列表失败: %w", err)
	}

	items := make([]channelItem, 0, len(channels))
	for _, c := range channels {
		items = append(items, channelItemFromDomain(c))
	}

	available, err := d.listAvailableModels()
	if err != nil {
		return channelListResult{}, err
	}
	return channelListResult{Total: total, Channels: items, AvailableModels: available}, nil
}

// maxAvailableModels 可选模型名上限（防御模型表异常膨胀拖垮上下文）
const maxAvailableModels = 200

// listAvailableModels 查询平台已录入的模型名列表（渠道提案的 models 只能从中选择）
func (d OpsToolDeps) listAvailableModels() ([]string, error) {
	var names []string
	if err := d.DB.Model(&domain.Model{}).Order("id ASC").Limit(maxAvailableModels).Pluck("name", &names).Error; err != nil {
		return nil, fmt.Errorf("查询可选模型列表失败: %w", err)
	}
	trimmed := make([]string, 0, len(names))
	for _, n := range names {
		if n = strings.TrimSpace(n); n != "" {
			trimmed = append(trimmed, n)
		}
	}
	return trimmed, nil
}

// previewChannel 校验并规范化渠道提案（只读，不写库）
func (d OpsToolDeps) previewChannel(ctx context.Context, req *channelPreviewRequest) (channelPreviewResult, error) {
	name := strings.TrimSpace(req.Name)
	channelType := strings.TrimSpace(req.ChannelType)
	endpoint := strings.TrimSpace(req.Endpoint)

	// 字段校验（与前端表单规则一致）
	if n := len([]rune(name)); n < 2 || n > 50 {
		return channelPreviewResult{Status: "invalid", Message: fmt.Sprintf("渠道名称长度需在 2-50 个字之间，当前 %d 个字", n)}, nil
	}
	if len(req.Models) == 0 {
		return channelPreviewResult{Status: "invalid", Message: "至少需要配置 1 个支持的模型"}, nil
	}
	req.Models = normalizeModels(req.Models)

	// 渠道类型：创建时必须是管理页可选类型；优化时保持原值透传（避免误改存量渠道的类型），
	// 非已知类型仅附加警告
	warn := ""
	if req.Id > 0 {
		if channelType == "" {
			return channelPreviewResult{Status: "invalid", Message: "渠道类型不能为空"}, nil
		}
		if !knownChannelType(channelType) {
			warn = fmt.Sprintf("注意：渠道类型 %q 不在管理页可选类型中，请确认是否需要调整", channelType)
		}
	} else if !knownChannelType(channelType) {
		return channelPreviewResult{Status: "invalid", Message: fmt.Sprintf("渠道类型必须是 %v 之一", AdminChannelTypes)}, nil
	}

	// 服务地址校验（与网站提案同规则）
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || !isValidHost(parsed) || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return channelPreviewResult{Status: "invalid", Message: "服务地址不合法，必须是完整的 http(s) 地址"}, nil
	}
	normalizedEndpoint := parsed.Scheme + "://" + parsed.Host + strings.TrimSuffix(parsed.Path, "/")

	// 模型映射结构校验
	if msg := validateMappings(req.Mappings); msg != "" {
		return channelPreviewResult{Status: "invalid", Message: msg}, nil
	}

	// 优化提案：先加载原值（同时验证目标存在）
	var original *channelItem
	if req.Id > 0 {
		var existing domain.Channel
		err := d.DB.First(&existing, req.Id).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return channelPreviewResult{Status: "invalid", Message: fmt.Sprintf("要优化的渠道不存在（id=%d）", req.Id)}, nil
			}
			return channelPreviewResult{}, fmt.Errorf("查询渠道失败: %w", err)
		}
		item := channelItemFromDomain(existing)
		original = &item
	}

	// 模型必须在平台模型管理中已录入：防止 LLM 编造模型名导致用户请求路由不到可用渠道。
	// 优化提案时原渠道已引用的模型放行（兼容历史上可能已删除的模型，避免未改动字段被拦下）
	if d.DB != nil {
		var allow []string
		if original != nil {
			allow = original.Models
		}
		unknown, err := d.unknownModels(req.Models, allow)
		if err != nil {
			return channelPreviewResult{}, err
		}
		if len(unknown) > 0 {
			available, _ := d.listAvailableModels()
			return channelPreviewResult{
				Status: "invalid",
				Message: fmt.Sprintf(
					"以下模型未在平台模型管理中录入：%s。渠道的模型列表只能从已录入模型（availableModels）中选择；"+
						"若确实需要新模型，请提示管理员先到模型管理页录入，再回到本页生成渠道提案",
					strings.Join(unknown, "、"),
				),
				AvailableModels: available,
			}, nil
		}
	}

	// 名称查重（只读查询；优化时排除自身）
	if d.DB != nil {
		query := d.DB.Model(&domain.Channel{}).Where("name = ?", name)
		if req.Id > 0 {
			query = query.Where("id <> ?", req.Id)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return channelPreviewResult{}, fmt.Errorf("查询重复渠道失败: %w", err)
		}
		if count > 0 {
			var dup domain.Channel
			_ = d.DB.Where("name = ?", name).First(&dup).Error
			return channelPreviewResult{
				Status:     "duplicate",
				Message:    "已存在同名渠道，请勿重复录入",
				Proposal:   buildChannelProposal(name, channelType, normalizedEndpoint, req),
				Original:   original,
				ExistingID: dup.Id,
			}, nil
		}
	}

	message := "校验通过。注意：提案不含密钥，请提醒管理员在确认卡片中手动填写密钥"
	if warn != "" {
		message = warn + "。" + message
	}
	return channelPreviewResult{
		Status:   "ok",
		Message:  message,
		Proposal: buildChannelProposal(name, channelType, normalizedEndpoint, req),
		Original: original,
	}, nil
}

// buildChannelProposal 构建规范渠道提案（无密钥字段）
func buildChannelProposal(name, channelType, endpoint string, req *channelPreviewRequest) *ChannelProposal {
	return &ChannelProposal{
		Type:        ProposalTypeChannel,
		Id:          req.Id,
		Name:        name,
		ChannelType: channelType,
		Priority:    req.Priority,
		Weight:      req.Weight,
		Retry:       req.Retry,
		Models:      req.Models,
		Endpoint:    endpoint,
		Status:      req.Status,
		Mappings:    req.Mappings,
	}
}

// channelItemFromDomain 领域渠道转列表条目（密钥脱敏为统计信息）
func channelItemFromDomain(c domain.Channel) channelItem {
	return channelItem{
		Id:          c.Id,
		Name:        c.Name,
		ChannelType: c.Type,
		Endpoint:    c.Endpoint,
		Priority:    c.Priority,
		Weight:      c.Weight,
		Retry:       c.Retry,
		Models:      c.Models,
		Status:      c.Status,
		Mappings:    c.Mappings,
		// 脱敏统计：只暴露是否配置与条数，永不暴露密钥内容
		SecretConfigured: strings.TrimSpace(c.Secret) != "",
		SecretKeyCount:   countSecretKeys(c.Secret),
	}
}

// countSecretKeys 统计密钥文本中非空行数（与 Channel.GetRandomSecret 的换行拆分规则一致）
func countSecretKeys(secret string) int {
	count := 0
	for _, line := range strings.Split(secret, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// normalizeModels 去除首尾空白、剔除空项并去重（保序）
func normalizeModels(models []string) []string {
	out := make([]string, 0, len(models))
	seen := make(map[string]bool, len(models))
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out
}

// unknownModels 返回 models 中未在平台模型表录入的名字（去重、保序）；
// allow 中的名字放行（优化提案时用于放行原渠道已引用的模型）
func (d OpsToolDeps) unknownModels(models, allow []string) ([]string, error) {
	var names []string
	if err := d.DB.Model(&domain.Model{}).Limit(maxAvailableModels).Pluck("name", &names).Error; err != nil {
		return nil, fmt.Errorf("查询可选模型列表失败: %w", err)
	}
	known := make(map[string]bool, len(names))
	for _, n := range names {
		known[strings.TrimSpace(n)] = true
	}
	allowed := make(map[string]bool, len(allow))
	for _, n := range allow {
		allowed[n] = true
	}
	var unknown []string
	seen := make(map[string]bool, len(models))
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		if !known[m] && !allowed[m] {
			unknown = append(unknown, m)
		}
	}
	return unknown, nil
}

// knownChannelType 是否为管理页可选的渠道类型
func knownChannelType(t string) bool {
	for _, known := range AdminChannelTypes {
		if known == t {
			return true
		}
	}
	return false
}

// validateMappings 校验模型映射结构，返回错误说明（空串表示通过）
func validateMappings(mappings []global.ModelMapping) string {
	for i, m := range mappings {
		if strings.TrimSpace(m.SourceModel) == "" {
			return fmt.Sprintf("第 %d 条模型映射缺少 sourceModel", i+1)
		}
		if len(m.Conditions) == 0 {
			return fmt.Sprintf("模型映射 %q 至少需要 1 个映射条件", m.SourceModel)
		}
		for j, cond := range m.Conditions {
			if strings.TrimSpace(cond.TargetModel) == "" {
				return fmt.Sprintf("模型映射 %q 的第 %d 个条件缺少 targetModel", m.SourceModel, j+1)
			}
			for key := range cond.Conditions {
				if !channelMappingConditionKeys[key] {
					return fmt.Sprintf("模型映射 %q 的条件 key %q 不支持，仅支持 enableWeb / type", m.SourceModel, key)
				}
			}
			if t, ok := cond.Conditions["type"].(string); ok && t != global.LLMTypeModel && t != global.LLMTypeAPP {
				return fmt.Sprintf("模型映射 %q 的条件 type 必须是 %q 或 %q", m.SourceModel, global.LLMTypeModel, global.LLMTypeAPP)
			}
		}
	}
	return ""
}
