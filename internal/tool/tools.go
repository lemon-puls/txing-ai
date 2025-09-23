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
			// 说明有速率限制，需要注意不要频繁调用
			"Search for information from Baidu Search Engine, has rate limit, please be careful",
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

		// 注册文件读取工具
		fileReadTool, err := utils.InferTool(
			"file read tool",
			"Read content from a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			readFile)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileReadTool)

		// 注册文件写入工具
		fileWriteTool, err := utils.InferTool(
			"file write tool",
			"Write content to a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			writeFile)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileWriteTool)

		// 注册文件内容替换工具
		fileReplaceTool, err := utils.InferTool(
			"file replace tool",
			"Replace content in a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			replaceFileContent)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileReplaceTool)

		// 注册文件删除工具
		fileDeleteTool, err := utils.InferTool(
			"file delete tool",
			"Delete a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			deleteFile)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileDeleteTool)

		// 注册文件列表工具
		fileListTool, err := utils.InferTool(
			"file list tool",
			"List files in a directory using relative path (only directories in runtime directory are allowed, e.g. runtime, runtime/folder)",
			listFiles)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileListTool)

		// 注册Markdown转PDF工具
		markdownToPDFTool, err := utils.InferTool(
			"markdown to pdf tool",
			"Convert markdown content to a PDF file with images support",
			saveMarkdownToPDF)
		if err != nil {
			panic(err)
		}
		tools = append(tools, markdownToPDFTool)
	})
	return tools
}
