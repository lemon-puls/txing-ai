package parallel

import (
	"context"
	"fmt"
	"sort"

	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

// IdentifyParallelGroups 从拓扑中识别所有并行组
// IdentifyParallelGroups identifies all parallel groups from the topology
func (e *ParallelExecutor) IdentifyParallelGroups(topo *types.Topology) ([]*ParallelGroup, error) {
	if topo == nil {
		return nil, fmt.Errorf("拓扑图为空 / Topology is nil")
	}

	// 创建节点映射 / Create node map
	nodeMap := make(map[string]*types.TopoNode)
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
func (e *ParallelExecutor) extractParallelGroup(parallelNode *types.TopoNode, edges []types.TopoEdge, nodeMap map[string]*types.TopoNode) *ParallelGroup {
	if parallelNode == nil {
		return nil
	}

	parallelID := parallelNode.Id

	// 解析并行配置 / Parse parallel configuration
	config := &types.ParallelConfig{
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
	branchNodes := make(map[string][]*types.TopoNode) // branchId -> nodes
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
	var branches [][]*types.TopoNode
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
func (e *ParallelExecutor) traceBranchPath(startNodeID string, edges []types.TopoEdge, nodeMap map[string]*types.TopoNode, parallelID string, joinNodeID *string) []*types.TopoNode {
	var path []*types.TopoNode
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
func (e *ParallelExecutor) findBranchesByParallelId(parallelID string, nodeMap map[string]*types.TopoNode, edges []types.TopoEdge) map[string][]*types.TopoNode {
	branchNodes := make(map[string][]*types.TopoNode)

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
func (e *ParallelExecutor) getNodeParallelId(node *types.TopoNode) string {
	// 优先使用 Label 作为 parallelId / Prefer using Label as parallelId
	if node.Data.Label != "" && node.Data.Label != node.Id {
		return node.Data.Label
	}
	return node.Id
}

// BuildParallelGraph 构建并行执行图（与现有 BuildGraph 集成）
// BuildParallelGraph builds a parallel execution graph (integrated with existing BuildGraph)
func (e *ParallelExecutor) BuildParallelGraph(ctx context.Context, topo *types.Topology,
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
func (e *ParallelExecutor) TopologicalSort(nodes []*types.TopoNode, edges []types.TopoEdge) ([]*types.TopoNode, error) {
	if len(nodes) == 0 {
		return nodes, nil
	}

	// 构建邻接表和入度表 / Build adjacency list and in-degree table
	nodeMap := make(map[string]*types.TopoNode)
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

	var sorted []*types.TopoNode
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
