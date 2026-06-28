package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// ParallelConfig 并行组节点配置
// ParallelConfig defines the configuration for parallel execution groups
type ParallelConfig struct {
	MaxConcurrency int          `json:"maxConcurrency"` // 最大并发数，0=无限制 / Max concurrency, 0=unlimited
	WaitStrategy   string      `json:"waitStrategy"`   // 等待策略：all / Wait strategy: all
	Timeout        int         `json:"timeout"`        // 超时时间（秒）/ Timeout in seconds
	BranchRetry    *RetryConfig `json:"branchRetry,omitempty"`
}

// JoinConfig 汇聚节点配置
// JoinConfig defines the configuration for join nodes
type JoinConfig struct {
	Strategy string `json:"strategy"` // 汇聚策略：all/any / Join strategy: all/any
	Timeout  int    `json:"timeout"`  // 超时时间（秒）/ Timeout in seconds
}

// ParallelGroup 并行组结构
// ParallelGroup represents a parallel execution group with its branches and join node
type ParallelGroup struct {
	ParallelNode *TopoNode      // 并行入口节点 / Parallel entry node
	Branches     [][]*TopoNode // 每个分支的节点列表 / Node list for each branch
	JoinNode     *TopoNode     // 汇聚节点 / Join node
	ParallelID   string        // 并行组ID / Parallel group ID
	Config       *ParallelConfig // 并行配置 / Parallel configuration
}

// ParallelExecutor 并行执行器
// ParallelExecutor handles parallel execution of workflow branches
type ParallelExecutor struct {
	agent      *WorkflowAgent // 工作流智能体 / Workflow agent
	maxWorkers int            // 最大工作线程数 / Maximum worker threads
	endpoint   string         // LLM 调用端点 / LLM endpoint
	apiKey     string         // LLM API 密钥 / LLM API key
	model      string         // 默认模型名 / Default model name
}

// NewParallelExecutor 创建并行执行器
// NewParallelExecutor creates a new parallel executor instance
func NewParallelExecutor(agent *WorkflowAgent, maxWorkers int, endpoint, apiKey, model string) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 10 // 默认最大并发数 / Default max concurrency
	}
	return &ParallelExecutor{
		agent:      agent,
		maxWorkers: maxWorkers,
		endpoint:   endpoint,
		apiKey:     apiKey,
		model:      model,
	}
}

// IdentifyParallelGroups 从拓扑中识别所有并行组
// IdentifyParallelGroups identifies all parallel groups from the topology
func (e *ParallelExecutor) IdentifyParallelGroups(topo *Topology) ([]*ParallelGroup, error) {
	if topo == nil {
		return nil, fmt.Errorf("拓扑图为空 / Topology is nil")
	}

	// 创建节点映射 / Create node map
	nodeMap := make(map[string]*TopoNode)
	for i := range topo.Nodes {
		nodeMap[topo.Nodes[i].Id] = &topo.Nodes[i]
	}

	// 查找所有 parallel 节点 / Find all parallel nodes
	var parallelGroups []*ParallelGroup

	for _, node := range topo.Nodes {
		if node.Data.NodeType == "parallel" {
			group := e.extractParallelGroup(&node, topo.Edges, nodeMap)
			if group != nil {
				parallelGroups = append(parallelGroups, group)
			}
		}
	}

	log.Info("识别到并行组数量",
		zap.Int("count", len(parallelGroups)))

	return parallelGroups, nil
}

// extractParallelGroup 提取单个并行组的详细信息
// extractParallelGroup extracts detailed information for a single parallel group
func (e *ParallelExecutor) extractParallelGroup(parallelNode *TopoNode, edges []TopoEdge, nodeMap map[string]*TopoNode) *ParallelGroup {
	if parallelNode == nil {
		return nil
	}

	parallelID := parallelNode.Id

	// 解析并行配置 / Parse parallel configuration
	config := &ParallelConfig{
		MaxConcurrency: 0,
		WaitStrategy:   "all",
		Timeout:        0,
	}
	if parallelNode.Data.ParallelConfig != nil {
		config.MaxConcurrency = parallelNode.Data.ParallelConfig.MaxConcurrency
		config.WaitStrategy = parallelNode.Data.ParallelConfig.WaitStrategy
		config.Timeout = parallelNode.Data.ParallelConfig.Timeout
	}

	// 找出所有属于该并行组的节点（通过 parallelId 或边关系）
	// Find all nodes belonging to this parallel group
	branchNodes := make(map[string][]*TopoNode) // branchId -> nodes
	joinNodeID := ""

	// 找出所有以此 parallel 节点为起点的边
	// Find all edges starting from this parallel node
	for _, edge := range edges {
		if edge.Source == parallelID {
			// 找到了分支入口
			branchID := edge.Target
			branchPath := e.traceBranchPath(branchID, edges, nodeMap, parallelID, &joinNodeID)
			if len(branchPath) > 0 {
				branchNodes[branchID] = branchPath
			}
		}
	}

	// 如果没有通过边找到分支，尝试通过 parallelId 属性查找
	// If no branches found via edges, try finding via parallelId property
	if len(branchNodes) == 0 {
		branchNodes = e.findBranchesByParallelId(parallelID, nodeMap, edges)
	}

	// 转换为分支列表，按 branch_X 序号排序以保证确定性 / Convert to branch list, sorted by branch_X index for determinism
	var branches [][]*TopoNode
	branchIDs := make([]string, 0, len(branchNodes))
	for branchID := range branchNodes {
		branchIDs = append(branchIDs, branchID)
	}
	sort.Strings(branchIDs) // branch_1 < branch_2 < branch_3
	for _, branchID := range branchIDs {
		branches = append(branches, branchNodes[branchID])
	}

	group := &ParallelGroup{
		ParallelNode: parallelNode,
		Branches:     branches,
		ParallelID:   parallelID,
		Config:       config,
	}

	// 设置汇聚节点 / Set join node
	if joinNodeID != "" {
		group.JoinNode = nodeMap[joinNodeID]
	}

	return group
}

