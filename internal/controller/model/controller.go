package model

import (
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/samber/lo"

	"github.com/gin-gonic/gin"
)

// Create 创建模型
// @Summary 创建模型
// @Description 创建新的模型
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param data body dto.CreateModelReq true "模型信息"
// @Success 200 {object} utils.Response{data=vo.ModelVO}
// @Router /api/admin/model [post]
func Create(ctx *gin.Context) {
	var req dto.CreateModelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)
	cosClient := utils.GetCosClientFromContext(ctx)

	model := &domain.Model{
		Name:        req.Name,
		Description: req.Description,
		Default:     req.Default,
		HighContext: req.HighContext,
		Avatar:      cosClient.ConvertObjectPath(req.Avatar),
		Tag:         req.Tag,
	}

	if err := db.Create(model).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建模型失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToModelVO(*model))
}

// Update 更新模型
// @Summary 更新模型
// @Description 更新模型信息
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Param data body dto.UpdateModelReq true "模型信息"
// @Success 200 {object} utils.Response{data=vo.ModelVO}
// @Router /api/admin/model/{id} [put]
func Update(ctx *gin.Context) {
	var req dto.UpdateModelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)
	cosClient := utils.GetCosClientFromContext(ctx)

	var model domain.Model
	if err := db.First(&model, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "模型不存在", err)
		return
	}

	if req.Name != "" {
		model.Name = req.Name
	}
	if req.Description != "" {
		model.Description = req.Description
	}
	if req.Default != nil {
		model.Default = *req.Default
	}
	if req.HighContext != nil {
		model.HighContext = *req.HighContext
	}
	if req.Avatar != "" {
		model.Avatar = cosClient.ConvertObjectPath(req.Avatar)
	}
	if req.Tag != "" {
		model.Tag = req.Tag
	}

	if err := db.Save(&model).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新模型失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToModelVO(model))
}

// Delete 删除模型
// @Summary 删除模型
// @Description 删除指定模型
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/model/{id} [delete]
func Delete(ctx *gin.Context) {
	var model domain.Model

	db := utils.GetDBFromContext(ctx)

	if err := db.First(&model, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "模型不存在", err)
		return
	}

	if err := db.Delete(&model).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除模型失败", err)
		return
	}

	utils.OkWithMsg(ctx, "删除成功")
}

// Get 获取模型详情
// @Summary 获取模型详情
// @Description 获取指定模型的详细信息
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param id path int true "模型ID"
// @Success 200 {object} utils.Response{data=vo.ModelVO}
// @Router /api/model/{id} [get]
func Get(ctx *gin.Context) {
	var model domain.Model

	db := utils.GetDBFromContext(ctx)
	if err := db.First(&model, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "模型不存在", err)
		return
	}

	if model.Avatar != "" {
		cosClient := utils.GetCosClientFromContext(ctx)
		model.Avatar, _ = cosClient.GenerateDownloadPresignedURL(model.Avatar)
	}

	utils.OkWithData(ctx, vo.ToModelVO(model))
}

// List 获取模型列表
// @Summary 获取模型列表
// @Description 获取模型列表，支持分页
// @Tags 模型管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式(asc/desc)"
// @Param tag query string false "标签"
// @Param default query bool false "是否默认"
// @Param name query string false "模型名称"
// @Success 200 {object} utils.Response
// @Router /api/model/list [get]
func List(ctx *gin.Context) {
	var req dto.ListModelReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)
	cosClient := utils.GetCosClientFromContext(ctx)

	// 构建查询条件
	query := db.Model(&domain.Model{})
	if req.Tag != "" {
		query = query.Where("tag LIKE ?", "%"+req.Tag+"%")
	}
	if req.Default != nil {
		query = query.Where("default = ?", *req.Default)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	var models []domain.Model

	// 获取分页数据
	pageVo, err := page.Paginate[domain.Model](query, req.PageRequest, &models)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取模型列表失败", err)
		return
	}

	// 转换为 VO
	modelVOs := vo.ToModelVOs(models)
	modelVOs = lo.Map(modelVOs, func(modelVO vo.ModelVO, _ int) vo.ModelVO {
		if modelVO.Avatar != "" {
			modelVO.Avatar, _ = cosClient.GenerateDownloadPresignedURL(modelVO.Avatar)
		}
		return modelVO
	})

	convert := page.Convert(pageVo, modelVOs)

	utils.OkWithData(ctx, convert)
}
