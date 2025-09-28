package tool

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"txing-ai/internal/global/logging/log"

	"github.com/russross/blackfriday/v2"
	"go.uber.org/zap"
)

// markdownToPDFParams 转换Markdown到PDF的参数
type markdownToPDFParams struct {
	Content  string `json:"content" jsonschema:"description=要转换的Markdown内容"`
	Filename string `json:"filename" jsonschema:"description=文件名(不含扩展名)"`
}

// saveMarkdownToPDF 将Markdown内容转换为PDF并保存到本地文件
func saveMarkdownToPDF(ctx context.Context, params *markdownToPDFParams) (string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("获取当前工作目录失败", zap.Error(err))
		return "", fmt.Errorf("获取当前工作目录失败: %v", err)
	}

	// 创建临时目录用于存放中间文件
	tempDir, err := ioutil.TempDir("", "markdown_to_pdf_")
	if err != nil {
		log.Error("创建临时目录失败", zap.Error(err))
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 处理Markdown中的图片
	content, imagePaths, err := processMarkdownImages(params.Content, tempDir)
	if err != nil {
		log.Error("处理Markdown图片失败", zap.Error(err))
		return "", fmt.Errorf("处理Markdown图片失败: %v", err)
	}
	defer cleanupTempImages(imagePaths)

	// 确保文件名有.pdf扩展名
	filename := params.Filename
	if filepath.Ext(filename) != ".pdf" {
		filename = filename + ".pdf"
	}

	// 构建保存路径
	savePath := filepath.Join(currentDir, "runtime", "temp")
	if err := os.MkdirAll(savePath, 0755); err != nil {
		log.Error("创建保存目录失败", zap.String("dir", savePath), zap.Error(err))
		return "", fmt.Errorf("创建保存目录失败: %v", err)
	}

	// 构建完整的文件路径
	fullPath := filepath.Join(savePath, filename)

	// 将Markdown转换为HTML
	html := markdownToHTML(content, filepath.Base(filename))

	// 将HTML保存到临时文件
	htmlPath := filepath.Join(tempDir, "temp.html")
	if err := ioutil.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		log.Error("保存HTML临时文件失败", zap.Error(err))
		return "", fmt.Errorf("保存HTML临时文件失败: %v", err)
	}

	// 将HTML转换为PDF
	if err := convertHTMLToPDF(htmlPath, fullPath); err != nil {
		log.Error("HTML转PDF失败", zap.Error(err))
		return "", fmt.Errorf("HTML转PDF失败: %v", err)
	}

	// 记录成功日志
	//log.Info("PDF文件生成成功",
	//	zap.String("path", fullPath),
	//	zap.Int("contentLength", len(params.Content)),
	//	zap.Time("timestamp", time.Now()))

	return fmt.Sprintf("PDF文件已成功生成: ./%s", filename), nil
}

// processMarkdownImages 处理Markdown中的图片，下载远程图片到本地
func processMarkdownImages(content string, tempDir string) (string, []string, error) {
	var imagePaths []string

	// 匹配Markdown中的图片链接
	re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		imageURL := match[1]
		if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
			// 下载远程图片
			localPath, err := downloadImageFromURL(imageURL, tempDir)
			if err != nil {
				log.Error("下载图片失败", zap.String("url", imageURL), zap.Error(err))
				continue
			}

			// 替换图片链接为本地路径
			content = strings.Replace(content, match[0], fmt.Sprintf("![image](%s)", localPath), 1)
			imagePaths = append(imagePaths, localPath)
		}
	}

	return content, imagePaths, nil
}

// downloadImageFromURL 从URL下载图片到本地
func downloadImageFromURL(url string, dir string) (string, error) {
	// 提取文件名
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// 确保文件名唯一
	filename = fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
	localPath := filepath.Join(dir, filename)

	// 使用curl下载图片
	cmd := exec.Command("curl", "-L", "-o", localPath, url)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return localPath, nil
}

// cleanupTempImages 清理临时图片文件
func cleanupTempImages(imagePaths []string) {
	for _, path := range imagePaths {
		os.Remove(path)
	}
}

