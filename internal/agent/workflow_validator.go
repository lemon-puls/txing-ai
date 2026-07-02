package agent

import (
	"encoding/json"
	"fmt"
)

// ValidationLevel 校验级别
type ValidationLevel string

const (
	LevelError   ValidationLevel = "error"
	LevelWarning ValidationLevel = "warning"
)

// ValidationError 校验错误
type ValidationError struct {
	Level   ValidationLevel `json:"level"`           // "error" | "warning"
	NodeID  string          `json:"nodeId,omitempty"` // 关联节点（可选）
	Code    string          `json:"code"`             // 错误码
	Message string          `json:"message"`          // 人类可读描述
}

// ValidationResult 校验结果
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// 错误码常量
const (
	CodeGraphEmpty                = "GRAPH_EMPTY"
	CodeMissingStartNode          = "MISSING_START_NODE"
	CodeMultipleStartNodes        = "MULTIPLE_START_NODES"
	CodeMissingEndNode            = "MISSING_END_NODE"
	CodeOrphanNode                = "ORPHAN_NODE"
	CodeCycleDetected             = "CYCLE_DETECTED"
	CodeLLMNodeNoModel            = "LLM_NODE_NO_MODEL"
	CodeLLMNodeEmptyPrompt        = "LLM_NODE_EMPTY_PROMPT"
	CodeToolNodeNoTools           = "TOOL_NODE_NO_TOOLS"
	CodeConditionNoExpression     = "CONDITION_NODE_NO_EXPRESSION"
	CodeConditionMissingTrue      = "CONDITION_MISSING_TRUE_BRANCH"
	CodeConditionMissingFalse     = "CONDITION_MISSING_FALSE_BRANCH"
	CodeEdgeInvalidSource         = "EDGE_INVALID_SOURCE"
	CodeEdgeInvalidTarget         = "EDGE_INVALID_TARGET"
	CodeStartHasIncoming          = "START_HAS_INCOMING"
	CodeEndHasOutgoing            = "END_HAS_OUTGOING"
	CodeConditionIncompleteConfig = "CONDITION_INCOMPLETE_CONFIG"
)

