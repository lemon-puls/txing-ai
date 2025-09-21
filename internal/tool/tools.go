package tool

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"sync"
)

var (
	toolRegisterOnce sync.Once
	tools            []tool.BaseTool
)

// ProvideTools 注册工具
func ProvideTools() []tool.BaseTool {
	toolRegisterOnce.Do(func() {
		// 注册网页搜索工具
		searchWebTool, err := utils.InferTool(
			"web search tool",
			"Search for information from Baidu Search Engine",
			searchWeb)
		if err != nil {
			panic(err)
		}
		tools = append(tools, searchWebTool)

		// 注册Markdown保存工具
		markdownSaveTool, err := utils.InferTool(
			"markdown save tool",
			"Save markdown content to a local file",
			saveMarkdown)
		if err != nil {
			panic(err)
		}
		tools = append(tools, markdownSaveTool)
	})
	return tools
}
