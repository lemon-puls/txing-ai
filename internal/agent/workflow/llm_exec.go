package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// LLMExecConfig 共享 LLM 执行配置
// LLMExecConfig is the shared configuration for executing an LLM call with
// an optional multi-round tool-calling loop. It is used by:
//   - LLM 节点 / the "llm" workflow node
//   - Agent 节点 / the "agent" workflow node
//   - 并行分支内的 LLM 节点 / LLM nodes inside parallel branches
type LLMExecConfig struct {
	NodeID    string // 节点 ID / Node ID (used to tag chunks)
	NodeLabel string // 节点标签 / Node label (used to tag chunks)
	NodeType  string // 节点类型 / Node type for chunk tagging: "llm" / "agent"

	// 模型解析 / Model resolution
	ModelResolver   types.ModelResolver // 模型解析器（可为 nil）/ Model resolver (may be nil)
	ModelName       string              // 节点配置的模型名，为空则使用默认 / Node-level model name; empty means defaults
	DefaultEndpoint string              // 默认 LLM 端点 / Default LLM endpoint
	DefaultAPIKey   string              // 默认 API 密钥 / Default API key
	DefaultModel    string              // 默认模型名 / Default model name
	// FallbackChatModel 创建节点模型失败时的兜底模型（可选）
	// FallbackChatModel is used when creating the node model fails (optional).
	FallbackChatModel *openai.ChatModel

	// 生成参数 / Generation parameters
	MaxTokens    int     // <=0 时使用默认值 8192 / defaults to 8192 when <=0
	Temperature  float32 // <=0 时使用默认值 0.7 / defaults to 0.7 when <=0
	SystemPrompt string  // 系统提示词 / System prompt

	// 对话历史（可选）：插入 system 之后、本次输入之前，用于多轮对话
	// Conversation history (optional): inserted after the system message and
	// before the current input, for multi-turn conversations.
	History []*schema.Message

	// 是否流式调用：开启后逐 token 推送 Content/ReasoningContent delta，
	// 并通过 ConcatMessages 聚合出完整消息供工具循环使用
	// When enabled, the model is called in streaming mode: content/reasoning
	// deltas are pushed via callback as they arrive, and the full message is
	// reconstructed via ConcatMessages for the tool-calling loop.
	Stream bool

	// 工具 / Tools
	AllTools             []tool.BaseTool // 全量工具注册表 / Full tool registry
	ToolNames            []string        // 需要绑定的工具名列表 / Tool names to bind
	UseAllToolsWhenEmpty bool            // Agent 节点语义：未配置工具时使用全部工具 / Agent-node semantics: use all tools when none configured
	MaxToolRounds        int             // 最大工具调用轮次，<=0 时使用默认值 5 / Max tool-call rounds; defaults to 5 when <=0
	Retry                *types.RetryConfig

	// EmitFinalContent 是否通过 callback 推送最终内容
	// EmitFinalContent controls whether the final content is pushed via callback.
	// 并行分支场景应设为 false，避免分支输出被前端渲染为最终结果。
	// Set it to false inside parallel branches so branch outputs are not
	// rendered as the final answer by the frontend.
	EmitFinalContent bool
}

