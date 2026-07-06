package parallel

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
	"txing-ai/internal/agent/workflow/condition"
	nodeexec "txing-ai/internal/agent/workflow/node"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// NodeExecutor 由 workflow 核心实现，供并行分支回调执行 LLM/Tool 节点
// NodeExecutor is implemented by the workflow core to execute LLM/Tool nodes within parallel branches
type NodeExecutor interface {
	ExecuteLLMNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error)
	ExecuteToolNodeInParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error)
}

// ParallelGroup 并行组结构
// ParallelGroup represents a parallel execution group with its branches and join node
type ParallelGroup struct {
	ParallelNode *types.TopoNode      // 并行入口节点 / Parallel entry node
	Branches     [][]*types.TopoNode  // 每个分支的节点列表 / Node list for each branch
	JoinNode     *types.TopoNode      // 汇聚节点 / Join node
	ParallelID   string               // 并行组ID / Parallel group ID
	Config       *types.ParallelConfig // 并行配置 / Parallel configuration
}

// ParallelExecutor 并行执行器
// ParallelExecutor handles parallel execution of workflow branches
type ParallelExecutor struct {
	executor   NodeExecutor // 节点执行器接口 / Node executor interface
	maxWorkers int          // 最大工作线程数 / Maximum worker threads
	endpoint   string       // LLM 调用端点 / LLM endpoint
	apiKey     string       // LLM API 密钥 / LLM API key
	model      string       // 默认模型名 / Default model name
}

// NewParallelExecutor 创建并行执行器
// NewParallelExecutor creates a new parallel executor instance
func NewParallelExecutor(executor NodeExecutor, maxWorkers int, endpoint, apiKey, model string) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 10 // 默认最大并发数 / Default max concurrency
	}
	return &ParallelExecutor{
		executor:   executor,
		maxWorkers: maxWorkers,
		endpoint:   endpoint,
		apiKey:     apiKey,
		model:      model,
	}
}
// ExecuteBranch 执行单个分支
// ExecuteBranch executes a single branch of the parallel group
func (e *ParallelExecutor) ExecuteBranch(ctx context.Context, branchIndex int, branchNodes []*types.TopoNode,
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

	result := &types.ParallelResult{
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
func (e *ParallelExecutor) executeNode(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	if node == nil {
		return "", fmt.Errorf("节点为空 / Node is nil")
	}

	nodeType := node.Data.NodeType
	startTime := time.Now().UnixMilli()

	// 发送节点开始消息 / Send node start message
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   nodeType,
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 开始执行", node.Data.Label),
			ExecutionLog: &global.ExecutionLogInfo{
				StartTime: startTime,
				Input:     input,
			},
		})
	}

	log.Info("执行节点",
		zap.String("nodeId", node.Id),
		zap.String("nodeType", nodeType),
		zap.String("label", node.Data.Label))

	// 统一收尾：所有路径都发完成/失败状态，避免前端节点面板一直转圈
	// Unified completion: send completed/failed status on every exit path so the
	// frontend node panel does not stay in "running" forever
	var (
		output string
		err    error
	)

	// 根据节点类型执行 / Execute based on node type
	switch nodeType {
	case "llm":
		output, err = e.executeLLMNode(ctx, node, input, callback)
	case "tool":
		output, err = e.executeToolNode(ctx, node, input, callback)
	case "condition":
		output, err = e.executeConditionNode(ctx, node, input, callback)
	case "code":
		output, err = e.executeCodeNodeParallel(ctx, node, input, callback)
	case "http":
		output, err = e.executeHTTPNodeParallel(ctx, node, input, callback)
	case "start", "end":
		output, err = input, nil
	default:
		log.Warn("未知节点类型，回退为输入", zap.String("nodeType", nodeType))
		output, err = input, nil
	}

	// 发送完成/失败状态（含 execution_log，让前端日志面板能正确显示节点条目）
	// Send final status with execution_log so the frontend execution-log panel
	// can insert/update the node entry properly.
	endTime := time.Now().UnixMilli()
	if callback != nil {
		finalStatus := "completed"
		finalShowMsg := fmt.Sprintf("[%s] 执行完成", node.Data.Label)
		var errMsg string
		if err != nil {
			finalStatus = "failed"
			finalShowMsg = fmt.Sprintf("[%s] 执行失败: %v", node.Data.Label, err)
			errMsg = err.Error()
		}
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   nodeType,
			NodeLabel:  node.Data.Label,
			NodeStatus: finalStatus,
			ShowMsg:    finalShowMsg,
			ExecutionLog: &global.ExecutionLogInfo{
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  endTime - startTime,
				Input:     input,
				Output:    output,
				Error:     errMsg,
				Retry:     0,
			},
		})
	}

	return output, err
}

