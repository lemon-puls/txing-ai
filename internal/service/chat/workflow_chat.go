package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"txing-ai/internal/agent/workflow"
	"txing-ai/internal/agent/workflow/resolver"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
	channelservice "txing-ai/internal/service/channel"
	workflowservice "txing-ai/internal/service/workflow"
	"txing-ai/internal/tool"
	"txing-ai/internal/utils"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NodeLog 节点执行日志（用于持久化，结构与前端 WorkflowMessage 的 nodeLogs 一致）
type NodeLog struct {
	NodeId    string     `json:"nodeId"`
	Type      string     `json:"type"`
	Label     string     `json:"label"`
	Status    string     `json:"status"`
	ToolCalls []ToolCall `json:"toolCalls"`
}

// ToolCall 工具调用记录
type ToolCall struct {
	ID     string `json:"id,omitempty"` // 工具调用 ID
	Name   string `json:"name"`
	Status string `json:"status"`
	Args   string `json:"args,omitempty"`   // 调用参数（截断后，仅用于展示）
	Result string `json:"result,omitempty"` // 执行结果（截断后，仅用于展示）
}

// 展示用截断阈值（args/result 可能很大，且 nodeLogs 会持久化进会话消息）
const (
	toolArgsMaxLen   = 1024
	toolResultMaxLen = 4096

	// workflowStatusInterrupted 用户停止生成时的工作流终态
	// （与 completed/failed 并列，前端据此显示"已中断"）
	workflowStatusInterrupted = "interrupted"
)

// isWorkflowCanceled 判断工作流执行错误是否由用户停止生成触发。
// 底层（eino/模型适配层）对 context 取消的包装方式不一，除 errors.Is 外
// 再兜底匹配错误文本，避免把"用户主动停止"误判为执行失败
func isWorkflowCanceled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "context canceled")
}

// truncateForDisplay 按 rune 截断超长文本，避免截断多字节字符
func truncateForDisplay(s string, max int) string {
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	return string(rs[:max]) + "\n…(内容过长已截断)"
}

// progressFromChunk 将执行 chunk 转换为前端工作流进度消息。
// 回调累积与写协程下发共用同一构建逻辑，保证 resume 回放与实时增量格式一致。
func progressFromChunk(chunk *global.Chunk) *dto.WorkflowProgress {
	return &dto.WorkflowProgress{
		Status:     "running",
		NodeID:     chunk.NodeId,
		NodeType:   chunk.NodeType,
		NodeLabel:  chunk.NodeLabel,
		NodeStatus: chunk.NodeStatus,
		ShowMsg:    chunk.ShowMsg,
		ToolName:   chunk.ToolName,
		ToolCallId: chunk.ToolCallId,
		ToolArgs:   truncateForDisplay(chunk.ToolParams, toolArgsMaxLen),
		ToolStatus: chunk.ToolStatus,
		ToolResult: truncateForDisplay(chunk.ToolResult, toolResultMaxLen),
	}
}

// workflowErrNoisePrefix 匹配错误链中逐层包装的内部前缀（eino 节点错误、LLM 调用、重试），
// 用于兜底分支从完整错误串中剥离噪音、提取根因
var workflowErrNoisePrefix = regexp.MustCompile(`^(?:\[(?:NodeRun|GraphRun)Error\] |LLM 调用失败: |执行失败（已重试 \d+ 次）: |执行失败: )`)

