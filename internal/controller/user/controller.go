package user

import (
	"txing-ai/internal/dto"
	"txing-ai/internal/service/user"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
)

// GetUserList 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页和条件筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式(asc/desc)"
// @Param userId query string false "用户ID"
// @Param status query int false "状态(0:禁用 1:启用)"
// @Success 200 {object} utils.Response
// @Router /api/user/list [get]
func GetUserList(ctx *gin.Context) {
	var req dto.UserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	pageVo := user.ListUser(req, utils.GetDBFromContext(ctx))

	userVOs := vo.ToUserVOs(pageVo.Records)

	convert := page.Convert(pageVo, userVOs)

	utils.OkWithData(ctx, convert)
}
