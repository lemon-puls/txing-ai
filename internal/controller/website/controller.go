package website

import (
	"fmt"
	"gorm.io/gorm"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// Create 创建网站
// @Summary 创建网站
// @Description 创建新的网站
// @Tags 网站管理
// @Accept json
// @Produce json
// @Param data body dto.CreateWebsiteReq true "网站信息"
// @Success 200 {object} utils.Response{data=vo.WebsiteVO}
// @Router /api/admin/websites [post]
func Create(ctx *gin.Context) {
	var req dto.CreateWebsiteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)

	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	// URL 查重（软删感知：GORM 自动排除已删除记录），兜底防止提案重复确认或手动重复新增
	normalizedURL := strings.TrimRight(strings.TrimSpace(req.Url), "/")
	if normalizedURL != "" {
		var count int64
		if err := db.Model(&domain.Website{}).Where("url = ?", normalizedURL).Count(&count).Error; err != nil {
			utils.ErrorWithMsg(ctx, "查询网站失败", err)
			return
		}
		if count > 0 {
			utils.ErrorWithMsg(ctx, "该网站地址已存在，请勿重复录入", nil)
			return
		}
	}

	website := &domain.Website{
		Name:        req.Name,
		Description: req.Description,
		Url:         req.Url,
		Avatar:      cosClient.ConvertObjectPath(req.Avatar),
		Tags:        req.Tags,
		Sort:        req.Sort,
		Status:      req.Status,
	}

	if err := db.Create(website).Error; err != nil {
		utils.ErrorWithMsg(ctx, "创建网站失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToWebsiteVO(*website))
}

// Update 更新网站
// @Summary 更新网站
// @Description 更新网站信息
// @Tags 网站管理
// @Accept json
// @Produce json
// @Param id path int true "网站ID"
// @Param data body dto.UpdateWebsiteReq true "网站信息"
// @Success 200 {object} utils.Response{data=vo.WebsiteVO}
// @Router /api/admin/websites/{id} [put]
func Update(ctx *gin.Context) {
	var req dto.UpdateWebsiteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	var website domain.Website
	if err := db.First(&website, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "网站不存在", err)
		return
	}

	if req.Name != "" {
		website.Name = req.Name
	}
	if req.Description != "" {
		website.Description = req.Description
	}
	if req.Url != "" {
		website.Url = req.Url
	}
	if req.Avatar != "" {
		website.Avatar = cosClient.ConvertObjectPath(req.Avatar)
	}
	if req.Tags != "" {
		website.Tags = req.Tags
	}
	if req.Sort != nil {
		website.Sort = *req.Sort
	}
	if req.Status != nil {
		website.Status = *req.Status
	}

	if err := db.Save(&website).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新网站失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToWebsiteVO(website))
}

// Delete 删除网站
// @Summary 删除网站
// @Description 删除网站
// @Tags 网站管理
// @Accept json
// @Produce json
// @Param id path int true "网站ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/websites/{id} [delete]
func Delete(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)

	var website domain.Website
	if err := db.First(&website, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "网站不存在", err)
		return
	}

	if err := db.Delete(&website).Error; err != nil {
		utils.ErrorWithMsg(ctx, "删除网站失败", err)
		return
	}

	utils.Ok(ctx)
}

// Get 获取网站详情
// @Summary 获取网站详情
// @Description 根据ID获取网站详情
// @Tags 网站管理
// @Accept json
// @Produce json
// @Param id path int true "网站ID"
// @Success 200 {object} utils.Response{data=vo.WebsiteVO}
// @Router /api/admin/websites/{id} [get]
func Get(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	var website domain.Website
	if err := db.First(&website, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "网站不存在", err)
		return
	}

	websiteVO := vo.ToWebsiteVO(website)
	if websiteVO.Avatar != "" {
		websiteVO.Avatar, _ = cosClient.GenerateDownloadPresignedURL(websiteVO.Avatar)
	}

	utils.OkWithData(ctx, websiteVO)
}