// friendlyWorkflowErrorMessage 将工作流底层错误转换为用户可理解的提示信息：
// 提取根因（API Key 无效、限流、超时、网络异常等）并给出操作指引，
// 不再把内部包装错误链（[NodeRunError] / node path 等）直接展示给用户
func friendlyWorkflowErrorMessage(err error) string {
	if err == nil {
		return "应用执行失败：执行过程中出现未知错误，请稍后重试。"
	}

	// 1. OpenAI 兼容 API 错误：按状态码给出明确原因与操作指引
	var apiErr *einoopenai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case http.StatusUnauthorized:
			return "应用执行失败：模型服务的 API Key 无效或已过期，请在「渠道管理」中检查并更新对应渠道的 API Key 后重试。"
		case http.StatusForbidden:
			return "应用执行失败：当前账号没有访问该模型的权限，请检查渠道的模型映射与账号权限。"
		case http.StatusTooManyRequests:
			return "应用执行失败：请求过于频繁（已触发限流），请稍后重试。"
		case http.StatusNotFound:
			return "应用执行失败：请求的模型不存在或服务地址有误，请检查渠道的模型配置。"
		}
		if apiErr.HTTPStatusCode >= 500 {
			return "应用执行失败：模型服务暂时不可用，请稍后重试。"
		}
		if msg := strings.TrimSpace(apiErr.Message); msg != "" {
			if apiErr.HTTPStatusCode > 0 {
				return fmt.Sprintf("应用执行失败：%s（错误码 %d）。请检查模型或渠道配置后重试。",
					truncateForDisplay(msg, 120), apiErr.HTTPStatusCode)
			}
			return fmt.Sprintf("应用执行失败：%s。请检查模型或渠道配置后重试。", truncateForDisplay(msg, 120))
		}
		return "应用执行失败：模型服务返回错误，请稍后重试。"
	}

	// 2. 超时
	if errors.Is(err, context.DeadlineExceeded) {
		return "应用执行失败：模型响应超时，请稍后重试或更换其他模型。"
	}
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "Client.Timeout") {
		return "应用执行失败：模型响应超时，请稍后重试或更换其他模型。"
	}

	// 3. 网络类错误
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection reset") || strings.Contains(msg, "TLS handshake") {
		return "应用执行失败：无法连接模型服务，请检查网络或渠道的服务地址配置。"
	}

	// 4. 兜底：剥离内部包装噪音，取最内层可读的错误消息
	if root := rootWorkflowError(err); root != "" {
		return fmt.Sprintf("应用执行失败：%s", root)
	}
	return "应用执行失败：执行过程中出现未知错误，请稍后重试。"
}

