package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/url"
	"strings"
	"txing-ai/internal/agent"
	"txing-ai/internal/dto"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/service/channel"
	"txing-ai/internal/utils"
)

// Generate 调用智能体
// @Summary 调用智能体
// @Description 调用智能体
// @Tags agent
// @Accept json
// @Produce json
// @Param data body dto.AgentExecReq true "请求信息"
// @Success 200 {object} utils.Response
// @Router /api/agent/exec [POST]
// exec 是一个处理智能体执行的函数
// 它接收一个 gin.Context 对象作为参数，用于处理 HTTP 请求和响应
func Exec(ctx *gin.Context) {

	// 定义一个 AgentExecReq 结构体变量 req，用于存储请求中的 JSON 数据
	var req dto.AgentExecReq
	// 将请求中的 JSON 数据绑定到 req 结构体上
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 如果绑定失败，则返回验证错误信息
		utils.ValidateError(ctx, err)
		return
	}

	// 从上下文中获取智能体工厂和数据库连接
	agentFactory := utils.GetAgentFactoryFromContext(ctx)
	db := utils.GetDBFromContext(ctx)

	// 将请求中的 AgentType 字符串转换为 AgentType 类型
	agentType := agent.AgentType(req.AgentType)
	// 使用智能体工厂创建指定类型的智能体实例
	agent, err := agentFactory.CreateAgent(agentType)
	if err != nil {
		// 如果创建智能体失败，记录错误日志并返回错误响应
		log.Error("create agent failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "创建智能体失败", err)
		return
	}
	// TODO 支持后台配置，然后从数据库中查询出要使用的模型，暂时先用默认模型qwen-plus
	// 设置默认使用的模型为 qwen-plus
	model := "qwen-plus"

	// 构建映射参数
	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}

	// 根据模型选择对应的渠道
	channel, mappingModel, err := channel.ChooseChannelAndModel(db, model, mappingParams)
	if err != nil {
		// 如果选择渠道失败，记录错误日志并返回错误响应
		log.Error("choose channel failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "选择渠道失败", err)
		return
	}
	// 获取请求中的内容
	content := req.Content
	// 执行智能体，传入上下文、渠道、模型和内容
	resp, err := agent.Execute(ctx, channel.GetEndpoint(), channel.GetRandomSecret(), mappingModel, content)

	if err != nil {
		// 如果执行智能体失败，记录错误日志并返回错误响应
		log.Error("execute agent failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "执行智能体失败", err)
		return
	}
	// 如果执行成功，返回成功响应和执行结果
	utils.OkWithData(ctx, resp)
}

// ExecStream 基于 SSE 调用智能体
// @Summary 基于 SSE 调用智能体
// @Description 使用 Server-Sent Events 流式调用智能体
// @Tags agent
// @Accept multipart/form-data
// @Produce text/event-stream
// @Param agentType formData string true "智能体类型"
// @Param content formData string false "请求内容"
// @Param file formData file false "上传文件"
// @Success 200 {object} utils.Response
// @Router /api/agent/exec/stream [POST]
func ExecStream(ctx *gin.Context) {

	// 定义一个 AgentExecReq 结构体变量 req，用于存储请求中的 form-data 数据
	var req dto.AgentExecReq
	// 从 form-data 中获取参数
	req.AgentType = ctx.PostForm("agentType")
	req.Content = ctx.PostForm("content")

	// 验证必填字段
	if req.AgentType == "" {
		utils.ErrorWithMsg(ctx, "智能体类型不能为空", nil)
		return
	}

	// 从上下文中获取智能体工厂和数据库连接
	agentFactory := utils.GetAgentFactoryFromContext(ctx)
	db := utils.GetDBFromContext(ctx)
	userId := utils.GetUIDFromContext(ctx)

	// 将请求中的 AgentType 字符串转换为 AgentType 类型
	agentType := agent.AgentType(req.AgentType)
	// 使用智能体工厂创建指定类型的智能体实例
	agent, err := agentFactory.CreateAgent(agentType)
	if err != nil {
		// 如果创建智能体失败，记录错误日志并返回错误响应
		log.Error("create agent failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "创建智能体失败", err)
		return
	}

	// TODO 支持后台配置，然后从数据库中查询出要使用的模型，暂时先用默认模型qwen-plus
	// 设置默认使用的模型为 qwen-plus
	model := "deepseek-v3"
	// 构建映射参数
	mappingParams := map[string]interface{}{
		"type": global.LLMTypeModel,
	}
	// 根据模型选择对应的渠道
	channel, mappingModel, err := channel.ChooseChannelAndModel(db, model, mappingParams)
	if err != nil {
		// 如果选择渠道失败，记录错误日志并返回错误响应
		log.Error("choose channel failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "选择渠道失败", err)
		return
	}

	// 设置 SSE 响应头
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 获取请求中的内容
	content := req.Content

	ctxWithCancel, cancel := context.WithCancel(ctx)

	// 创建一个回调函数，用于处理流式响应
	callback := func(chunk *global.Chunk) error {
		// 构建 SSE 消息
		data := map[string]interface{}{
			"content":          chunk.Content,
			"reasoningContent": chunk.ReasoningContent,
			"toolCallId":       chunk.ToolCallId,
			"toolName":         chunk.ToolName,
			"toolParams":       chunk.ToolParams,
			"toolResult":       chunk.ToolResult,
			"showMsg":          chunk.ShowMsg,
			"end":              false,
		}

		// 发送 SSE 消息
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Error("json marshal data failed, cancel context", zap.Error(err))
			cancel()
			return err
		}

		_, err = fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
		if err != nil {
			log.Error("write sse message failed, cancel context", zap.Error(err))
			cancel()
			return err
		}
		ctx.Writer.Flush()
		return nil
	}

	// 获取上传的文件
	file, header, err := ctx.Request.FormFile("file")
	// 文件路径
	filePath := ""
	if err == nil && file != nil {
		// 保存文件到本地
		saveFilePath, fileSize, saveErr := utils.SaveUploadedFile(file, header.Filename, userId, "", "")
		if saveErr != nil {
			log.Error("save uploaded file failed", zap.Error(saveErr))
		} else {
			log.Info("file saved successfully", zap.String("path", filePath), zap.Int64("size", fileSize))
			filePath = saveFilePath
		}
		defer file.Close()
	}

	// 执行智能体，传入上下文、渠道、模型、内容和回调函数
	response := ""
	response, err = agent.ExecuteStream(ctxWithCancel, channel.GetEndpoint(), channel.GetRandomSecret(), mappingModel, content, filePath, callback)
	if err != nil {
		// 如果执行智能体失败，记录错误日志
		log.Error("execute agent stream failed", zap.Error(err))
		//// 判断上下文是否已经被取消
		//if ctxWithCancel.Err() != nil {
		//	log.Info("context canceled")
		//	return
		//}

		// 发送错误消息
		errData := map[string]interface{}{
			"error": err.Error(),
			"end":   true,
		}

		jsonData, err := json.Marshal(errData)
		if err != nil {
			log.Error("json marshal data failed", zap.Error(err))
		}

		_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
		ctx.Writer.Flush()
		return
	}

	downloadURL := ""
	// 从响应中获取文件路径
	if response != "" {
		// response 中包含文件的格式："文件：优化简历_lzw_腾讯后台开发工程师.pdf"
		// 从中获取文件名，然后拼接下载路径，再更新到 response 中
		// 下载路径 /api/file/download?filePath={文件名}
		if strings.Contains(response, "文件：") {
			// 提取文件名
			filePrefix := "文件："
			startIndex := strings.Index(response, filePrefix) + len(filePrefix)
			fileName := strings.TrimSpace(response[startIndex:])

			// 拼接下载路径
			downloadURL = fmt.Sprintf("/api/file/download?filePath=%s", url.QueryEscape(fileName))
		}
	}

	// 发送结束消息
	endData := map[string]interface{}{
		"content": downloadURL,
		"end":     true,
	}
	jsonData, err := json.Marshal(endData)
	if err != nil {
		log.Error("json marshal data failed", zap.Error(err))
	}

	_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
	ctx.Writer.Flush()
}
