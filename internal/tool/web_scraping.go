package tool

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"time"
)

// webScrapingRequest 网页抓取请求参数
type webScrapingRequest struct {
	URL string `json:"url" binding:"required"`
}

// webScrapingResponse 网页抓取响应
type webScrapingResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// scrapeWebPage 网页抓取工具
// 从指定URL抓取网页内容
func scrapeWebPage(ctx context.Context, req *webScrapingRequest) (webScrapingResponse, error) {
	var response webScrapingResponse

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送HTTP请求
	resp, err := client.Get(req.URL)
	if err != nil {
		response.Error = fmt.Sprintf("请求网页失败: %s", err.Error())
		return response, err
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		response.Error = fmt.Sprintf("HTTP请求失败，状态码: %d", resp.StatusCode)
		return response, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		response.Error = fmt.Sprintf("解析HTML失败: %s", err.Error())
		return response, err
	}

	// 获取HTML内容
	html, err := doc.Html()
	if err != nil {
		response.Error = fmt.Sprintf("获取HTML内容失败: %s", err.Error())
		return response, err
	}

	response.Content = html
	return response, nil
}