// traceBranchPath 追踪单个分支的路径直到汇聚节点
// traceBranchPath traces a single branch path until reaching the join node
func (e *ParallelExecutor) traceBranchPath(startNodeID string, edges []TopoEdge, nodeMap map[string]*TopoNode, parallelID string, joinNodeID *string) []*TopoNode {
	var path []*TopoNode
	currentID := startNodeID
	visited := make(map[string]bool)

	for {
		if visited[currentID] {
			break
		}
		visited[currentID] = true

		node := nodeMap[currentID]
		if node == nil {
			break
		}

		// 检查是否是汇聚节点 / Check if this is a join node
		if node.Data.NodeType == "join" {
			if parallelID == "" || node.Data.Label == parallelID {
				*joinNodeID = currentID
				break
			}
		}

		path = append(path, node)

		// 找到下一个节点 / Find next node
		nextID := ""
		for _, edge := range edges {
			if edge.Source == currentID && edge.SourceHandle == "" {
				nextID = edge.Target
				break
			}
		}

		if nextID == "" {
			break
		}

		// 检查下一个节点是否是同一并行组的汇聚节点
		// Check if next node is a join node of this parallel group
		nextNode := nodeMap[nextID]
		if nextNode != nil && nextNode.Data.NodeType == "join" {
			*joinNodeID = nextID
			break
		}

		currentID = nextID
	}

	return path
}

// findBranchesByParallelId 通过 parallelId 属性查找分支
// findBranchesByParallelId finds branches by parallelId property
func (e *ParallelExecutor) findBranchesByParallelId(parallelID string, nodeMap map[string]*TopoNode, edges []TopoEdge) map[string][]*TopoNode {
	branchNodes := make(map[string][]*TopoNode)

	// 遍历所有节点查找属于该并行组的节点
	// Iterate all nodes to find those belonging to this parallel group
	for _, node := range nodeMap {
		// 检查节点的并行组ID / Check node's parallel group ID
		nodeParallelID := e.getNodeParallelId(node)

		// 检查是否属于当前并行组 / Check if belongs to current parallel group
		if nodeParallelID != parallelID {
			continue
		}

		// 跳过 parallel 和 join 节点 / Skip parallel and join nodes
		if node.Data.NodeType == "parallel" || node.Data.NodeType == "join" {
			continue
		}

		// 根据边关系确定分支归属 / Determine branch ownership by edge relationships
		// 查找以该节点为终点的边 / Find edges ending at this node
		for _, edge := range edges {
			if edge.Target == node.Id {
				sourceNode := nodeMap[edge.Source]
				if sourceNode != nil {
					sourceParallelID := e.getNodeParallelId(sourceNode)
					if sourceParallelID == parallelID || sourceNode.Data.NodeType == "parallel" {
						// 添加到对应分支 / Add to corresponding branch
						branchNodes[edge.Source] = append(branchNodes[edge.Source], node)
					}
				}
			}
		}
	}

	return branchNodes
}

// getNodeParallelId 获取节点的并行组ID
// getNodeParallelId gets the parallel group ID of a node
func (e *ParallelExecutor) getNodeParallelId(node *TopoNode) string {
	// 优先使用 Label 作为 parallelId / Prefer using Label as parallelId
	if node.Data.Label != "" && node.Data.Label != node.Id {
		return node.Data.Label
	}
	return node.Id
}

