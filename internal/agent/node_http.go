package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
)

// executeHTTPNode 执行 HTTP 节点
func executeHTTPNode(
	ctx context.Context,
	nodeId string,
	nodeLabel string,
	config *HTTPConfig,
	input *schema.Message,
	callback func(chunk *global.Chunk) error,
) (*schema.Message, error) {
	execLog := &NodeExecutionLog{
		NodeID:    nodeId,
		NodeType:  "http",
		NodeLabel: nodeLabel,
		StartTime: time.Now().UnixMilli(),
	}

	if callback != nil {
		callback(&global.Chunk{
			NodeId:     nodeId,
			NodeType:   "http",
			NodeLabel:  nodeLabel,
			NodeStatus: "running",
			ShowMsg:    fmt.Sprintf("[%s] 发送 HTTP 请求...", nodeLabel),
		})
	}

	log.Info("执行 HTTP 节点", zap.String("nodeId", nodeId), zap.String("method", config.Method), zap.String("url", config.URL))

	// 获取输入内容
	inputContent := ""
	if input != nil {
		inputContent = input.Content
	}
	execLog.Input = fmt.Sprintf("%s %s", config.Method, config.URL)

	// 替换 URL 和 Body 中的变量
	url := config.URL
	url = strings.ReplaceAll(url, "{{input}}", inputContent)
	url = strings.ReplaceAll(url, "{{output}}", inputContent)

	body := config.Body
	body = strings.ReplaceAll(body, "{{input}}", inputContent)
	body = strings.ReplaceAll(body, "{{output}}", inputContent)

	// 设置超时
	timeout := 30
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// 创建请求
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, url, reqBody)
	if err != nil {
		log.Error("创建 HTTP 请求失败", zap.String("nodeId", nodeId), zap.Error(err))
		execLog.Status = "failed"
		execLog.Error = err.Error()
		execLog.EndTime = time.Now().UnixMilli()
		execLog.Duration = execLog.EndTime - execLog.StartTime
		sendExecutionLog(callback, execLog)
		return nil, err
	}

	// 设置请求头
	for key, value := range config.Headers {
		value = strings.ReplaceAll(value, "{{input}}", inputContent)
		value = strings.ReplaceAll(value, "{{output}}", inputContent)
		req.Header.Set(key, value)
	}

	// 如果没有设置 Content-Type 且有 body，默认设置为 application/json
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("HTTP 请求失败", zap.String("nodeId", nodeId), zap.Error(err))
		execLog.Status = "failed"
		execLog.Error = err.Error()
		execLog.EndTime = time.Now().UnixMilli()
		execLog.Duration = execLog.EndTime - execLog.StartTime
		sendExecutionLog(callback, execLog)
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("读取 HTTP 响应失败", zap.String("nodeId", nodeId), zap.Error(err))
		execLog.Status = "failed"
		execLog.Error = err.Error()
		execLog.EndTime = time.Now().UnixMilli()
		execLog.Duration = execLog.EndTime - execLog.StartTime
		sendExecutionLog(callback, execLog)
		return nil, err
	}

	result := string(respBody)

	// 检查状态码
	if resp.StatusCode >= 400 {
		errMsg := fmt.Sprintf("HTTP 请求返回错误状态码: %d, 响应: %s", resp.StatusCode, result)
		log.Error("HTTP 请求错误", zap.String("nodeId", nodeId), zap.Int("statusCode", resp.StatusCode))
		execLog.Status = "failed"
		execLog.Error = errMsg
		execLog.EndTime = time.Now().UnixMilli()
		execLog.Duration = execLog.EndTime - execLog.StartTime
		sendExecutionLog(callback, execLog)
		return nil, fmt.Errorf("%s", errMsg)
	}

	execLog.Status = "completed"
	execLog.Output = result
	execLog.EndTime = time.Now().UnixMilli()
	execLog.Duration = execLog.EndTime - execLog.StartTime
	sendExecutionLog(callback, execLog)

	if callback != nil {
		callback(&global.Chunk{
			NodeId:     nodeId,
			NodeType:   "http",
			NodeLabel:  nodeLabel,
			NodeStatus: "completed",
			ShowMsg:    fmt.Sprintf("[%s] HTTP 请求完成 (状态码: %d)", nodeLabel, resp.StatusCode),
		})
	}

	log.Info("HTTP 请求完成", zap.String("nodeId", nodeId), zap.Int("statusCode", resp.StatusCode))

	return schema.AssistantMessage(result, nil), nil
}
