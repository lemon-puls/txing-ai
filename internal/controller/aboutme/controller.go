package aboutme

import (
	"sort"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"
)

// ============================================================================
// Hero 单例（只允许 1 条；不存在的表返回空对象）
// ============================================================================

// GetHero 获取 Hero 配置
// @Summary 获取 Hero 配置
// @Description 获取关于我页面 Hero 区配置
// @Tags 关于我管理
// @Produce json
// @Success 200 {object} utils.Response{data=vo.AboutMeHeroVO}
// @Router /api/admin/about/hero [get]
func GetHero(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var hero domain.AboutMeHero
	if err := db.First(&hero).Error; err != nil {
		// 记录不存在时返回空对象
		utils.OkWithData(ctx, vo.AboutMeHeroVO{})
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeHeroVO(hero))
}

// UpsertHero 创建或更新 Hero 配置
// @Summary 创建或更新 Hero 配置
// @Description 不存在则创建，存在则更新
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.UpdateAboutMeHeroReq true "Hero 配置"
// @Success 200 {object} utils.Response{data=vo.AboutMeHeroVO}
// @Router /api/admin/about/hero [put]
func UpsertHero(ctx *gin.Context) {
	var req dto.UpdateAboutMeHeroReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var hero domain.AboutMeHero
	err := db.First(&hero).Error
	if err == nil {
		hero.AvatarText = req.AvatarText
		hero.StatusText = req.StatusText
		hero.Name = req.Name
		hero.Subtitle = req.Subtitle
		if err := db.Save(&hero).Error; err != nil {
			utils.ErrorWithMsg(ctx, "更新 Hero 失败", err)
			return
		}
	} else {
		hero = domain.AboutMeHero{
			AvatarText: req.AvatarText,
			StatusText: req.StatusText,
			Name:       req.Name,
			Subtitle:   req.Subtitle,
		}
		if err := db.Create(&hero).Error; err != nil {
			utils.ErrorWithMsg(ctx, "创建 Hero 失败", err)
			return
		}
	}
	utils.OkWithData(ctx, vo.ToAboutMeHeroVO(hero))
}

// ============================================================================
// Floating Icon
// ============================================================================

// ListFloatingIcons 浮动图标列表
// @Summary 浮动图标列表
// @Tags 关于我管理
// @Produce json
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/floating-icon/list [get]
func ListFloatingIcons(ctx *gin.Context) {
	var req dto.ListAboutMeFloatingIconReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	query := db.Model(&domain.AboutMeFloatingIcon{})
	var items []domain.AboutMeFloatingIcon
	pageVo, err := page.Paginate[domain.AboutMeFloatingIcon](query, req.PageRequest, &items)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取浮动图标失败", err)
		return
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Sort < items[j].Sort })
	convert := page.Convert(pageVo, vo.ToAboutMeFloatingIconVOs(items))
	utils.OkWithData(ctx, convert)
}

// CreateFloatingIcon 新增浮动图标
// @Summary 新增浮动图标
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.CreateAboutMeFloatingIconReq true "浮动图标"
// @Success 200 {object} utils.Response{data=vo.AboutMeFloatingIconVO}
// @Router /api/admin/about/floating-icon [post]
func CreateFloatingIcon(ctx *gin.Context) {
	var req dto.CreateAboutMeFloatingIconReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	item := domain.AboutMeFloatingIcon{
		Name:   req.Name,
		Symbol: req.Symbol,
		Sort:   req.Sort,
	}
	if err := db.Create(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建浮动图标失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeFloatingIconVO(item))
}

// UpdateFloatingIcon 更新浮动图标
// @Summary 更新浮动图标
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param data body dto.UpdateAboutMeFloatingIconReq true "浮动图标"
// @Success 200 {object} utils.Response{data=vo.AboutMeFloatingIconVO}
// @Router /api/admin/about/floating-icon/{id} [put]
func UpdateFloatingIcon(ctx *gin.Context) {
	var req dto.UpdateAboutMeFloatingIconReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var item domain.AboutMeFloatingIcon
	if err := db.First(&item, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "记录不存在", err)
		return
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Symbol != "" {
		item.Symbol = req.Symbol
	}
	item.Sort = req.Sort
	if err := db.Save(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新浮动图标失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeFloatingIconVO(item))
}