// rootWorkflowError 剥离 eino 节点错误包装（node path 尾巴）与逐层前缀，
// 提取最内层可读的错误消息，避免内部噪音直接暴露给用户
func rootWorkflowError(err error) string {
	s := err.Error()
	// 去掉 eino 的 node path 尾巴
	if idx := strings.Index(s, "\n------------------------\n"); idx >= 0 {
		s = s[:idx]
	}
	// 反复剥离内部包装前缀
	for {
		loc := workflowErrNoisePrefix.FindString(s)
		if loc == "" {
			break
		}
		s = strings.TrimPrefix(s, loc)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return truncateForDisplay(s, 200)
}

// HandleWorkflowChat 处理 AI 对话中的工作流执行
func HandleWorkflowChat(ctx context.Context, conn *utils.Connection, conversation *domain.Conversation, db *gorm.DB, msg *dto.WsMessageRequest, resProvider iface.ResourceProvider) {
	workflowID := *msg.WorkflowID

	// 1. 查询工作流（先查询以获取应用名称）
	flow, err := workflowservice.Get(workflowID, db)
	if err != nil {
		log.Error("获取工作流失败", zap.Int64("workflowId", workflowID), zap.Error(err))
		conn.Send(dto.WsMessageResponse{
			Content:        "应用不存在或已下架",
			End:            true,
			ConversationId: conversation.Id,
		})
		return
	}

	if flow.Status != "published" {
		conn.Send(dto.WsMessageResponse{
			Content:        "应用未发布，暂时无法使用",
			End:            true,
			ConversationId: conversation.Id,
		})
		return
	}

	// 2. 保存用户消息到会话（包含应用名称和文件信息）
	if err := conversation.HandleWorkflowMessage(msg, flow.Name, db); err != nil {
		log.Error("保存用户消息失败", zap.Error(err))
		conn.Send(dto.WsMessageResponse{
			Content:        "消息保存失败",
			End:            true,
			ConversationId: conversation.Id,
		})
		return
	}

	// 3. 解析拓扑获取 inputSchema
	var topo types.Topology
	json.Unmarshal([]byte(flow.Topology), &topo)

	// 4. 查询最近一次执行记录（用于迭代）
	lastExecution := findLastExecution(db, conversation.Id, workflowID)

	// 5. 构建工作流输入
	content := buildWorkflowInput(msg, lastExecution, topo.Config)

	// 6. 构建文件引用
	fileRefs := buildFileRefs(msg, lastExecution)

	// 7. 创建模型解析器
	modelResolver := resolver.NewChannelModelResolver(db)

	// 8. 初始化 WorkflowAgent
	workflowAgent := workflow.NewWorkflowAgent(resProvider, flow.Topology, modelResolver)

	// 9. 获取默认模型
	defaultModel := "deepseek-v3"
	if topo.Config != nil && topo.Config.DefaultModel != "" {
		defaultModel = topo.Config.DefaultModel
	}

	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}
	channel, mappingModel, err := channelservice.ChooseChannelAndModel(db, defaultModel, mappingParams)
	if err != nil {
		log.Error("选择渠道失败", zap.Error(err))
		conn.Send(dto.WsMessageResponse{
			Content:        "模型服务不可用，请稍后重试",
			End:            true,
			ConversationId: conversation.Id,
		})
		return
	}

	// 10. 启动工作流会话（与 WS 连接解耦：客户端断开后工作流继续执行，支持刷新后 resume）。
	// 传入 userId：工具保存文件时按用户目录落盘，与下载接口路径一致
	uid, _ := utils.GetUIDFromContextAllowEmpty(ctx)
	ws, wctx := workflowStreamManager.Start(conversation.Id, uid)
	if err := ws.Attach(conn); err != nil {
		log.Error("attach workflow stream consumer failed", zap.Error(err))
	}

	// 11. 收集输出、产物和节点日志
	var outputBuilder strings.Builder
	var artifacts []dto.ArtifactInfo
	var nodeLogs []NodeLog
	nodeLogMap := make(map[string]int) // nodeId -> index in nodeLogs
	var workflowErrMsg string          // 执行失败时的原始错误详情（持久化后供前端"查看详情"）

	// 工作流异常结束时兜底收尾会话，避免残留"执行中"状态
	defer func() {
		if r := recover(); r != nil {
			log.Error("workflow chat panic", zap.Any("err", r))
			ws.finish(fmt.Errorf("workflow chat panic: %v", r), "", artifacts, fmt.Sprint(r))
		}
	}()

	// 12. 执行工作流
	callback := func(chunk *global.Chunk) error {
		// 构建工作流进度消息（与写协程共用同一构建逻辑）
		progress := progressFromChunk(chunk)

		// 收集输出内容（resume 快照与最终正文）
		if chunk.Content != "" {
			outputBuilder.WriteString(chunk.Content)
			ws.appendContent(chunk.Content)
		}
		// 累积进度（resume 时回放给前端重建节点/工具展示）
		ws.appendProgress(*progress)

		// 收集产物（URL 携带实际保存目录相对路径，如 2/2026-08-20/xxx.pdf，
		// 下载端按此定位，避免跨天后按"今天"猜目录 404）
		if chunk.ToolResult != "" && chunk.ToolName != "" {
			if artifact := extractArtifactFromChunk(chunk, tool.SaveDirRelPath(wctx)); artifact != nil {
				artifacts = append(artifacts, *artifact)
			}
		}

		// 收集节点日志（与前端 WorkflowMessage watcher 逻辑一致）
		if chunk.NodeId != "" {
			if idx, exists := nodeLogMap[chunk.NodeId]; exists {
				// 更新已有节点
				existing := &nodeLogs[idx]
				if chunk.NodeStatus != "" {
					existing.Status = chunk.NodeStatus
				}
				if chunk.ToolName != "" {
					// 优先按 ToolCallId 精确匹配（同轮同名多次调用不互相覆盖）
					matched := -1
					if chunk.ToolCallId != "" {
						for i := range existing.ToolCalls {
							if existing.ToolCalls[i].ID == chunk.ToolCallId {
								matched = i
								break
							}
						}
					} else if chunk.ToolStatus != "running" {
						// 无 ToolCallId 兜底：结果类 chunk 回填最近一条"同名且仍在执行中"的记录；
						// running 视为新调用起点，直接追加新记录，避免同名多次调用被合并
						for i := len(existing.ToolCalls) - 1; i >= 0; i-- {
							if existing.ToolCalls[i].Name == chunk.ToolName &&
								existing.ToolCalls[i].Status == "running" {
								matched = i
								break
							}
						}
					}

					if matched >= 0 {
						tc := &existing.ToolCalls[matched]
						if chunk.ToolStatus != "" {
							tc.Status = chunk.ToolStatus
						}
						if chunk.ToolParams != "" {
							tc.Args = truncateForDisplay(chunk.ToolParams, toolArgsMaxLen)
						}
						if chunk.ToolResult != "" {
							tc.Result = truncateForDisplay(chunk.ToolResult, toolResultMaxLen)
						}
					} else {
						// 新工具调用：追加
						existing.ToolCalls = append(existing.ToolCalls, ToolCall{
							ID:     chunk.ToolCallId,
							Name:   chunk.ToolName,
							Status: chunk.ToolStatus,
							Args:   truncateForDisplay(chunk.ToolParams, toolArgsMaxLen),
							Result: truncateForDisplay(chunk.ToolResult, toolResultMaxLen),
						})
					}
				}
				// 节点完成时，将所有未完成的工具调用标记为完成
				if chunk.NodeStatus == "completed" || chunk.NodeStatus == "failed" {
					for i := range existing.ToolCalls {
						if existing.ToolCalls[i].Status == "running" {
							existing.ToolCalls[i].Status = chunk.NodeStatus
						}
					}
				}
			} else {
				// 新增节点
				var toolCalls []ToolCall
				if chunk.ToolName != "" {
					toolCalls = append(toolCalls, ToolCall{
						ID:     chunk.ToolCallId,
						Name:   chunk.ToolName,
						Status: chunk.ToolStatus,
						Args:   truncateForDisplay(chunk.ToolParams, toolArgsMaxLen),
						Result: truncateForDisplay(chunk.ToolResult, toolResultMaxLen),
					})
				}
				nodeLogMap[chunk.NodeId] = len(nodeLogs)
				nodeLogs = append(nodeLogs, NodeLog{
					NodeId:    chunk.NodeId,
					Type:      chunk.NodeType,
					Label:     chunk.NodeLabel,
					Status:    chunk.NodeStatus,
					ToolCalls: toolCalls,
				})
			}
		}

		// 广播增量给所有在线消费者（由写协程下发，保证同一连接单写者；
		// 进度信息放在 Workflow 字段，不发送 Content 避免前端显示为纯文本）
		ws.broadcast(partialChunk{Chunk: chunk, End: false})
		return nil
	}

	// 12. 执行工作流（使用解耦上下文：连接断开不影响执行，直到完成或被取消）
	_, execErr := workflowAgent.ExecuteStream(wctx, channel.GetEndpoint(), channel.GetRandomSecret(), mappingModel, content, "", callback)

	// 13. 处理执行结果
	output := outputBuilder.String()

	// 只保留最终交付产物：本次执行已生成 PDF 时，剔除中间过程的 Markdown 分片文件
	// （如 zhanjiang_1day_part1.md 等），避免把过程文件当产物展示给用户
	artifacts = filterFinalArtifacts(artifacts)

	if execErr != nil {
		log.Error("工作流执行失败", zap.Error(execErr))
		if isWorkflowCanceled(execErr) {
			// 用户点击停止生成：把已生成的部分内容 + "已中断"终态下发给前端，
			// 并保存到会话消息（刷新后仍可见），界面不再停留在"执行中"
			if output == "" {
				output = "已中断生成"
			}
			ws.finish(execErr, output, artifacts, "")
			saveWorkflowResponse(db, conversation, output, workflowStatusInterrupted, artifacts, flow.Name, nodeLogs, "")
			return
		}
		// 用户看到的是提取根因后的友好提示，原始错误链放入 Workflow.Error 供"查看详情"
		output = friendlyWorkflowErrorMessage(execErr)
		workflowErrMsg = execErr.Error()
	}

	// 通知消费者发送最终状态（Content/Workflow/Artifacts，由各连接写协程下发）
	ws.finish(execErr, output, artifacts, workflowErrMsg)

	// 14. 保存执行记录
	saveExecution(db, conversation.Id, workflowID, msg, fileRefs, output, artifacts, execErr == nil)

	// 15. 保存 AI 响应到会话（包含工作流状态、原始错误、产物信息和节点日志）
	if output != "" {
		workflowStatus := "completed"
		if execErr != nil {
			workflowStatus = "failed"
		}
		saveWorkflowResponse(db, conversation, output, workflowStatus, artifacts, flow.Name, nodeLogs, workflowErrMsg)
	}
}