// ValidateTopology 对工作流拓扑进行结构校验
// 返回 ValidationResult，Valid=true 表示校验通过
func ValidateTopology(topologyJSON string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// 解析拓扑
	var topo Topology
	if err := json.Unmarshal([]byte(topologyJSON), &topo); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			Code:    CodeGraphEmpty,
			Message: fmt.Sprintf("拓扑数据解析失败: %v", err),
		})
		return result
	}

	// 1. 空图检查
	if len(topo.Nodes) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			Code:    CodeGraphEmpty,
			Message: "工作流中没有任何节点",
		})
		return result
	}

	// 构建节点索引
	nodeMap := make(map[string]*TopoNode, len(topo.Nodes))
	for i := range topo.Nodes {
		nodeMap[topo.Nodes[i].Id] = &topo.Nodes[i]
	}

	// 构建入边/出边索引
	inEdges := make(map[string][]TopoEdge)  // nodeId -> 入边列表
	outEdges := make(map[string][]TopoEdge) // nodeId -> 出边列表
	for _, edge := range topo.Edges {
		inEdges[edge.Target] = append(inEdges[edge.Target], edge)
		outEdges[edge.Source] = append(outEdges[edge.Source], edge)
	}

	// 2. Start/End 节点检查
	var startNodeIds, endNodeIds []string
	for _, node := range topo.Nodes {
		switch node.Data.NodeType {
		case "start":
			startNodeIds = append(startNodeIds, node.Id)
		case "end":
			endNodeIds = append(endNodeIds, node.Id)
		}
	}

	if len(startNodeIds) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			Code:    CodeMissingStartNode,
			Message: "工作流缺少「开始」节点",
		})
	} else if len(startNodeIds) > 1 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			Code:    CodeMultipleStartNodes,
			Message: fmt.Sprintf("工作流存在 %d 个「开始」节点，只能有 1 个", len(startNodeIds)),
		})
	}

	if len(endNodeIds) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			Code:    CodeMissingEndNode,
			Message: "工作流缺少「结束」节点",
		})
	}

	// 3. Start 节点不应有入边，End 节点不应有出边
	for _, sid := range startNodeIds {
		if len(inEdges[sid]) > 0 {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  sid,
				Code:    CodeStartHasIncoming,
				Message: "「开始」节点不应有入边连接",
			})
			result.Valid = false
		}
	}
	for _, eid := range endNodeIds {
		if len(outEdges[eid]) > 0 {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  eid,
				Code:    CodeEndHasOutgoing,
				Message: "「结束」节点不应有出边连接",
			})
			result.Valid = false
		}
	}

	// 4. 边的合法性检查
	for _, edge := range topo.Edges {
		if _, ok := nodeMap[edge.Source]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  edge.Source,
				Code:    CodeEdgeInvalidSource,
				Message: fmt.Sprintf("边 %s 的源节点 %s 不存在", edge.Id, edge.Source),
			})
			result.Valid = false
		}
		if _, ok := nodeMap[edge.Target]; !ok {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  edge.Target,
				Code:    CodeEdgeInvalidTarget,
				Message: fmt.Sprintf("边 %s 的目标节点 %s 不存在", edge.Id, edge.Target),
			})
			result.Valid = false
		}
	}

	// 5. 孤立节点检查（排除 start/end）
	for _, node := range topo.Nodes {
		if node.Data.NodeType == "start" || node.Data.NodeType == "end" {
			continue
		}
		hasIn := len(inEdges[node.Id]) > 0
		hasOut := len(outEdges[node.Id]) > 0
		if !hasIn && !hasOut {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  node.Id,
				Code:    CodeOrphanNode,
				Message: fmt.Sprintf("节点「%s」是孤立的，没有任何连接", node.Data.Label),
			})
			result.Valid = false
		}
	}

	// 6. 环路检测（DFS 拓扑排序）
	if hasCycle, cycleNodes := detectCycle(topo.Nodes, topo.Edges); hasCycle {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			Code:    CodeCycleDetected,
			Message: fmt.Sprintf("工作流存在环路，涉及节点: %v", cycleNodes),
		})
		result.Valid = false
	}

	// 7. 节点配置完整性检查
	for _, node := range topo.Nodes {
		switch node.Data.NodeType {
		case "llm":
			validateLLMNode(&node, result)
		case "tool":
			validateToolNode(&node, result)
		case "condition":
			validateConditionNode(&node, outEdges[node.Id], result)
		}
	}

	return result
}

// validateLLMNode 校验 LLM 节点配置
func validateLLMNode(node *TopoNode, result *ValidationResult) {
	if node.Data.ModelConfig == nil || node.Data.ModelConfig.Model == "" {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  node.Id,
			Code:    CodeLLMNodeNoModel,
			Message: fmt.Sprintf("LLM 节点「%s」未配置模型", node.Data.Label),
		})
		result.Valid = false
	}
	if node.Data.ModelConfig != nil && node.Data.ModelConfig.SystemPrompt == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Level:   LevelWarning,
			NodeID:  node.Id,
			Code:    CodeLLMNodeEmptyPrompt,
			Message: fmt.Sprintf("LLM 节点「%s」的系统提示词为空，建议添加以获得更好的效果", node.Data.Label),
		})
	}
}

// validateToolNode 校验工具节点配置
func validateToolNode(node *TopoNode, result *ValidationResult) {
	if node.Data.ToolConfig == nil {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  node.Id,
			Code:    CodeToolNodeNoTools,
			Message: fmt.Sprintf("工具节点「%s」未选择任何工具", node.Data.Label),
		})
		result.Valid = false
		return
	}
	// ToolName 为单工具模式（直接执行），Tools 为多工具模式（LLM 调用）
	if node.Data.ToolConfig.ToolName == "" && len(node.Data.ToolConfig.Tools) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  node.Id,
			Code:    CodeToolNodeNoTools,
			Message: fmt.Sprintf("工具节点「%s」未选择任何工具", node.Data.Label),
		})
		result.Valid = false
	}
}

