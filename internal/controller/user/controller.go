package user

import (
	"fmt"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
// @Summary 用户注册
// @Description 新用户注册
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body dto.RegisterReq true "注册信息"
// @Success 200 {object} utils.Response{data=vo.UserVO}
// @Router /api/user/register [post]
func userRegister(ctx *gin.Context) {
	var req dto.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	// 检查用户名是否已存在
	var count int64
	if err := db.Model(&domain.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		utils.ErrorWithMsg(ctx, "系统错误", err)
		return
	}
	if count > 0 {
		utils.ErrorWithMsg(ctx, "用户名已存在", nil)
		return
	}

	// 创建用户
	user := &domain.User{
		Username: req.Username,
		//Password: utils.EncryptPassword(req.Password), TODO 实现加密密码
		Password: req.Password,
		Email:    req.Email,
		Phone:    req.Phone,
		Gender:   req.Gender,
		Age:      req.Age,
		Avatar:   req.Avatar,
		Status:   0, // 正常状态
		Role:     0, // 普通用户
	}

	if err := db.Create(user).Error; err != nil {
		utils.ErrorWithMsg(ctx, "注册失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToUserVO(*user))
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并返回token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body dto.LoginReq true "登录信息"
// @Success 200 {object} utils.Response{data=vo.LoginVO}
// @Router /api/user/login [post]
func Login(ctx *gin.Context) {
	var req dto.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	var user domain.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.ErrorWithMsg(ctx, "用户不存在", err)
		return
	}

	// TODO 实现验证密码
	//if !utils.ValidatePassword(req.Password, user.Password) {
	//	utils.ErrorWithMsg(ctx, "密码错误", nil)
	//	return
	//}

	// 检查用户状态
	if user.Status == 1 {
		utils.ErrorWithMsg(ctx, "账号已被禁用", nil)
		return
	}

	// TODO 实现生成token逻辑
	//token, err := utils.GenerateToken(user.Id)
	//if err != nil {
	//	utils.ErrorWithMsg(ctx, "登录失败", err)
	//	return
	//}

	utils.OkWithData(ctx, vo.ToLoginVO(user, "token"))
}

// UpdateProfile 更新个人信息
// @Summary 更新个人信息
// @Description 更新当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body dto.UpdateProfileReq true "个人信息"
// @Success 200 {object} utils.Response{data=vo.UserVO}
// @Router /api/user/profile [put]
func UpdateProfile(ctx *gin.Context) {
	var req dto.UpdateProfileReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)
	// TODO 获取用户信息
	//userId := utils.GetUserIdFromContext(ctx)

	var user domain.User
	if err := db.First(&user, "userId").Error; err != nil {
		utils.ErrorWithMsg(ctx, "用户不存在", err)
		return
	}

	user.Email = req.Email
	user.Phone = req.Phone
	user.Gender = req.Gender
	user.Age = req.Age
	user.Avatar = req.Avatar

	if err := db.Save(&user).Error; err != nil {
		utils.ErrorWithMsg(ctx, "更新失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToUserVO(user))
}

// UpdatePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body dto.UpdatePasswordReq true "密码信息"
// @Success 200 {object} utils.Response
// @Router /api/user/password [put]
func UpdatePassword(ctx *gin.Context) {
	var req dto.UpdatePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)
	// TODO 获取用户信息
	//userId := utils.GetUserIdFromContext(ctx)

	var user domain.User
	if err := db.First(&user, "userId").Error; err != nil {
		utils.ErrorWithMsg(ctx, "用户不存在", err)
		return
	}

	// TODO 验证旧密码
	//if !utils.ValidatePassword(req.OldPassword, user.Password) {
	//	utils.ErrorWithMsg(ctx, "原密码错误", nil)
	//	return
	//}

	// TODO 更新密码
	//user.Password = utils.EncryptPassword(req.NewPassword)
	//if err := db.Save(&user).Error; err != nil {
	//	utils.ErrorWithMsg(ctx, "修改密码失败", err)
	//	return
	//}

	utils.OkWithMsg(ctx, "修改密码成功")
}

// ResetPassword 重置密码
// @Summary 重置密码
// @Description 通过邮箱重置密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body dto.ResetPasswordReq true "重置密码信息"
// @Success 200 {object} utils.Response
// @Router /api/user/reset-password [post]
func ResetPassword(ctx *gin.Context) {
	var req dto.ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	var user domain.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.ErrorWithMsg(ctx, "用户不存在", err)
		return
	}

	// TODO: 发送重置密码邮件
	// 这里需要集成邮件服务

	utils.OkWithMsg(ctx, "重置密码邮件已发送")
}

// Logout 退出登录
// @Summary 退出登录
// @Description 清除用户登录状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/user/logout [post]
func Logout(ctx *gin.Context) {
	// TODO: 清除token
	// 这里需要根据具体的token存储方式来实现

	utils.OkWithMsg(ctx, "退出成功")
}

// List 获取用户列表(管理员接口)
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式(asc/desc)"
// @Success 200 {object} utils.Response
// @Router /api/admin/user/list [get]
func List(ctx *gin.Context) {
	var req dto.ListUserReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext(ctx)

	// 构建查询条件
	query := db.Model(&domain.User{})
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var users []domain.User

	// 获取分页数据
	pageVo, err := page.Paginate[domain.User](query, req.PageRequest, &users)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取用户列表失败", err)
		return
	}

	// 转换为 VO
	userVOs := vo.ToUserVOs(users)
	convert := page.Convert(pageVo, userVOs)

	utils.OkWithData(ctx, convert)
}

// ToggleUserStatus 切换用户状态(管理员接口)
// @Summary 切换用户状态
// @Description 启用或禁用指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Router /api/admin/user/status/{id} [put]
func ToggleUserStatus(ctx *gin.Context) {
	db := utils.GetDBFromContext(ctx)

	var user domain.User
	if err := db.First(&user, ctx.Param("id")).Error; err != nil {
		utils.ErrorWithMsg(ctx, "用户不存在", err)
		return
	}

	// 不能修改超级管理员状态
	if user.Role == 1 {
		utils.ErrorWithMsg(ctx, "不能修改超级管理员状态", nil)
		return
	}

	// 切换状态: 0 -> 1 或 1 -> 0
	user.Status = 1 - user.Status
	if err := db.Save(&user).Error; err != nil {
		utils.ErrorWithMsg(ctx, "修改用户状态失败", err)
		return
	}

	statusMsg := "启用"
	if user.Status == 1 {
		statusMsg = "禁用"
	}
	utils.OkWithMsg(ctx, fmt.Sprintf("用户%s成功", statusMsg))
}
