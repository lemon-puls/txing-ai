package tool

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// webScrapingRequest 网页抓取请求参数
type webScrapingRequest struct {
	URL           string `json:"url" binding:"required"`
	MaxTextLength int    `json:"max_text_length,omitempty"` // 最大文本长度限制
}

// webScrapingResponse 网页抓取响应
type webScrapingResponse struct {
	Content     string `json:"content"`      // 提取的文本内容
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"` // 网页描述
	Error       string `json:"error,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"` // 内容是否被截断
}

// 常用浏览器User-Agent列表，用于随机切换
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Edge/120.0.0.0",
}

// 默认最大文本长度 (约8000个token)
const defaultMaxTextLength = 16000

// 清理HTML文本的正则表达式
var (
	spaceRegex      = regexp.MustCompile(`\s+`)
	scriptRegex     = regexp.MustCompile(`(?s)<script.*?</script>`)
	styleRegex      = regexp.MustCompile(`(?s)<style.*?</style>`)
	htmlTagRegex    = regexp.MustCompile(`<[^>]*>`)
)

// scrapeWebPage 网页抓取工具
// 从指定URL抓取网页内容，使用Colly库处理反爬机制，并提取有效文本
func scrapeWebPage(ctx context.Context, req *webScrapingRequest) (webScrapingResponse, error) {
	var response webScrapingResponse
	var textContent strings.Builder
	var pageTitle, pageDescription string
	
	// 设置最大文本长度
	maxLength := defaultMaxTextLength
	if req.MaxTextLength > 0 {
		maxLength = req.MaxTextLength
	}

	// 创建新的Colly收集器
	c := colly.NewCollector(
		// 允许访问HTTPS网站
		colly.AllowURLRevisit(),
		// 设置超时时间
		colly.MaxDepth(1),
	)

	// 配置收集器
	c.WithTransport(&http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 禁用TLS证书验证
		},
	})

	// 设置请求超时
	c.SetRequestTimeout(15 * time.Second)

	// 随机设置User-Agent
	c.OnRequest(func(r *colly.Request) {
		// 设置随机User-Agent
		r.Headers.Set("User-Agent", userAgents[time.Now().UnixNano()%int64(len(userAgents))])
		// 设置Referer
		r.Headers.Set("Referer", "https://www.google.com/")
		// 设置Accept头
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	})

	// 获取页面标题
	c.OnHTML("title", func(e *colly.HTMLElement) {
		pageTitle = strings.TrimSpace(e.Text)
	})

	// 获取页面描述
	c.OnHTML("meta[name=description]", func(e *colly.HTMLElement) {
		pageDescription = e.Attr("content")
	})

	// 提取正文内容
	c.OnHTML("body", func(e *colly.HTMLElement) {
		// 提取所有段落文本
		e.ForEach("p, h1, h2, h3, h4, h5, h6, li", func(_ int, el *colly.HTMLElement) {
			text := strings.TrimSpace(el.Text)
			if text != "" {
				textContent.WriteString(text)
				textContent.WriteString("\n\n")
			}
		})

		// 如果没有足够的结构化内容，尝试提取所有文本
		if textContent.Len() < 100 {
			text := extractText(e.DOM.Text())
			textContent.WriteString(text)
		}
	})

	// 处理错误
	c.OnError(func(r *colly.Response, err error) {
		response.Error = fmt.Sprintf("抓取网页失败: %s, 状态码: %d", err.Error(), r.StatusCode)
	})

	// 访问网页
	err := c.Visit(req.URL)
	if err != nil {
		response.Error = fmt.Sprintf("访问网页失败: %s", err.Error())
		return response, err
	}

	// 等待所有异步请求完成
	c.Wait()

	// 设置响应内容
	content := textContent.String()
	
	// 检查是否需要截断内容
	truncated := false
	if len(content) > maxLength {
		content = content[:maxLength]
		truncated = true
	}
	
	response.Content = content
	response.Title = pageTitle
	response.URL = req.URL
	response.Description = pageDescription
	response.Truncated = truncated

	return response, nil
}

// extractText 从HTML中提取纯文本
func extractText(html string) string {
	// 移除script标签及其内容
	html = scriptRegex.ReplaceAllString(html, "")
	
	// 移除style标签及其内容
	html = styleRegex.ReplaceAllString(html, "")
	
	// 移除HTML标签
	html = htmlTagRegex.ReplaceAllString(html, " ")
	
	// 规范化空白字符
	html = spaceRegex.ReplaceAllString(html, " ")
	
	return strings.TrimSpace(html)
}
