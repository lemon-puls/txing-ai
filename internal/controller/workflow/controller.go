package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
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
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Create 创建工作流
// @Summary 创建工作流
// @Description 创建一个新的 AgentFlow 工作流
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param data body dto.CreateAgentFlowReq true "创建信息"
// @Success 200 {object} utils.Response
// @Router /api/workflow [post]
func Create(ctx *gin.Context) {
	var req dto.CreateAgentFlowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := workflowservice.Create(req, db); err != nil {
		utils.ErrorWithMsg(ctx, "创建工作流失败", err)
		return
	}

	utils.OkWithMsg(ctx, "创建成功")
}

// Update 更新工作流
// @Summary 更新工作流
// @Description 更新现有的 AgentFlow 工作流
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param data body dto.UpdateAgentFlowReq true "更新信息"
// @Success 200 {object} utils.Response
// @Router /api/workflow/{id} [put]
func Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	var req dto.UpdateAgentFlowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := workflowservice.Update(id, req, db); err != nil {
		utils.ErrorWithMsg(ctx, "更新工作流失败", err)
		return
	}

	utils.OkWithMsg(ctx, "更新成功")
}

// Delete 删除工作流
// @Summary 删除工作流
// @Description 删除现有的 AgentFlow 工作流
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} utils.Response
// @Router /api/workflow/{id} [delete]
func Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := workflowservice.Delete(id, db); err != nil {
		utils.ErrorWithMsg(ctx, "删除工作流失败", err)
		return
	}

	utils.OkWithMsg(ctx, "删除成功")
}

// Get 获取单个工作流
// @Summary 获取单个工作流
// @Description 获取工作流详情
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} utils.Response{data=vo.AgentFlowVO}
// @Router /api/workflow/{id} [get]
func Get(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	flow, err := workflowservice.Get(id, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取工作流失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToAgentFlowVO(*flow))
}

// List 获取工作流列表
// @Summary 获取工作流列表
// @Description 获取工作流分页列表
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param name query string false "工作流名称"
// @Success 200 {object} utils.Response
// @Router /api/workflow [get]
func List(ctx *gin.Context) {
	var req dto.ListAgentFlowReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	pageVo, err := workflowservice.List(req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取列表失败", err)
		return
	}

	// 转换为 VO
	vos := vo.ToAgentFlowVOs(pageVo.Records)
	result := page.Convert(pageVo, vos)

	utils.OkWithData(ctx, result)
}

// GetModels 获取可用模型列表
// @Summary 获取可用模型列表
// @Description 获取工作流可用的模型列表
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/workflow/models [get]
func GetModels(ctx *gin.Context) {
	db := utils.GetDBFromContext[*gorm.DB](ctx)
	var models []domain.Model
	if err := db.Find(&models).Error; err != nil {
		utils.ErrorWithMsg(ctx, "获取模型列表失败", err)
		return
	}

	// 转换为前端需要的格式
	type ModelOption struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}
	var result []ModelOption
	for _, m := range models {
		result = append(result, ModelOption{
			Name:        m.Name,
			DisplayName: m.Name,
		})
	}

	utils.OkWithData(ctx, result)
}

// GetTools 获取可用工具列表
// @Summary 获取可用工具列表
// @Description 获取工作流可用的工具列表
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/workflow/tools [get]
func GetTools(ctx *gin.Context) {
	// 返回预定义的工具列表
	tools := []map[string]string{
		{"name": "web_search_tool", "displayName": "网页搜索"},
		{"name": "markdown_save_tool", "displayName": "保存Markdown"},
		{"name": "markdown_to_pdf_file_tool", "displayName": "转PDF"},
		{"name": "image_download_tool", "displayName": "图片下载"},
		{"name": "image_search_tool", "displayName": "图片搜索"},
		{"name": "web_scraping_tool", "displayName": "网页抓取"},
		{"name": "pdf_read_tool", "displayName": "PDF读取"},
		// Maps 相关工具
		{"name": "maps_geo", "displayName": "地理位置查询"},
		{"name": "maps_text_search", "displayName": "地点文本搜索"},
		{"name": "maps_direction_transit_integrated", "displayName": "公交换乘规划"},
		{"name": "maps_direction_driving", "displayName": "驾车路线规划"},
		{"name": "maps_direction_walking", "displayName": "步行路线规划"},
		{"name": "maps_distance", "displayName": "距离测算"},
		{"name": "maps_search_detail", "displayName": "地点详情查询"},
		{"name": "maps_weather", "displayName": "天气预报查询"},
		{"name": "maps_around_search", "displayName": "周边搜索"},
	}
	utils.OkWithData(ctx, tools)
}

