package channel

import (
	"gorm.io/gorm"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
)

// Create 创建渠道
// @Summary 创建渠道
// @Description 创建新的渠道
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Param data body dto.CreateChannelReq true "渠道信息"
// @Success 200 {object} utils.Response{data=vo.ChannelVO}
// @Router /api/admin/channel [post]
func Create(ctx *gin.Context) {
	var req dto.CreateChannelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)

	channel := &domain.Channel{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Weight:   req.Weight,
		Models:   req.Models,
		Retry:    req.Retry,
		Secret:   req.Secret,
		Endpoint: req.Endpoint,
		Status:   req.Status,
		Mappings: req.Mappings,
	}

	if err := db.Create(channel).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建渠道失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToChannelVO(*channel))
}

// Update 更新渠道
// @Summary 更新渠道
// @Description 更新渠道信息
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Param id path int true "渠道ID"
// @Param data body dto.UpdateChannelReq true "渠道信息"
// @Success 200 {object} utils.Response{data=vo.ChannelVO}
// @Router /api/admin/channel/{id} [put]
func Update(ctx *gin.Context) {
	var req dto.UpdateChannelReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)

	var channel domain.Channel
	if err := db.First(&channel, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "渠道不存在", err)
		return
	}

	channel.Name = req.Name
	channel.Type = req.Type
	channel.Priority = req.Priority
	channel.Weight = req.Weight
	channel.Models = req.Models
	channel.Retry = req.Retry
	channel.Secret = req.Secret
	channel.Endpoint = req.Endpoint
	channel.Status = req.Status
	// 转换 Mappings
	channel.Mappings = req.Mappings

	if err := db.Save(&channel).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新渠道失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToChannelVO(channel))
}

// Delete 删除渠道
// @Summary 删除渠道
// @Description 删除指定渠道
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Param id path int true "渠道ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/channel/{id} [delete]
func Delete(ctx *gin.Context) {
	var channel domain.Channel

	db := utils.GetDBFromContext[*gorm.DB](ctx)

	if err := db.First(&channel, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "渠道不存在", err)
		return
	}

	if err := db.Delete(&channel).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除渠道失败", err)
		return
	}

	utils.OkWithMsg(ctx, "删除成功")
}

// Get 获取渠道详情
// @Summary 获取渠道详情
// @Description 获取指定渠道的详细信息
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Param id path int true "渠道ID"
// @Success 200 {object} utils.Response{data=vo.ChannelVO}
// @Router /api/admin/channel/{id} [get]
func Get(ctx *gin.Context) {
	var channel domain.Channel

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := db.First(&channel, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "渠道不存在", err)
		return
	}

	utils.OkWithData(ctx, vo.ToChannelVO(channel))
}

// List 获取渠道列表
// @Summary 获取渠道列表
// @Description 获取渠道列表，支持分页
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式(asc/desc)"
// @Param type query string false "渠道类型"
// @Param status query bool false "状态"
// @Param name query string false "渠道名称"
// @Success 200 {object} utils.Response
// @Router /api/admin/channel/list [get]
func List(ctx *gin.Context) {
	var req dto.ListChannelReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)

	// 构建查询条件
	query := db.Model(&domain.Channel{})
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.Name != "" {
		query = query.Where("name like ?", "%"+req.Name+"%")
	}

	var channels []domain.Channel

	// 获取分页数据
	pageVo, err := page.Paginate[domain.Channel](query, req.PageRequest, &channels)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取渠道列表失败", err)
		return
	}

	// 转换为 VO
	channelVOs := vo.ToChannelVOs(channels)
	convert := page.Convert(pageVo, channelVOs)

	utils.OkWithData(ctx, convert)
}