// findLastExecution 查询会话中某工作流的最近一次执行记录
func findLastExecution(db *gorm.DB, conversationID, workflowID int64) *domain.WorkflowExecution {
	var execution domain.WorkflowExecution
	err := db.Where("conversation_id = ? AND workflow_id = ? AND status = ?",
		conversationID, workflowID, "completed").
		Order("create_time DESC").
		First(&execution).Error
	if err != nil {
		return nil
	}
	return &execution
}

// buildWorkflowInput 构建工作流输入内容
func buildWorkflowInput(msg *dto.WsMessageRequest, lastExecution *domain.WorkflowExecution, config *types.WorkflowConfig) string {
	if config == nil || len(config.InputSchema) == 0 {
		// 无 inputSchema：直接使用用户消息
		if lastExecution != nil {
			// 迭代模式：注入上次结果
			return fmt.Sprintf("【上次执行结果】\n%s\n\n【修改要求】\n%s", lastExecution.Output, msg.Content)
		}
		return msg.Content
	}

	// 有 inputSchema：按字段构建
	inputs := make(map[string]string)
	if lastExecution != nil {
		json.Unmarshal([]byte(lastExecution.Inputs), &inputs)
	}

	// 构建文件引用映射
	fileRefMap := make(map[string]string)
	for _, f := range msg.Files {
		fileRefMap[f.FieldName] = f.FileURL
	}

	isFirstText := true
	parts := make([]string, 0)

	for _, field := range config.InputSchema {
		if field.Type == "file" {
			// 文件字段：读取文件内容并添加
			fileURL := fileRefMap[field.Name]
			if fileURL == "" && lastExecution != nil {
				// 迭代时复用上次的文件
				var refs map[string]string
				json.Unmarshal([]byte(lastExecution.FileRefs), &refs)
				fileURL = refs[field.Name]
			}
			if fileURL != "" {
				fileContent := readWorkflowFileContent(fileURL)
				if fileContent != "" {
					parts = append(parts, field.Label+"：\n"+fileContent)
				}
			}
			continue
		}

		var value string
		if isFirstText {
			// 第一个文本字段
			if lastExecution != nil {
				value = fmt.Sprintf("【上次执行结果】\n%s\n\n【修改要求】\n%s", lastExecution.Output, msg.Content)
			} else {
				value = msg.Content
			}
			isFirstText = false
		} else {
			// 其他文本字段
			if lastExecution != nil {
				value = inputs[field.Name]
			} else {
				value = field.Default
			}
		}

		if value != "" {
			parts = append(parts, field.Label+"："+value)
		}
	}

	return strings.Join(parts, "\n\n")
}