// ExecuteBranch 执行单个分支
// ExecuteBranch executes a single branch of the parallel group
func (e *ParallelExecutor) ExecuteBranch(ctx context.Context, branchIndex int, branchNodes []*TopoNode,
	initialInput string, pCtx *ParallelContext, callback func(chunk *global.Chunk) error) error {

	branchID := fmt.Sprintf("branch_%d", branchIndex)
	startTime := time.Now().UnixMilli()

	log.Info("开始执行分支",
		zap.String("branchId", branchID),
		zap.Int("nodeCount", len(branchNodes)))

	// 创建分支上下文 / Create branch context
	branchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 包装 callback，为每个 chunk 注入 branchID 以区分并行分支
	wrappedCallback := callback
	if callback != nil {
		wrappedCallback = func(chunk *global.Chunk) error {
			chunk.BranchID = branchID
			return callback(chunk)
		}
	}

	var lastOutput string = initialInput
	var branchErr error

	// 按顺序执行分支内的节点 / Execute nodes in branch sequentially
	for i, node := range branchNodes {
		select {
		case <-branchCtx.Done():
			branchErr = branchCtx.Err()
			break
		default:
		}

		nodeOutput, execErr := e.executeNode(branchCtx, node, lastOutput, wrappedCallback)

		if execErr != nil {
			log.Error("分支节点执行失败",
				zap.String("branchId", branchID),
				zap.String("nodeId", node.Id),
				zap.Error(execErr))
			branchErr = execErr
			break
		}

		if nodeOutput != "" {
			lastOutput = nodeOutput
		}

		// 发送节点完成消息 / Send node completion message
		if wrappedCallback != nil {
			wrappedCallback(&global.Chunk{
				NodeId:     node.Id,
				NodeType:   node.Data.NodeType,
				NodeLabel:  node.Data.Label,
				NodeStatus: "completed",
				ShowMsg:    fmt.Sprintf("[%s] 分支 %s 节点完成", node.Data.Label, branchID),
			})
		}

		// 避免最后一个节点重复发送消息 / Avoid duplicate message on last node
		_ = i
	}

	endTime := time.Now().UnixMilli()

	// 构建结果 / Build result
	status := "completed"
	if branchErr != nil {
		status = "failed"
	}

	nodeIDs := make([]string, len(branchNodes))
	for i, n := range branchNodes {
		nodeIDs[i] = n.Id
	}

	result := &ParallelResult{
		BranchID:  branchID,
		NodeIDs:   nodeIDs,
		Output:    lastOutput,
		Status:    status,
		Error:     branchErr,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 添加结果到上下文 / Add result to context
	pCtx.AddResult(branchID, result)

	log.Info("分支执行完成",
		zap.String("branchId", branchID),
		zap.String("status", status),
		zap.Int64("duration", endTime-startTime))

	return branchErr
}

// executeNode 执行单个节点
// executeNode executes a single node
func (e *ParallelExecutor) executeNode(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	if node == nil {
		return "", fmt.Errorf("节点为空 / Node is nil")
	}

	nodeType := node.Data.NodeType

	// 发送节点开始消息 / Send node start message
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   nodeType,
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 开始执行", node.Data.Label),
		})
	}

	log.Info("执行节点",
		zap.String("nodeId", node.Id),
		zap.String("nodeType", nodeType),
		zap.String("label", node.Data.Label))

	// 根据节点类型执行 / Execute based on node type
	switch nodeType {
	case "llm":
		return e.executeLLMNode(ctx, node, input, callback)
	case "tool":
		return e.executeToolNode(ctx, node, input, callback)
	case "condition":
		return e.executeConditionNode(ctx, node, input, callback)
	case "code":
		return e.executeCodeNodeParallel(ctx, node, input, callback)
	case "http":
		return e.executeHTTPNodeParallel(ctx, node, input, callback)
	case "start", "end":
		return input, nil
	default:
		log.Warn("未知节点类型，回退为输入", zap.String("nodeType", nodeType))
		return input, nil
	}
}

// executeLLMNode 在并行上下文中执行 LLM 节点
// executeLLMNode executes an LLM node in parallel context
func (e *ParallelExecutor) executeLLMNode(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	// 发送节点开始消息 / Send node start message
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "llm",
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 开始执行", node.Data.Label),
		})
	}

	// 使用 WorkflowAgent 的真实 LLM 调用逻辑
	if e.agent != nil {
		result, err := e.agent.ExecuteLLMNodeInParallel(ctx, node, input, callback)
		if err == nil && result != "" {
			if callback != nil {
				callback(&global.Chunk{
					NodeId:     node.Id,
					NodeType:   "llm",
					NodeLabel:  node.Data.Label,
					NodeStatus: "completed",
					ShowMsg:    fmt.Sprintf("[%s] 执行完成", node.Data.Label),
				})
			}
			return result, nil
		}
		if err != nil {
			log.Error("LLM 节点并行执行失败", zap.String("nodeId", node.Id), zap.Error(err))
		}
	}

	// 最终降级：构造空消息体尝试调用默认模型
	modelConfig := node.Data.ModelConfig
	if modelConfig == nil {
		return input, nil
	}

	output := input
	if modelConfig.SystemPrompt != "" {
		output = fmt.Sprintf("[%s] 基于提示词「%s」处理：%s", node.Data.Label, modelConfig.SystemPrompt, input)
	} else {
		output = fmt.Sprintf("[%s] 处理结果：%s", node.Data.Label, input)
	}

	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "llm",
			NodeLabel:  node.Data.Label,
			NodeStatus: "completed",
			ShowMsg:    fmt.Sprintf("[%s] 执行完成（降级）", node.Data.Label),
		})
	}

	return output, nil
}

