package agent

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"txing-ai/internal/agent"
	"txing-ai/internal/dto"
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
	// 根据模型选择对应的渠道
	channel, err := channel.ChooseChannelByModel(db, model)
	if err != nil {
		// 如果选择渠道失败，记录错误日志并返回错误响应
		log.Error("choose channel failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "选择渠道失败", err)
		return
	}
	// 获取请求中的内容
	content := req.Content
	// 执行智能体，传入上下文、渠道、模型和内容
	resp, err := agent.Execute(ctx, channel, model, content)

	if err != nil {
		// 如果执行智能体失败，记录错误日志并返回错误响应
		log.Error("execute agent failed", zap.Error(err))
		utils.ErrorWithMsg(ctx, "执行智能体失败", err)
		return
	}
	// 如果执行成功，返回成功响应和执行结果
	utils.OkWithData(ctx, resp)
}