// ExecuteLLM 执行带多轮工具调用的 LLM 调用（共享执行核心）
// ExecuteLLM performs an LLM call with an optional multi-round tool-calling
// loop. This is the single shared implementation previously duplicated in
// the llm node, the agent node and the parallel executor.
// 返回模型最终输出内容 / Returns the final model output content.
func ExecuteLLM(ctx context.Context, cfg *LLMExecConfig, input string, callback func(chunk *global.Chunk) error) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("LLM 执行配置为空 / LLM exec config is nil")
	}

	// 参数默认值 / Apply defaults
	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 8192
	}
	temperature := cfg.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}
	maxToolRounds := cfg.MaxToolRounds
	if maxToolRounds <= 0 {
		maxToolRounds = 5
	}

	// 解析节点模型信息（支持节点级别覆盖）
	// Resolve node model info (node-level override supported)
	nodeEndpoint, nodeAPIKey, nodeModel := resolveLLMModel(
		cfg.ModelResolver, cfg.ModelName,
		cfg.DefaultEndpoint, cfg.DefaultAPIKey, cfg.DefaultModel)

	// 创建节点专属模型 / Create the node-specific chat model
	nodeChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     nodeEndpoint,
		Model:       nodeModel,
		APIKey:      nodeAPIKey,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	})
	usingFallback := false
	if err != nil {
		log.Error("创建节点模型失败",
			zap.String("nodeId", cfg.NodeID), zap.Error(err))
		if cfg.FallbackChatModel == nil {
			return "", fmt.Errorf("创建节点模型失败: %w", err)
		}
		// 使用兜底模型；兜底模型已绑定全量工具，跳过重新绑定避免污染共享模型
		// Use the fallback model. It already has all tools bound, so skip
		// rebinding to avoid mutating the shared model instance.
		nodeChatModel = cfg.FallbackChatModel
		usingFallback = true
	}

	// 选择并绑定工具 / Select and bind tools
	nodeTools := selectLLMTools(ctx, cfg)
	var llmToolNode *compose.ToolsNode
	if len(nodeTools) > 0 {
		nodeToolInfos := make([]*schema.ToolInfo, 0, len(nodeTools))
		for _, t := range nodeTools {
			info, infoErr := t.Info(ctx)
			if infoErr != nil {
				continue
			}
			nodeToolInfos = append(nodeToolInfos, info)
		}
		if !usingFallback {
			if bindErr := nodeChatModel.BindTools(nodeToolInfos); bindErr != nil {
				log.Warn("LLM 节点绑定工具失败",
					zap.String("nodeId", cfg.NodeID), zap.Error(bindErr))
			}
		}
		toolNode, toolNodeErr := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
			Tools:               nodeTools,
			ExecuteSequentially: true,
		})
		if toolNodeErr != nil {
			log.Error("创建 LLM 节点工具执行器失败",
				zap.String("nodeId", cfg.NodeID), zap.Error(toolNodeErr))
		} else {
			llmToolNode = toolNode
		}
	}

	// 构建消息 / Build messages
	var messages []*schema.Message
	if cfg.SystemPrompt != "" {
		messages = append(messages, schema.SystemMessage(cfg.SystemPrompt))
	}
	messages = append(messages, cfg.History...)
	if input != "" {
		messages = append(messages, schema.UserMessage(input))
	}

	// 首次 LLM 调用 / First LLM call
	response, execErr := cfg.callChatModel(ctx, nodeChatModel, messages, callback)
	if execErr != nil {
		log.Error("LLM generate error",
			zap.String("nodeId", cfg.NodeID), zap.Error(execErr))
		return "", fmt.Errorf("LLM 调用失败: %w", execErr)
	}

	// 多轮工具调用循环 / Multi-round tool-calling loop
	if llmToolNode != nil {
		// 同一工具连续生成非法参数（如大段正文 JSON 被截断）时，错误回喂很难让
		// 模型自愈，继续循环只会空转消耗轮次；达到上限后中止并做收尾生成
		const maxConsecutiveInvalidToolCalls = 3
		consecutiveInvalid := 0

		for round := 0; round < maxToolRounds; round++ {
			if response == nil || len(response.ToolCalls) == 0 {
				break
			}

			// 用户停止生成（上下文已取消）后不再继续调用工具：
			// 避免出现"工具执行失败: context canceled"之类的噪声展示
			if ctxErr := ctx.Err(); ctxErr != nil {
				return "", fmt.Errorf("LLM 节点执行被中断: %w", ctxErr)
			}

			// 兜底补齐工具调用 ID：部分 provider/中转不返回 ID，缺失会导致
			// 前后端按 ID 精确匹配失效、同名多次调用被合并成一条
			ensureToolCallIDs(response.ToolCalls, cfg.NodeID, round)

			log.Info("LLM 节点工具调用",
				zap.String("nodeId", cfg.NodeID),
				zap.Int("round", round+1),
				zap.Int("toolCallCount", len(response.ToolCalls)))

			// 推送每个工具调用的详细信息 / Emit tool call details
			if callback != nil {
				for _, tc := range response.ToolCalls {
					callback(&global.Chunk{
						NodeId:     cfg.NodeID,
						NodeType:   cfg.NodeType,
						NodeLabel:  cfg.NodeLabel,
						ToolCallId: tc.ID,
						ToolName:   tc.Function.Name,
						ToolParams: tc.Function.Arguments,
						ToolStatus: "running",
						ShowMsg:    fmt.Sprintf("[%s] 调用工具: %s", cfg.NodeLabel, tc.Function.Name),
					})
				}
			}

			// 将 assistant 消息（含 ToolCalls）加入消息列表
			// Append the assistant message (with ToolCalls) to the history
			messages = append(messages, response)

			// 参数合法性校验：非法 JSON 参数不执行、错误信息回喂给模型自修复；
			// 参数合法的子集照常执行 —— provider 要求 assistant 消息里的每个
			// tool_call_id 都有对应 tool 消息，漏喂会导致下一次请求 400
			// Validate tool arguments: invalid JSON arguments are not executed and
			// an error is fed back for self-repair, while the valid subset still runs.
			validCalls, errMsgs := splitToolCalls(response.ToolCalls)
			if len(errMsgs) > 0 {
				consecutiveInvalid++
				// 仅对参数非法的调用推送 failed，合法调用的结果随后照常推送，
				// 避免前端工具行状态与实际执行情况不符
				if callback != nil {
					for _, em := range errMsgs {
						callback(&global.Chunk{
							NodeId:     cfg.NodeID,
							NodeType:   cfg.NodeType,
							NodeLabel:  cfg.NodeLabel,
							ToolCallId: em.ToolCallID,
							ToolName:   em.ToolName,
							ToolResult: "工具调用参数非法，本次调用未执行",
							ToolStatus: "failed",
							ShowMsg:    fmt.Sprintf("[%s] 工具 %s 参数非法", cfg.NodeLabel, em.ToolName),
						})
					}
				}
				messages = append(messages, errMsgs...)

				// 执行参数合法的子集并回喂结果，保证消息历史完整
				if len(validCalls) > 0 {
					toolResults, toolErr := llmToolNode.Invoke(ctx, cloneWithToolCalls(response, validCalls))
					if toolErr != nil {
						messages = append(messages, toolErrorMessages("工具执行失败: "+toolErr.Error(), validCalls)...)
					} else {
						emitToolResults(cfg, callback, toolResults)
						messages = append(messages, capToolResultsForContext(toolResults)...)
					}
				}

				if consecutiveInvalid >= maxConsecutiveInvalidToolCalls {
					log.Warn("工具调用参数连续非法，中止工具循环",
						zap.String("nodeId", cfg.NodeID),
						zap.Int("consecutiveInvalid", consecutiveInvalid))
					break
				}
			} else {
				consecutiveInvalid = 0
				// 执行工具调用 / Invoke tools
				toolResults, toolErr := llmToolNode.Invoke(ctx, response)
				if toolErr != nil {
					log.Error("LLM 节点工具执行失败",
						zap.String("nodeId", cfg.NodeID), zap.Error(toolErr))
					// 用户停止生成导致的工具中断：标记为"已中断"而非"失败"，
					// 避免前端工具行显示红色失败误导用户
					toolStatus := "failed"
					toolResultMsg := "工具执行失败: " + toolErr.Error()
					if ctx.Err() != nil {
						toolStatus = "interrupted"
						toolResultMsg = "工具执行已被中断"
					}
					if callback != nil {
						// 逐个工具推送状态（带身份与原因），让前端对应行能翻成对应状态并展示详情
						for _, tc := range response.ToolCalls {
							callback(&global.Chunk{
								NodeId:     cfg.NodeID,
								NodeType:   cfg.NodeType,
								NodeLabel:  cfg.NodeLabel,
								ToolCallId: tc.ID,
								ToolName:   tc.Function.Name,
								ToolResult: toolResultMsg,
								ToolStatus: toolStatus,
								ShowMsg:    fmt.Sprintf("[%s] 工具 %s 执行被中断", cfg.NodeLabel, tc.Function.Name),
							})
						}
					}
					// 每个工具调用各回喂一条失败消息（provider 强校验每个 tool_call_id）
					messages = append(messages, toolErrorMessages(toolResultMsg, response.ToolCalls)...)
				} else {
					emitToolResults(cfg, callback, toolResults)
					// 回喂前截断超长结果（展示 chunk 已用原始全量内容推送，不受影响）
					messages = append(messages, capToolResultsForContext(toolResults)...)
				}
			}

			if callback != nil {
				callback(&global.Chunk{
					NodeId:    cfg.NodeID,
					NodeType:  cfg.NodeType,
					NodeLabel: cfg.NodeLabel,
					ShowMsg:   fmt.Sprintf("[%s] 继续思考... (第%d轮)", cfg.NodeLabel, round+1),
				})
			}

			// 再次调用 LLM / Call the LLM again
			response, execErr = cfg.callChatModel(ctx, nodeChatModel, messages, callback)
			if execErr != nil {
				log.Error("LLM 多轮调用 generate error",
					zap.String("nodeId", cfg.NodeID),
					zap.Error(execErr), zap.Int("round", round+1))
				break
			}
		}
	}

	// 用户主动停止生成（或执行超时）：工具循环因上下文取消而中断时，
	// 立即把取消错误向上传播，不再执行收尾工具/续写/兜底文案。
	// 否则响应为空时兜底文案（"执行未能完成，可能因输出超出长度限制…"）
	// 会被当成正常输出展示，误导用户以为执行真的失败了
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", fmt.Errorf("LLM 节点执行被中断: %w", ctxErr)
	}

	// [TOOL-ROUND-LIMIT] 轮次耗尽但模型仍在请求工具调用：
	// 工具调用响应通常无文本内容（Content 为空），直接返回会导致最终输出为空。
	// 这里执行最后一批工具并做收尾生成，让模型基于结果产出最终答复。
	if response != nil && len(response.ToolCalls) > 0 && llmToolNode != nil {
		// 主循环因连续非法中止时 response 已入列（同一指针），重复追加会产生
		// 两条相同的 assistant(tool_calls) 消息，破坏消息历史
		if len(messages) == 0 || messages[len(messages)-1] != response {
			// 与主循环一致：补齐缺失的工具调用 ID，保证 ToolMessage 引用与展示匹配稳定
			ensureToolCallIDs(response.ToolCalls, cfg.NodeID, maxToolRounds)
			messages = append(messages, response)
			validCalls, errMsgs := splitToolCalls(response.ToolCalls)
			messages = append(messages, errMsgs...)
			if len(validCalls) > 0 {
				if toolResults, toolErr := llmToolNode.Invoke(ctx, cloneWithToolCalls(response, validCalls)); toolErr != nil {
					messages = append(messages, toolErrorMessages("工具执行失败: "+toolErr.Error(), validCalls)...)
				} else {
					emitToolResults(cfg, callback, toolResults)
					messages = append(messages, capToolResultsForContext(toolResults)...)
				}
			}
		}
		// 收尾生成：本次结果作为最终输出，不再继续工具循环
		if response, execErr = cfg.callChatModel(ctx, nodeChatModel, messages, callback); execErr != nil {
			log.Error("LLM 收尾生成失败", zap.String("nodeId", cfg.NodeID), zap.Error(execErr))
		}
	}

	result := ""
	if response != nil {
		result = response.Content
	}

	// 截断续写：输出撞 token 上限时 finish_reason=length，内容会停在半截（如表格中间）。
	// 以半成品为上下文让模型从截断处继续，最多续写 maxContinuations 次，内容拼接为完整结果。
	// Truncation continuation: when generation stops at the token limit, resume
	// from the partial answer and concatenate the pieces.
	const maxContinuations = 2
	for cont := 0; cont < maxContinuations; cont++ {
		if response == nil || response.ResponseMeta == nil || response.ResponseMeta.FinishReason != "length" {
			break
		}
		// 响应为空但 finish_reason=length（如推理模型的思考过程耗尽 token 配额、
		// 未产出正文）：无可续写内容；且把空 assistant 消息追加进上下文会被渠道拒绝
		// （400 Invalid assistant message: content or tool_calls must be set），直接跳过续写
		if response.Content == "" {
			log.Warn("LLM 输出为空且 finish_reason=length，无可续写内容",
				zap.String("nodeId", cfg.NodeID))
			break
		}
		log.Warn("LLM 输出撞 token 上限被截断，尝试续写",
			zap.String("nodeId", cfg.NodeID),
			zap.Int("continuation", cont+1),
			zap.Int("partialLen", len(response.Content)))
		messages = append(messages, response, schema.UserMessage(
			"你上面的输出因长度限制被截断。请严格从截断处继续输出：不要重复已输出的内容，不要添加任何前言或说明，直接续写正文。"))
		var contErr error
		response, contErr = cfg.callChatModel(ctx, nodeChatModel, messages, callback)
		if contErr != nil {
			log.Error("LLM 截断续写生成失败", zap.String("nodeId", cfg.NodeID), zap.Error(contErr))
			break
		}
		if response != nil {
			result += response.Content
		}
	}

	// 用户主动停止生成（或执行超时）：不产出兜底文案，取消错误向上传播，
	// 由上层识别为"已中断"（interrupted）。若已累积部分内容（如截断续写被打断），
	// 先推送出去，中断消息仍能展示已生成的部分（流式模式 delta 已推送过，跳过）
	if ctxErr := ctx.Err(); ctxErr != nil {
		if result != "" && cfg.EmitFinalContent && !cfg.Stream && callback != nil {
			callback(&global.Chunk{
				Content: result,
				ShowMsg: fmt.Sprintf("[%s] 思考中...", cfg.NodeLabel),
			})
		}
		return result, fmt.Errorf("LLM 节点执行被中断: %w", ctxErr)
	}

	// 最终兜底：仍未产出内容时给出提示，避免返回空字符串
	if result == "" && cfg.EmitFinalContent {
		result = "执行未能完成，可能因输出超出长度限制或工具调用轮次达到上限，请稍后重试或调整生成参数。"
	}

	// 推送最终内容 / Emit final content
	// 流式模式下 delta 已实时推送，不再整段重复推送
	if cfg.EmitFinalContent && !cfg.Stream && callback != nil && result != "" {
		callback(&global.Chunk{
			Content: result,
			ShowMsg: fmt.Sprintf("[%s] 思考中...", cfg.NodeLabel),
		})
	}

	return result, nil
}

