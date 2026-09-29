package ops

import (
	"encoding/json"
	"strconv"
	"strings"

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

	// 回填各会话最近一次请求的页面标识（Context JSON 中的 page，copier 无法自动映射），
	// 前端据此优先恢复与当前页面同页的会话，避免跨页串台
	for i := range result.Data {
		pageVO.Data[i].Page = extractSessionPage(result.Data[i].Context)
	}

	utils.OkWithData(c, pageVO)
}

// extractSessionPage 从会话的上下文 JSON 中提取页面标识（解析失败返回空串）
func extractSessionPage(context string) string {
	if context == "" {
		return ""
	}
	var ctx struct {
		Page string `json:"page"`
	}
	if err := json.Unmarshal([]byte(context), &ctx); err != nil {
		return ""
	}
	return ctx.Page
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
// @Description 持久化前端本地产生的消息（如提案确认提示）；markProposalConfirmed 时将未确认提案置为已确认；rewindLastRound 时回退最后一轮问答（重新生成/失败重试用）
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
	// 回退模式下允许空 Content（不追加消息），普通追加仍要求非空
	if !req.RewindLastRound && strings.TrimSpace(req.Content) == "" {
		utils.ErrorWithMsg(c, "消息内容不能为空", nil)
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

	// 回退最后一轮问答：移除末尾的助手消息（成功回复或失败/中断占位）与其对应用户消息。
	// 重新生成/失败重试前由前端调用，随后重发原问题，避免会话历史与前端界面出现重复轮次
	if req.RewindLastRound && len(session.FormattedMessages) > 0 {
		msgs := session.FormattedMessages
		if msgs[len(msgs)-1].Role == "assistant" {
			msgs = msgs[:len(msgs)-1]
		}
		if len(msgs) > 0 && msgs[len(msgs)-1].Role == "user" {
			msgs = msgs[:len(msgs)-1]
		}
		session.FormattedMessages = msgs
		// 纯回退请求：保存后直接返回，不追加消息
		if strings.TrimSpace(req.Content) == "" {
			if err := session.Save(db); err != nil {
				log.Error("保存运营助手会话失败", zap.Int64("sessionId", session.Id), zap.Error(err))
				utils.ErrorWithCode(c, global.CodeServerInternalError, err)
				return
			}
			utils.Ok(c)
			return
		}
	}

	// 将未确认的提案标记为已确认：优先按前端传来的提案 name 精确匹配
	// （批量录入一条消息可携带多张提案卡片），未传 name 时回退为最后一条未确认提案
	if req.MarkProposalConfirmed {
	matched:
		for i := len(session.FormattedMessages) - 1; i >= 0; i-- {
			m := &session.FormattedMessages[i]
			for j := len(m.Proposals) - 1; j >= 0; j-- {
				item := &m.Proposals[j]
				if len(item.Proposal) == 0 || item.Status == "" || item.Status == "confirmed" {
					continue
				}
				if req.ProposalName != "" {
					var named struct {
						Name string `json:"name"`
					}
					if err := json.Unmarshal(item.Proposal, &named); err == nil && named.Name != "" && named.Name != req.ProposalName {
						continue
					}
				}
				item.Status = "confirmed"
				break matched
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