// DeleteFloatingIcon 删除浮动图标
// @Summary 删除浮动图标
// @Tags 关于我管理
// @Param id path int true "ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/floating-icon/{id} [delete]
func DeleteFloatingIcon(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := db.Delete(&domain.AboutMeFloatingIcon{}, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除浮动图标失败", err)
		return
	}
	utils.OkWithMsg(ctx, "删除成功")
}

// ============================================================================
// Reason
// ============================================================================

// ListReasons "为什么选择我"卡片列表
// @Summary "为什么选择我"列表
// @Tags 关于我管理
// @Produce json
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/reason/list [get]
func ListReasons(ctx *gin.Context) {
	var req dto.ListAboutMeReasonReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	query := db.Model(&domain.AboutMeReason{})
	var items []domain.AboutMeReason
	pageVo, err := page.Paginate[domain.AboutMeReason](query, req.PageRequest, &items)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取列表失败", err)
		return
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Sort < items[j].Sort })
	convert := page.Convert(pageVo, vo.ToAboutMeReasonVOs(items))
	utils.OkWithData(ctx, convert)
}

// CreateReason @Summary 新增"为什么选择我"卡片
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.CreateAboutMeReasonReq true "卡片内容"
// @Success 200 {object} utils.Response{data=vo.AboutMeReasonVO}
// @Router /api/admin/about/reason [post]
func CreateReason(ctx *gin.Context) {
	var req dto.CreateAboutMeReasonReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	item := domain.AboutMeReason{
		Emoji: req.Emoji,
		Title: req.Title,
		Desc:  req.Desc,
		Tags:  req.Tags,
		Sort:  req.Sort,
	}
	// GORM 不支持直接给嵌套结构体赋值，需要序列化到 JSON 列
	if err := db.Create(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建失败", err)
		return
	}
	// 二阶段更新 Stats（嵌套结构体的 JSON 列）
	if len(req.Stats) > 0 {
		if err := db.Model(&item).Update("stats", req.Stats).Error; err != nil {
			utils.ErrorWithMsg(ctx, "更新 stats 失败", err)
			return
		}
	}
	utils.OkWithData(ctx, vo.ToAboutMeReasonVO(item))
}

// UpdateReason @Summary 更新"为什么选择我"卡片
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param data body dto.UpdateAboutMeReasonReq true "卡片内容"
// @Success 200 {object} utils.Response{data=vo.AboutMeReasonVO}
// @Router /api/admin/about/reason/{id} [put]
func UpdateReason(ctx *gin.Context) {
	var req dto.UpdateAboutMeReasonReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var item domain.AboutMeReason
	if err := db.First(&item, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "记录不存在", err)
		return
	}
	item.Emoji = req.Emoji
	item.Title = req.Title
	item.Desc = req.Desc
	item.Tags = req.Tags
	item.Sort = req.Sort
	if req.Stats != nil {
		item.Stats = req.Stats
	}
	if err := db.Save(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeReasonVO(item))
}

// DeleteReason @Summary 删除"为什么选择我"卡片
// @Tags 关于我管理
// @Param id path int true "ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/reason/{id} [delete]
func DeleteReason(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := db.Delete(&domain.AboutMeReason{}, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除失败", err)
		return
	}
	utils.OkWithMsg(ctx, "删除成功")
}

// ============================================================================
// Skill
// ============================================================================

// ListSkills 核心能力分类列表
// @Summary 核心能力列表
// @Tags 关于我管理
// @Produce json
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/skill/list [get]
func ListSkills(ctx *gin.Context) {
	var req dto.ListAboutMeSkillReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	query := db.Model(&domain.AboutMeSkill{})
	var items []domain.AboutMeSkill
	pageVo, err := page.Paginate[domain.AboutMeSkill](query, req.PageRequest, &items)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取列表失败", err)
		return
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Sort < items[j].Sort })
	convert := page.Convert(pageVo, vo.ToAboutMeSkillVOs(items))
	utils.OkWithData(ctx, convert)
}

