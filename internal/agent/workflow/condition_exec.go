package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"txing-ai/internal/agent/workflow/condition"
	"txing-ai/internal/global"
)

// executeLLMCondition 使用 LLM 执行条件判断
func (a *WorkflowAgent) executeLLMCondition(ctx context.Context, endpoint, apiKey, model string, prompt string, input string, callback func(chunk *global.Chunk) error) (*condition.ConditionResult, error) {
	if prompt == "" {
		return nil, fmt.Errorf("LLM 判断提示词为空")
	}

	// 创建 LLM 模型
	maxTokens := 1024 // 判断结果不需要太长
	temperature := float32(0.1) // 低温度使结果更确定

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     endpoint,
		Model:       model,
		APIKey:      apiKey,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建判断模型失败: %w", err)
	}

	// 构建判断提示
	systemPrompt := `你是一个条件判断助手。请根据用户的输入内容，判断是否满足指定条件。

你必须严格按照以下 JSON 格式返回结果，不要包含任何其他内容：
{"result": true, "reason": "判断原因"}
或
{"result": false, "reason": "判断原因"}

result 必须是布尔值 true 或 false。
reason 简要说明判断原因。`

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(fmt.Sprintf("判断条件：%s\n\n输入内容：%s\n\n请判断输入内容是否满足条件。", prompt, input)),
	}

	if callback != nil {
		callback(&global.Chunk{ShowMsg: "[条件判断] AI 正在分析..."})
	}

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 判断失败: %w", err)
	}

	// 解析 JSON 响应
	content := strings.TrimSpace(response.Content)
	// 尝试提取 JSON（处理可能的 markdown 代码块）
	if strings.HasPrefix(content, "```") {
		// 移除 markdown 代码块
		lines := strings.Split(content, "\n")
		var jsonLines []string
		inBlock := false
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				inBlock = !inBlock
				continue
			}
			if inBlock {
				jsonLines = append(jsonLines, line)
			}
		}
		content = strings.Join(jsonLines, "\n")
	}

	var llmResponse condition.LLMJudgmentResponse
	if err := json.Unmarshal([]byte(content), &llmResponse); err != nil {
		// 尝试从文本中提取 true/false
		lowerContent := strings.ToLower(content)
		if strings.Contains(lowerContent, "true") || strings.Contains(lowerContent, "是") || strings.Contains(lowerContent, "yes") {
			return condition.NewConditionResult(true, "从响应文本中推断: "+content), nil
		}
		if strings.Contains(lowerContent, "false") || strings.Contains(lowerContent, "否") || strings.Contains(lowerContent, "no") {
			return condition.NewConditionResult(false, "从响应文本中推断: "+content), nil
		}
		return nil, fmt.Errorf("解析 LLM 判断结果失败: %w, 响应内容: %s", err, content)
	}

	if callback != nil {
		callback(&global.Chunk{ShowMsg: fmt.Sprintf("[条件判断] 结果: %v, 原因: %s", llmResponse.Result, llmResponse.Reason)})
	}

	return condition.NewConditionResult(llmResponse.Result, llmResponse.Reason), nil
}

// executeToolResultCondition 基于工具结果执行条件判断
func (a *WorkflowAgent) executeToolResultCondition(ctx context.Context, config *condition.ConditionConfigV2, input *schema.Message) *condition.ConditionResult {
	// 工具结果判断需要从消息的 Extra 中获取工具执行结果
	if input == nil || input.Extra == nil {
		return condition.NewConditionError(fmt.Errorf("无法获取工具执行结果"), config)
	}

	// 尝试从 Extra 中获取工具结果
	var toolResult interface{}
	var found bool

	// 查找工具结果
	if config.ToolResultKey != "" {
		if result, ok := input.Extra[config.ToolResultKey]; ok {
			toolResult = result
			found = true
		}
	}

	// 如果没有指定 key，尝试从常见字段获取
	if !found {
		for _, key := range []string{"result", "output", "data", "tool_result"} {
			if result, ok := input.Extra[key]; ok {
				toolResult = result
				found = true
				break
			}
		}
	}

	if !found {
		// 如果 Extra 中没有工具结果，尝试使用消息内容
		toolResult = input.Content
	}

	// 将结果转为字符串进行比较
	resultStr := fmt.Sprintf("%v", toolResult)

	// 与期望值比较
	if config.ExpectedValue != "" {
		if resultStr == config.ExpectedValue {
			return condition.NewConditionResult(true, fmt.Sprintf("工具结果 '%s' 等于期望值 '%s'", resultStr, config.ExpectedValue))
		}
		return condition.NewConditionResult(false, fmt.Sprintf("工具结果 '%s' 不等于期望值 '%s'", resultStr, config.ExpectedValue))
	}

	// 如果没有期望值，检查是否有表达式
	if config.Expression != "" {
		eval := condition.NewExpressionEvaluator()
		return eval.Evaluate(config.Expression, resultStr)
	}

	// 默认：非空即为 true
	if resultStr != "" && resultStr != "null" && resultStr != "nil" {
		return condition.NewConditionResult(true, "工具结果非空: "+resultStr)
	}

	return condition.NewConditionResult(false, "工具结果为空")
}