// Run 执行工作流
// @Summary 执行工作流
// @Description 动态执行工作流，支持流式输出和文件上传
// @Tags 工作流管理
// @Accept multipart/form-data
// @Produce text/event-stream
// @Param id path int true "工作流ID"
// @Param content formData string false "请求内容"
// @Param file formData file false "上传文件（支持 PDF、TXT、MD）"
// @Success 200 {string} string "SSE stream"
// @Router /api/workflow/{id}/run [post]
func Run(ctx *gin.Context, resProvider iface.ResourceProvider) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	flow, err := workflowservice.Get(id, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取工作流失败", err)
		return
	}

	// 解析拓扑获取 inputSchema
	var topo agent.Topology
	json.Unmarshal([]byte(flow.Topology), &topo)

	content := ""

	// 根据 inputSchema 构建结构化输入
	if topo.Config != nil && len(topo.Config.InputSchema) > 0 {
		var parts []string
		for _, field := range topo.Config.InputSchema {
			if field.Type == "file" {
				// 文件字段：处理上传的文件
				file, header, fileErr := ctx.Request.FormFile(field.Name)
				if fileErr == nil && file != nil {
					defer file.Close()
					savePath, _, saveErr := utils.SaveUploadedFile(file, header.Filename, 0, "workflow_uploads", "")
					if saveErr != nil {
						log.Error("工作流上传文件保存失败", zap.String("field", field.Name), zap.Error(saveErr))
						continue
					}
					fileContent := extractFileContent(ctx, savePath, header.Filename)
					if fileContent != "" {
						parts = append(parts, field.Label+"：\n"+fileContent)
					}
				}
			} else {
				// 文本字段：从表单获取值
				value := ctx.PostForm(field.Name)
				if value == "" {
					value = field.Default
				}
				if value != "" {
					parts = append(parts, field.Label+"："+value)
				}
			}
		}
		content = strings.Join(parts, "\n\n")
	} else {
		// 无 inputSchema 时，使用默认的 content 字段（兼容旧逻辑）
		content = ctx.PostForm("content")
		if content == "" {
			var req struct {
				Content string `json:"content"`
			}
			if ctx.ShouldBindJSON(&req) == nil {
				content = req.Content
			}
		}

		// 处理默认文件上传
		file, header, fileErr := ctx.Request.FormFile("file")
		if fileErr == nil && file != nil {
			defer file.Close()
			savePath, _, saveErr := utils.SaveUploadedFile(file, header.Filename, 0, "workflow_uploads", "")
			if saveErr != nil {
				log.Error("工作流上传文件保存失败", zap.Error(saveErr))
			} else {
				fileContent := extractFileContent(ctx, savePath, header.Filename)
				if fileContent != "" {
					if content != "" {
						content = content + "\n\n文件内容：\n\n" + fileContent
					} else {
						content = "文件内容：\n\n" + fileContent
					}
				}
			}
		}
	}

	if content == "" {
		utils.ErrorWithMsg(ctx, "请输入内容", nil)
		return
	}

	// 创建模型解析器
	modelResolver := agent.NewChannelModelResolver(db)

	// 初始化 WorkflowAgent
	workflowAgent := agent.NewWorkflowAgent(resProvider, flow.Topology, modelResolver)

	// 获取默认模型（优先使用拓扑配置中的模型，否则使用系统默认）
	defaultModel := "deepseek-v3"
	if topo.Config != nil && topo.Config.DefaultModel != "" {
		defaultModel = topo.Config.DefaultModel
	}

	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}
	channel, mappingModel, err := channelservice.ChooseChannelAndModel(db, defaultModel, mappingParams)
	if err != nil {
		utils.ErrorWithMsg(ctx, "选择渠道失败", err)
		return
	}

	// 设置 SSE 响应头
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// 文件产物收集（工具节点生成的文件）
	var artifacts []map[string]string
	// 文件生成工具及其结果前缀
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

	callback := func(chunk *global.Chunk) error {
		data := map[string]interface{}{
			"content":          chunk.Content,
			"reasoningContent": chunk.ReasoningContent,
			"toolCallId":       chunk.ToolCallId,
			"toolName":         chunk.ToolName,
			"toolParams":       chunk.ToolParams,
			"toolResult":       chunk.ToolResult,
			"showMsg":          chunk.ShowMsg,
			"nodeId":           chunk.NodeId,
			"nodeType":         chunk.NodeType,
			"nodeLabel":        chunk.NodeLabel,
			"nodeStatus":       chunk.NodeStatus,
			"end":              false,
		}

		// 添加执行日志信息
		if chunk.ExecutionLog != nil {
			data["execution_log"] = chunk.ExecutionLog
		}

		// 检测文件生成工具的结果，收集产物信息
		if chunk.ToolResult != "" && chunk.ToolName != "" {
			if prefix, ok := fileGenToolPrefixes[chunk.ToolName]; ok {
				if fileName := extractArtifactFileName(chunk.ToolResult, prefix); fileName != "" {
					downloadURL := fmt.Sprintf("/api/file/download?filePath=%s", url.QueryEscape(fileName))
					artifacts = append(artifacts, map[string]string{
						"name":     fileName,
						"url":      downloadURL,
						"category": fileGenToolCategories[chunk.ToolName],
					})
				}
			}
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			cancel()
			return err
		}

		_, err = fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
		if err != nil {
			cancel()
			return err
		}
		ctx.Writer.Flush()
		return nil
	}

	// 执行
	_, err = workflowAgent.ExecuteStream(ctxWithCancel, channel.GetEndpoint(), channel.GetRandomSecret(), mappingModel, content, "", callback)
	if err != nil {
		errData := map[string]interface{}{
			"error": err.Error(),
			"end":   true,
		}
		jsonData, _ := json.Marshal(errData)
		_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
		ctx.Writer.Flush()
		return
	}

	// 发送结束标志（附带文件产物信息）
	endData := map[string]interface{}{
		"end": true,
	}
	if len(artifacts) > 0 {
		endData["artifacts"] = artifacts
	}
	endJsonData, _ := json.Marshal(endData)
	_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", endJsonData)
	ctx.Writer.Flush()
}

