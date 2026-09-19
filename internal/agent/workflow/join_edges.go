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
func buildParallelJoinEdges(graph *compose.Graph[[]*schema.Message, *schema.Message], topo *types.Topology) {
	// 先一次性算出所有并行组
	groups, _ := parallel.NewParallelExecutor(nil, 0, "", "", "").IdentifyParallelGroups(topo)

	// 直接从并行组信息中获取 join 节点，添加 parallel → join 边
	for _, g := range groups {
		if g.JoinNode == nil {
			log.Warn("并行组缺少 join 节点", zap.String("parallelId", g.ParallelID))
			continue
		}

		err := graph.AddEdge(g.ParallelID, g.JoinNode.Id)
		if err != nil {
			log.Warn("添加并行→汇聚边失败",
				zap.String("source", g.ParallelID),
				zap.String("target", g.JoinNode.Id),
				zap.Error(err))
			continue
		}
		log.Info("已建立并行→汇聚边", zap.String("source", g.ParallelID), zap.String("target", g.JoinNode.Id))
	}
}
