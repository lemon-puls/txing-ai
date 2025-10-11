package tool

import (
	"encoding/json"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
)

const (
	// 工具名称
	webSearchToolName     = "web_search_tool"
	markdownToPDFToolName = "markdown_to_pdf_file_tool"
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

		// 注册文件读取工具
		fileReadTool, err := utils.InferTool(
			"file_read_tool",
			"Read content from a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			readFile)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileReadTool)

		// 注册文件写入工具
		fileWriteTool, err := utils.InferTool(
			"file_write_tool",
			"Write content to a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			writeFile)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileWriteTool)

		// 注册文件内容替换工具
		fileReplaceTool, err := utils.InferTool(
			"file_replace_tool",
			"Replace content in a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			replaceFileContent)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileReplaceTool)

		// 注册文件删除工具
		fileDeleteTool, err := utils.InferTool(
			"file_delete_tool",
			"Delete a file using relative path (only files in runtime directory are allowed, e.g. runtime/test.txt, runtime/folder/file.txt)",
			deleteFile)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileDeleteTool)

		// 注册文件列表工具
		fileListTool, err := utils.InferTool(
			"file_list_tool",
			"List files in a directory using relative path (only directories in runtime directory are allowed, e.g. runtime, runtime/folder)",
			listFiles)
		if err != nil {
			panic(err)
		}
		tools = append(tools, fileListTool)

		// 注册Markdown转PDF工具
		markdownToPDFTool, err := utils.InferTool(
			"markdown_to_pdf_file_tool",
			"Convert markdown content to PDF, and save it to local file",
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
		pdfValidateTool, err := utils.InferTool(
			"pdf_validate_tool",
			"Validate if a file is a valid PDF file (only files in runtime directory are allowed, e.g. runtime/test.pdf, runtime/folder/file.pdf)",
			validatePdf)
		if err != nil {
			panic(err)
		}
		tools = append(tools, pdfValidateTool)

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
	if toolName == webSearchToolName {
		// 将 paramsStr 解析为 webSearchParams
		var searchParams webSearchParams
		if err := json.Unmarshal([]byte(paramsStr), &searchParams); err != nil {
			log.Error("构建网页搜搜请求显示信息失败", zap.Error(err))
			return "", err
		}
		return "网页搜索：" + searchParams.Query, nil
	} else if toolName == markdownToPDFToolName {
		var params markdownToPDFParams
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			log.Error("构建Markdown转PDF请求显示信息失败", zap.Error(err))
			return "", err
		}
		return "保存为 PDF 文件：" + params.Filename, nil
	} else {
		return "", nil
	}
}

// 构建指定工具的调用响应信息（用于前端展示）

func BuildResponseShowMsg(toolName string, response string) (string, error) {
	if toolName == webSearchToolName {
		// 处理网页搜索工具的响应，统计结果数量
		r := strings.TrimSpace(response)
		if r == "" {
			return "共找到 0 个相关网页", nil
		}

		// 如果是提示语或错误
		if strings.Contains(r, "搜索结果为空") {
			return "共找到 0 个相关网页", nil
		}

		// 优先尝试按JSON数组解析；如果不是数组则包裹为数组
		var arr []map[string]interface{}
		raw := r
		// 去掉可能的末尾逗号
		raw = strings.TrimSuffix(raw, ",")
		if !strings.HasPrefix(strings.TrimLeft(raw, " \n\t"), "[") {
			raw = "[" + raw + "]"
		}

		if err := json.Unmarshal([]byte(raw), &arr); err != nil {
			log.Error("构建网页搜搜响应显示信息失败, 无法解析JSON", zap.Error(err))
			return "", err
		}

		return "共找到 " + strconv.Itoa(len(arr)) + " 个相关网页", nil
	} else if toolName == markdownToPDFToolName {
		return response, nil
	}
	return "", nil
}