// extractFileContent 根据文件类型提取文本内容
func extractFileContent(ctx *gin.Context, filePath, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		text, err := tool.ReadPdfText(ctx, &tool.PdfReadParams{FilePath: filePath})
		if err != nil {
			log.Error("PDF 文本提取失败", zap.Error(err))
			return ""
		}
		return text
	case ".txt", ".md":
		data, err := utils.ReadFileContent(filePath)
		if err != nil {
			log.Error("文本文件读取失败", zap.Error(err))
			return ""
		}
		return data
	default:
		log.Warn("不支持的文件类型", zap.String("ext", ext))
		return ""
	}
}

// ValidateTopology 校验工作流拓扑（结构校验，无需保存）
// @Summary 校验工作流拓扑
// @Description 对传入的工作流拓扑 JSON 进行结构合法性校验
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param data body dto.ValidateWorkflowReq true "拓扑数据"
// @Success 200 {object} utils.Response{data=vo.ValidationResultVO}
// @Router /api/workflow/validate [post]
func ValidateTopology(ctx *gin.Context) {
	var req dto.ValidateWorkflowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	if req.Topology == "" {
		utils.ErrorWithMsg(ctx, "拓扑数据不能为空", nil)
		return
	}

	// 结构校验
	result := agent.ValidateTopology(req.Topology)

	// 转换为 VO
	utils.OkWithData(ctx, toValidationResultVO(result))
}