// CreateSkill @Summary 新增核心能力
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.CreateAboutMeSkillReq true "技能"
// @Success 200 {object} utils.Response{data=vo.AboutMeSkillVO}
// @Router /api/admin/about/skill [post]
func CreateSkill(ctx *gin.Context) {
	var req dto.CreateAboutMeSkillReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	item := domain.AboutMeSkill{
		Category: req.Category,
		IconKey:  req.IconKey,
		Tags:     req.Tags,
		Level:    req.Level,
		Sort:     req.Sort,
	}
	if err := db.Create(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeSkillVO(item))
}

// UpdateSkill @Summary 更新核心能力
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param data body dto.UpdateAboutMeSkillReq true "技能"
// @Success 200 {object} utils.Response{data=vo.AboutMeSkillVO}
// @Router /api/admin/about/skill/{id} [put]
func UpdateSkill(ctx *gin.Context) {
	var req dto.UpdateAboutMeSkillReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var item domain.AboutMeSkill
	if err := db.First(&item, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "记录不存在", err)
		return
	}
	if req.Category != "" {
		item.Category = req.Category
	}
	if req.IconKey != "" {
		item.IconKey = req.IconKey
	}
	item.Tags = req.Tags
	item.Level = req.Level
	item.Sort = req.Sort
	if err := db.Save(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeSkillVO(item))
}

// DeleteSkill @Summary 删除核心能力
// @Tags 关于我管理
// @Param id path int true "ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/skill/{id} [delete]
func DeleteSkill(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := db.Delete(&domain.AboutMeSkill{}, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除失败", err)
		return
	}
	utils.OkWithMsg(ctx, "删除成功")
}

// ============================================================================
// Project
// ============================================================================

// ListProjects 精选作品列表
// @Summary 精选作品列表
// @Tags 关于我管理
// @Produce json
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/project/list [get]
func ListProjects(ctx *gin.Context) {
	var req dto.ListAboutMeProjectReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	query := db.Model(&domain.AboutMeProject{})
	var items []domain.AboutMeProject
	pageVo, err := page.Paginate[domain.AboutMeProject](query, req.PageRequest, &items)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取列表失败", err)
		return
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Sort < items[j].Sort })
	convert := page.Convert(pageVo, vo.ToAboutMeProjectVOs(items))
	utils.OkWithData(ctx, convert)
}

// GetProject @Summary 获取单个项目
// @Tags 关于我管理
// @Param id path int true "ID"
// @Success 200 {object} utils.Response{data=vo.AboutMeProjectVO}
// @Router /api/admin/about/project/{id} [get]
func GetProject(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var item domain.AboutMeProject
	if err := db.First(&item, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "项目不存在", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeProjectVO(item))
}

// CreateProject @Summary 新增项目
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.CreateAboutMeProjectReq true "项目"
// @Success 200 {object} utils.Response{data=vo.AboutMeProjectVO}
// @Router /api/admin/about/project [post]
func CreateProject(ctx *gin.Context) {
	var req dto.CreateAboutMeProjectReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if req.Gradient == "" {
		req.Gradient = "1"
	}
	item := domain.AboutMeProject{
		Name:       req.Name,
		Desc:       req.Desc,
		IconKey:    req.IconKey,
		Gradient:   req.Gradient,
		Tags:       req.Tags,
		Link:       req.Link,
		Badge:      req.Badge,
		Highlights: req.Highlights,
		Sort:       req.Sort,
	}
	if err := db.Create(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建项目失败", err)
		return
	}
	// 更新嵌套 JSON 列
	updates := map[string]interface{}{}
	if req.Media != nil {
		updates["media"] = req.Media
	}
	if req.TechStack != nil {
		updates["tech_stack"] = req.TechStack
	}
	if req.Features != nil {
		updates["features"] = req.Features
	}
	if len(updates) > 0 {
		if err := db.Model(&item).Updates(updates).Error; err != nil {
			utils.ErrorWithMsg(ctx, "更新嵌套字段失败", err)
			return
		}
	}
	// 重新查询以包含所有字段
	db.First(&item, item.Id)
	utils.OkWithData(ctx, vo.ToAboutMeProjectVO(item))
}