// callChatModel 统一的模型调用入口，屏蔽 Generate 与 Stream 两种模式
// callChatModel is the single entry point for model calls, hiding the
// difference between Generate and Stream modes from the tool-calling loop.
//   - 非 Stream：与原实现一致，重试包裹整个 Generate 调用
//   - Stream：重试仅包裹流的建立阶段；Recv 中途出错直接返回（不重试，
//     避免 delta 重复推送），逐 token 通过 callback 推送增量内容，
//     最后用 ConcatMessages 聚合出完整消息（含按 index 聚合的 ToolCalls）
func (cfg *LLMExecConfig) callChatModel(ctx context.Context, model *openai.ChatModel,
	messages []*schema.Message, callback func(chunk *global.Chunk) error) (*schema.Message, error) {

	if !cfg.Stream {
		var response *schema.Message
		execErr := executeWithRetry(cfg.Retry, func() error {
			var genErr error
			response, genErr = model.Generate(ctx, messages)
			return genErr
		})
		return response, execErr
	}

	// 流式建立阶段（带重试）/ Stream establishment (with retry)
	var stream *schema.StreamReader[*schema.Message]
	execErr := executeWithRetry(cfg.Retry, func() error {
		var streamErr error
		stream, streamErr = model.Stream(ctx, messages)
		return streamErr
	})
	if execErr != nil {
		return nil, execErr
	}
	defer stream.Close()

	chunks := make([]*schema.Message, 0, 32)
	for {
		chunk, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		chunks = append(chunks, chunk)
		if callback != nil && (chunk.Content != "" || chunk.ReasoningContent != "") {
			if err := callback(&global.Chunk{
				NodeId:           cfg.NodeID,
				NodeType:         cfg.NodeType,
				NodeLabel:        cfg.NodeLabel,
				Content:          chunk.Content,
				ReasoningContent: chunk.ReasoningContent,
			}); err != nil {
				return nil, err
			}
		}
	}

	// 无任何 chunk 时返回空 assistant 消息，避免 ConcatMessages 报错
	if len(chunks) == 0 {
		return &schema.Message{Role: schema.Assistant}, nil
	}
	return schema.ConcatMessages(chunks)
}