// readWorkflowFileContent 从文件 URL 读取文件内容
// URL 格式: /api/file/download?filePath=xxx 或直接是相对路径如 2/2026-06-07/xxx.pdf
func readWorkflowFileContent(fileURL string) string {
	// 提取文件路径
	filePath := ""
	if strings.Contains(fileURL, "filePath=") {
		idx := strings.Index(fileURL, "filePath=")
		filePath = fileURL[idx+9:]
		// URL 解码
		filePath = strings.ReplaceAll(filePath, "%2F", "/")
		filePath = strings.ReplaceAll(filePath, "%5C", "\\")
		filePath = strings.ReplaceAll(filePath, "+", " ")
	} else if strings.HasPrefix(fileURL, "/api/file/") {
		filePath = strings.TrimPrefix(fileURL, "/api/file/")
	} else {
		// 直接是相对路径，如 2/2026-06-07/xxx.pdf
		filePath = fileURL
	}

	if filePath == "" {
		return ""
	}

	// 构建完整的文件路径（加上 uploads 目录前缀）
	config := global.LoadConfig().LocalUploadConfig
	fullPath := filePath
	if config.Dir != "" {
		// 获取当前工作目录
		currentDir, err := os.Getwd()
		if err != nil {
			log.Error("获取工作目录失败", zap.Error(err))
			return ""
		}
		fullPath = filepath.Join(currentDir, config.Dir, filePath)
	}

	// 根据文件类型提取内容
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".pdf":
		// PDF 文件使用专门的提取工具
		text, err := tool.ReadPdfText(nil, &tool.PdfReadParams{FilePath: fullPath})
		if err != nil {
			log.Error("PDF 文本提取失败", zap.String("path", fullPath), zap.Error(err))
			return ""
		}
		return text
	case ".txt", ".md":
		// 文本文件直接读取
		data, err := utils.ReadFileContent(fullPath)
		if err != nil {
			log.Error("读取工作流文件失败", zap.String("path", fullPath), zap.Error(err))
			return ""
		}
		return data
	default:
		log.Warn("不支持的文件类型", zap.String("ext", ext))
		return ""
	}
}

