package workflowservice

import (
	"errors"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils/page"

	"gorm.io/gorm"
)

// CreateTemplate 从工作流创建模板
func CreateTemplate(req dto.CreateTemplateReq, db *gorm.DB) (*domain.AgentFlow, error) {
	// 获取原工作流
	var flow domain.AgentFlow
	if err := db.First(&flow, req.FlowID).Error; err != nil {
		return nil, errors.New("工作流不存在")
	}

	// 创建模板（复制工作流）
	template := &domain.AgentFlow{
		Name:             req.Name,
		Description:      req.Description,
		Topology:         flow.Topology,
		IsTemplate:        true,
		TemplateCategory: req.Category,
	}

	if err := db.Create(template).Error; err != nil {
		return nil, err
	}

	return template, nil
}

// ListTemplates 获取模板列表
func ListTemplates(req dto.ListTemplateReq, db *gorm.DB) (*page.PageVo[domain.AgentFlow], error) {
	var templates []domain.AgentFlow
	query := db.Model(&domain.AgentFlow{}).Where("is_template = ?", true)

	if req.Category != "" {
		query = query.Where("template_category = ?", req.Category)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	pageVo, err := page.Paginate[domain.AgentFlow](query, req.PageRequest, &templates)
	return pageVo, err
}

// GetTemplate 获取模板详情
func GetTemplate(id int64, db *gorm.DB) (*domain.AgentFlow, error) {
	var template domain.AgentFlow
	err := db.Where("id = ? AND is_template = ?", id, true).First(&template).Error
	return &template, err
}

// CloneTemplate 克隆模板为新工作流
func CloneTemplate(req dto.CloneTemplateReq, db *gorm.DB) (*domain.AgentFlow, error) {
	// 获取模板
	var template domain.AgentFlow
	if err := db.Where("id = ? AND is_template = ?", req.TemplateID, true).First(&template).Error; err != nil {
		return nil, errors.New("模板不存在")
	}

	// 设置新名称
	name := req.Name
	if name == "" {
		name = template.Name + " (副本)"
	}

	// 创建新工作流
	newFlow := &domain.AgentFlow{
		Name:        name,
		Description: template.Description,
		Topology:    template.Topology,
	}

	if err := db.Create(newFlow).Error; err != nil {
		return nil, err
	}

	return newFlow, nil
}

// DeleteTemplate 删除模板
func DeleteTemplate(id int64, db *gorm.DB) error {
	return db.Model(&domain.AgentFlow{}).Where("id = ? AND is_template = ?", id, true).Update("is_template", false).Error
}