// resolveLLMModel 解析节点模型信息，支持节点级别覆盖
// resolveLLMModel resolves the endpoint/apiKey/model for a node, supporting
// node-level overrides via the ModelResolver.
func resolveLLMModel(resolver types.ModelResolver, nodeModel, defaultEndpoint, defaultAPIKey, defaultModel string) (endpoint, apiKey, model string) {
	endpoint = defaultEndpoint
	apiKey = defaultAPIKey
	model = defaultModel

	if nodeModel == "" {
		return endpoint, apiKey, model
	}

	if resolver != nil {
		info, err := resolver.Resolve(nodeModel)
		if err != nil {
			log.Warn("解析节点模型失败，使用默认模型",
				zap.String("nodeModel", nodeModel), zap.Error(err))
			return endpoint, apiKey, model
		}
		return info.Endpoint, info.APIKey, info.Model
	}

	// 没有 ModelResolver 但节点指定了模型，仅覆盖模型名称
	// No resolver available: override the model name only
	model = nodeModel
	return endpoint, apiKey, model
}

// selectLLMTools 根据配置筛选节点要使用的工具
// selectLLMTools selects the tools a node should use based on its config.
func selectLLMTools(ctx context.Context, cfg *LLMExecConfig) []tool.BaseTool {
	if len(cfg.ToolNames) == 0 {
		if cfg.UseAllToolsWhenEmpty {
			return cfg.AllTools
		}
		return nil
	}

	toolMap := make(map[string]tool.BaseTool)
	for _, t := range cfg.AllTools {
		info, err := t.Info(ctx)
		if err != nil || info == nil {
			continue
		}
		toolMap[info.Name] = t
	}

	var result []tool.BaseTool
	for _, name := range cfg.ToolNames {
		if t, ok := toolMap[name]; ok {
			result = append(result, t)
		}
	}
	return result
}

