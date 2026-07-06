package types

// ParallelResult 分支执行结果结构体
// ParallelResult represents the execution result of a single branch
type ParallelResult struct {
	BranchID  string   `json:"branchId"`  // 分支ID / Branch ID
	NodeIDs   []string `json:"nodeIds"`   // 该分支包含的节点ID列表 / List of node IDs in this branch
	Output    string   `json:"output"`    // 最终输出 / Final output
	Status    string   `json:"status"`    // completed/failed/running
	Error     error    `json:"error"`     // 错误信息 / Error information
	StartTime int64    `json:"startTime"` // 开始时间戳 / Start timestamp
	EndTime   int64    `json:"endTime"`   // 结束时间戳 / End timestamp
}

// JoinResult 汇聚结果结构体
// JoinResult represents the result of a join operation
type JoinResult struct {
	CompletedBranches map[string]*ParallelResult `json:"completedBranches"` // 已完成的分支结果 / Completed branch results
	AllResultsMerged string                      `json:"allResultsMerged"`  // 合并后的结果 / Merged results
	CompletedCount   int                         `json:"completedCount"`    // 已完成数量 / Completed count
	TotalCount       int                         `json:"totalCount"`        // 总数量 / Total count
	Strategy         string                      `json:"strategy"`          // 汇聚策略 / Join strategy
	TimedOut         bool                        `json:"timedOut"`          // 是否超时 / Whether timed out
}
