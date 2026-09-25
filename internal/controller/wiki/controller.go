package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	wikisvc "txing-ai/internal/service/wiki"
	"txing-ai/internal/utils"
)

// getRedis 获取 gin context 中的 redis 客户端（BuiltinMiddleWare 注入）
func getRedis(ctx *gin.Context) *redis.Client {
	return utils.GetFromContext[*redis.Client](ctx, "redis")
}

// newService 按请求构造服务（DB/COS 来自 gin context，对齐 ops 范式）
func newService(ctx *gin.Context) (*wikisvc.Service, error) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)
	return wikisvc.NewService(db, cosClient)
}

// pageParams 解析分页参数
func pageParams(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	return page, pageSize
}

// --- 公开问答 ---

// Ask 面试官匿名问答（SSE）
// @Summary 知识库问答（公开）
// @Description 面试官/访客匿名向知识库 AI 提问；SSE 流式返回内容与工具调用进度。多轮上下文由前端携带；IP 限流 + 全站日配额；问答会被记录（前端须展示"对话将被记录"声明）
// @Tags 知识库
// @Accept json
// @Produce text/event-stream
// @Param data body dto.WikiAskReq true "会话标识与提问"
// @Success 200 {string} string "SSE stream"
// @Router /api/wiki/ask [post]
func Ask(ctx *gin.Context) {
	if !wikisvc.AskEnabled() {
		utils.ErrorWithMsg(ctx, "问答功能未开放", nil)
		return
	}

	var req dto.WikiAskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	// 限流（SSE 头输出前完成，失败走 JSON 错误）
	if err := wikisvc.CheckAskLimit(ctx, getRedis(ctx), ctx.ClientIP()); err != nil {
		utils.ErrorWithMsg(ctx, err.Error(), nil)
		return
	}

	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}

	// SSE 响应头（对齐 ops）
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓存

	// 可取消上下文：客户端断开时主动取消
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	writeFrame := func(data map[string]interface{}) error {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData); err != nil {
			cancel() // 客户端已断开（broken pipe）
			return err
		}
		ctx.Writer.Flush()
		return nil
	}

	_ = writeFrame(map[string]interface{}{"type": "start"})

	callback := func(chunk *global.Chunk) error {
		if chunk.ToolCallId != "" {
			if chunk.ToolStatus == "running" {
				return writeFrame(map[string]interface{}{
					"type": "tool_call", "toolCallId": chunk.ToolCallId,
					"toolName": chunk.ToolName, "toolParams": chunk.ToolParams,
					"status": "running", "showMsg": chunk.ShowMsg,
				})
			}
			return writeFrame(map[string]interface{}{
				"type": "tool_result", "toolCallId": chunk.ToolCallId,
				"toolName": chunk.ToolName, "toolResult": chunk.ToolResult,
				"status": chunk.ToolStatus, "showMsg": chunk.ShowMsg,
			})
		}
		if chunk.Content != "" || chunk.ReasoningContent != "" {
			return writeFrame(map[string]interface{}{
				"type": "content",
				"content": chunk.Content, "reasoningContent": chunk.ReasoningContent,
			})
		}
		if chunk.ShowMsg != "" {
			return writeFrame(map[string]interface{}{"type": "show_msg", "showMsg": chunk.ShowMsg})
		}
		return nil
	}

	if err := svc.Ask(ctxWithCancel, &req, ctx.ClientIP(), callback); err != nil {
		log.Error("wiki ask 执行失败", zap.Error(err))
		_ = writeFrame(map[string]interface{}{"type": "error", "error": err.Error(), "end": true})
		return
	}
	_ = writeFrame(map[string]interface{}{"type": "end", "end": true})
}

// --- 源管理 ---

// CreateMDSource 新建 markdown 源
// @Summary 新建 markdown 源
// @Description 上传 markdown 原文创建 Raw 源（服务端转存 COS，只存 key）
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param data body dto.WikiMDSourceReq true "标题与正文"
// @Success 200 {object} domain.WikiSource
// @Router /api/admin/wiki/sources/md [post]
func CreateMDSource(ctx *gin.Context) {
	var req dto.WikiMDSourceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	src, err := svc.CreateMDSource(ctx, req.Title, req.Content)
	if err != nil {
		utils.ErrorWithMsg(ctx, "创建源失败", err)
		return
	}
	utils.OkWithData(ctx, src)
}

