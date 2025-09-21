package httputils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient 通用HTTP客户端
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建一个新的HTTP客户端
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// DefaultHTTPClient 返回默认配置的HTTP客户端(10秒超时)
func DefaultHTTPClient() *HTTPClient {
	return NewHTTPClient(10 * time.Second)
}

// Get 发送GET请求
func (c *HTTPClient) Get(ctx context.Context, url string, params url.Values, headers map[string]string) ([]byte, error) {
	// 构建请求URL
	reqURL := url
	if params != nil && len(params) > 0 {
		if strings.Contains(url, "?") {
			reqURL = fmt.Sprintf("%s&%s", url, params.Encode())
		} else {
			reqURL = fmt.Sprintf("%s?%s", url, params.Encode())
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建GET请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送GET请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	return body, nil
}

// Post 发送POST请求
func (c *HTTPClient) Post(ctx context.Context, url string, data interface{}, headers map[string]string) ([]byte, error) {
	// 将数据转换为JSON
	var reqBody io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %v", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建POST请求失败: %v", err)
	}

	// 设置默认Content-Type
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送POST请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	return body, nil
}

// GetJSON 发送GET请求并解析JSON响应
func (c *HTTPClient) GetJSON(ctx context.Context, url string, params url.Values, headers map[string]string, result interface{}) error {
	body, err := c.Get(ctx, url, params, headers)
	if err != nil {
		return err
	}

	// 解析JSON响应
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("解析JSON响应失败: %v", err)
	}

	return nil
}

// PostJSON 发送POST请求并解析JSON响应
func (c *HTTPClient) PostJSON(ctx context.Context, url string, data interface{}, headers map[string]string, result interface{}) error {
	body, err := c.Post(ctx, url, data, headers)
	if err != nil {
		return err
	}

	// 解析JSON响应
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("解析JSON响应失败: %v", err)
	}

	return nil
}
