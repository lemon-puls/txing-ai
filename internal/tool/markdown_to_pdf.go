package tool

import (
	"context"
	"fmt"
	"github.com/russross/blackfriday/v2"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"txing-ai/internal/global/logging/log"
)

// markdownToPDFParams 将Markdown转换为PDF的参数
type markdownToPDFParams struct {
	Content  string `json:"content" jsonschema:"description=要转换的Markdown内容"`
	Filename string `json:"filename" jsonschema:"description=文件名(不含扩展名)"`
	Title    string `json:"title,omitempty" jsonschema:"description=PDF文档标题(可选)"`
}

// saveMarkdownToPDF 将Markdown内容转换为PDF并保存
func saveMarkdownToPDF(ctx context.Context, params *markdownToPDFParams) (string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("获取当前工作目录失败", zap.Error(err))
		return "获取当前工作目录失败", nil
	}

	log.Info("开始Markdown转PDF", zap.String("content", params.Content))

	// 创建保存目录
	saveDir := filepath.Join(currentDir, "runtime", "temp")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Error("创建保存目录失败", zap.String("dir", saveDir), zap.Error(err))
		return "创建保存目录失败", fmt.Errorf("创建保存目录失败: %v", err)
	}

	// 确保文件名有.pdf扩展名
	filename := params.Filename
	if filepath.Ext(filename) != ".pdf" {
		filename = filename + ".pdf"
	}

	// 构建完整的文件路径
	fullPath := filepath.Join(saveDir, filename)

	// 处理Markdown内容中的图片
	content, imagePaths, err := processMarkdownImages(params.Content, saveDir)
	if err != nil {
		log.Error("处理Markdown图片失败", zap.Error(err))
		return fmt.Sprintf("处理Markdown图片失败：%v", err), nil
	}
	defer cleanupTempImages(imagePaths)

	// 将Markdown转换为HTML
	htmlContent := markdownToHTML(content, params.Title)

	// 保存HTML到临时文件
	htmlFilePath := filepath.Join(saveDir, "temp.html")
	if err := os.WriteFile(htmlFilePath, []byte(htmlContent), 0644); err != nil {
		log.Error("保存临时HTML文件失败", zap.String("path", htmlFilePath), zap.Error(err))
		return fmt.Sprintf("保存临时HTML文件失败：%v"), nil
	}
	defer os.Remove(htmlFilePath)

	// 使用wkhtmltopdf将HTML转换为PDF
	if err := convertHTMLToPDF(htmlFilePath, fullPath); err != nil {
		log.Error("HTML转PDF失败", zap.Error(err))
		return fmt.Sprintf("HTML转PDF失败：%v"), nil
	}

	// 记录成功日志
	log.Info("PDF文件生成成功",
		zap.String("path", fullPath),
		zap.Int("contentLength", len(params.Content)),
		zap.Time("timestamp", time.Now()))

	return fmt.Sprintf("PDF文件已成功生成: ./%s", filename), nil
}

// processMarkdownImages 处理Markdown内容中的图片
// 返回处理后的Markdown内容和下载的图片路径列表
func processMarkdownImages(content string, tempDir string) (string, []string, error) {
	var imagePaths []string

	// 创建图片目录
	imageDir := tempDir
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return "", nil, fmt.Errorf("创建图片目录失败: %v", err)
	}

	// 匹配Markdown中的图片链接
	imageRegex := regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

	// 替换图片链接为本地路径
	processedContent := imageRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		altText := submatches[1]
		imageURL := submatches[2]

		// 如果是网络图片，下载到本地
		if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
			localPath, err := downloadImageFromURL(imageURL, imageDir)
			if err != nil {
				log.Warn("下载图片失败", zap.String("url", imageURL), zap.Error(err))
				return match // 如果下载失败，保持原样
			}

			imagePaths = append(imagePaths, localPath)
			return fmt.Sprintf("![%s](%s)", altText, localPath)
		}

		return match
	})

	return processedContent, imagePaths, nil
}

// downloadImage 下载图片到本地
func downloadImageFromURL(url string, dir string) (string, error) {
	// 从URL中提取文件名
	filename := filepath.Base(url)
	if filename == "" || filename == "." || filename == "/" {
		// 如果无法从URL中提取有效文件名，生成一个随机文件名
		filename = fmt.Sprintf("image_%d%s", time.Now().UnixNano(), filepath.Ext(url))
	}

	// 确保文件名有扩展名
	if filepath.Ext(filename) == "" {
		filename = filename + ".jpg" // 默认扩展名
	}

	// 构建本地文件路径
	localPath := filepath.Join(dir, filename)

	// 创建文件
	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	// 将响应内容写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
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

// markdownToHTML 将Markdown转换为HTML
func markdownToHTML(content string, title string) string {
	// 使用blackfriday将Markdown转换为HTML
	htmlContent := blackfriday.Run([]byte(content))

	// 添加HTML头和样式
	htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3, h4, h5, h6 {
            color: #2c3e50;
            margin-top: 24px;
            margin-bottom: 16px;
            font-weight: 600;
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
            font-family: Consolas, Monaco, 'Andale Mono', monospace;
            background-color: #f6f8fa;
            padding: 0.2em 0.4em;
            border-radius: 3px;
        }
        blockquote {
            border-left: 4px solid #dfe2e5;
            padding: 0 1em;
            color: #6a737d;
            margin: 0;
        }
        img {
            max-width: 100%%;
            height: auto;
        }
        table {
            border-collapse: collapse;
            width: 100%%;
            margin-bottom: 16px;
        }
        table, th, td {
            border: 1px solid #dfe2e5;
        }
        th, td {
            padding: 8px 16px;
            text-align: left;
        }
        tr:nth-child(even) {
            background-color: #f6f8fa;
        }
    </style>
</head>
<body>
    %s
</body>
</html>`

	if title == "" {
		title = "Markdown Document"
	}

	return fmt.Sprintf(htmlTemplate, title, string(htmlContent))
}

// convertHTMLToPDF 使用wkhtmltopdf将HTML转换为PDF
func convertHTMLToPDF(htmlPath, pdfPath string) error {
	// 检查wkhtmltopdf是否安装
	_, err := exec.LookPath("wkhtmltopdf")
	if err != nil {
		return fmt.Errorf("未找到wkhtmltopdf，请确保已安装: %v", err)
	}

	// 构建命令
	cmd := exec.Command("wkhtmltopdf",
		"--enable-local-file-access",
		"--page-size", "A4",
		"--margin-top", "20mm",
		"--margin-bottom", "20mm",
		"--margin-left", "20mm",
		"--margin-right", "20mm",
		"--encoding", "UTF-8",
		"--quiet",
		htmlPath, pdfPath)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wkhtmltopdf执行失败: %v, 输出: %s", err, string(output))
	}

	return nil
}