// buildFileRefs 构建文件引用
func buildFileRefs(msg *dto.WsMessageRequest, lastExecution *domain.WorkflowExecution) map[string]string {
	refs := make(map[string]string)

	// 从上次执行记录中恢复文件引用
	if lastExecution != nil && lastExecution.FileRefs != "" {
		json.Unmarshal([]byte(lastExecution.FileRefs), &refs)
	}

	// 用新上传的文件覆盖
	for _, f := range msg.Files {
		refs[f.FieldName] = f.FileURL
	}

	return refs
}

// extractArtifactFromChunk 从 chunk 中提取产物信息。
// relPath 是工具实际保存目录相对上传根目录的路径（如 "2/2026-08-20" 或 "2026-08-20"），
// 拼进下载 URL 后，下载端可按此路径精确定位文件，避免仅凭文件名 + "今天" 猜日期目录
// 导致跨天后下载 404（文件保存在执行当天目录，下载可能在次日发生）
func extractArtifactFromChunk(chunk *global.Chunk, relPath string) *dto.ArtifactInfo {
	fileGenToolPrefixes := map[string]string{
		"markdown_save_tool":        "Markdown文件已成功保存到: ./",
		"markdown_to_pdf_file_tool": "PDF已成功保存: ./",
		"image_download_tool":       "下载完成: ./",
	}
	fileGenToolCategories := map[string]string{
		"markdown_save_tool":        "markdown",
		"markdown_to_pdf_file_tool": "pdf",
		"image_download_tool":       "image",
	}

	prefix, ok := fileGenToolPrefixes[chunk.ToolName]
	if !ok {
		return nil
	}

	idx := strings.Index(chunk.ToolResult, prefix)
	if idx < 0 {
		return nil
	}
	start := idx + len(prefix)
	if start >= len(chunk.ToolResult) {
		return nil
	}

	end := strings.IndexAny(chunk.ToolResult[start:], "\n\r ")
	var fileName string
	if end < 0 {
		fileName = strings.TrimSpace(chunk.ToolResult[start:])
	} else {
		fileName = strings.TrimSpace(chunk.ToolResult[start : start+end])
	}

	if fileName == "" {
		return nil
	}

	// 下载 URL 携带相对目录（如 2/2026-08-20/xxx.pdf），
	// 下载端按此路径解析，不依赖"今天"猜测
	fileRelPath := fileName
	if relPath != "" {
		fileRelPath = filepath.ToSlash(filepath.Join(relPath, fileName))
	}

	return &dto.ArtifactInfo{
		Name:     fileName,
		URL:      fmt.Sprintf("/api/file/download?filePath=%s", url.QueryEscape(fileRelPath)),
		Category: fileGenToolCategories[chunk.ToolName],
	}
}