// ValidateById 校验已保存的工作流（支持 LLM 语义校验）
// @Summary 校验已保存的工作流
// @Description 校验指定 ID 的工作流，可选启用 LLM 语义校验
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param data body dto.ValidateWorkflowReq true "校验选项"
// @Success 200 {object} utils.Response{data=vo.ValidationResultVO}
// @Router /api/workflow/{id}/validate [post]
func ValidateById(ctx *gin.Context, resProvider iface.ResourceProvider) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	var req dto.ValidateWorkflowReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 允许空 body，默认不使用 LLM
		req.UseLLM = false
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	flow, err := workflowservice.Get(id, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取工作流失败", err)
		return
	}

	if flow.Topology == "" {
		utils.ErrorWithMsg(ctx, "工作流拓扑数据为空", nil)
		return
	}

	// 第一层：结构校验
	structResult := agent.ValidateTopology(flow.Topology)

	// 如果不需要 LLM 校验，直接返回结构校验结果
	if !req.UseLLM {
		utils.OkWithData(ctx, toValidationResultVO(structResult))
		return
	}

	// 第二层：LLM 语义校验
	// 获取模型配置
	model := "deepseek-v3"
	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}
	channel, mappingModel, err := channelservice.ChooseChannelAndModel(db, model, mappingParams)
	if err != nil {
		// LLM 校验失败不影响结构校验结果，返回结构校验结果 + warning
		structResult.Warnings = append(structResult.Warnings, agent.ValidationError{
			Level:   "warning",
			Code:    "LLM_VALIDATION_UNAVAILABLE",
			Message: fmt.Sprintf("LLM 校验不可用: %v，已返回结构校验结果", err),
		})
		utils.OkWithData(ctx, toValidationResultVO(structResult))
		return
	}

	llmResult, err := agent.ValidateTopologyWithLLM(
		ctx, channel.GetEndpoint(), channel.GetRandomSecret(), mappingModel, flow.Topology,
	)
	if err != nil {
		structResult.Warnings = append(structResult.Warnings, agent.ValidationError{
			Level:   "warning",
			Code:    "LLM_VALIDATION_FAILED",
			Message: fmt.Sprintf("LLM 校验执行失败: %v，已返回结构校验结果", err),
		})
		utils.OkWithData(ctx, toValidationResultVO(structResult))
		return
	}

	// 合并结构校验和 LLM 校验结果
	merged := mergeValidationResults(structResult, llmResult)
	utils.OkWithData(ctx, toValidationResultVO(merged))
}

// toValidationResultVO 将 agent.ValidationResult 转换为 vo.ValidationResultVO
func toValidationResultVO(result *agent.ValidationResult) vo.ValidationResultVO {
	voResult := vo.ValidationResultVO{
		Valid: result.Valid,
	}

	for _, e := range result.Errors {
		voResult.Errors = append(voResult.Errors, vo.ValidationErrorVO{
			Level:   string(e.Level),
			NodeID:  e.NodeID,
			Code:    e.Code,
			Message: e.Message,
		})
	}

	for _, w := range result.Warnings {
		voResult.Warnings = append(voResult.Warnings, vo.ValidationErrorVO{
			Level:   string(w.Level),
			NodeID:  w.NodeID,
			Code:    w.Code,
			Message: w.Message,
		})
	}

	return voResult
}