// validateConditionNode 校验条件节点配置
func validateConditionNode(node *TopoNode, nodeOutEdges []TopoEdge, result *ValidationResult) {
	config := node.Data.ConditionConf

	if config == nil {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  node.Id,
			Code:    CodeConditionIncompleteConfig,
			Message: fmt.Sprintf("条件节点「%s」未配置判断条件", node.Data.Label),
		})
		result.Valid = false
		return
	}

	// 检查条件类型对应的配置是否完整
	switch config.Type {
	case "expression":
		if config.Expression == "" {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  node.Id,
				Code:    CodeConditionNoExpression,
				Message: fmt.Sprintf("条件节点「%s」的表达式为空", node.Data.Label),
			})
			result.Valid = false
		}
	case "llm":
		if config.LLMPrompt == "" {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  node.Id,
				Code:    CodeConditionNoExpression,
				Message: fmt.Sprintf("条件节点「%s」的 AI 判断提示词为空", node.Data.Label),
			})
			result.Valid = false
		}
	case "tool_result":
		if config.ToolResultKey == "" {
			result.Errors = append(result.Errors, ValidationError{
				Level:   LevelError,
				NodeID:  node.Id,
				Code:    CodeConditionNoExpression,
				Message: fmt.Sprintf("条件节点「%s」的工具结果字段为空", node.Data.Label),
			})
			result.Valid = false
		}
	}

	// 检查 true/false 分支出边（使用默认 handle 名称）
	hasTrue, hasFalse := false, false
	for _, edge := range nodeOutEdges {
		switch edge.SourceHandle {
		case "true":
			hasTrue = true
		case "false":
			hasFalse = true
		}
	}

	if !hasTrue {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  node.Id,
			Code:    CodeConditionMissingTrue,
			Message: fmt.Sprintf("条件节点「%s」缺少「True」分支的出边连接", node.Data.Label),
		})
		result.Valid = false
	}
	if !hasFalse {
		result.Errors = append(result.Errors, ValidationError{
			Level:   LevelError,
			NodeID:  node.Id,
			Code:    CodeConditionMissingFalse,
			Message: fmt.Sprintf("条件节点「%s」缺少「False」分支的出边连接", node.Data.Label),
		})
		result.Valid = false
	}
}

// detectCycle 使用 DFS 检测环路
// 返回是否有环，以及环中涉及的节点 ID
func detectCycle(nodes []TopoNode, edges []TopoEdge) (bool, []string) {
	// 构建邻接表
	adj := make(map[string][]string)
	for _, node := range nodes {
		adj[node.Id] = []string{}
	}
	for _, edge := range edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
	}

	const (
		white = 0 // 未访问
		gray  = 1 // 正在访问（在当前 DFS 栈中）
		black = 2 // 已完成
	)

	color := make(map[string]int)
	parent := make(map[string]string)

	var dfs func(nodeId string) (bool, []string)
	dfs = func(nodeId string) (bool, []string) {
		color[nodeId] = gray

		for _, neighbor := range adj[nodeId] {
			if color[neighbor] == gray {
				// 找到环，提取环中的节点
				cycle := []string{neighbor}
				cur := nodeId
				for cur != neighbor {
					cycle = append(cycle, cur)
					cur = parent[cur]
				}
				return true, cycle
			}
			if color[neighbor] == white {
				parent[neighbor] = nodeId
				if found, cycle := dfs(neighbor); found {
					return true, cycle
				}
			}
		}

		color[nodeId] = black
		return false, nil
	}

	for _, node := range nodes {
		if color[node.Id] == white {
			if found, cycle := dfs(node.Id); found {
				return true, cycle
			}
		}
	}

	return false, nil
}
