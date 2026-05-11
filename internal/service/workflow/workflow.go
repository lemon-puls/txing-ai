package workflowservice

import (
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils/page"

	"gorm.io/gorm"
)

func Create(req dto.CreateAgentFlowReq, db *gorm.DB) error {
	flow := &domain.AgentFlow{
		Name:        req.Name,
		Description: req.Description,
		Topology:    req.Topology,
	}
	return db.Create(flow).Error
}

func Update(id int64, req dto.UpdateAgentFlowReq, db *gorm.DB) error {
	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"topology":    req.Topology,
	}
	return db.Model(&domain.AgentFlow{}).Where("id = ?", id).Updates(updates).Error
}

func Delete(id int64, db *gorm.DB) error {
	return db.Delete(&domain.AgentFlow{}, id).Error
}

func Get(id int64, db *gorm.DB) (*domain.AgentFlow, error) {
	var flow domain.AgentFlow
	err := db.First(&flow, id).Error
	return &flow, err
}

func List(req dto.ListAgentFlowReq, db *gorm.DB) (*page.PageVo[domain.AgentFlow], error) {
	var flows []domain.AgentFlow
	query := db.Model(&domain.AgentFlow{})

	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	pageVo, err := page.Paginate[domain.AgentFlow](query, req.PageRequest, &flows)
	return pageVo, err
}
