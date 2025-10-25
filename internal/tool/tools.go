package tool

import (
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"sync"
	"txing-ai/internal/iface"
)

const (
	// 工具名称
	webSearchToolName            = "web_search_tool"
	markdownToPDFToolName        = "markdown_to_pdf_file_tool"
	mapsGeoToolName              = "maps_geo"
	mapsTextSearchToolName       = "maps_text_search"
	mapsDirectionToolName        = "maps_direction_transit_integrated"
	imageSearchToolName          = "image_search_tool"
	imageDownloadToolName        = "image_download_tool"
	markdownSaveToolName         = "markdown_save_tool"
	mapsSearchDetailToolName     = "maps_search_detail"
	mapsDistanceToolName         = "maps_distance"
	mapsWeatherToolName          = "maps_weather"
	mapsDirectionDrivingToolName = "maps_direction_driving"
	mapsDirectionWalkingToolName = "maps_direction_walking"
	mapsAroundSearchToolName     = "maps_around_search"
)

var (
	toolRegisterOnce sync.Once
	tools            []tool.BaseTool
)

// ProvideTools 注册工具
func ProvideTools(res iface.ResourceProvider) []tool.BaseTool {
	toolRegisterOnce.Do(func() {
		// 注册网页搜索工具
		searchWebTool, err := utils.InferTool(
			webSearchToolName,
			// 说明有速率限制，需要注意不要频繁调用
			"Search for information from Baidu Search Engine, has rate limit, please be careful",
			searchWeb)
		if err != nil {
			panic(err)
		}
		tools = append(tools, searchWebTool)

		// 注册Markdown保存工具
		markdownSaveTool, err := utils.InferTool(
			"markdown_save_tool",
			"Save markdown content to a local file",
			saveMarkdown)
		if err != nil {
			panic(err)
		}
		tools = append(tools, markdownSaveTool)

		// 注册图片下载工具
		imageDownloadTool, err := utils.InferTool(
			"image_download_tool",
			"Download image from URL to local file",
			downloadImage)
		if err != nil {
			panic(err)
		}
		tools = append(tools, imageDownloadTool)

		// 注册图片搜索工具
		imageSearchTool, err := utils.InferTool(
			"image_search_tool",
			"Search images from Sougou Search Engine",
			searchImageSougou)
		if err != nil {
			panic(err)
		}
		tools = append(tools, imageSearchTool)

		// 注册网页抓取工具
		webScrapingTool, err := utils.InferTool(
			"web_scraping_tool",
			"Scrape the content of a web page",
			scrapeWebPage)
		if err != nil {
			panic(err)
		}
		tools = append(tools, webScrapingTool)

		// TODO 文件操作相关工具还需要完善，暂时注释掉，避免 LLM 误用
		//// 注册文件读取工具
		//fileReadTool, err := utils.InferTool(
		//	"file_read_tool",
		//	"Read content from a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
		//	readFile)
		//if err != nil {
		//	panic(err)
		//}
		//tools = append(tools, fileReadTool)
		//
		//// 注册文件写入工具
		//fileWriteTool, err := utils.InferTool(
		//	"file_write_tool",
		//	"Write content to a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
		//	writeFile)
		//if err != nil {
		//	panic(err)
		//}
		//tools = append(tools, fileWriteTool)
		//
		//// 注册文件内容替换工具
		//fileReplaceTool, err := utils.InferTool(
		//	"file_replace_tool",
		//	"Replace content in a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
		//	replaceFileContent)
		//if err != nil {
		//	panic(err)
		//}
		//tools = append(tools, fileReplaceTool)
		//
		//// 注册文件删除工具
		//fileDeleteTool, err := utils.InferTool(
		//	"file_delete_tool",
		//	"Delete a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
		//	deleteFile)
		//if err != nil {
		//	panic(err)
		//}
		//tools = append(tools, fileDeleteTool)
		//
		//// 注册文件列表工具
		//fileListTool, err := utils.InferTool(
		//	"file_list_tool",
		//	"List files in a directory using relative path (only directories in runtime directory are allowed, e.g. runtime, runtime/folder)",
		//	listFiles)
		//if err != nil {
		//	panic(err)
		//}
		//tools = append(tools, fileListTool)

		// 注册Markdown转PDF工具
		markdownToPDFTool, err := utils.InferTool(
			"markdown_to_pdf_file_tool",
			"Convert markdown content to PDF, and save it to local file. Note: use \\n\\n for line breaks; single \\n will not render as a line break in the PDF."+
				"fileName will be automatically appended with timestamp",
			saveMarkdownToPDF)
		if err != nil {
			panic(err)
		}
		tools = append(tools, markdownToPDFTool)

		//注册PDF文本提取工具
		pdfReadTool, err := utils.InferTool(
			"pdf_read_tool",
			"Extract text content from a PDF file (only files in runtime directory are allowed, e.g. runtime/test.pdf, runtime/folder/file.pdf)",
			ReadPdfText)
		if err != nil {
			panic(err)
		}
		tools = append(tools, pdfReadTool)

		// 注册PDF验证工具
		//pdfValidateTool, err := utils.InferTool(
		//	"pdf_validate_tool",
		//	"Validate if a file is a valid PDF file (only files in runtime directory are allowed, e.g. runtime/test.pdf, runtime/folder/file.pdf)",
		//	validatePdf)
		//if err != nil {
		//	panic(err)
		//}
		//tools = append(tools, pdfValidateTool)

		// 添加MCP工具（如果已初始化）
		mcpClientManager := res.GetMCPClientManager()
		mcpTools := mcpClientManager.GetAllMCPTools()
		if len(mcpTools) > 0 {
			tools = append(tools, mcpTools...)
		}
	})
	return tools
}

// 构建指定工具的调用请求信息（用于前端展示）
func BuildRequestShowMsg(toolName string, paramsStr string) (string, error) {
	return buildRequestShowMsgInner(toolName, paramsStr)
}

// 构建指定工具的调用响应信息（用于前端展示）
func BuildResponseShowMsg(toolName string, response string) (string, error) {
	return buildResponseShowMsgInner(toolName, response)
}