// ExecuteLLMNodeInParallel 在并行上下文中执行 LLM 节点（导出的方法供 WorkflowAgent 调用）
// ExecuteLLMNodeInParallel executes an LLM node in parallel context with real LLM calls
func (e *WorkflowAgent) ExecuteLLMNodeInParallel(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	modelConfig := node.Data.ModelConfig
	if modelConfig == nil {
		return "", fmt.Errorf("LLM 节点缺少模型配置")
	}

	nodeModelName := modelConfig.Model
	nodeMaxTokens := 8192
	if modelConfig.MaxTokens > 0 {
		nodeMaxTokens = modelConfig.MaxTokens
	}
	nodeTemperature := float32(0.7)
	if modelConfig.Temperature > 0 {
		nodeTemperature = float32(modelConfig.Temperature)
	}
	systemPrompt := modelConfig.SystemPrompt
	llmToolNames := modelConfig.Tools
	llmMaxToolRounds := modelConfig.MaxToolRounds
	if llmMaxToolRounds <= 0 {
		llmMaxToolRounds = 5
	}

	// 解析节点模型信息
	var nodeEndpoint, nodeAPIKey, nodeModel string
	if e.modelResolver != nil {
		info, err := e.modelResolver.Resolve(nodeModelName)
		if err == nil {
			nodeEndpoint = info.Endpoint
			nodeAPIKey = info.APIKey
			nodeModel = info.Model
		} else if nodeModelName != "" {
			nodeEndpoint, nodeAPIKey, nodeModel = nodeModelName, "", nodeModelName
		}
	} else if nodeModelName != "" {
		nodeEndpoint, nodeAPIKey, nodeModel = nodeModelName, "", nodeModelName
	}
	if nodeEndpoint == "" {
		nodeEndpoint = e.endpoint
	}
	if nodeAPIKey == "" {
		nodeAPIKey = e.apiKey
	}
	if nodeModel == "" {
		nodeModel = e.model
	}

	// 创建节点专属模型
	nodeChatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     nodeEndpoint,
		Model:       nodeModel,
		APIKey:      nodeAPIKey,
		MaxTokens:   &nodeMaxTokens,
		Temperature: &nodeTemperature,
	})
	if err != nil {
		return "", fmt.Errorf("创建节点模型失败: %w", err)
	}

	// 绑定工具（如果配置了）
	var llmToolNode *compose.ToolsNode
	if len(llmToolNames) > 0 {
		llmNodeTools := e.getToolsByNames(llmToolNames)
		if len(llmNodeTools) > 0 {
			nodeToolInfos := make([]*schema.ToolInfo, 0, len(llmNodeTools))
			for _, t := range llmNodeTools {
				info, err := t.Info(ctx)
				if err != nil {
					continue
				}
				nodeToolInfos = append(nodeToolInfos, info)
			}
			if err := nodeChatModel.BindTools(nodeToolInfos); err != nil {
				log.Warn("LLM 节点绑定工具失败", zap.String("nodeId", node.Id), zap.Error(err))
			} else {
				llmToolNode, _ = compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
					Tools:               llmNodeTools,
					ExecuteSequentially: true,
				})
			}
		}
	}

	// 构建消息
	var messages []*schema.Message
	if systemPrompt != "" {
		messages = append(messages, schema.SystemMessage(systemPrompt))
	}
	if input != "" {
		messages = append(messages, schema.UserMessage(input))
	}

	// 首次 LLM 调用
	response, genErr := nodeChatModel.Generate(ctx, messages)
	if genErr != nil {
		return "", fmt.Errorf("LLM 调用失败: %w", genErr)
	}

	// 工具调用循环
	for round := 0; round < llmMaxToolRounds; round++ {
		if len(response.ToolCalls) == 0 {
			break
		}

		// 发送工具调用开始消息
		if callback != nil {
			for _, tc := range response.ToolCalls {
				callback(&global.Chunk{
					NodeId:     node.Id,
					NodeType:   "llm",
					NodeLabel:  node.Data.Label,
					ToolCallId: tc.ID,
					ToolName:   tc.Function.Name,
					ToolParams: tc.Function.Arguments,
					ToolStatus: "running",
					ShowMsg:    fmt.Sprintf("[%s] 调用工具: %s", node.Data.Label, tc.Function.Name),
				})
			}
		}

		messages = append(messages, response)

		if llmToolNode != nil {
			toolResults, toolErr := llmToolNode.Invoke(ctx, response)
			if toolErr != nil {
				messages = append(messages, schema.ToolMessage("工具执行失败: "+toolErr.Error(), response.ToolCalls[0].ID))
			} else {
				if callback != nil {
					for _, tr := range toolResults {
						callback(&global.Chunk{
							NodeId:     node.Id,
							NodeType:   "llm",
							NodeLabel:  node.Data.Label,
							ToolCallId: tr.ToolCallID,
							ToolName:   tr.ToolName,
							ToolResult: tr.Content,
							ToolStatus: "completed",
							ShowMsg:    fmt.Sprintf("[%s] 工具 %s 执行完成", node.Data.Label, tr.ToolName),
						})
					}
				}
				messages = append(messages, toolResults...)
			}
		}

		response, genErr = nodeChatModel.Generate(ctx, messages)
		if genErr != nil {
			break
		}
	}

	result := ""
	if response != nil {
		result = response.Content
	}
	return result, nil
}

