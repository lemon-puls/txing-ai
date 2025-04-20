package preset

import (
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
)

// Create 创建预设
// @Summary 创建预设
// @Description 创建新的预设
// @Tags 预设管理
// @Accept json
// @Produce json
// @Param data body dto.CreatePresetReq true "预设信息"
// @Success 200 {object} utils.Response{data=vo.PresetVO}
// @Router /api/admin/preset [post]
func Create(ctx *gin.Context) {
	var req dto.CreatePresetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	preset := &domain.Preset{
		UserID:      req.UserID,
		Avatar:      req.Avatar,
		Name:        req.Name,
		Description: req.Description,
		Context:     req.Context,
		Official:    req.Official,
	}

	if err := db.Create(preset).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建预设失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToPresetVO(*preset))
}

// Update 更新预设
// @Summary 更新预设
// @Description 更新预设信息
// @Tags 预设管理
// @Accept json
// @Produce json
// @Param id path int true "预设ID"
// @Param data body dto.UpdatePresetReq true "预设信息"
// @Success 200 {object} utils.Response{data=vo.PresetVO}
// @Router /api/admin/preset/{id} [put]
func Update(ctx *gin.Context) {
	var req dto.UpdatePresetReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	var preset domain.Preset
	if err := db.First(&preset, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "预设不存在", err)
		return
	}

	preset.UserID = req.UserID
	preset.Avatar = req.Avatar
	preset.Name = req.Name
	preset.Description = req.Description
	preset.Context = req.Context
	preset.Official = req.Official

	if err := db.Save(&preset).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新预设失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToPresetVO(preset))
}

// Delete 删除预设
// @Summary 删除预设
// @Description 删除指定预设
// @Tags 预设管理
// @Accept json
// @Produce json
// @Param id path int true "预设ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/preset/{id} [delete]
func Delete(ctx *gin.Context) {
	var preset domain.Preset

	db := utils.GetDBFromContext(ctx)

	if err := db.First(&preset, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "预设不存在", err)
		return
	}

	if err := db.Delete(&preset).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除预设失败", err)
		return
	}

	utils.OkWithMsg(ctx, "删除成功")
}

// Get 获取预设详情
// @Summary 获取预设详情
// @Description 获取指定预设的详细信息
// @Tags 预设管理
// @Accept json
// @Produce json
// @Param id path int true "预设ID"
// @Success 200 {object} utils.Response{data=vo.PresetVO}
// @Router /api/admin/preset/{id} [get]
func Get(ctx *gin.Context) {
	var preset domain.Preset

	db := utils.GetDBFromContext(ctx)
	if err := db.First(&preset, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "预设不存在", err)
		return
	}

	utils.OkWithData(ctx, vo.ToPresetVO(preset))
}

// List 获取预设列表
// @Summary 获取预设列表
// @Description 获取预设列表，支持分页
// @Tags 预设管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式(asc/desc)"
// @Param official query bool false "是否官方预设"
// @Param user_id query int false "用户ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/preset/list [get]
func List(ctx *gin.Context) {
	var req dto.ListPresetReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	// 构建查询条件
	query := db.Model(&domain.Preset{})
	if req.Official != nil {
		query = query.Where("official = ?", *req.Official)
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	var presets []domain.Preset

	// 获取分页数据
	pageVo, err := page.Paginate[domain.Preset](query, req.PageRequest, &presets)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取预设列表失败", err)
		return
	}

	// 转换为 VO
	presetVOs := vo.ToPresetVOs(presets)
	convert := page.Convert(pageVo, presetVOs)

	utils.OkWithData(ctx, convert)
}
