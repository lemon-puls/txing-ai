package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CheckURLExists 检查URL是否存在
// 从 internal/controller/website 抽取，供管理端与运营助手工具复用
func CheckURLExists(url string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second, // 5秒超时
	}

	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer func() {
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	return resp.StatusCode == http.StatusOK
}

// ExtractFaviconFromHTML 从网页内容中提取favicon链接
// 从 internal/controller/website 抽取，供管理端与运营助手工具复用
func ExtractFaviconFromHTML(urlStr string) (string, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送GET请求获取网页内容
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	// 设置User-Agent以模拟浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取网页内容失败，状态码：%d", resp.StatusCode)
	}

	// 读取网页内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析网页内容
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// 提取favicon链接
	// 1. 尝试查找link标签中的favicon
	patterns := []string{
		`<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+\.(?:ico|png|jpg|jpeg|svg|gif))["']`,
		`<link[^>]*href=["']([^"']+\.(?:ico|png|jpg|jpeg|svg|gif))["'][^>]*rel=["'](?:shortcut )?icon["']`,
		`<link[^>]*rel=["']apple-touch-icon["'][^>]*href=["']([^"']+\.(?:ico|png|jpg|jpeg|svg|gif))["']`,
		`<link[^>]*href=["']([^"']+\.(?:ico|png|jpg|jpeg|svg|gif))["'][^>]*rel=["']apple-touch-icon["']`,
		`<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["'](?:shortcut )?icon["']`,
		`<link[^>]*rel=["']apple-touch-icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["']apple-touch-icon["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) > 1 {
				faviconPath := match[1]
				// 处理相对路径
				if !strings.HasPrefix(faviconPath, "http") {
					if strings.HasPrefix(faviconPath, "//") {
						// 处理协议相对URL
						return parsedURL.Scheme + ":" + faviconPath, nil
					} else if strings.HasPrefix(faviconPath, "/") {
						// 处理根路径
						return fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, faviconPath), nil
					} else {
						// 处理相对路径
						base := urlStr
						if !strings.HasSuffix(base, "/") {
							base = base[:strings.LastIndex(base, "/")+1]
						}
						return base + faviconPath, nil
					}
				}
				return faviconPath, nil
			}
		}
	}

	// 2. 如果没有找到，尝试使用Google的favicon服务
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=64", parsedURL.Host), nil
}
