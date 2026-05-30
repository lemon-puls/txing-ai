package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"txing-ai/internal/agent"
	"txing-ai/internal/domain"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/iface"
	channelservice "txing-ai/internal/service/channel"
	workflowservice "txing-ai/internal/service/workflow"
	"txing-ai/internal/utils"
	"txing-ai/internal/utils/page"
	"txing-ai/internal/vo"

	"github.com/gin-gonic/gin"
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
	if err := db.Where("status = ?", 1).Find(&models).Error; err != nil {
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
	}
	utils.OkWithData(ctx, tools)
}

// Run 执行工作流
// @Summary 执行工作流
// @Description 动态执行工作流，支持流式输出
// @Tags 工作流管理
// @Accept json
// @Produce text/event-stream
// @Param id path int true "工作流ID"
// @Param content formData string false "请求内容"
// @Success 200 {string} string "SSE stream"
// @Router /api/workflow/{id}/run [post]
func Run(ctx *gin.Context, resProvider iface.ResourceProvider) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorWithMsg(ctx, "参数错误", err)
		return
	}

	content := ctx.PostForm("content")
	if content == "" {
		// 也可以从 JSON 中取，这里兼容 form
		var req struct {
			Content string `json:"content"`
		}
		if ctx.ShouldBindJSON(&req) == nil {
			content = req.Content
		}
	}

	db := utils.GetDBFromContext[*gorm.DB](ctx)
	flow, err := workflowservice.Get(id, db)
	if err != nil {
		utils.ErrorWithMsg(ctx, "获取工作流失败", err)
		return
	}

	// 初始化 WorkflowAgent
	workflowAgent := agent.NewWorkflowAgent(resProvider, flow.Topology)

	// 获取默认模型
	model := "deepseek-v3"
	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}
	channel, mappingModel, err := channelservice.ChooseChannelAndModel(db, model, mappingParams)
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

	// 发送结束标志
	endData := map[string]interface{}{
		"end": true,
	}
	endJsonData, _ := json.Marshal(endData)
	_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", endJsonData)
	ctx.Writer.Flush()
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