// mergeValidationResults 合并两个校验结果
func mergeValidationResults(a, b *agent.ValidationResult) *agent.ValidationResult {
	merged := &agent.ValidationResult{
		Valid: a.Valid && b.Valid,
	}
	merged.Errors = append(a.Errors, b.Errors...)
	merged.Warnings = append(a.Warnings, b.Warnings...)
	return merged
}

// extractArtifactFileName 从工具执行结果中提取文件名
// 工具结果格式示例: "Markdown文件已成功保存到: ./filename.md"
func extractArtifactFileName(toolResult string, prefix string) string {
	idx := strings.Index(toolResult, prefix)
	if idx < 0 {
		return ""
	}
	start := idx + len(prefix)
	if start >= len(toolResult) {
		return ""
	}
	// 截取到行尾或字符串末尾
	end := strings.IndexAny(toolResult[start:], "\n\r ")
	if end < 0 {
		return strings.TrimSpace(toolResult[start:])
	}
	return strings.TrimSpace(toolResult[start : start+end])
}

// ==================== 版本管理 ====================

// CreateVersion 创建工作流版本
// @Summary 创建工作流版本
// @Description 为指定工作流创建新版本
// @Tags 工作流版本管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param data body dto.CreateVersionReq true "版本信息"
// @Success 200 {object} utils.Response{data=vo.AgentFlowVersionVO}
// @Router /api/workflow/{id}/versions [post]
func CreateVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	var req dto.CreateVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	version, err := workflowservice.CreateVersion(id, req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "创建版本失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToAgentFlowVersionVO(*version))
}

// ListVersions 获取版本列表
// @Summary 获取版本列表
// @Description 获取工作流的版本历史
// @Tags 工作流版本管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} utils.Response
// @Router /api/workflow/{id}/versions [get]
func ListVersions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	var req dto.ListVersionReq
	req.FlowID = id
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	pageVo, err := workflowservice.ListVersions(req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取版本列表失败", err)
		return
	}

	vos := vo.ToAgentFlowVersionVOs(pageVo.Records)
	result := page.Convert(pageVo, vos)
	utils.OkWithData(ctx, result)
}

// GetVersion 获取指定版本
// @Summary 获取指定版本
// @Description 获取工作流的指定版本详情
// @Tags 工作流版本管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param version path int true "版本号"
// @Success 200 {object} utils.Response{data=vo.AgentFlowVersionVO}
// @Router /api/workflow/{id}/versions/{version} [get]
func GetVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	versionStr := ctx.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		utils.ErrorWithMsg(ctx, "版本号错误", err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	v, err := workflowservice.GetVersion(id, version, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取版本失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToAgentFlowVersionVO(*v))
}

// PublishVersion 发布版本
// @Summary 发布版本
// @Description 发布工作流的指定版本
// @Tags 工作流版本管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param data body dto.PublishVersionReq true "发布信息"
// @Success 200 {object} utils.Response
// @Router /api/workflow/{id}/versions/publish [post]
func PublishVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	var req dto.PublishVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := workflowservice.PublishVersion(id, req.Version, db); err != nil {
		utils.ErrorWithMsg(ctx, "发布版本失败", err)
		return
	}

	utils.OkWithMsg(ctx, "发布成功")
}

// RollbackVersion 回滚版本
// @Summary 回滚版本
// @Description 回滚工作流到指定版本
// @Tags 工作流版本管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param version path int true "版本号"
// @Success 200 {object} utils.Response
// @Router /api/workflow/{id}/versions/{version}/rollback [post]
func RollbackVersion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	versionStr := ctx.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		utils.ErrorWithMsg(ctx, "版本号错误", err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := workflowservice.RollbackToVersion(id, version, db); err != nil {
		utils.ErrorWithMsg(ctx, "回滚失败", err)
		return
	}

	utils.OkWithMsg(ctx, "回滚成功")
}

// ==================== 模板管理 ====================