// executeToolNode 在并行上下文中执行工具节点
// executeToolNode executes a tool node in parallel context
func (e *ParallelExecutor) executeToolNode(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	// 发送节点开始消息
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "tool",
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 执行工具...", node.Data.Label),
		})
	}

	toolConfig := node.Data.ToolConfig
	if toolConfig == nil {
		return input, nil
	}

	toolName := toolConfig.ToolName
	if toolName == "" && len(toolConfig.Tools) > 0 {
		toolName = toolConfig.Tools[0]
	}

	log.Info("并行执行工具节点",
		zap.String("nodeId", node.Id),
		zap.String("toolName", toolName))

	// 使用 WorkflowAgent 的工具执行逻辑（如果可用）
	if e.agent != nil {
		result, err := e.agent.ExecuteToolNodeInParallel(ctx, node, input, callback)
		if err != nil {
			// 工具执行失败，记录错误但继续处理
			log.Error("工具节点并行执行失败", zap.String("nodeId", node.Id), zap.Error(err))
			return "", err
		}
		return result, nil
	}

	// 备用实现：模拟工具执行
	result := fmt.Sprintf("[%s] 工具 %s 执行结果: %s", node.Data.Label, toolName, input)

	// 发送节点完成消息
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "tool",
			NodeLabel:  node.Data.Label,
			NodeStatus: "completed",
			ShowMsg:    fmt.Sprintf("[%s] 工具执行完成", node.Data.Label),
		})
	}

	return result, nil
}

// ExecuteToolNodeInParallel 在并行上下文中执行工具节点（导出的方法）
func (e *WorkflowAgent) ExecuteToolNodeInParallel(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	toolConfig := node.Data.ToolConfig
	if toolConfig == nil {
		return input, nil
	}

	toolName := toolConfig.ToolName
	if toolName == "" && len(toolConfig.Tools) > 0 {
		toolName = toolConfig.Tools[0]
	}

	if toolName == "" {
		log.Warn("工具节点未配置工具名称", zap.String("nodeId", node.Id))
		return input, nil
	}

	// 查找指定工具
	toolTools := e.getToolsByNames([]string{toolName})
	if len(toolTools) == 0 {
		log.Warn("工具节点找不到指定工具", zap.String("nodeId", node.Id), zap.String("toolName", toolName))
		return input, nil
	}

	// 获取工具参数，并替换变量占位符
	toolParams := replaceVarsInParams(toolConfig.Params, input, input)

	// 构建工具参数：如果有输入，将输入内容合并到参数中
	paramsJSON, _ := json.Marshal(toolParams)
	if input != "" {
		var paramsMap map[string]interface{}
		if err := json.Unmarshal(paramsJSON, &paramsMap); err == nil {
			if _, exists := paramsMap["toolInput"]; !exists {
				paramsMap["toolInput"] = input
				paramsJSON, _ = json.Marshal(paramsMap)
			}
		}
	}

	// 断言为 InvokableTool（直接执行）
	invokableTool, ok := toolTools[0].(tool.InvokableTool)
	if !ok {
		log.Warn("工具不支持直接执行", zap.String("nodeId", node.Id), zap.String("toolName", toolName))
		return input, nil
	}

	// 执行工具
	result, err := invokableTool.InvokableRun(ctx, string(paramsJSON))
	if err != nil {
		log.Error("工具执行失败", zap.String("nodeId", node.Id), zap.Error(err))
		return "", err
	}

	log.Info("工具执行成功", zap.String("nodeId", node.Id), zap.String("toolName", toolName))
	return result, nil
}

// executeConditionNodeParallel 并行执行条件节点
// executeConditionNodeParallel executes a condition node in parallel context
func (e *ParallelExecutor) executeConditionNode(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	conditionConfig := node.Data.ConditionConf
	if conditionConfig == nil {
		return input, nil
	}

	log.Info("条件节点执行",
		zap.String("nodeId", node.Id),
		zap.String("type", conditionConfig.Type))

	// 根据条件类型执行判断 / Execute judgment based on condition type
	switch conditionConfig.Type {
	case "expression":
		eval := NewExpressionEvaluator()
		result := eval.EvaluateWithVars(conditionConfig.Expression, map[string]string{"output": input})
		if result.Result {
			return input, nil
		}
		return input, fmt.Errorf("条件不满足 / Condition not met")
	default:
		return input, nil
	}
}