// markdownToHTML 将Markdown内容转换为HTML
func markdownToHTML(content string, title string) string {
	// 设置Blackfriday扩展选项
	extensions := blackfriday.CommonExtensions | blackfriday.AutoHeadingIDs

	// 创建HTML渲染器
	htmlFlags := blackfriday.CommonHTMLFlags
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: htmlFlags,
	})

	// 将Markdown转换为HTML
	html := blackfriday.Run([]byte(content), blackfriday.WithExtensions(extensions), blackfriday.WithRenderer(renderer))

	// 处理HTML中的emoji
	processedHTML := processEmojis(string(html))

	// 生成完整的HTML文档
	return generateHTML(processedHTML, title)
}

// processEmojis 处理HTML内容中的emoji字符，为其添加特殊样式类
func processEmojis(htmlContent string) string {
	// 匹配常见的emoji字符范围
	// Go的正则表达式不支持\p{Emoji}，所以使用Unicode范围来匹配常见emoji
	emojiRegex := regexp.MustCompile(`[\x{1F300}-\x{1F6FF}\x{1F900}-\x{1F9FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]`)

	// 为每个emoji添加span标签和emoji类
	processedHTML := emojiRegex.ReplaceAllStringFunc(htmlContent, func(emoji string) string {
		return fmt.Sprintf("<span class=\"emoji\">%s</span>", emoji)
	})

	return processedHTML
}

// generateHTML 生成完整的HTML文档
func generateHTML(body string, title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        @font-face {
            font-family: 'NotoEmoji';
            src: url('https://cdn.jsdelivr.net/npm/emoji-datasource-apple@14.0.0/img/apple/sheets/32.png') format('png');
            font-display: swap;
        }
        
        body {
            font-family: 'Segoe UI', 'Apple Color Emoji', 'Noto Color Emoji', 'NotoEmoji', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 0px 20px;
        }
        
        h1, h2, h3, h4, h5, h6 {
            margin-top: 24px;
            margin-bottom: 16px;
            font-weight: 600;
            line-height: 1.25;
        }
        
        h1 {
            font-size: 2em;
            border-bottom: 1px solid #eaecef;
            padding-bottom: 0.3em;
        }
        
        h2 {
            font-size: 1.5em;
            border-bottom: 1px solid #eaecef;
            padding-bottom: 0.3em;
        }
        
        a {
            color: #0366d6;
            text-decoration: none;
        }
        
        a:hover {
            text-decoration: underline;
        }
        
        pre {
            background-color: #f6f8fa;
            border-radius: 3px;
            padding: 16px;
            overflow: auto;
        }
        
        code {
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
            background-color: #f6f8fa;
            padding: 0.2em 0.4em;
            border-radius: 3px;
        }
        
        img {
            max-width: 100%%
        }
        
        blockquote {
            border-left: 4px solid #dfe2e5;
            padding: 0 1em;
            color: #6a737d;
            margin: 0;
        }
        
        table {
            border-collapse: collapse;
            width: 100%%
            margin-bottom: 16px;
        }
        
        table th, table td {
            border: 1px solid #dfe2e5;
            padding: 6px 13px;
        }
        
        table tr {
            background-color: #fff;
            border-top: 1px solid #c6cbd1;
        }
        
        table tr:nth-child(2n) {
            background-color: #f6f8fa;
        }
        
        .emoji {
            font-family: 'Apple Color Emoji', 'Segoe UI Emoji', 'Noto Color Emoji', 'NotoEmoji', sans-serif;
            font-size: 1.2em;
            line-height: 1;
            vertical-align: middle;
        }
    </style>
</head>
<body>
    %s
</body>
</html>`, title, body)
}

// convertHTMLToPDF 将HTML转换为PDF
func convertHTMLToPDF(htmlPath string, pdfPath string) error {
	// 检查wkhtmltopdf是否安装
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		return fmt.Errorf("wkhtmltopdf未安装: %v", err)
	}

	// 构建wkhtmltopdf命令
	cmd := exec.Command(
		"wkhtmltopdf",
		"--enable-local-file-access",
		"--enable-javascript",
		"--javascript-delay", "1000",
		"--no-stop-slow-scripts",
		"--disable-smart-shrinking",
		"--page-size", "A4",
		"--margin-top", "20",
		"--margin-right", "20",
		"--margin-bottom", "20",
		"--margin-left", "20",
		"--encoding", "UTF-8",
		htmlPath,
		pdfPath,
	)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wkhtmltopdf执行失败: %v, 输出: %s", err, string(output))
	}

	return nil
}