// List 获取网站列表
// @Summary 获取网站列表
// @Description 获取网站列表，支持分页和搜索
// @Tags 网站管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式(asc/desc)"
// @Param name query string false "网站名称"
// @Param tag query string false "标签"
// @Param status query int false "状态"
// @Success 200 {object} utils.Response
// @Router /api/admin/websites/list [get]
func List(ctx *gin.Context) {
	var req dto.ListWebsiteReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	// 构建查询条件
	query := db.Model(&domain.Website{})
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Tag != "" {
		query = query.Where("tags LIKE ?", "%"+req.Tag+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var websites []domain.Website

	// 获取分页数据
	pageVo, err := page.Paginate[domain.Website](query, req.PageRequest, &websites)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取网站列表失败", err)
		return
	}

	// 转换为 VO
	websiteVOs := vo.ToWebsiteVOs(websites)
	websiteVOs = lo.Map(websiteVOs, func(websiteVO vo.WebsiteVO, _ int) vo.WebsiteVO {
		if websiteVO.Avatar != "" {
			websiteVO.Avatar, _ = cosClient.GenerateDownloadPresignedURL(websiteVO.Avatar)
		}
		return websiteVO
	})

	convert := page.Convert(pageVo, websiteVOs)

	utils.OkWithData(ctx, convert)
}

// UserList 用户端获取网站列表
// @Summary 用户端获取网站列表
// @Description 用户端获取启用状态的网站列表，支持搜索和过滤
// @Tags 网站
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param limit query int false "每页数量" minimum(1) default(20)
// @Param name query string false "网站名称"
// @Param tag query string false "标签"
// @Success 200 {object} utils.Response
// @Router /api/websites/list [get]
func UserList(ctx *gin.Context) {
	var req dto.ListWebsiteReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	// 构建查询条件，只查询启用状态的网站
	query := db.Model(&domain.Website{}).Where("status = ?", 1)
	if req.Name != "" {
		query = query.Where("(name LIKE ? OR description LIKE ?)", "%"+req.Name+"%", "%"+req.Name+"%")
	}
	if req.Tag != "" {
		query = query.Where("tags LIKE ?", "%"+req.Tag+"%")
	}

	// 按排序字段排序
	query = query.Order("sort ASC, id DESC")

	var websites []domain.Website

	// 获取分页数据
	pageVo, err := page.Paginate[domain.Website](query, req.PageRequest, &websites)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取网站列表失败", err)
		return
	}

	// 转换为 VO
	websiteVOs := vo.ToWebsiteVOs(websites)
	websiteVOs = lo.Map(websiteVOs, func(websiteVO vo.WebsiteVO, _ int) vo.WebsiteVO {
		if websiteVO.Avatar != "" {
			websiteVO.Avatar, _ = cosClient.GenerateDownloadPresignedURL(websiteVO.Avatar)
		}
		return websiteVO
	})

	convert := page.Convert(pageVo, websiteVOs)

	utils.OkWithData(ctx, convert)
}

// GetFavicon 获取网站图标
// @Summary 获取网站图标
// @Description 自动获取网站的favicon图标
// @Tags 网站管理
// @Accept json
// @Produce json
// @Param data body dto.GetFaviconReq true "网站地址"
// @Success 200 {object} utils.Response{data=vo.GetFaviconVO}
// @Router /api/admin/websites/favicon [post]
func GetFavicon(ctx *gin.Context) {
	var req dto.GetFaviconReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)

	// 解析URL
	parsedURL, err := url.Parse(req.Url)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的URL", err)
		return
	}

	// 先尝试从网页内容中解析favicon链接
	faviconURL, err := extractFaviconFromHTML(req.Url)
	if err != nil || faviconURL == "" {
		// 如果解析失败，尝试常见路径
		faviconURLs := []string{
			fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host),
			fmt.Sprintf("%s://%s/favicon.png", parsedURL.Scheme, parsedURL.Host),
			fmt.Sprintf("%s://%s/apple-touch-icon.png", parsedURL.Scheme, parsedURL.Host),
			fmt.Sprintf("%s://%s/apple-touch-icon-precomposed.png", parsedURL.Scheme, parsedURL.Host),
		}

		for _, fURL := range faviconURLs {
			if checkURLExists(fURL) {
				faviconURL = fURL
				break
			}
		}

		if faviconURL == "" {
			// 如果都没找到，返回默认的favicon.ico路径
			faviconURL = fmt.Sprintf("%s://%s/favicon.ico", parsedURL.Scheme, parsedURL.Host)
		}
	}

	userId := utils.GetUIDFromContext(ctx)
	currentDate := time.Now().Format("2006-01-02")
	// 拼接文件保存地址 <用户 id>/<日期>/文件名
	// 解析原始文件名并生成唯一文件名
	u, _ := url.Parse(faviconURL)
	base := path.Base(u.Path)
	if base == "" || base == "/" {
		base = "favicon.ico"
	}
	if path.Ext(base) == "" {
		base = base + ".ico"
	}
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), base)
	keyPath := path.Join(strconv.FormatInt(userId, 10), currentDate, fileName)

	key, _ := cosClient.PutFromURL(ctx, keyPath, faviconURL)
	presignedURL, _ := cosClient.GenerateDownloadPresignedURL(key)

	utils.OkWithData(ctx, vo.GetFaviconVO{
		Favicon: presignedURL,
	})
}

// checkURLExists 检查URL是否存在（委托给 utils 包，供多处复用）
func checkURLExists(url string) bool {
	return utils.CheckURLExists(url)
}

// extractFaviconFromHTML 从网页内容中提取favicon链接（委托给 utils 包，供多处复用）
func extractFaviconFromHTML(urlStr string) (string, error) {
	return utils.ExtractFaviconFromHTML(urlStr)
}
