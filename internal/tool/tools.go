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

		// 注册图片下载工具
		imageDownloadTool, err := utils.InferTool(
			"image download tool",
			"Download image from URL to local file",
			downloadImage)
		if err != nil {
			panic(err)
		}
		tools = append(tools, imageDownloadTool)

		// 注册图片搜索工具
		imageSearchTool, err := utils.InferTool(
			"image search tool",
			"Search images from Sougou Search Engine",
			searchImageSougou)
		if err != nil {
			panic(err)
		}
		tools = append(tools, imageSearchTool)

		// 注册网页抓取工具
		webScrapingTool, err := utils.InferTool(
			"web scraping tool",
			"Scrape the content of a web page",
			scrapeWebPage)
		if err != nil {
			panic(err)
		}
		tools = append(tools, webScrapingTool)
	})
	return tools
}
