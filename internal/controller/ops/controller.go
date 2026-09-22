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
	defaultModelName   = "deepseek-v3"
	defaultMaxRounds   = 8
	auditStatusFailed  = "failed"
	auditStatusSuccess = "completed"
)

// ChatStream 运营助手 SSE 对话
// @Summary 运营助手流式对话
// @Description 管理后台运营助手，基于 SSE 流式返回内容、工具调用进度与结构化提案（如网站录入提案，确认后才入库）
// @Tags 运营助手
// @Accept json
// @Produce text/event-stream
// @Param data body dto.OpsChatStreamReq true "对话消息与页面上下文"
// @Success 200 {string} string "SSE stream"
// @Router /api/admin/ops/chat/stream [post]
func ChatStream(ctx *gin.Context) {
	var req dto.OpsChatStreamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	// 对话必须以 user 消息结尾（作为本次输入）
	lastMsg := req.Messages[len(req.Messages)-1]
	if lastMsg.Role != "user" {
		utils.ErrorWithMsg(ctx, "最后一条消息必须是 user 角色", nil)
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

	// 页面上下文注入系统提示词
	systemPrompt := ops.OpsSystemPrompt
	if req.Context != nil && (req.Context.Page != "" || len(req.Context.Draft) > 0) {
		if ctxJSON, err := json.Marshal(req.Context); err == nil {
			systemPrompt += "\n\n## 当前页面上下文\n" + string(ctxJSON)
		}
	}

	// 构建多轮历史（最后一条 user 消息作为本次输入，不重复放入 History）
	var history []*schema.Message
	for i := 0; i < len(req.Messages)-1; i++ {
		m := req.Messages[i]
		if m.Role == "user" {
			history = append(history, schema.UserMessage(m.Content))
		} else {
			history = append(history, schema.AssistantMessage(m.Content, nil))
		}
	}

	// 审计记录（流结束后落库）
	startTime := time.Now()
	var (
		toolCallRecords []map[string]string
		capturedProposal *ops.WebsiteProposal
		emittedContent  bool
	)
	opLog := domain.OpsAgentLog{
		UserID: userId,
		Model:  modelName,
		Status: "running",
	}
	if req.Context != nil {
		opLog.Page = req.Context.Page
		if ctxJSON, err := json.Marshal(req.Context); err == nil {
			opLog.Context = string(ctxJSON)
		}
	}
	if msgJSON, err := json.Marshal(req.Messages); err == nil {
		opLog.Input = string(msgJSON)
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

	// SSE 帧写入（ExecuteLLM 为顺序执行，无需加锁）
	writeFrame := func(data map[string]interface{}) error {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData); err != nil {
			return err
		}
		ctx.Writer.Flush()
		return nil
	}

	// 工具集按请求构造，捕获当前请求的 DB / COS / userId
	tools := ops.ProvideOpsTools(ops.OpsToolDeps{DB: db, COS: cosClient, UserID: userId})
	modelResolver := resolver.NewChannelModelResolver(db)

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

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
					Proposal *ops.WebsiteProposal `json:"proposal"`
				}
				if err := json.Unmarshal([]byte(chunk.ToolResult), &previewPayload); err != nil {
					log.Warn("解析网站提案失败", zap.String("toolResult", chunk.ToolResult), zap.Error(err))
					return nil
				}
				if previewPayload.Proposal != nil {
					capturedProposal = previewPayload.Proposal
					return writeFrame(map[string]interface{}{
						"type": "proposal", "proposal": previewPayload.Proposal,
						"status": previewPayload.Status,
					})
				}
			}
			return nil
		}

		// 内容 delta（流式）或整段内容
		if chunk.Content != "" || chunk.ReasoningContent != "" {
			emittedContent = true
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
	}, lastMsg.Content, callback)

	if err != nil {
		opLog.Error = err.Error()
		opLog.Status = auditStatusFailed
		if ctxWithCancel.Err() != nil {
			opLog.Status = "interrupted"
		}
		opLog.Output = result
		log.Error("运营助手执行失败", zap.Error(err))
		_ = writeFrame(map[string]interface{}{"type": "error", "error": err.Error(), "end": true})
		return
	}

	// 兜底：全程无内容时补发一次（如流式模式下产出为空走兜底文案）
	if !emittedContent && result != "" {
		_ = writeFrame(map[string]interface{}{"type": "content", "content": result, "reasoningContent": ""})
	}

	opLog.Status = auditStatusSuccess
	opLog.Output = result
	_ = writeFrame(map[string]interface{}{"type": "end", "end": true})
}