// CreateURLSource 新建 URL 源
// @Summary 新建网页 URL 源
// @Description 登记网页地址，ingest 时由服务端抓取正文
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param data body dto.WikiURLSourceReq true "标题与地址"
// @Success 200 {object} domain.WikiSource
// @Router /api/admin/wiki/sources/url [post]
func CreateURLSource(ctx *gin.Context) {
	var req dto.WikiURLSourceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	src, err := svc.CreateURLSource(ctx, req.Title, req.Url)
	if err != nil {
		utils.ErrorWithMsg(ctx, "创建源失败", err)
		return
	}
	utils.OkWithData(ctx, src)
}

// ListSources 源列表
// @Summary 分页列出知识库源
// @Tags 知识库管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/wiki/sources/list [get]
func ListSources(ctx *gin.Context) {
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	page, pageSize := pageParams(ctx)
	list, total, err := svc.ListSources(ctx, page, pageSize)
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询源列表失败", err)
		return
	}
	utils.OkWithData(ctx, gin.H{"list": list, "total": total})
}

// DeleteSource 删除源
// @Summary 删除知识库源
// @Tags 知识库管理
// @Produce json
// @Param id path int true "源ID"
// @Success 200 {string} string
// @Router /api/admin/wiki/sources/{id} [delete]
func DeleteSource(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的源ID", err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.DeleteSource(ctx, id); err != nil {
		utils.ErrorWithMsg(ctx, "删除源失败", err)
		return
	}
	utils.Ok(ctx)
}

// TriggerIngest 触发源编译
// @Summary 触发源的 ingest 编译
// @Description 异步执行：抓取/拉取正文 → agent 编译为草稿；通过源列表的 status 轮询进度
// @Tags 知识库管理
// @Produce json
// @Param id path int true "源ID"
// @Success 200 {string} string
// @Router /api/admin/wiki/sources/{id}/ingest [post]
func TriggerIngest(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的源ID", err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.TriggerIngest(id); err != nil {
		utils.ErrorWithMsg(ctx, "触发编译失败", err)
		return
	}
	utils.Ok(ctx)
}

// --- 草稿审核 ---

// ListDrafts 草稿列表
// @Summary 分页列出待审核草稿
// @Tags 知识库管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param keyword query string false "关键词（标题/slug/摘要）"
// @Param pageType query string false "页面类型 summary/entity/concept"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/wiki/drafts/list [get]
func ListDrafts(ctx *gin.Context) {
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	page, pageSize := pageParams(ctx)
	list, total, err := svc.ListDrafts(ctx, page, pageSize, ctx.Query("keyword"), ctx.Query("pageType"))
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询草稿列表失败", err)
		return
	}
	utils.OkWithData(ctx, gin.H{"list": list, "total": total})
}

// UpdateDraft 编辑草稿
// @Summary 编辑草稿（可改标题/类型/别名/摘要/正文）
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param id path int true "草稿ID"
// @Param data body dto.WikiUpdateDraftReq true "更新字段（只更新传入项）"
// @Success 200 {string} string
// @Router /api/admin/wiki/drafts/{id} [put]
func UpdateDraft(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的草稿ID", err)
		return
	}
	var req dto.WikiUpdateDraftReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.UpdateDraft(ctx, id, &req); err != nil {
		utils.ErrorWithMsg(ctx, "更新草稿失败", err)
		return
	}
	utils.Ok(ctx)
}

// ConfirmDraft 确认草稿
// @Summary 确认草稿为已发布（同 slug 已有发布页则合并更新）
// @Tags 知识库管理
// @Produce json
// @Param id path int true "草稿ID"
// @Success 200 {string} string
// @Router /api/admin/wiki/drafts/{id}/confirm [post]
func ConfirmDraft(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的草稿ID", err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.ConfirmDraft(ctx, id); err != nil {
		utils.ErrorWithMsg(ctx, "确认草稿失败", err)
		return
	}
	utils.Ok(ctx)
}

