package workflow

import (
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/agent/workflow/parallel"
	"txing-ai/internal/agent/workflow/types"
	"txing-ai/internal/global/logging/log"
)

// buildParallelJoinEdges 构建并行节点到汇聚节点的边
func buildParallelJoinEdges(graph *compose.Graph[[]*schema.Message, *schema.Message], topo *types.Topology, parallelBranchNodes map[string]struct{}) {
	// 先一次性算出所有并行组
	groups, _ := parallel.NewParallelExecutor(nil, 0, "", "", "").IdentifyParallelGroups(topo)

	// branchId -> parallelId
	branchToParallel := make(map[string]string)
	for _, g := range groups {
		for _, branch := range g.Branches {
			for _, b := range branch {
				branchToParallel[b.Id] = g.ParallelNode.Id
			}
		}
	}

	// 用 set 去重：每个 (parallelId, joinId) 只加一次
	addedPairs := make(map[string]struct{})

	// 对每条 (branch -> X) 边，如果 branch 属于某个 parallel，则加 (parallel -> X) 边
	for _, edge := range topo.Edges {
		if _, isBranch := parallelBranchNodes[edge.Source]; !isBranch {
			continue
		}
		parallelId, ok := branchToParallel[edge.Source]
		if !ok {
			continue
		}
		pairKey := parallelId + "->" + edge.Target
		if _, already := addedPairs[pairKey]; already {
			continue
		}
		addedPairs[pairKey] = struct{}{}

		err := graph.AddEdge(parallelId, edge.Target)
		if err != nil {
			log.Warn("添加并行→汇聚边失败",
				zap.String("source", parallelId),
				zap.String("target", edge.Target),
				zap.Error(err))
			continue
		}
		log.Info("已建立并行→汇聚边", zap.String("source", parallelId), zap.String("target", edge.Target))
	}
}
