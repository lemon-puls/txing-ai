package tool

import (
	"errors"
)

// ToolShowMsgBuilder 定义每个工具的请求/响应展示构造接口
// 将具体逻辑内聚到各工具实现中，通过注册表统一分发
type ToolShowMsgBuilder interface {
	BuildRequest(paramsStr string) (string, error)
	BuildResponse(response string) (string, error)
}

var builders = map[string]ToolShowMsgBuilder{}

// RegisterShowMsgBuilder 在包初始化时由各工具文件调用进行注册
func RegisterShowMsgBuilder(name string, b ToolShowMsgBuilder) {
	builders[name] = b
}

// buildRequestShowMsgInner 统一分发到各工具的实现
func buildRequestShowMsgInner(toolName string, paramsStr string) (string, error) {
	b, ok := builders[toolName]
	if !ok {
		return "", nil
	}
	return b.BuildRequest(paramsStr)
}

// buildResponseShowMsgInner 统一分发到各工具的实现
func buildResponseShowMsgInner(toolName string, response string) (string, error) {
	b, ok := builders[toolName]
	if !ok {
		return "", nil
	}
	return b.BuildResponse(response)
}

// ErrInvalidJSON 是各工具在解析失败时返回的统一错误，便于上层处理
var ErrInvalidJSON = errors.New("invalid json for show message building")