// ConfirmAllDrafts 全部确认
// @Summary 一键确认全部草稿
// @Tags 知识库管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/wiki/drafts/confirm_all [post]
func ConfirmAllDrafts(ctx *gin.Context) {
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	confirmed, err := svc.ConfirmAllDrafts(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, fmt.Sprintf("已确认 %d 条后失败", confirmed), err)
		return
	}
	utils.OkWithData(ctx, gin.H{"confirmed": confirmed})
}

// DeleteDraft 丢弃草稿
// @Summary 丢弃草稿
// @Tags 知识库管理
// @Produce json
// @Param id path int true "草稿ID"
// @Success 200 {string} string
// @Router /api/admin/wiki/drafts/{id} [delete]
func DeleteDraft(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的草稿ID", err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.DeleteDraft(ctx, id); err != nil {
		utils.ErrorWithMsg(ctx, "丢弃草稿失败", err)
		return
	}
	utils.Ok(ctx)
}

// --- 已发布页管理 ---

// ListPublished 已发布页列表
// @Summary 分页列出已发布页面
// @Tags 知识库管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param keyword query string false "关键词（标题/slug/摘要）"
// @Param pageType query string false "页面类型 summary/entity/concept"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/wiki/pages/list [get]
func ListPublished(ctx *gin.Context) {
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	page, pageSize := pageParams(ctx)
	list, total, err := svc.ListPublished(ctx, page, pageSize, ctx.Query("keyword"), ctx.Query("pageType"))
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询页面列表失败", err)
		return
	}
	utils.OkWithData(ctx, gin.H{"list": list, "total": total})
}

// GetPublished 页面详情
// @Summary 获取单个已发布页
// @Tags 知识库管理
// @Produce json
// @Param id path int true "页面ID"
// @Success 200 {object} domain.WikiPage
// @Router /api/admin/wiki/pages/{id} [get]
func GetPublished(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的页面ID", err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	page, err := svc.GetPublished(ctx, id)
	if err != nil {
		utils.ErrorWithMsg(ctx, "查询页面失败", err)
		return
	}
	utils.OkWithData(ctx, page)
}

// UpdatePublished 编辑已发布页
// @Summary 编辑已发布页（version+1 并重建出入链与 index）
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param id path int true "页面ID"
// @Param data body dto.WikiUpdateDraftReq true "更新字段（只更新传入项）"
// @Success 200 {string} string
// @Router /api/admin/wiki/pages/{id} [put]
func UpdatePublished(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的页面ID", err)
		return
	}
	var req dto.WikiUpdateDraftReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.UpdatePublished(ctx, id, &req); err != nil {
		utils.ErrorWithMsg(ctx, "更新页面失败", err)
		return
	}
	utils.Ok(ctx)
}

// OfflinePage 下线页面
// @Summary 下线已发布页
// @Tags 知识库管理
// @Produce json
// @Param id path int true "页面ID"
// @Success 200 {string} string
// @Router /api/admin/wiki/pages/{id}/offline [post]
func OfflinePage(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "无效的页面ID", err)
		return
	}
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	if err := svc.OfflinePage(ctx, id); err != nil {
		utils.ErrorWithMsg(ctx, "下线页面失败", err)
		return
	}
	utils.Ok(ctx)
}

// --- 导出 ---

// ExportAll 导出全部已发布页
// @Summary 导出全部已发布页为 md 压缩包（含 index.md）
// @Tags 知识库管理
// @Produce application/zip
// @Success 200 {file} file "zip 压缩包"
// @Router /api/admin/wiki/export [get]
func ExportAll(ctx *gin.Context) {
	svc, err := newService(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "服务初始化失败", err)
		return
	}
	data, filename, err := svc.ExportAll(ctx)
	if err != nil {
		utils.ErrorWithMsg(ctx, "导出失败", err)
		return
	}
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	ctx.Data(http.StatusOK, "application/zip", data)
}