// filterFinalArtifacts 只保留最终交付产物。
// 常见产文件流程是 Agent 先用 markdown_save_tool 分片保存中间 Markdown
// （输出过长时按部分落盘），再用 markdown_to_pdf_file_tool 合并生成最终 PDF；
// 此时 Markdown 分片只是过程文件，不应作为"生成的文件产物"展示给用户。
// 规则：本次执行已生成 PDF 时，剔除 markdown 类产物；未生成 PDF 则全部保留
// （纯 Markdown 交付类应用不受影响）。
func filterFinalArtifacts(artifacts []dto.ArtifactInfo) []dto.ArtifactInfo {
	hasPDF := false
	for _, a := range artifacts {
		if a.Category == "pdf" {
			hasPDF = true
			break
		}
	}
	if !hasPDF {
		return artifacts
	}

	filtered := make([]dto.ArtifactInfo, 0, len(artifacts))
	for _, a := range artifacts {
		if a.Category != "markdown" {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// saveExecution 保存工作流执行记录
func saveExecution(db *gorm.DB, conversationID, workflowID int64, msg *dto.WsMessageRequest, fileRefs map[string]string, output string, artifacts []dto.ArtifactInfo, success bool) {
	inputsJSON, _ := json.Marshal(map[string]string{"content": msg.Content})
	fileRefsJSON, _ := json.Marshal(fileRefs)
	artifactsJSON, _ := json.Marshal(artifacts)

	status := "completed"
	if !success {
		status = "failed"
	}

	execution := &domain.WorkflowExecution{
		ConversationID: conversationID,
		WorkflowID:     workflowID,
		Inputs:         string(inputsJSON),
		FileRefs:       string(fileRefsJSON),
		Output:         output,
		Artifacts:      string(artifactsJSON),
		Status:         status,
	}

	if err := db.Create(execution).Error; err != nil {
		log.Error("保存工作流执行记录失败", zap.Error(err))
	}
}

// saveWorkflowResponse 保存工作流响应到会话消息（包含工作流状态、原始错误、产物信息和节点日志）
func saveWorkflowResponse(db *gorm.DB, conversation *domain.Conversation, content string, workflowStatus string, artifacts []dto.ArtifactInfo, appName string, nodeLogs []NodeLog, workflowError string) {
	artifactsJSON, _ := json.Marshal(artifacts)
	executionLogsJSON, _ := json.Marshal(nodeLogs)
	assistantMsg := global.Message{
		Role:           global.Assistant,
		Content:        content,
		WorkflowStatus: workflowStatus,
		WorkflowError:  workflowError,
		Artifacts:      string(artifactsJSON),
		AppName:        appName,
		ExecutionLogs:  string(executionLogsJSON),
	}
	conversation.FormattedMessage = append(conversation.FormattedMessage, assistantMsg)

	msgJSON, err := json.Marshal(conversation.FormattedMessage)
	if err != nil {
		log.Error("序列化消息失败", zap.Error(err))
		return
	}

	if err := db.Model(conversation).Update("message", string(msgJSON)).Error; err != nil {
		log.Error("保存工作流响应消息失败", zap.Error(err))
	}
}