// UpdateProject @Summary 更新项目
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param data body dto.UpdateAboutMeProjectReq true "项目"
// @Success 200 {object} utils.Response{data=vo.AboutMeProjectVO}
// @Router /api/admin/about/project/{id} [put]
func UpdateProject(ctx *gin.Context) {
	var req dto.UpdateAboutMeProjectReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var item domain.AboutMeProject
	if err := db.First(&item, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "项目不存在", err)
		return
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Desc != "" {
		item.Desc = req.Desc
	}
	if req.IconKey != "" {
		item.IconKey = req.IconKey
	}
	if req.Gradient != "" {
		item.Gradient = req.Gradient
	}
	if req.Tags != nil {
		item.Tags = req.Tags
	}
	item.Link = req.Link
	item.Badge = req.Badge
	if req.Highlights != nil {
		item.Highlights = req.Highlights
	}
	if req.Media != nil {
		item.Media = req.Media
	}
	if req.TechStack != nil {
		item.TechStack = req.TechStack
	}
	if req.Features != nil {
		item.Features = req.Features
	}
	item.Sort = req.Sort
	if err := db.Save(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新项目失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeProjectVO(item))
}

// DeleteProject @Summary 删除项目
// @Tags 关于我管理
// @Param id path int true "ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/project/{id} [delete]
func DeleteProject(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := db.Delete(&domain.AboutMeProject{}, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除失败", err)
		return
	}
	utils.OkWithMsg(ctx, "删除成功")
}

// ============================================================================
// Timeline
// ============================================================================

// ListTimelines 成长轨迹列表
// @Summary 成长轨迹列表
// @Tags 关于我管理
// @Produce json
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/timeline/list [get]
func ListTimelines(ctx *gin.Context) {
	var req dto.ListAboutMeTimelineReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	query := db.Model(&domain.AboutMeTimeline{})
	var items []domain.AboutMeTimeline
	pageVo, err := page.Paginate[domain.AboutMeTimeline](query, req.PageRequest, &items)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取列表失败", err)
		return
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Sort < items[j].Sort })
	convert := page.Convert(pageVo, vo.ToAboutMeTimelineVOs(items))
	utils.OkWithData(ctx, convert)
}

// CreateTimeline @Summary 新增时间线
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.CreateAboutMeTimelineReq true "时间线"
// @Success 200 {object} utils.Response{data=vo.AboutMeTimelineVO}
// @Router /api/admin/about/timeline [post]
func CreateTimeline(ctx *gin.Context) {
	var req dto.CreateAboutMeTimelineReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	item := domain.AboutMeTimeline{
		Time:  req.Time,
		Title: req.Title,
		Desc:  req.Desc,
		Tags:  req.Tags,
		Sort:  req.Sort,
	}
	if err := db.Create(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeTimelineVO(item))
}

// UpdateTimeline @Summary 更新时间线
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param data body dto.UpdateAboutMeTimelineReq true "时间线"
// @Success 200 {object} utils.Response{data=vo.AboutMeTimelineVO}
// @Router /api/admin/about/timeline/{id} [put]
func UpdateTimeline(ctx *gin.Context) {
	var req dto.UpdateAboutMeTimelineReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var item domain.AboutMeTimeline
	if err := db.First(&item, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "记录不存在", err)
		return
	}
	if req.Time != "" {
		item.Time = req.Time
	}
	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Desc != "" {
		item.Desc = req.Desc
	}
	if req.Tags != nil {
		item.Tags = req.Tags
	}
	item.Sort = req.Sort
	if err := db.Save(&item).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新失败", err)
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeTimelineVO(item))
}

