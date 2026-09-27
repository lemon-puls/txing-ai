package ops

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"txing-ai/internal/domain"
)

// modelListRequest model_list_tool 请求参数
type modelListRequest struct {
	// 按模型名称模糊检索的关键字，可为空
	Keyword string `json:"keyword,omitempty"`
	// 返回条数上限，缺省 20，最大 50
	Limit int `json:"limit,omitempty"`
}

// modelItem 模型列表条目（字段名与 domain.Model 的 JSON 标签对齐）
type modelItem struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Default     bool   `json:"default"`
	HighContext bool   `json:"high_context"`
	Multimodal  bool   `json:"multimodal"`
	Avatar      string `json:"avatar,omitempty"`
	Tag         string `json:"tag,omitempty"`
}

// modelListResult model_list_tool 返回结果
type modelListResult struct {
	Total  int64       `json:"total"`
	Models []modelItem `json:"models"`
}

// modelPreviewRequest model_preview_tool 请求参数（LLM 拟提交的模型资料）。
// Id > 0 表示对已有模型的优化提案，其余字段为优化后的完整目标值。
type modelPreviewRequest struct {
	Id          int64    `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Default     bool     `json:"default"`
	HighContext bool     `json:"high_context"`
	Multimodal  bool     `json:"multimodal"`
	Avatar      string   `json:"avatar,omitempty"`
	Tags        []string `json:"tags"`
}

// modelPreviewResult model_preview_tool 返回结果（信封与 websitePreviewResult 同形：
// controller 按统一信封拦截并下发 proposal 帧）
type modelPreviewResult struct {
	// "ok" | "duplicate" | "invalid"
	Status string `json:"status"`
	// 人类可读的校验说明（中文）
	Message string `json:"message,omitempty"`
	// 校验通过后的规范提案（invalid 时为 nil；duplicate 时附带便于用户决策）
	Proposal *ModelProposal `json:"proposal,omitempty"`
	// 优化提案时的原值（Id>0 时返回，供 LLM 说明改动与管理员对照）
	Original *modelItem `json:"original,omitempty"`
	// 重复记录的 ID（duplicate 时返回）
	ExistingID int64 `json:"existingId,omitempty"`
}

// listModels 查询模型列表（只读），供 LLM 了解已有模型、避免重复录入
func (d OpsToolDeps) listModels(ctx context.Context, req *modelListRequest) (modelListResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	query := d.DB.Model(&domain.Model{})
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return modelListResult{}, fmt.Errorf("查询模型总数失败: %w", err)
	}

	var models []domain.Model
	if err := query.Order("id DESC").Limit(limit).Find(&models).Error; err != nil {
		return modelListResult{}, fmt.Errorf("查询模型列表失败: %w", err)
	}

	items := make([]modelItem, 0, len(models))
	for _, m := range models {
		items = append(items, modelItemFromDomain(m))
	}
	return modelListResult{Total: total, Models: items}, nil
}

// previewModel 校验并规范化模型提案（只读，不写库）
func (d OpsToolDeps) previewModel(ctx context.Context, req *modelPreviewRequest) (modelPreviewResult, error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)

	// 字段长度校验（与前端表单规则一致）
	if n := len([]rune(name)); n < 2 || n > 50 {
		return modelPreviewResult{Status: "invalid", Message: fmt.Sprintf("模型名称长度需在 2-50 个字之间，当前 %d 个字", n)}, nil
	}
	if n := len([]rune(description)); n > 500 {
		return modelPreviewResult{Status: "invalid", Message: fmt.Sprintf("模型描述不能超过 500 个字，当前 %d 个字", n)}, nil
	}
	tags := reorderTagsFirst(normalizeTags(req.Tags), PresetModelTags)

	// 优化提案：先加载原值（同时验证目标存在）
	var original *modelItem
	if req.Id > 0 {
		var existing domain.Model
		err := d.DB.First(&existing, req.Id).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return modelPreviewResult{Status: "invalid", Message: fmt.Sprintf("要优化的模型不存在（id=%d）", req.Id)}, nil
			}
			return modelPreviewResult{}, fmt.Errorf("查询模型失败: %w", err)
		}
		item := modelItemFromDomain(existing)
		original = &item
	}

	// 名称查重（只读查询；优化时排除自身）
	if d.DB != nil {
		query := d.DB.Model(&domain.Model{}).Where("name = ?", name)
		if req.Id > 0 {
			query = query.Where("id <> ?", req.Id)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return modelPreviewResult{}, fmt.Errorf("查询重复模型失败: %w", err)
		}
		if count > 0 {
			var dup domain.Model
			_ = d.DB.Where("name = ?", name).First(&dup).Error
			return modelPreviewResult{
				Status:     "duplicate",
				Message:    "已存在同名模型，请勿重复录入",
				Proposal:   buildModelProposal(name, description, req, tags),
				Original:   original,
				ExistingID: dup.Id,
			}, nil
		}
	}

	return modelPreviewResult{
		Status:   "ok",
		Message:  "校验通过",
		Proposal: buildModelProposal(name, description, req, tags),
		Original: original,
	}, nil
}

// buildModelProposal 构建规范模型提案
func buildModelProposal(name, description string, req *modelPreviewRequest, tags []string) *ModelProposal {
	return &ModelProposal{
		Type:        ProposalTypeModel,
		Id:          req.Id,
		Name:        name,
		Description: description,
		Default:     req.Default,
		HighContext: req.HighContext,
		Multimodal:  req.Multimodal,
		Avatar:      strings.TrimSpace(req.Avatar),
		Tags:        strings.Join(tags, ","),
	}
}

// modelItemFromDomain 领域模型转列表条目
func modelItemFromDomain(m domain.Model) modelItem {
	return modelItem{
		Id:          m.Id,
		Name:        m.Name,
		Description: m.Description,
		Default:     m.Default,
		HighContext: m.HighContext,
		Multimodal:  m.Multimodal,
		Avatar:      m.Avatar,
		Tag:         m.Tag,
	}
}
