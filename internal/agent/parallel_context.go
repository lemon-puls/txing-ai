package agent

import (
	"context"
	"sync"
)

// ParallelContext 并行执行上下文结构体，管理并行分支的结果收集和状态
// ParallelContext manages parallel execution context, including result collection and state tracking
type ParallelContext struct {
	mu sync.RWMutex
	results         map[string]*ParallelResult // 分支结果 / Branch results
	completedCount  int32                    // 已完成计数 / Completed count
	totalCount      int                      // 总数 / Total count
	cancelled       bool                     // 取消标志 / Cancellation flag
	resultCh         chan struct{}           // 结果变更通知 / Result change notification
	cancelFunc      context.CancelFunc       // 取消函数 / Cancellation function
}

// NewParallelContext 创建一个新的并行上下文
// NewParallelContext creates a new parallel context instance
func NewParallelContext(total int) *ParallelContext {
	return &ParallelContext{
		results:    make(map[string]*ParallelResult),
		totalCount:  total,
		resultCh:    make(chan struct{}),
	}
}

// SetCancelFunc 设置取消函数
// SetCancelFunc sets the cancellation function
func (pc *ParallelContext) SetCancelFunc(cancel context.CancelFunc) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cancelFunc = cancel
}

// AddResult 添加分支执行结果
// AddResult adds a branch execution result
func (pc *ParallelContext) AddResult(branchId string, result *ParallelResult) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.results[branchId] = result
	pc.completedCount++

	// 通知结果变更 / Notify result change
	select {
	case pc.resultCh <- struct{}{}:
	default:
	}
}

// GetResults 获取所有结果（返回副本）
// GetResults returns all results (returns a copy)
func (pc *ParallelContext) GetResults() map[string]*ParallelResult {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	// 返回副本以保证线程安全 / Return a copy for thread safety
	resultsCopy := make(map[string]*ParallelResult, len(pc.results))
	for k, v := range pc.results {
		resultsCopy[k] = v
	}
	return resultsCopy
}

// GetCompletedCount 获取已完成数量
// GetCompletedCount returns the number of completed branches
func (pc *ParallelContext) GetCompletedCount() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return int(pc.completedCount)
}

// GetTotalCount 获取总数
// GetTotalCount returns the total number of branches
func (pc *ParallelContext) GetTotalCount() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.totalCount
}

// IsAllCompleted 检查是否全部完成
// IsAllCompleted checks if all branches are completed
func (pc *ParallelContext) IsAllCompleted() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return int(pc.completedCount) >= pc.totalCount
}

// IsCancelled 检查是否已取消
// IsCancelled checks if execution was cancelled
func (pc *ParallelContext) IsCancelled() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.cancelled
}

// Cancel 取消执行
// Cancel cancels the execution
func (pc *ParallelContext) Cancel() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cancelled = true
	if pc.cancelFunc != nil {
		pc.cancelFunc()
	}
}

// ResultCh 返回结果变更 channel
// ResultCh returns the result change notification channel
func (pc *ParallelContext) ResultCh() <-chan struct{} {
	return pc.resultCh
}
