package presetservice

import (
	"txing-ai/internal/domain"
	"txing-ai/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPresetByID 根据ID获取预设信息
func GetPresetByID(db *gorm.DB, cosClient *utils.COSClient, id int64) (*domain.Preset, error) {
	preset := domain.Preset{}
	if err := db.First(&preset, id).Error; err != nil {
		return nil, err
	}

	preset.Avatar, _ = cosClient.GenerateDownloadPresignedURL(preset.Avatar)

	return &preset, nil
}

// GetPresetsByIds 批量获取预设信息
func GetPresetsByIds(ctx *gin.Context, ids []int64) ([]*domain.Preset, error) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	var presets []*domain.Preset
	if err := db.Where("id IN ?", ids).Find(&presets).Error; err != nil {
		return nil, err
	}

	// 处理头像URL
	for _, preset := range presets {
		preset.Avatar, _ = cosClient.GenerateDownloadPresignedURL(preset.Avatar)
	}

	return presets, nil
}
