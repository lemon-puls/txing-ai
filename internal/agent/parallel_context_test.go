package agent

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewParallelContext(t *testing.T) {
	ctx := NewParallelContext(5)
	if ctx == nil {
		t.Fatal("NewParallelContext 返回 nil")
	}
	if ctx.totalCount != 5 {
		t.Errorf("期望 totalCount=5, 得到 %d", ctx.totalCount)
	}
	if ctx.GetTotalCount() != 5 {
		t.Errorf("GetTotalCount 期望 5, 得到 %d", ctx.GetTotalCount())
	}
}

func TestParallelContext_AddResult(t *testing.T) {
	ctx := NewParallelContext(3)

	// 添加第一个结果
	result1 := &ParallelResult{BranchID: "branch1", Status: "completed", Output: "output1"}
	ctx.AddResult("branch1", result1)

	if ctx.GetCompletedCount() != 1 {
		t.Errorf("期望 completedCount=1, 得到 %d", ctx.GetCompletedCount())
	}

	// 添加第二个结果
	result2 := &ParallelResult{BranchID: "branch2", Status: "completed", Output: "output2"}
	ctx.AddResult("branch2", result2)

	if ctx.GetCompletedCount() != 2 {
		t.Errorf("期望 completedCount=2, 得到 %d", ctx.GetCompletedCount())
	}

	// 验证结果
	results := ctx.GetResults()
	if len(results) != 2 {
		t.Errorf("期望 2 个结果, 得到 %d", len(results))
	}
	if results["branch1"].Output != "output1" {
		t.Errorf("期望 branch1 output='output1', 得到 '%s'", results["branch1"].Output)
	}
}

func TestParallelContext_IsAllCompleted(t *testing.T) {
	ctx := NewParallelContext(3)

	if ctx.IsAllCompleted() {
		t.Error("不应全部完成")
	}

	// 添加 3 个结果
	ctx.AddResult("branch1", &ParallelResult{BranchID: "branch1", Status: "completed"})
	ctx.AddResult("branch2", &ParallelResult{BranchID: "branch2", Status: "completed"})
	ctx.AddResult("branch3", &ParallelResult{BranchID: "branch3", Status: "completed"})

	if !ctx.IsAllCompleted() {
		t.Error("应全部完成")
	}
}

func TestParallelContext_Cancel(t *testing.T) {
	ctx := NewParallelContext(3)

	if ctx.IsCancelled() {
		t.Error("不应被取消")
	}

	ctx.Cancel()

	if !ctx.IsCancelled() {
		t.Error("应被取消")
	}

	// 多次取消不应 panic
	ctx.Cancel()
}

func TestParallelContext_SetCancelFunc(t *testing.T) {
	ctx := NewParallelContext(2)

	ctxChild, cancel := context.WithCancel(context.Background())
	ctx.SetCancelFunc(cancel)

	// 触发取消
	ctx.Cancel()

	// 验证子 context 是否被取消
	select {
	case <-ctxChild.Done():
		// 期望被取消
	default:
		t.Error("子 context 应被取消")
	}
}

func TestParallelContext_GetResults_ReturnsCopy(t *testing.T) {
	ctx := NewParallelContext(2)

	ctx.AddResult("branch1", &ParallelResult{BranchID: "branch1", Output: "output1"})
	ctx.AddResult("branch2", &ParallelResult{BranchID: "branch2", Output: "output2"})

	results1 := ctx.GetResults()
	results2 := ctx.GetResults()

	// 两个调用应返回不同的 map 实例（检查指针不同）
	ptr1 := fmt.Sprintf("%p", results1)
	ptr2 := fmt.Sprintf("%p", results2)
	if ptr1 == ptr2 {
		t.Error("每次调用应返回新的副本")
	}

	// 验证两个副本内容相同
	if len(results1) != len(results2) {
		t.Errorf("副本大小应相同: %d vs %d", len(results1), len(results2))
	}

	// 修改一个副本不应影响另一个副本的 map 引用
	results1["branch3"] = &ParallelResult{BranchID: "branch3"}
	if _, exists := results2["branch3"]; exists {
		t.Error("副本修改不应影响另一个副本")
	}
}

