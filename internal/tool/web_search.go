package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils/httputils"

	"go.uber.org/zap"
)

// webSearchParams 替换段落参数
type webSearchParams struct {
	Query string `json:"query" jsonschema:"description=搜索关键词"`
}

// 搜索结果结构体
type SearchResponse struct {
	OrganicResults []map[string]interface{} `json:"organic_results"`
}

func searchWeb(ctx context.Context, params *webSearchParams) (string, error) {
	config := global.LoadConfig()

	// 创建URL参数
	values := url.Values{}
	values.Add("q", params.Query)
	values.Add("engine", config.SearchAPIConfig.Engine)
	values.Add("api_key", config.SearchAPIConfig.ApiKey) // 需要替换为实际的API密钥

	// 使用通用HTTP客户端
	httpClient := httputils.DefaultHTTPClient()

	// 发送请求并解析JSON响应
	var searchResponse SearchResponse
	err := httpClient.GetJSON(ctx, config.SearchAPIConfig.Endpoint, values, nil, &searchResponse)
	if err != nil {
		log.Error("Web搜索请求失败", zap.Error(err))
		return fmt.Sprintf("Web搜索请求失败: %v", err), nil
	}

	// 提取前5条结果
	resultCount := len(searchResponse.OrganicResults)
	if resultCount > 5 {
		resultCount = 5
	}

	// 如果没有结果，返回空字符串
	if resultCount == 0 {
		log.Debug("Web搜索无结果", zap.String("query", params.Query))
		return "搜索结果为空", nil
	}

	// 将结果转换为JSON字符串并拼接
	var resultStrings []string
	for i := 0; i < resultCount; i++ {
		resultJSON, err := json.Marshal(searchResponse.OrganicResults[i])
		if err != nil {
			log.Warn("结果序列化失败", zap.Error(err))
			continue
		}
		resultStrings = append(resultStrings, string(resultJSON))
	}

	// 拼接结果
	log.Debug("Web搜索完成", zap.Int("resultCount", resultCount))
	return strings.Join(resultStrings, ","), nil
}