// ensureToolCallIDs 为缺失 ID 的工具调用生成唯一兜底 ID
// ensureToolCallIDs assigns a synthetic unique ID to tool calls whose ID is
// empty. Some providers/relays omit tool call IDs; without one, the ID-based
// matching on both backend and frontend degrades and multiple calls of the
// same tool get merged into a single row.
// 注意：会原地修改 toolCalls（同一 response 随后交给 ToolsNode 执行，
// 结果消息将携带相同 ID，保证 running/completed chunk 能精确对应）。
func ensureToolCallIDs(toolCalls []schema.ToolCall, nodeID string, round int) {
	for i := range toolCalls {
		if toolCalls[i].ID == "" {
			toolCalls[i].ID = fmt.Sprintf("tc-%s-r%d-%d", nodeID, round, i)
		}
	}
}

// 回喂 LLM 上下文的工具结果长度上限（按 rune 计）
// toolContextMaxLen caps how much of each tool result is fed back into the
// LLM context. Search-type tools can return huge payloads (e.g. image search
// returns dozens of items per call); appending them in full bloats the
// context, squeezes the output budget and slows generation down.
const toolContextMaxLen = 8000

// capToolResultsForContext 截断超长的工具结果以保护上下文预算
// capToolResultsForContext returns copies of toolResults whose Content is
// capped at toolContextMaxLen. Original messages are not mutated (display
// chunks are emitted from the raw content elsewhere).
func capToolResultsForContext(toolResults []*schema.Message) []*schema.Message {
	capped := make([]*schema.Message, 0, len(toolResults))
	for _, tr := range toolResults {
		if tr == nil || len([]rune(tr.Content)) <= toolContextMaxLen {
			capped = append(capped, tr)
			continue
		}
		clone := *tr
		clone.Content = truncateRunes(tr.Content, toolContextMaxLen) + "\n…(内容过长已截断)"
		capped = append(capped, &clone)
	}
	return capped
}