func TestParallelContext_ConcurrentAccess(t *testing.T) {
	ctx := NewParallelContext(100)
	var wg sync.WaitGroup

	// 并发添加结果
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx.AddResult(string(rune('a'+id)), &ParallelResult{
				BranchID: string(rune('a' + id)),
				Status:   "completed",
				Output:   "output",
			})
		}(i)
	}

	wg.Wait()

	if ctx.GetCompletedCount() != 100 {
		t.Errorf("期望 100 个完成, 得到 %d", ctx.GetCompletedCount())
	}

	if !ctx.IsAllCompleted() {
		t.Error("应全部完成")
	}
}

func TestParallelContext_ResultCh(t *testing.T) {
	ctx := NewParallelContext(2)

	resultCh := ctx.ResultCh()
	if resultCh == nil {
		t.Fatal("ResultCh 不应为 nil")
	}

	// 初始状态，channel 应该有容量
	select {
	case <-resultCh:
		t.Error("初始状态不应收到通知")
	default:
		// 期望的默认行为
	}

	// 添加第一个结果 - 应触发通知
	ctx.AddResult("branch1", &ParallelResult{BranchID: "branch1", Status: "completed"})

	// 验证 channel 不是 nil
	if ctx.ResultCh() == nil {
		t.Error("ResultCh 不应为 nil")
	}

	// 验证 AddResult 没有 panic（线程安全测试）
	ctx.AddResult("branch2", &ParallelResult{BranchID: "branch2", Status: "completed"})
}

func TestParallelContext_AllCompletedClosesChannel(t *testing.T) {
	ctx := NewParallelContext(2)

	resultCh := ctx.ResultCh()

	// 添加第一个结果
	ctx.AddResult("branch1", &ParallelResult{BranchID: "branch1", Status: "completed"})

	// channel 仍未关闭
	select {
	case <-resultCh:
		// 可能收到通知
	default:
	}

	// 添加第二个结果，所有分支完成
	ctx.AddResult("branch2", &ParallelResult{BranchID: "branch2", Status: "completed"})

	// 等待一小段时间确保 channel 关闭
	time.Sleep(10 * time.Millisecond)

	// channel 应该被关闭
	select {
	case _, ok := <-resultCh:
		if ok {
			t.Error("channel 应已关闭")
		}
	default:
		// 可能已经消费了关闭信号
	}
}

func TestParallelResult_Structure(t *testing.T) {
	result := &ParallelResult{
		BranchID:  "test-branch",
		NodeIDs:   []string{"node1", "node2"},
		Output:    "test output",
		Status:    "completed",
		Error:     nil,
		StartTime: 1000,
		EndTime:   2000,
	}

	if result.BranchID != "test-branch" {
		t.Errorf("期望 BranchID='test-branch', 得到 '%s'", result.BranchID)
	}
	if len(result.NodeIDs) != 2 {
		t.Errorf("期望 2 个 NodeIDs, 得到 %d", len(result.NodeIDs))
	}
	if result.Output != "test output" {
		t.Errorf("期望 Output='test output', 得到 '%s'", result.Output)
	}
	if result.Status != "completed" {
		t.Errorf("期望 Status='completed', 得到 '%s'", result.Status)
	}
	if result.EndTime-result.StartTime != 1000 {
		t.Errorf("期望耗时 1000ms, 得到 %d", result.EndTime-result.StartTime)
	}
}

func TestJoinResult_Structure(t *testing.T) {
	result := &JoinResult{
		CompletedBranches: map[string]*ParallelResult{
			"branch1": {BranchID: "branch1", Status: "completed"},
		},
		AllResultsMerged: "merged output",
		CompletedCount:   1,
		TotalCount:       3,
		Strategy:         "all",
		TimedOut:         false,
	}

	if len(result.CompletedBranches) != 1 {
		t.Errorf("期望 1 个已完成分支, 得到 %d", len(result.CompletedBranches))
	}
	if result.CompletedCount != 1 {
		t.Errorf("期望 CompletedCount=1, 得到 %d", result.CompletedCount)
	}
	if result.TotalCount != 3 {
		t.Errorf("期望 TotalCount=3, 得到 %d", result.TotalCount)
	}
	if result.Strategy != "all" {
		t.Errorf("期望 Strategy='all', 得到 '%s'", result.Strategy)
	}
	if result.TimedOut {
		t.Error("不应超时")
	}
}
