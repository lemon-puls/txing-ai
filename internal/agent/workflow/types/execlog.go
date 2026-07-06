package types

import (
	"fmt"

	"txing-ai/internal/global"
)

// NodeExecutionLog 节点执行日志
type NodeExecutionLog struct {
	NodeID    string `json:"nodeId"`
	NodeType  string `json:"nodeType"`
	NodeLabel string `json:"nodeLabel"`
	Status    string `json:"status"` // running, completed, failed
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Duration  int64  `json:"duration"` // 毫秒
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	Retry     int    `json:"retry,omitempty"` // 当前重试次数
}

// SendExecutionLog 发送执行日志到回调
func SendExecutionLog(callback func(chunk *global.Chunk) error, execLog *NodeExecutionLog) {
	if callback == nil || execLog == nil {
		return
	}
	callback(&global.Chunk{
		NodeId:     execLog.NodeID,
		NodeType:   execLog.NodeType,
		NodeLabel:  execLog.NodeLabel,
		NodeStatus: execLog.Status,
		ShowMsg:    fmt.Sprintf("[%s] %s (耗时: %dms)", execLog.NodeLabel, execLog.Status, execLog.Duration),
		ExecutionLog: &global.ExecutionLogInfo{
			StartTime: execLog.StartTime,
			EndTime:   execLog.EndTime,
			Duration:  execLog.Duration,
			Input:     execLog.Input,
			Output:    execLog.Output,
			Error:     execLog.Error,
			Retry:     execLog.Retry,
		},
	})
}
