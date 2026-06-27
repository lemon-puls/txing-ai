package workflowservice

import (
	"errors"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils/page"

	"gorm.io/gorm"
)

// CreateVersion 创建工作流版本
func CreateVersion(flowID int64, req dto.CreateVersionReq, db *gorm.DB) (*domain.AgentFlowVersion, error) {
	// 获取当前工作流
	var flow domain.AgentFlow
	if err := db.First(&flow, flowID).Error; err != nil {
		return nil, errors.New("工作流不存在")
	}

	// 获取当前最大版本号
	var maxVersion int
	db.Model(&domain.AgentFlowVersion{}).Where("flow_id = ?", flowID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

	// 创建新版本
	version := &domain.AgentFlowVersion{
		FlowID:      flowID,
		Version:     maxVersion + 1,
		Name:        req.Name,
		Description: req.Description,
		Topology:    flow.Topology,
		ChangeLog:   req.ChangeLog,
	}

	if err := db.Create(version).Error; err != nil {
		return nil, err
	}

	// 更新工作流的当前版本号
	db.Model(&domain.AgentFlow{}).Where("id = ?", flowID).Update("current_version", version.Version)

	return version, nil
}

// GetVersion 获取指定版本
func GetVersion(flowID int64, version int, db *gorm.DB) (*domain.AgentFlowVersion, error) {
	var v domain.AgentFlowVersion
	err := db.Where("flow_id = ? AND version = ?", flowID, version).First(&v).Error
	return &v, err
}

// ListVersions 获取版本列表
func ListVersions(req dto.ListVersionReq, db *gorm.DB) (*page.PageVo[domain.AgentFlowVersion], error) {
	var versions []domain.AgentFlowVersion
	query := db.Model(&domain.AgentFlowVersion{}).Where("flow_id = ?", req.FlowID).Order("version DESC")

	pageVo, err := page.Paginate[domain.AgentFlowVersion](query, req.PageRequest, &versions)
	return pageVo, err
}

// PublishVersion 发布指定版本
func PublishVersion(flowID int64, version int, db *gorm.DB) error {
	// 检查版本是否存在
	var v domain.AgentFlowVersion
	if err := db.Where("flow_id = ? AND version = ?", flowID, version).First(&v).Error; err != nil {
		return errors.New("版本不存在")
	}

	// 将该版本标记为已发布
	if err := db.Model(&v).Update("is_published", true).Error; err != nil {
		return err
	}

	// 更新工作流的已发布版本号
	if err := db.Model(&domain.AgentFlow{}).Where("id = ?", flowID).Update("published_version", version).Error; err != nil {
		return err
	}

	return nil
}

// RollbackToVersion 回滚到指定版本
func RollbackToVersion(flowID int64, version int, db *gorm.DB) error {
	// 获取指定版本
	var v domain.AgentFlowVersion
	if err := db.Where("flow_id = ? AND version = ?", flowID, version).First(&v).Error; err != nil {
		return errors.New("版本不存在")
	}

	// 更新工作流的拓扑数据
	updates := map[string]interface{}{
		"topology":        v.Topology,
		"current_version": version,
	}

	return db.Model(&domain.AgentFlow{}).Where("id = ?", flowID).Updates(updates).Error
}