// truncateRunes 按 rune 截断，避免截断多字节字符
func truncateRunes(s string, max int) string {
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	return string(rs[:max])
}

// emitToolResults 推送工具执行结果 chunk（completed 状态）
func emitToolResults(cfg *LLMExecConfig, callback func(*global.Chunk) error, toolResults []*schema.Message) {
	if callback == nil {
		return
	}
	for _, tr := range toolResults {
		callback(&global.Chunk{
			NodeId:     cfg.NodeID,
			NodeType:   cfg.NodeType,
			NodeLabel:  cfg.NodeLabel,
			ToolCallId: tr.ToolCallID,
			ToolName:   tr.ToolName,
			ToolResult: tr.Content,
			ToolStatus: "completed",
			ShowMsg:    fmt.Sprintf("[%s] 工具 %s 执行完成", cfg.NodeLabel, tr.ToolName),
		})
	}
}

// splitToolCalls 按参数是否为合法 JSON 把工具调用分成合法/非法两组，
// 并为非法组生成回喂错误消息（ToolMessage）。
// splitToolCalls partitions tool calls into a valid subset and error messages
// (as ToolMessages) for the invalid ones.
// 错误信息附带可操作指引：内容过长导致 JSON 被截断时，建议先落盘再传文件路径，
// 避免模型反复生成同样被截断的参数空转多轮。
func splitToolCalls(toolCalls []schema.ToolCall) (valid []schema.ToolCall, errMsgs []*schema.Message) {
	for _, tc := range toolCalls {
		var hint string
		switch {
		case tc.Function.Arguments == "":
			// 空参数同样回喂错误，避免模型以为调用已成功
			hint = "工具调用参数为空: " + tc.Function.Name + ", ToolCallID: " + tc.ID +
				"。请提供完整合法的 JSON 参数"
			log.Warn(hint)
		default:
			var jsonObj map[string]interface{}
			err := json.Unmarshal([]byte(tc.Function.Arguments), &jsonObj)
			if err == nil {
				valid = append(valid, tc)
				continue
			}
			hint = fmt.Sprintf("工具调用参数不是合法的 JSON: %s, ToolCallID: %s, 错误: %v",
				tc.Function.Name, tc.ID, err)
			// 大段正文塞进工具调用参数时，模型生成的 JSON 容易被截断/转义错误；
			// 给出明确的修复指引，避免同一失败反复重试
			switch tc.Function.Name {
			case "markdown_to_pdf_file_tool":
				hint += "。若内容过长导致 JSON 被截断，请先用 markdown_save_tool 保存 Markdown 文件，" +
					"再调用本工具并传入 filePath 参数（不要再把整段正文放进 content）"
			case "markdown_save_tool":
				hint += "。若内容过长导致 JSON 被截断，请将内容拆分为多段分多次保存到同一文件：" +
					"第一次调用传 is_append=false，后续调用传 is_append=true 追加（每次只传一小段，避免参数超长被截断）"
			default:
				hint += "。请重新生成，确保参数是完整合法的 JSON（注意双引号、换行需正确转义）"
			}
			log.Error(hint)
		}
		errMsgs = append(errMsgs, &schema.Message{
			Role:       schema.Tool,
			Content:    hint,
			ToolName:   tc.Function.Name,
			ToolCallID: tc.ID,
		})
	}
	return valid, errMsgs
}

// cloneWithToolCalls 返回仅携带 subset 工具调用的 response 副本（不修改原消息，
// 历史中已入列的 assistant 消息保持完整 tool_calls）
func cloneWithToolCalls(response *schema.Message, subset []schema.ToolCall) *schema.Message {
	clone := *response
	clone.ToolCalls = subset
	return &clone
}

// toolErrorMessages 为每个工具调用生成一条失败 ToolMessage。
// OpenAI 协议要求 assistant 消息里的每个 tool_call_id 都必须有对应 tool 消息，
// 只回喂首条会让下一次请求被 provider 以 400 拒绝。
func toolErrorMessages(msg string, toolCalls []schema.ToolCall) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(toolCalls))
	for _, tc := range toolCalls {
		msgs = append(msgs, &schema.Message{
			Role:       schema.Tool,
			Content:    msg,
			ToolName:   tc.Function.Name,
			ToolCallID: tc.ID,
		})
	}
	return msgs
}