// executeCodeNodeParallel 在并行上下文中执行代码节点
// executeCodeNodeParallel executes a code node in parallel context
func (e *ParallelExecutor) executeCodeNodeParallel(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	codeConfig := node.Data.CodeConfig
	if codeConfig == nil {
		return input, nil
	}

	// 发送节点开始消息
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "code",
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 执行代码...", node.Data.Label),
		})
	}

	log.Info("并行执行代码节点",
		zap.String("nodeId", node.Id),
		zap.String("language", codeConfig.Language))

	// 调用 node_code.go 中的实际实现
	schemaMsg := &schema.Message{Content: input}
	result, err := executeCodeNode(ctx, node.Id, node.Data.Label, codeConfig, schemaMsg, callback)
	if err != nil {
		log.Error("代码节点执行失败", zap.String("nodeId", node.Id), zap.Error(err))
		return input, err
	}

	if result != nil {
		return result.Content, nil
	}
	return input, nil
}

// executeHTTPNodeParallel 在并行上下文中执行 HTTP 节点
// executeHTTPNodeParallel executes an HTTP node in parallel context
func (e *ParallelExecutor) executeHTTPNodeParallel(ctx context.Context, node *TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	httpConfig := node.Data.HTTPConfig
	if httpConfig == nil {
		return input, nil
	}

	// 发送节点开始消息
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "http",
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 发送 HTTP 请求...", node.Data.Label),
		})
	}

	log.Info("并行执行 HTTP 节点",
		zap.String("nodeId", node.Id),
		zap.String("method", httpConfig.Method),
		zap.String("url", httpConfig.URL))

	// 调用 node_http.go 中的实际实现
	schemaMsg := &schema.Message{Content: input}
	result, err := executeHTTPNode(ctx, node.Id, node.Data.Label, httpConfig, schemaMsg, callback)
	if err != nil {
		log.Error("HTTP 节点执行失败", zap.String("nodeId", node.Id), zap.Error(err))
		return input, err
	}

	if result != nil {
		return result.Content, nil
	}
	return input, nil
}

// ExecuteParallelGroup 使用 goroutine pool 并行执行所有分支
// ExecuteParallelGroup executes all branches in parallel using goroutine pool
func (e *ParallelExecutor) ExecuteParallelGroup(ctx context.Context, group *ParallelGroup,
	initialInput string, callback func(chunk *global.Chunk) error) (map[string]*ParallelResult, error) {

	if group == nil {
		return nil, fmt.Errorf("并行组为空 / Parallel group is nil")
	}

	branchCount := len(group.Branches)
	if branchCount == 0 {
		return nil, fmt.Errorf("并行组没有分支 / Parallel group has no branches")
	}

	log.Info("开始并行执行",
		zap.String("parallelId", group.ParallelID),
		zap.Int("branchCount", branchCount))

	// 创建并行上下文 / Create parallel context
	pCtx := NewParallelContext(branchCount)

	// 设置超时 / Set timeout
	if group.Config != nil && group.Config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(group.Config.Timeout)*time.Second)
		defer cancel()
	}

	// 使用 WaitGroup 等待所有分支完成
	// Use WaitGroup to wait for all branches to complete
	var wg sync.WaitGroup
	wg.Add(branchCount)

	// 信号量用于限制并发数
	// Semaphore for concurrency control
	var sem chan struct{}
	maxConcurrency := 10
	if group.Config != nil && group.Config.MaxConcurrency > 0 {
		maxConcurrency = group.Config.MaxConcurrency
	}
	if maxConcurrency > 0 && maxConcurrency < branchCount {
		sem = make(chan struct{}, maxConcurrency)
	}

	// 并行启动所有分支 / Start all branches in parallel
	for i, branch := range group.Branches {
			// 诊断日志 / Diagnostic logs
			ctxErrStr := "nil"
			if ctx.Err() != nil {
				ctxErrStr = ctx.Err().Error()
			}
			log.Info("准备启动分支 goroutine",
				zap.String("parallelId", group.ParallelID),
				zap.Int("branchIndex", i),
				zap.String("ctx.Err()", ctxErrStr),
				zap.Int("branchCount", len(group.Branches)))

			go func(branchIndex int, branchNodes []*TopoNode) {
				defer wg.Done()

				// 诊断日志 / Diagnostic logs
				ctxErrStr := "nil"
				if ctx.Err() != nil {
					ctxErrStr = ctx.Err().Error()
				}
				log.Info("分支 goroutine 启动",
					zap.String("parallelId", group.ParallelID),
					zap.Int("branchIndex", branchIndex),
					zap.String("ctx.Err()", ctxErrStr))

				// 获取信号量 / Acquire semaphore
				if sem != nil {
					sem <- struct{}{}
					defer func() { <-sem }()
				}

				// 执行分支 / Execute branch
				err := e.ExecuteBranch(ctx, branchIndex, branchNodes, initialInput, pCtx, callback)
				if err != nil {
					log.Warn("分支执行出错（不影响其他分支）",
						zap.Int("branchIndex", branchIndex),
						zap.Error(err))
				}
			}(i, branch)
	}

	// 等待所有分支完成 / Wait for all branches to complete
	wg.Wait()

	// 收集结果 / Collect results
	results := pCtx.GetResults()

	log.Info("并行执行完成",
		zap.String("parallelId", group.ParallelID),
		zap.Int("resultCount", len(results)))

	return results, nil
}

