package tool

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/http"
	"strings"
	"time"
)

// webScrapingRequest 网页抓取请求参数
type webScrapingRequest struct {
	URL string `json:"url" binding:"required"`
}

// webScrapingResponse 网页抓取响应
type webScrapingResponse struct {
	Content string `json:"content"`
	Title   string `json:"title,omitempty"`
	Error   string `json:"error,omitempty"`
}

// 常用浏览器User-Agent列表，用于随机切换
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Edge/120.0.0.0",
}

// scrapeWebPage 网页抓取工具
// 从指定URL抓取网页内容，使用Colly库处理反爬机制
func scrapeWebPage(ctx context.Context, req *webScrapingRequest) (webScrapingResponse, error) {
	var response webScrapingResponse
	var htmlContent strings.Builder
	var pageTitle string

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

	// 处理响应
	c.OnResponse(func(r *colly.Response) {
		htmlContent.Write(r.Body)
	})

	// 获取页面标题
	c.OnHTML("title", func(e *colly.HTMLElement) {
		pageTitle = e.Text
	})

	// 处理错误
	c.OnError(func(r *colly.Response, err error) {
		response.Error = fmt.Sprintf("抓取网页失败: %s, 状态码: %d", err.Error(), r.StatusCode)
	})

	// 设置上下文，支持取消操作
	err := c.Visit(req.URL)
	if err != nil {
		response.Error = fmt.Sprintf("访问网页失败: %s", err.Error())
		return response, err
	}

	// 等待所有异步请求完成
	c.Wait()

	// 设置响应内容
	response.Content = htmlContent.String()
	response.Title = pageTitle

	return response, nil
}