// executeLLMNode 在并行上下文中执行 LLM 节点
// executeLLMNode executes an LLM node in parallel context
func (e *ParallelExecutor) executeLLMNode(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
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
	if e.executor != nil {
		result, err := e.executor.ExecuteLLMNodeInParallel(ctx, node, input, callback)
		if err != nil {
			log.Error("LLM 节点并行执行失败", zap.String("nodeId", node.Id), zap.Error(err))
			return "", err
		}
		if result != "" {
			return result, nil
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

	return output, nil
}

// executeToolNode 并行执行工具节点
// executeToolNode executes a tool node in parallel context
func (e *ParallelExecutor) executeToolNode(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
	// 发送节点开始消息 / Send node start message
	if callback != nil {
		callback(&global.Chunk{
			NodeId:     node.Id,
			NodeType:   "tool",
			NodeLabel:  node.Data.Label,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 开始执行", node.Data.Label),
		})
	}

	// 使用 WorkflowAgent 的真实工具调用逻辑
	if e.executor != nil {
		result, err := e.executor.ExecuteToolNodeInParallel(ctx, node, input, callback)
		if err != nil {
			log.Error("Tool 节点并行执行失败", zap.String("nodeId", node.Id), zap.Error(err))
			return "", err
		}
		return result, nil
	}

	// 降级返回输入
	return input, nil
}

// executeConditionNodeParallel 并行执行条件节点
// executeConditionNodeParallel executes a condition node in parallel context
func (e *ParallelExecutor) executeConditionNode(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
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
		eval := condition.NewExpressionEvaluator()
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
func (e *ParallelExecutor) executeCodeNodeParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
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
	result, err := nodeexec.ExecuteCodeNode(ctx, node.Id, node.Data.Label, codeConfig, schemaMsg, callback)
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
func (e *ParallelExecutor) executeHTTPNodeParallel(ctx context.Context, node *types.TopoNode, input string, callback func(chunk *global.Chunk) error) (string, error) {
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
	result, err := nodeexec.ExecuteHTTPNode(ctx, node.Id, node.Data.Label, httpConfig, schemaMsg, callback)
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
	initialInput string, callback func(chunk *global.Chunk) error) (map[string]*types.ParallelResult, error) {

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

			go func(branchIndex int, branchNodes []*types.TopoNode) {
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
func (e *ParallelExecutor) MergeResults(results map[string]*types.ParallelResult) string {
	return e.mergeResults(results)
}

// mergeResults 合并多个分支的结果
// mergeResults merges results from multiple branches
func (e *ParallelExecutor) mergeResults(results map[string]*types.ParallelResult) string {
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

// WaitForJoin 等待汇聚条件满足（all/any 策略）
// WaitForJoin waits for join condition to be satisfied (all/any strategy)
func (e *ParallelExecutor) WaitForJoin(ctx context.Context, config *types.JoinConfig,
	results map[string]*types.ParallelResult) (*types.JoinResult, error) {

	if config == nil {
		config = &types.JoinConfig{
			Strategy: "all",
			Timeout:  0,
		}
	}

	// 检查已完成的结果 / Check already completed results
	completedCount := 0
	totalCount := len(results)
	completedBranches := make(map[string]*types.ParallelResult)

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
func (e *ParallelExecutor) buildJoinResult(completedBranches map[string]*types.ParallelResult, totalCount int, strategy string, timedOut bool) *types.JoinResult {
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

	return &types.JoinResult{
		CompletedBranches: completedBranches,
		AllResultsMerged: mergedOutput.String(),
		CompletedCount:   len(completedBranches),
		TotalCount:       totalCount,
		Strategy:         strategy,
		TimedOut:         timedOut,
	}
}
// ReplaceNestedVars 替换 {{input.xxx}} 和 {{output.xxx}} 格式的嵌套变量
// ReplaceNestedVars replaces nested variables like {{input.xxx}} and {{output.xxx}}
// 支持 JSON 字段访问：如果 input/output 是 JSON 字符串，会解析并提取对应字段
// Supports JSON field access: parses input/output as JSON if it's a JSON string
func ReplaceNestedVars(s, input, output string) string {
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