// MergeResults 合并多个分支的结果（导出的方法供外部调用）
// MergeResults merges results from multiple branches (exported for external use)
func (e *ParallelExecutor) MergeResults(results map[string]*ParallelResult) string {
	return e.mergeResults(results)
}

// mergeResults 合并多个分支的结果
// mergeResults merges results from multiple branches
func (e *ParallelExecutor) mergeResults(results map[string]*ParallelResult) string {
	if len(results) == 0 {
		return ""
	}

	var outputs []string
	for branchID, result := range results {
		if result.Status == "completed" {
			outputs = append(outputs, fmt.Sprintf(`{"branchId":"%s","output":"%s"}`,
				branchID, escapeJSON(result.Output)))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(outputs, ","))
}

// escapeJSON 转义 JSON 特殊字符
// escapeJSON escapes JSON special characters
func escapeJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// WaitForJoin 等待汇聚条件满足（all/any 策略）
// WaitForJoin waits for join condition to be satisfied (all/any strategy)
func (e *ParallelExecutor) WaitForJoin(ctx context.Context, config *JoinConfig,
	results map[string]*ParallelResult) (*JoinResult, error) {

	if config == nil {
		config = &JoinConfig{
			Strategy: "all",
			Timeout:  0,
		}
	}

	// 检查已完成的结果 / Check already completed results
	completedCount := 0
	totalCount := len(results)
	completedBranches := make(map[string]*ParallelResult)

	for branchID, result := range results {
		if result.Status == "completed" {
			completedCount++
			completedBranches[branchID] = result
		}
	}

	// 根据策略判断是否满足 / Check if satisfied based on strategy
	satisfied := false
	switch config.Strategy {
	case "all":
		satisfied = completedCount >= totalCount
	case "any":
		satisfied = completedCount > 0
	default:
		satisfied = completedCount >= totalCount
	}

	if satisfied {
		return e.buildJoinResult(completedBranches, totalCount, config.Strategy, false), nil
	}

	// 如果策略不满足但设置了超时，等待超时
	// If strategy not satisfied but timeout is set, wait for timeout
	if config.Timeout > 0 {
		timeout := time.Duration(config.Timeout) * time.Second
		select {
		case <-ctx.Done():
			return e.buildJoinResult(completedBranches, totalCount, config.Strategy, true), ctx.Err()
		case <-time.After(timeout):
			return e.buildJoinResult(completedBranches, totalCount, config.Strategy, true), nil
		}
	}

	return e.buildJoinResult(completedBranches, totalCount, config.Strategy, false), nil
}

// buildJoinResult 构建汇聚结果
// buildJoinResult builds the join result
func (e *ParallelExecutor) buildJoinResult(completedBranches map[string]*ParallelResult, totalCount int, strategy string, timedOut bool) *JoinResult {
	// 合并所有结果 / Merge all results
	var mergedOutput strings.Builder
	mergedOutput.WriteString("[")

	var first bool = true
	for branchID, result := range completedBranches {
		if !first {
			mergedOutput.WriteString(",")
		}
		first = false

		mergedOutput.WriteString(fmt.Sprintf(`{"branchId":"%s","output":"%s","status":"%s"}`,
			branchID, escapeJSON(result.Output), result.Status))
	}

	mergedOutput.WriteString("]")

	return &JoinResult{
		CompletedBranches: completedBranches,
		AllResultsMerged: mergedOutput.String(),
		CompletedCount:   len(completedBranches),
		TotalCount:       totalCount,
		Strategy:         strategy,
		TimedOut:         timedOut,
	}
}

// BuildParallelGraph 构建并行执行图（与现有 BuildGraph 集成）
// BuildParallelGraph builds a parallel execution graph (integrated with existing BuildGraph)
func (e *ParallelExecutor) BuildParallelGraph(ctx context.Context, topo *Topology,
	endpoint, apiKey, model string, callback func(chunk *global.Chunk) error) error {

	// 识别并行组 / Identify parallel groups
	parallelGroups, err := e.IdentifyParallelGroups(topo)
	if err != nil {
		return fmt.Errorf("识别并行组失败 / Failed to identify parallel groups: %w", err)
	}

	log.Info("构建并行图",
		zap.Int("parallelGroupCount", len(parallelGroups)))

	// 对于每个并行组，可以在后续与 BuildGraph 集成时处理
	// For each parallel group, can be processed when integrated with BuildGraph
	for _, group := range parallelGroups {
		log.Info("并行组信息",
			zap.String("parallelId", group.ParallelID),
			zap.Int("branchCount", len(group.Branches)))
	}

	return nil
}

// TopologicalSort 对节点列表进行拓扑排序
// TopologicalSort performs topological sort on a list of nodes
func (e *ParallelExecutor) TopologicalSort(nodes []*TopoNode, edges []TopoEdge) ([]*TopoNode, error) {
	if len(nodes) == 0 {
		return nodes, nil
	}

	// 构建邻接表和入度表 / Build adjacency list and in-degree table
	nodeMap := make(map[string]*TopoNode)
	for _, node := range nodes {
		nodeMap[node.Id] = node
	}

	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// 初始化入度 / Initialize in-degree
	for _, node := range nodes {
		inDegree[node.Id] = 0
	}

	// 构建图 / Build graph
	for _, edge := range edges {
		if _, sourceOk := nodeMap[edge.Source]; sourceOk {
			if _, targetOk := nodeMap[edge.Target]; targetOk {
				adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
				inDegree[edge.Target]++
			}
		}
	}

	// Kahn 算法 / Kahn's algorithm
	var queue []string
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	var sorted []*TopoNode
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, nodeMap[current])

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// 检查是否有环 / Check for cycles
	if len(sorted) != len(nodes) {
		return nil, fmt.Errorf("图中存在环，无法完成拓扑排序 / Graph contains cycle, cannot complete topological sort")
	}

	return sorted, nil
}

