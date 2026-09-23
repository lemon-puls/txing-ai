package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow"
	"txing-ai/internal/agent/workflow/resolver"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/tool/ops"
	"txing-ai/internal/utils"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 默认配置（config.yaml 的 ops_agent 段缺省时生效）
const (
	defaultModelName = "deepseek-v3"
	defaultMaxRounds = 8
	// opsHistoryWindow 发送给 LLM 的历史消息窗口大小（从最新往前数）
	opsHistoryWindow = 20

	auditStatusFailed  = "failed"
	auditStatusSuccess = "completed"
)

// proposalOnlyPlaceholder 纯提案消息（无正文）发给 LLM 时的占位文本，避免空 content 请求部分渠道报 400
const proposalOnlyPlaceholder = "（生成了录入提案）"

// ChatStream 运营助手 SSE 对话
// @Summary 运营助手流式对话
// @Description 管理后台运营助手，基于 SSE 流式返回内容、工具调用进度与结构化提案（如网站录入提案，确认后才入库）；会话由服务端持久化，首帧返回 sessionId
// @Tags 运营助手
// @Accept json
// @Produce text/event-stream
// @Param data body dto.OpsChatStreamReq true "会话ID与本次输入"
// @Success 200 {string} string "SSE stream"
// @Router /api/admin/ops/chat/stream [post]
func ChatStream(ctx *gin.Context) {
	var req dto.OpsChatStreamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	cosClient := utils.GetCosClientFromContext[*utils.COSClient](ctx)
	userId := utils.GetUIDFromContext(ctx)
	role := utils.GetRoleFromContext(ctx)
	_ = role // /api/admin 前缀已在 AuthMiddleware 中校验超管角色

	// 配置（ops_agent 段可能缺省或热更新后被置空，需防御性判空）
	modelName := defaultModelName
	maxRounds := defaultMaxRounds
	if opsCfg := global.LoadConfig().OpsAgentConfig; opsCfg != nil {
		if opsCfg.Model != "" {
			modelName = opsCfg.Model
		}
		if opsCfg.MaxToolRounds > 0 {
			maxRounds = opsCfg.MaxToolRounds
		}
	}

	// 加载或创建会话（SSE 头输出前失败可直接走 JSON 错误响应）
	var session *domain.OpsChatSession
	if req.SessionId > 0 {
		var err error
		session, err = domain.QueryOpsChatSessionById(db, req.SessionId)
		if err != nil {
			utils.ErrorWithMsg(ctx, "会话不存在", err)
			return
		}
		if session.UserID != userId {
			utils.ErrorWithCode(ctx, global.CodeNotPermission, nil)
			return
		}
	} else {
		session = &domain.OpsChatSession{UserID: userId}
	}

	// 流前落库用户消息（对齐用户端 HandleMessage 时点），历史窗口不含本次输入
	session.AppendUserMessage(req.Content)
	session.Model = modelName
	session.SetContext(req.Context)
	if err := session.Save(db); err != nil {
		log.Error("保存运营助手会话失败", zap.Error(err))
		utils.ErrorWithMsg(ctx, "保存会话失败", err)
		return
	}
	history := buildOpsHistory(session.FormattedMessages[:len(session.FormattedMessages)-1])

	// 页面上下文注入系统提示词
	systemPrompt := ops.OpsSystemPrompt
	if req.Context != nil && (req.Context.Page != "" || len(req.Context.Draft) > 0) {
		if ctxJSON, err := json.Marshal(req.Context); err == nil {
			systemPrompt += "\n\n## 当前页面上下文\n" + string(ctxJSON)
		}
	}

	// 审计记录（流结束后落库）
	startTime := time.Now()
	var (
		toolCallRecords   []map[string]string
		capturedProposal  *ops.WebsiteProposal
		capturedStatus    string
		capturedMessage   string
		streamedContent   string
		streamedReasoning string
		liveToolCalls     []domain.OpsChatToolCall
		emittedContent    bool
	)
	opLog := domain.OpsAgentLog{
		UserID:    userId,
		SessionId: session.Id,
		Model:     modelName,
		Input:     req.Content,
		Status:    "running",
	}
	if req.Context != nil {
		opLog.Page = req.Context.Page
		if ctxJSON, err := json.Marshal(req.Context); err == nil {
			opLog.Context = string(ctxJSON)
		}
	}
	defer func() {
		opLog.DurationMs = time.Since(startTime).Milliseconds()
		if len(toolCallRecords) > 0 {
			if tcJSON, err := json.Marshal(toolCallRecords); err == nil {
				opLog.ToolCalls = string(tcJSON)
			}
		}
		if capturedProposal != nil {
			if pJSON, err := json.Marshal(capturedProposal); err == nil {
				opLog.Proposal = string(pJSON)
			}
		}
		if err := db.Create(&opLog).Error; err != nil {
			log.Error("保存运营助手审计日志失败", zap.Error(err))
		}
	}()

	// 设置 SSE 响应头（与 agent exec 流式端点保持一致）
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 可取消上下文：客户端断开（SSE 写失败）时主动取消，工作流据此识别中断并停止工具循环
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// SSE 帧写入（ExecuteLLM 为顺序执行，无需加锁）
	writeFrame := func(data map[string]interface{}) error {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData); err != nil {
			// 写失败（如客户端已断开，broken pipe）：立即取消上下文，
			// 否则请求 ctx 可能一直不取消，工作流会继续空跑工具轮次且中断无法被正确分类
			cancel()
			return err
		}
		ctx.Writer.Flush()
		return nil
	}

	// 首帧回传会话ID，前端据此绑定后续请求
	_ = writeFrame(map[string]interface{}{"type": "session", "sessionId": session.Id})

	// 工具集按请求构造，捕获当前请求的 DB / COS / userId
	tools := ops.ProvideOpsTools(ops.OpsToolDeps{DB: db, COS: cosClient, UserID: userId})
	modelResolver := resolver.NewChannelModelResolver(db)

	// chunk → SSE 帧分发
	callback := func(chunk *global.Chunk) error {
		// 工具调用生命周期
		if chunk.ToolCallId != "" {
			record := map[string]string{
				"id": chunk.ToolCallId, "name": chunk.ToolName,
				"params": chunk.ToolParams, "result": chunk.ToolResult,
				"status": chunk.ToolStatus,
			}
			toolCallRecords = append(toolCallRecords, record)
			upsertLiveToolCall(&liveToolCalls, chunk)

			if chunk.ToolStatus == "running" {
				return writeFrame(map[string]interface{}{
					"type": "tool_call", "toolCallId": chunk.ToolCallId,
					"toolName": chunk.ToolName, "toolParams": chunk.ToolParams,
					"status": "running", "showMsg": chunk.ShowMsg,
				})
			}

			// completed / failed / interrupted
			if err := writeFrame(map[string]interface{}{
				"type": "tool_result", "toolCallId": chunk.ToolCallId,
				"toolName": chunk.ToolName, "toolResult": chunk.ToolResult,
				"status": chunk.ToolStatus, "showMsg": chunk.ShowMsg,
			}); err != nil {
				return err
			}

			// 拦截网站提案：服务端解析校验后以独立帧下发
			if chunk.ToolName == ops.WebsitePreviewToolName && chunk.ToolStatus == "completed" {
				var previewPayload struct {
					Status   string               `json:"status"`
					Message  string               `json:"message"`
					Proposal *ops.WebsiteProposal `json:"proposal"`
				}
				if err := json.Unmarshal([]byte(chunk.ToolResult), &previewPayload); err != nil {
					log.Warn("解析网站提案失败", zap.String("toolResult", chunk.ToolResult), zap.Error(err))
					return nil
				}
				if previewPayload.Proposal != nil {
					capturedProposal = previewPayload.Proposal
					capturedStatus = previewPayload.Status
					capturedMessage = previewPayload.Message
					return writeFrame(map[string]interface{}{
						"type": "proposal", "proposal": previewPayload.Proposal,
						"status": previewPayload.Status, "message": previewPayload.Message,
					})
				}
			}
			return nil
		}

		// 内容 delta（流式）或整段内容
		if chunk.Content != "" || chunk.ReasoningContent != "" {
			emittedContent = true
			streamedContent += chunk.Content
			streamedReasoning += chunk.ReasoningContent
			return writeFrame(map[string]interface{}{
				"type": "content",
				"content": chunk.Content, "reasoningContent": chunk.ReasoningContent,
			})
		}

		// 过程提示
		if chunk.ShowMsg != "" {
			return writeFrame(map[string]interface{}{"type": "show_msg", "showMsg": chunk.ShowMsg})
		}
		return nil
	}

	result, err := workflow.ExecuteLLM(ctxWithCancel, &workflow.LLMExecConfig{
		NodeID:           "ops-agent",
		NodeLabel:        "运营助手",
		NodeType:         "agent",
		ModelResolver:    modelResolver,
		ModelName:        modelName,
		SystemPrompt:     systemPrompt,
		History:          history,
		AllTools:         tools,
		ToolNames:        []string{ops.WebsiteFetchToolName, ops.WebsitePreviewToolName},
		MaxToolRounds:    maxRounds,
		EmitFinalContent: true,
		Stream:           true,
	}, req.Content, callback)

	// 流后组装助手消息并落库（对齐用户端 SaveResponse 时点；错误/中断路径也持久化）
	assistantMsg := domain.OpsChatMessage{
		Role:      "assistant",
		Content:   streamedContent,
		Reasoning: streamedReasoning,
		ToolCalls: liveToolCalls,
	}
	if capturedProposal != nil {
		assistantMsg.Proposal = toDomainProposal(capturedProposal)
		assistantMsg.ProposalStatus = capturedStatus
		assistantMsg.ProposalMessage = capturedMessage
	}
	if err != nil {
		opLog.Error = err.Error()
		opLog.Status = auditStatusFailed
		opLog.Output = result
		if ctxWithCancel.Err() != nil {
			// 客户端主动中断：流式模式下 result 为空，只存 callback 累积的部分内容
			opLog.Status = "interrupted"
			assistantMsg.Interrupted = true
		} else {
			assistantMsg.Error = err.Error()
		}
		log.Error("运营助手执行失败", zap.Error(err))
		appendAndSaveSession(db, session, assistantMsg)
		_ = writeFrame(map[string]interface{}{"type": "error", "error": err.Error(), "end": true})
		return
	}

	// 兜底：全程无内容时补发一次（如流式模式下产出为空走兜底文案）
	if !emittedContent && result != "" {
		assistantMsg.Content = result
		_ = writeFrame(map[string]interface{}{"type": "content", "content": result, "reasoningContent": ""})
	}

	opLog.Status = auditStatusSuccess
	opLog.Output = result
	appendAndSaveSession(db, session, assistantMsg)
	_ = writeFrame(map[string]interface{}{"type": "end", "end": true})
}

