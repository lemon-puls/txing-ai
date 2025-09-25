package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	einomcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"os"
	_ "path/filepath"
	"txing-ai/internal/global"
)

// mCPServerConfig 定义单个MCP服务器配置
type mCPServerConfig struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Environment map[string]string `json:"environment"`
}

// mCPServerConfigs 定义多个MCP服务器配置
type mCPServerConfigs struct {
	Servers []mCPServerConfig `json:"servers"`
}

// MCPClientManager 管理MCP客户端
type MCPClientManager struct {
	clients map[string]*client.Client
	tools   map[string][]tool.BaseTool
	ctx     context.Context
}

// Init 初始化所有MCP服务器
func NewMCPClientManager(ctx context.Context) *MCPClientManager {

	// 全局MCP客户端管理器
	var mcpManager = &MCPClientManager{
		clients: make(map[string]*client.Client),
		tools:   make(map[string][]tool.BaseTool),
		ctx:     ctx,
	}

	return mcpManager
}

// Init 初始化所有MCP服务器
func (m *MCPClientManager) Init(configPath string) error {

	// 读取配置文件
	configs, err := loadMCPServerConfigs(configPath)
	if err != nil {
		return fmt.Errorf("加载MCP服务器配置失败: %w", err)
	}

	// 初始化每个MCP服务器
	// 以 stdio 方式启动 MCP 服务端
	for _, serverConfig := range configs.Servers {
		if err := initMCPServer(m, serverConfig); err != nil {
			return fmt.Errorf("初始化MCP服务器 %s 失败: %w", serverConfig.Name, err)
		}
	}

	return nil
}

// GetMCPTools 获取指定MCP服务器的工具
func (m *MCPClientManager) GetMCPTools(serverName string) ([]tool.BaseTool, error) {
	tools, exists := m.tools[serverName]
	if !exists {
		return nil, fmt.Errorf("MCP服务器 %s 不存在或未初始化", serverName)
	}
	return tools, nil
}

// GetAllMCPTools 获取所有MCP服务器的工具
func (m *MCPClientManager) GetAllMCPTools() []tool.BaseTool {
	var allTools []tool.BaseTool
	for _, tools := range m.tools {
		allTools = append(allTools, tools...)
	}
	return allTools
}

// CloseMCPServer 关闭指定MCP服务器
func (m *MCPClientManager) CloseMCPServer(serverName string) error {
	cli, exists := m.clients[serverName]
	if !exists {
		return fmt.Errorf("MCP服务器 %s 不存在或未初始化", serverName)
	}

	// 关闭客户端
	if err := cli.Close(); err != nil {
		return fmt.Errorf("关闭MCP客户端失败: %w", err)
	}

	// 删除客户端和工具
	delete(m.clients, serverName)
	delete(m.tools, serverName)

	return nil
}

// CloseAllMCPServers 关闭所有MCP服务器
func (m *MCPClientManager) CloseAllMCPServers() error {
	var lastErr error
	for name := range m.clients {
		if err := m.CloseMCPServer(name); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// loadMCPServerConfigs 加载MCP服务器配置
func loadMCPServerConfigs(configPath string) (*mCPServerConfigs, error) {
	if configPath == "" {
		configPath = "./" + global.RuntimeDir + "/mcp_config.json.sample"
	}

	// 确保配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var configs mCPServerConfigs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &configs, nil
}

// initMCPServer 初始化单个MCP服务器
func initMCPServer(m *MCPClientManager, config mCPServerConfig) error {
	// 将 config.Environment 转换为字符串数组
	var envs []string
	for key, value := range config.Environment {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	// 创建stdio MCP客户端
	cli, err := client.NewStdioMCPClient(config.Command, envs, config.Args...)
	if err != nil {
		return fmt.Errorf("创建MCP客户端失败: %w", err)
	}

	// 初始化MCP客户端
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    config.Name,
		Version: "1.0.0",
	}

	_, err = cli.Initialize(m.ctx, initRequest)
	if err != nil {
		return fmt.Errorf("初始化MCP客户端失败: %w", err)
	}

	// 获取MCP工具
	mcpTools, err := einomcp.GetTools(m.ctx, &einomcp.Config{Cli: cli})
	if err != nil {
		return fmt.Errorf("获取MCP工具失败: %w", err)
	}

	// 存储客户端和工具
	m.clients[config.Name] = cli
	m.tools[config.Name] = mcpTools

	return nil
}
