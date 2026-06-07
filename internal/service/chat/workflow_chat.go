package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"txing-ai/internal/agent"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
	channelservice "txing-ai/internal/service/channel"
	workflowservice "txing-ai/internal/service/workflow"
	"txing-ai/internal/tool"
	"txing-ai/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HandleWorkflowChat 处理 AI 对话中的工作流执行
func HandleWorkflowChat(ctx context.Context, conn *utils.Connection, conversation *domain.Conversation, db *gorm.DB, msg *dto.WsMessageRequest, resProvider iface.ResourceProvider) {
	workflowID := *msg.WorkflowID

	// 1. 保存用户消息到会话
	if err := conversation.HandleMessage(msg, db); err != nil {
		log.Error("保存用户消息失败", zap.Error(err))
		conn.Send(dto.WsMessageResponse{
			Content:        "消息保存失败",
			End:            true,
			ConversationId: conversation.Id,
		})
		return
	}

	// 2. 查询工作流
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

	// 3. 解析拓扑获取 inputSchema
	var topo agent.Topology
	json.Unmarshal([]byte(flow.Topology), &topo)

	// 4. 查询最近一次执行记录（用于迭代）
	lastExecution := findLastExecution(db, conversation.Id, workflowID)

	// 5. 构建工作流输入
	content := buildWorkflowInput(msg, lastExecution, topo.Config)

	// 6. 构建文件引用
	fileRefs := buildFileRefs(msg, lastExecution)

	// 7. 创建模型解析器
	modelResolver := agent.NewChannelModelResolver(db)

	// 8. 初始化 WorkflowAgent
	workflowAgent := agent.NewWorkflowAgent(resProvider, flow.Topology, modelResolver)

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

	// 10. 发送执行开始状态
	conn.Send(dto.WsMessageResponse{
		Content:        "",
		End:            false,
		ConversationId: conversation.Id,
		Workflow: &dto.WorkflowProgress{
			Status: "running",
		},
	})

	// 11. 收集输出和产物
	var outputBuilder strings.Builder
	var artifacts []dto.ArtifactInfo

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// 12. 执行工作流
	callback := func(chunk *global.Chunk) error {
		// 构建工作流进度消息
		progress := &dto.WorkflowProgress{
			Status:     "running",
			NodeID:     chunk.NodeId,
			NodeType:   chunk.NodeType,
			NodeLabel:  chunk.NodeLabel,
			NodeStatus: chunk.NodeStatus,
			ShowMsg:    chunk.ShowMsg,
			ToolName:   chunk.ToolName,
			ToolStatus: chunk.ToolStatus,
			ToolResult: chunk.ToolResult,
		}

		// 收集输出内容
		if chunk.Content != "" {
			outputBuilder.WriteString(chunk.Content)
		}

		// 收集产物
		if chunk.ToolResult != "" && chunk.ToolName != "" {
			if artifact := extractArtifactFromChunk(chunk); artifact != nil {
				artifacts = append(artifacts, *artifact)
			}
		}

		// 发送进度到客户端（进度信息放在 Workflow 字段，不发送 Content 避免前端显示为纯文本）
		resp := dto.WsMessageResponse{
			End:            false,
			ConversationId: conversation.Id,
			Workflow:       progress,
		}

		return conn.Send(resp)
	}

	_, execErr := workflowAgent.ExecuteStream(ctxWithCancel, channel.GetEndpoint(), channel.GetRandomSecret(), mappingModel, content, "", callback)

	// 13. 处理执行结果
	output := outputBuilder.String()

	if execErr != nil {
		log.Error("工作流执行失败", zap.Error(execErr))
		conn.Send(dto.WsMessageResponse{
			Content:        fmt.Sprintf("应用执行失败：%s", execErr.Error()),
			End:            true,
			ConversationId: conversation.Id,
			Workflow: &dto.WorkflowProgress{
				Status: "failed",
			},
		})
		output = fmt.Sprintf("执行失败：%s", execErr.Error())
	} else {
		// 发送完成状态（包含格式化的结果文本和产物信息）
		conn.Send(dto.WsMessageResponse{
			Content:        output,
			End:            true,
			ConversationId: conversation.Id,
			Workflow: &dto.WorkflowProgress{
				Status: "completed",
			},
			Artifacts: artifacts,
		})
	}

	// 14. 保存执行记录
	saveExecution(db, conversation.Id, workflowID, msg, fileRefs, output, artifacts, execErr == nil)

	// 15. 保存 AI 响应到会话
	if output != "" {
		saveWorkflowResponse(db, conversation, output)
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
func buildWorkflowInput(msg *dto.WsMessageRequest, lastExecution *domain.WorkflowExecution, config *agent.WorkflowConfig) string {
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

// extractArtifactFromChunk 从 chunk 中提取产物信息
func extractArtifactFromChunk(chunk *global.Chunk) *dto.ArtifactInfo {
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

	return &dto.ArtifactInfo{
		Name:     fileName,
		URL:      fmt.Sprintf("/api/file/download?filePath=%s", fileName),
		Category: fileGenToolCategories[chunk.ToolName],
	}
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

// saveWorkflowResponse 保存工作流响应到会话消息
func saveWorkflowResponse(db *gorm.DB, conversation *domain.Conversation, content string) {
	assistantMsg := global.Message{
		Role:    global.Assistant,
		Content: content,
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