// DeleteTimeline @Summary 删除时间线
// @Tags 关于我管理
// @Param id path int true "ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/about/timeline/{id} [delete]
func DeleteTimeline(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := db.Delete(&domain.AboutMeTimeline{}, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除失败", err)
		return
	}
	utils.OkWithMsg(ctx, "删除成功")
}

// ============================================================================
// Contact
// ============================================================================

// GetContact 获取联系区配置
// @Summary 获取联系区配置
// @Tags 关于我管理
// @Produce json
// @Success 200 {object} utils.Response{data=vo.AboutMeContactVO}
// @Router /api/admin/about/contact [get]
func GetContact(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var c domain.AboutMeContact
	if err := db.First(&c).Error; err != nil {
		utils.OkWithData(ctx, vo.AboutMeContactVO{})
		return
	}
	utils.OkWithData(ctx, vo.ToAboutMeContactVO(c))
}

// UpsertContact 创建或更新联系区配置
// @Summary 创建或更新联系区配置
// @Tags 关于我管理
// @Accept json
// @Produce json
// @Param data body dto.UpdateAboutMeContactReq true "联系区配置"
// @Success 200 {object} utils.Response{data=vo.AboutMeContactVO}
// @Router /api/admin/about/contact [put]
func UpsertContact(ctx *gin.Context) {
	var req dto.UpdateAboutMeContactReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var c domain.AboutMeContact
	err := db.First(&c).Error
	if err == nil {
		c.Title = req.Title
		c.Desc = req.Desc
		c.Links = req.Links
		if err := db.Save(&c).Error; err != nil {
			utils.ErrorWithMsg(ctx, "更新联系区失败", err)
			return
		}
	} else {
		c = domain.AboutMeContact{
			Title: req.Title,
			Desc:  req.Desc,
			Links: req.Links,
		}
		if err := db.Create(&c).Error; err != nil {
			utils.ErrorWithMsg(ctx, "创建联系区失败", err)
			return
		}
	}
	utils.OkWithData(ctx, vo.ToAboutMeContactVO(c))
}

// ============================================================================
// 公开接口：聚合快照
// ============================================================================

// PublicSnapshot 公开的关于我页面聚合快照
// @Summary 关于我页面聚合数据（公开）
// @Description 返回 Hero/Reasons/Skills/Projects/Timeline/Contact 完整数据，无需鉴权
// @Tags 关于我
// @Produce json
// @Success 200 {object} utils.Response{data=vo.AboutMeSnapshotVO}
// @Router /api/about [get]
func PublicSnapshot(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)

	var hero domain.AboutMeHero
	_ = db.First(&hero).Error

	var floatingIcons []domain.AboutMeFloatingIcon
	db.Order("sort asc, id asc").Find(&floatingIcons)

	var reasons []domain.AboutMeReason
	db.Order("sort asc, id asc").Find(&reasons)

	var skills []domain.AboutMeSkill
	db.Order("sort asc, id asc").Find(&skills)

	var projects []domain.AboutMeProject
	db.Order("sort asc, id asc").Find(&projects)

	var timelines []domain.AboutMeTimeline
	db.Order("sort asc, id asc").Find(&timelines)

	var contact domain.AboutMeContact
	_ = db.First(&contact).Error

	snapshot := vo.AboutMeSnapshotVO{
		Hero:         vo.ToAboutMeHeroVO(hero),
		FloatingIcon: vo.ToAboutMeFloatingIconVOs(floatingIcons),
		Reasons:      vo.ToAboutMeReasonVOs(reasons),
		Skills:       vo.ToAboutMeSkillVOs(skills),
		Projects:     vo.ToAboutMeProjectVOs(projects),
		Timeline:     vo.ToAboutMeTimelineVOs(timelines),
		Contact:      vo.ToAboutMeContactVO(contact),
	}
	utils.OkWithData(ctx, snapshot)
}