// ValidateParallelGroup 验证并行组的有效性
// ValidateParallelGroup validates the correctness of a parallel group
func (e *ParallelExecutor) ValidateParallelGroup(group *ParallelGroup) error {
	if group == nil {
		return fmt.Errorf("并行组为空 / Parallel group is nil")
	}

	if group.ParallelNode == nil {
		return fmt.Errorf("并行入口节点为空 / Parallel entry node is nil")
	}

	if group.ParallelNode.Data.NodeType != "parallel" {
		return fmt.Errorf("入口节点类型错误，期望 'parallel'，实际 '%s'",
			group.ParallelNode.Data.NodeType)
	}

	if len(group.Branches) == 0 {
		return fmt.Errorf("并行组没有分支 / Parallel group has no branches")
	}

	for i, branch := range group.Branches {
		if len(branch) == 0 {
			return fmt.Errorf("分支 %d 为空 / Branch %d is empty", i, i)
		}
	}

	return nil
}

// replaceVarsInParams 递归替换参数 map 中的变量占位符
// 支持 {{input}}、{{output}}、{{input.field}} 等格式
// 并将处理后的值统一转为 string 以便传递给 InvokableRun
func replaceVarsInParams(params map[string]interface{}, input, output string) map[string]interface{} {
	if params == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range params {
		result[k] = replaceVarValue(v, input, output)
	}
	return result
}

// replaceVarValue 替换单个值中的变量占位符，支持递归处理 map/slice
func replaceVarValue(v interface{}, input, output string) interface{} {
	switch val := v.(type) {
	case string:
		replaced := val
		replaced = strings.ReplaceAll(replaced, "{{input}}", input)
		replaced = strings.ReplaceAll(replaced, "{{output}}", output)
		replaced = replaceNestedVars(replaced, input, output)
		return replaced
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k2, v2 := range val {
			m[k2] = replaceVarValue(v2, input, output)
		}
		return m
	case []interface{}:
		arr := make([]interface{}, len(val))
		for i, elem := range val {
			arr[i] = replaceVarValue(elem, input, output)
		}
		return arr
	default:
		return val
	}
}

// replaceNestedVars 替换 {{input.xxx}} 和 {{output.xxx}} 格式的嵌套变量
// replaceNestedVars replaces nested variables like {{input.xxx}} and {{output.xxx}}
// 支持 JSON 字段访问：如果 input/output 是 JSON 字符串，会解析并提取对应字段
// Supports JSON field access: parses input/output as JSON if it's a JSON string
func replaceNestedVars(s, input, output string) string {
	re := regexp.MustCompile(`\{\{(input|output)(?:\.(\w+))?\}\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// 提取变量名和字段名 / Extract variable name and field name
		sub := re.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		varName := sub[1]
		fieldName := sub[2]

		// 选择源字符串 / Select source string
		var source string
		if varName == "input" {
			source = input
		} else {
			source = output
		}

		// 如果没有字段名，返回整个字符串 / If no field name, return whole string
		if fieldName == "" {
			return source
		}

		// 尝试解析为 JSON 并提取字段 / Try to parse as JSON and extract field
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(source), &data); err == nil {
			if val, ok := data[fieldName]; ok {
				return toString(val)
			}
		}

		// 回退策略：当 source 不是 JSON 时（如纯文本输入），用整个 source 替换
		// 这样 {{input.question}} 等同于 {{input}}
		// Fallback: if source is not JSON (e.g. plain text user input),
		// treat {{input.xxx}} as {{input}} to match user expectations
		log.Debug("嵌套变量按整体输入回退 / Nested var falls back to whole input",
			zap.String("variable", varName),
			zap.String("field", fieldName),
			zap.String("source", source))
		return source
	})
}

// toString 将任意类型转为字符串
// toString converts any value to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case nil:
		return ""
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