// appendAndSaveSession 持久化助手消息，失败只记日志不阻断响应
func appendAndSaveSession(db *gorm.DB, session *domain.OpsChatSession, msg domain.OpsChatMessage) {
	session.AppendAssistantMessage(msg)
	if err := session.Save(db); err != nil {
		log.Error("保存运营助手会话失败", zap.Int64("sessionId", session.Id), zap.Error(err))
	}
}

// buildOpsHistory 将持久化消息转换为 LLM 历史窗口（跳过失败消息，纯提案消息用占位文本）。
// toolCalls 不回喂：当前仅两个查询类工具，结果已在上一轮 assistant 正文中被模型总结。
func buildOpsHistory(messages []domain.OpsChatMessage) []*schema.Message {
	start := 0
	if len(messages) > opsHistoryWindow {
		start = len(messages) - opsHistoryWindow
	}
	window := messages[start:]

	history := make([]*schema.Message, 0, len(window))
	for _, m := range window {
		if m.Role == "user" {
			history = append(history, schema.UserMessage(m.Content))
			continue
		}
		// 失败的空消息不进历史，避免误导模型
		if m.Content == "" && m.Error != "" {
			continue
		}
		content := m.Content
		if content == "" {
			content = proposalOnlyPlaceholder
		}
		history = append(history, schema.AssistantMessage(content, nil))
	}
	return history
}

// upsertLiveToolCall 按 ToolCallId 更新或追加工具调用记录（running → completed 各一帧）
func upsertLiveToolCall(list *[]domain.OpsChatToolCall, chunk *global.Chunk) {
	for i := range *list {
		if (*list)[i].Id == chunk.ToolCallId {
			(*list)[i].Name = chunk.ToolName
			(*list)[i].Params = chunk.ToolParams
			(*list)[i].Result = chunk.ToolResult
			(*list)[i].Status = chunk.ToolStatus
			return
		}
	}
	*list = append(*list, domain.OpsChatToolCall{
		Id:     chunk.ToolCallId,
		Name:   chunk.ToolName,
		Params: chunk.ToolParams,
		Result: chunk.ToolResult,
		Status: chunk.ToolStatus,
	})
}

// toDomainProposal 将 tool 层提案转换为 domain 镜像结构（tool→domain 不可反向 import）
func toDomainProposal(p *ops.WebsiteProposal) *domain.OpsChatProposal {
	return &domain.OpsChatProposal{
		Type:        p.Type,
		Name:        p.Name,
		Description: p.Description,
		Url:         p.Url,
		Avatar:      p.Avatar,
		Tags:        p.Tags,
	}
}
