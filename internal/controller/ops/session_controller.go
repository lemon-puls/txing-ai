package ops

import (
	"strconv"

	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/jinzhu/copier"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetOpsSessionList 运营助手会话列表
// @Summary 运营助手会话列表
// @Description 游标分页获取当前管理员的运营助手会话列表（按更新时间倒序）
// @Tags 运营助手
// @Accept json
// @Produce json
// @Param data body dto.OpsChatSessionListReq true "游标分页参数"
// @Success 200 {object} utils.Response "成功"
// @Router /api/admin/ops/chat/sessions/list [post]
func GetOpsSessionList(c *gin.Context) {
	userId := utils.GetUIDFromContext(c)

	var req dto.OpsChatSessionListReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithCode(c, global.CodeInvalidParams, err)
		return
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	db := utils.GetDBFromContext[*gorm.DB](c)

	result, err := page.GetCursorPageByMySQL[domain.OpsChatSession](
		db,
		req.CursorPageBaseRequest,
		func(db *gorm.DB) {
			db.Where("user_id = ?", userId)
		},
		func(t *domain.OpsChatSession) interface{} {
			return &t.UpdateTime
		},
	)
	if err != nil {
		log.Error("query ops chat sessions failed", zap.Error(err))
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	pageVO, err := page.ConvertCursorPageVO[domain.OpsChatSession, vo.OpsChatSessionSimpleVO](result)
	if err != nil {
		log.Error("ConvertCursorPageVO failed", zap.Error(err))
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	utils.OkWithData(c, pageVO)
}

// GetOpsSessionDetail 运营助手会话详情
// @Summary 运营助手会话详情
// @Description 获取指定会话的完整消息列表，用于刷新后回放
// @Tags 运营助手
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} utils.Response{data=vo.OpsChatSessionDetailVO} "成功"
// @Failure 403 {object} utils.Response "无权限"
// @Router /api/admin/ops/chat/sessions/{id} [get]
func GetOpsSessionDetail(c *gin.Context) {
	userId := utils.GetUIDFromContext(c)

	sessionId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, global.CodeInvalidParams, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](c)

	session, err := domain.QueryOpsChatSessionById(db, sessionId)
	if err != nil {
		utils.ErrorWithMsg(c, "会话不存在", err)
		return
	}
	if session.UserID != userId {
		utils.ErrorWithCode(c, global.CodeNotPermission, nil)
		return
	}

	result := &vo.OpsChatSessionDetailVO{}
	if err := copier.Copy(result, session); err != nil {
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}
	result.Messages = session.FormattedMessages
	if result.Messages == nil {
		result.Messages = []domain.OpsChatMessage{}
	}

	utils.OkWithData(c, result)
}

// DeleteOpsSession 删除运营助手会话
// @Summary 删除运营助手会话
// @Description 软删除指定的会话（仅限本人会话）
// @Tags 运营助手
// @Produce json
// @Param id path int true "会话ID"
// @Success 200 {object} utils.Response "成功"
// @Failure 403 {object} utils.Response "无权限"
// @Router /api/admin/ops/chat/sessions/{id} [delete]
func DeleteOpsSession(c *gin.Context) {
	userId := utils.GetUIDFromContext(c)

	sessionId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, global.CodeInvalidParams, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](c)

	result := db.Where("id = ? AND user_id = ?", sessionId, userId).Delete(&domain.OpsChatSession{})
	if result.Error != nil {
		utils.ErrorWithCode(c, global.CodeServerInternalError, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		utils.ErrorWithMsg(c, "会话不存在或无权限删除", nil)
		return
	}

	utils.Ok(c)
}

// AppendOpsSession 追加运营助手会话消息
// @Summary 追加运营助手会话消息
// @Description 持久化前端本地产生的消息（如提案确认提示）；markProposalConfirmed 时将最后一条未确认提案置为已确认
// @Tags 运营助手
// @Accept json
// @Produce json
// @Param id path int true "会话ID"
// @Param data body dto.OpsChatAppendReq true "追加的消息"
// @Success 200 {object} utils.Response "成功"
// @Failure 403 {object} utils.Response "无权限"
// @Router /api/admin/ops/chat/sessions/{id}/append [post]
func AppendOpsSession(c *gin.Context) {
	userId := utils.GetUIDFromContext(c)

	sessionId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithCode(c, global.CodeInvalidParams, err)
		return
	}

	var req dto.OpsChatAppendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(c, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](c)

	session, err := domain.QueryOpsChatSessionById(db, sessionId)
	if err != nil {
		utils.ErrorWithMsg(c, "会话不存在", err)
		return
	}
	if session.UserID != userId {
		utils.ErrorWithCode(c, global.CodeNotPermission, nil)
		return
	}

	// 将最后一条未确认的提案标记为已确认（从尾向前找，命中即止）
	if req.MarkProposalConfirmed {
		for i := len(session.FormattedMessages) - 1; i >= 0; i-- {
			m := &session.FormattedMessages[i]
			if m.Proposal != nil && m.ProposalStatus != "" && m.ProposalStatus != "confirmed" {
				m.ProposalStatus = "confirmed"
				break
			}
		}
	}

	session.AppendAssistantMessage(domain.OpsChatMessage{
		Role:      req.Role,
		Content:   req.Content,
		Confirmed: true,
	})
	if err := session.Save(db); err != nil {
		log.Error("保存运营助手会话失败", zap.Int64("sessionId", session.Id), zap.Error(err))
		utils.ErrorWithCode(c, global.CodeServerInternalError, err)
		return
	}

	utils.Ok(c)
}
