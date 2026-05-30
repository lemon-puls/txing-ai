package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/global/logging/log"
)

// LLMValidationResponse LLM 校验响应结构
type LLMValidationResponse struct {
	Valid    bool                       `json:"valid"`
	Errors   []LLMValidationIssue       `json:"errors,omitempty"`
	Warnings []LLMValidationIssue       `json:"warnings,omitempty"`
	Summary  string                     `json:"summary,omitempty"`
}

// LLMValidationIssue LLM 校验发现的问题
type LLMValidationIssue struct {
	NodeID  string `json:"nodeId,omitempty"`
	Message string `json:"message"`
}

// ValidateTopologyWithLLM 使用 LLM 对工作流进行语义层面的校验
// 会先去除布局信息，只保留逻辑结构，然后交给 LLM 分析
func ValidateTopologyWithLLM(ctx context.Context, endpoint, apiKey, model string, topologyJSON string) (*ValidationResult, error) {
	// 解析拓扑，提取逻辑摘要
	summary, err := buildTopologySummary(topologyJSON)
	if err != nil {
		return nil, fmt.Errorf("解析拓扑失败: %w", err)
	}

	// 创建 LLM 模型
	maxTokens := 2048
	temperature := float32(0.1) // 低温度确保结果稳定

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     endpoint,
		Model:       model,
		APIKey:      apiKey,
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("创建校验模型失败: %w", err)
	}

	// 构建校验提示
	systemPrompt := `你是一个工作流校验专家。你的任务是审查 AI 工作流的逻辑结构，发现潜在问题并给出改进建议。

请从以下角度分析工作流：
1. 逻辑完整性：工作流是否能完成其设计目的？是否存在逻辑断裂？
2. 节点配置合理性：LLM 节点的提示词是否清晰？工具选择是否合适？
3. 条件分支合理性：条件判断逻辑是否正确？是否有遗漏的分支？
4. 数据流连贯性：节点之间的数据传递是否合理？
5. 潜在风险：是否有死循环、无限递归、资源浪费等风险？

你必须严格按照以下 JSON 格式返回结果，不要包含任何其他内容：
{
  "valid": true/false,
  "errors": [{"nodeId": "节点ID（可选）", "message": "问题描述"}],
  "warnings": [{"nodeId": "节点ID（可选）", "message": "建议描述"}],
  "summary": "整体评估摘要"
}

注意：
- "valid" 为 false 表示存在严重问题，建议不要执行
- "errors" 是必须修复的问题
- "warnings" 是优化建议，不影响执行
- 如果工作流没有问题，返回 {"valid": true, "summary": "..."} 即可`

	userMessage := fmt.Sprintf("请审查以下 AI 工作流：\n\n%s", summary)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	}

	log.Info("开始 LLM 工作流校验", zap.String("model", model))

	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 校验调用失败: %w", err)
	}

	// 解析响应
	content := strings.TrimSpace(response.Content)
	content = extractJSON(content)

	var llmResp LLMValidationResponse
	if err := json.Unmarshal([]byte(content), &llmResp); err != nil {
		// 解析失败时，将整个响应作为 summary 返回
		log.Warn("LLM 校验响应解析失败，使用文本结果", zap.String("content", content))
		return &ValidationResult{
			Valid: true,
			Warnings: []ValidationError{
				{
					Level:   LevelWarning,
					Code:    "LLM_VALIDATION_PARSE_ERROR",
					Message: fmt.Sprintf("LLM 返回了非结构化结果: %s", content),
				},
			},
		}, nil
	}

	// 转换为通用 ValidationResult
	result := &ValidationResult{
		Valid: llmResp.Valid,
	}

	for _, e := range llmResp.Errors {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  e.NodeID,
			Code:    "LLM_VALIDATION_ERROR",
			Message: e.Message,
		})
	}

	for _, w := range llmResp.Warnings {
		result.Warnings = append(result.Warnings, ValidationError{
			Level:   LevelWarning,
			NodeID:  w.NodeID,
			Code:    "LLM_VALIDATION_WARNING",
			Message: w.Message,
		})
	}

	// 将 summary 作为信息性 warning 返回
	if llmResp.Summary != "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Level:   LevelWarning,
			Code:    "LLM_VALIDATION_SUMMARY",
			Message: llmResp.Summary,
		})
	}

	return result, nil
}

// topologySummary 用于生成给 LLM 的拓扑摘要（去除布局信息）
type topologySummary struct {
	Nodes []nodeSummary `json:"nodes"`
	Edges []edgeSummary `json:"edges"`
}

type nodeSummary struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Label    string            `json:"label"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type edgeSummary struct {
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"`
}

// buildTopologySummary 从拓扑 JSON 中提取逻辑摘要，去除位置等无关信息
func buildTopologySummary(topologyJSON string) (string, error) {
	var topo Topology
	if err := json.Unmarshal([]byte(topologyJSON), &topo); err != nil {
		return "", err
	}

	summary := topologySummary{
		Nodes: make([]nodeSummary, 0, len(topo.Nodes)),
		Edges: make([]edgeSummary, 0, len(topo.Edges)),
	}

	for _, node := range topo.Nodes {
		ns := nodeSummary{
			ID:    node.Id,
			Type:  node.Data.NodeType,
			Label: node.Data.Label,
		}

		// 根据节点类型提取关键配置
		config := make(map[string]interface{})
		switch node.Data.NodeType {
		case "llm":
			if node.Data.ModelConfig != nil {
				config["model"] = node.Data.ModelConfig.Model
				config["systemPrompt"] = node.Data.ModelConfig.SystemPrompt
				config["temperature"] = node.Data.ModelConfig.Temperature
				config["maxTokens"] = node.Data.ModelConfig.MaxTokens
			}
		case "tool":
			if node.Data.ToolConfig != nil {
				config["tools"] = node.Data.ToolConfig.Tools
			}
		case "condition":
			if node.Data.ConditionConf != nil {
				config["type"] = node.Data.ConditionConf.Type
				config["expression"] = node.Data.ConditionConf.Expression
				config["llmPrompt"] = node.Data.ConditionConf.LLMPrompt
				config["expectedValue"] = node.Data.ConditionConf.ExpectedValue
			}
		}
		if len(config) > 0 {
			ns.Config = config
		}

		summary.Nodes = append(summary.Nodes, ns)
	}

	for _, edge := range topo.Edges {
		summary.Edges = append(summary.Edges, edgeSummary{
			Source:       edge.Source,
			Target:       edge.Target,
			SourceHandle: edge.SourceHandle,
		})
	}

	jsonBytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// extractJSON 从可能包含 markdown 代码块的文本中提取 JSON
func extractJSON(content string) string {
	// 去除 markdown 代码块
	if strings.HasPrefix(content, "```") {
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
		if len(jsonLines) > 0 {
			return strings.Join(jsonLines, "\n")
		}
	}

	// 尝试找到第一个 { 和最后一个 }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		return content[start : end+1]
	}

	return content
}