// CreateTemplate 创建模板
// @Summary 创建模板
// @Description 从工作流创建模板
// @Tags 工作流模板管理
// @Accept json
// @Produce json
// @Param data body dto.CreateTemplateReq true "模板信息"
// @Success 200 {object} utils.Response{data=vo.TemplateVO}
// @Router /api/workflow/templates [post]
func CreateTemplate(ctx *gin.Context) {
	var req dto.CreateTemplateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	template, err := workflowservice.CreateTemplate(req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "创建模板失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToTemplateVO(*template))
}

// ListTemplates 获取模板列表
// @Summary 获取模板列表
// @Description 获取工作流模板市场
// @Tags 工作流模板管理
// @Accept json
// @Produce json
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Param category query string false "模板分类"
// @Param name query string false "模板名称"
// @Success 200 {object} utils.Response
// @Router /api/workflow/templates [get]
func ListTemplates(ctx *gin.Context) {
	var req dto.ListTemplateReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	pageVo, err := workflowservice.ListTemplates(req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取模板列表失败", err)
		return
	}

	vos := vo.ToTemplateVOs(pageVo.Records)
	result := page.Convert(pageVo, vos)
	utils.OkWithData(ctx, result)
}

// CloneTemplate 克隆模板
// @Summary 克隆模板
// @Description 从模板创建工作流
// @Tags 工作流模板管理
// @Accept json
// @Produce json
// @Param data body dto.CloneTemplateReq true "克隆信息"
// @Success 200 {object} utils.Response
// @Router /api/workflow/templates/clone [post]
func CloneTemplate(ctx *gin.Context) {
	var req dto.CloneTemplateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	flow, err := workflowservice.CloneTemplate(req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "克隆模板失败", err)
		return
	}

	utils.OkWithData(ctx, vo.ToAgentFlowVO(*flow))
}

// ListPublished 获取已发布工作流列表（客户端用）
// @Summary 获取已发布工作流列表
// @Description 获取状态为 published 的工作流列表，供客户端浏览
// @Tags 工作流公开接口
// @Accept json
// @Produce json
// @Param page query int true "页码" minimum(1)
// @Param limit query int true "每页数量" minimum(1)
// @Param name query string false "工作流名称"
// @Param category query string false "分类"
// @Success 200 {object} utils.Response
// @Router /api/workflow/public [get]
func ListPublished(ctx *gin.Context) {
	var req dto.ListPublishedWorkflowReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	pageVo, err := workflowservice.ListPublished(req, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取列表失败", err)
		return
	}

	vos := vo.ToPublishedWorkflowVOs(pageVo.Records)
	result := page.Convert(pageVo, vos)
	utils.OkWithData(ctx, result)
}

// GetPublishedDetail 获取已发布工作流详情（客户端用）
// @Summary 获取已发布工作流详情
// @Description 获取已发布工作流详情，包含拓扑数据用于解析 inputSchema
// @Tags 工作流公开接口
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} utils.Response{data=vo.PublishedWorkflowVO}
// @Router /api/workflow/public/{id} [get]
func GetPublishedDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	flow, err := workflowservice.Get(id, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取工作流失败", err)
		return
	}

	if flow.Status != "published" {
		utils.ErrorWithMsg(ctx, "工作流未发布", nil)
		return
	}

	utils.OkWithData(ctx, vo.ToPublishedWorkflowVO(*flow))
}

// UpdateStatus 更新工作流状态
// @Summary 更新工作流状态
// @Description 更新工作流状态（draft/published）
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param data body dto.UpdateWorkflowStatusReq true "状态信息"
// @Success 200 {object} utils.Response
// @Router /api/workflow/{id}/status [put]
func UpdateStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	var req dto.UpdateWorkflowStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidateError(ctx, err)
		return
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	if err := workflowservice.UpdateStatus(id, req.Status, db); err != nil {
		utils.ErrorWithMsg(ctx, "更新状态失败", err)
		return
	}

	utils.OkWithMsg(ctx, "状态更新成功")
}
