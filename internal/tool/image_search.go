package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils/httputils"

	"go.uber.org/zap"
)

// imageSearchParams 图片搜索参数
type imageSearchParams struct {
	Words string `json:"words" jsonschema:"description=搜索关键词"`
	Page  int    `json:"page,omitempty" jsonschema:"description=页码,默认为1"`
	Type  int    `json:"type,omitempty" jsonschema:"description=返回类型,1=预览图地址,2=源图地址,默认为1"`
}

// 图片搜索结果
type ImageSearchResult struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Res  []string `json:"res"`
	Page string   `json:"page"`
}

// searchImageSougou 搜索图片
func searchImageSougou(ctx context.Context, params *imageSearchParams) (string, error) {

	config := global.LoadConfig()

	// 验证搜索关键词
	if params.Words == "" {
		return "搜索关键词不能为空", nil
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Type != 1 && params.Type != 2 {
		params.Type = 1
	}

	// 创建URL参数
	values := url.Values{}
	values.Add("id", config.ImageSearchConfig.Sougou.ID)   // 使用公共ID
	values.Add("key", config.ImageSearchConfig.Sougou.Key) // 使用公共KEY
	values.Add("words", params.Words)
	values.Add("page", fmt.Sprintf("%d", params.Page))
	values.Add("type", fmt.Sprintf("%d", params.Type))

	// 使用通用HTTP客户端
	httpClient := httputils.DefaultHTTPClient()

	// 记录请求日志
	log.Debug("发起图片搜索请求",
		zap.String("words", params.Words),
		zap.Int("page", params.Page),
		zap.Int("type", params.Type))

	// 发送请求并解析JSON响应
	apiURL := config.ImageSearchConfig.Sougou.Endpoint
	var searchResponse ImageSearchResult
	err := httpClient.GetJSON(ctx, apiURL, values, nil, &searchResponse)
	if err != nil {
		log.Error("图片搜索请求失败", zap.Error(err))
		return "图片搜索请求失败", nil
	}

	// 检查响应状态
	if searchResponse.Code != 200 {
		log.Error("图片搜索API返回错误",
			zap.Int("code", searchResponse.Code),
			zap.String("msg", searchResponse.Msg))
		return searchResponse.Msg, nil
	}

	// 如果没有结果，返回提示信息
	if len(searchResponse.Res) == 0 {
		log.Debug("图片搜索无结果", zap.String("words", params.Words))
		return "未找到相关图片", nil
	}

	// 将结果转换为JSON字符串
	resultJSON, err := json.Marshal(searchResponse.Res)
	if err != nil {
		log.Error("结果序列化失败", zap.Error(err))
		return "结果序列化失败", nil
	}

	//记录成功日志
	log.Info("图片搜索完成",
		zap.String("words", params.Words),
		zap.Int("resultCount", len(searchResponse.Res)),
		zap.Time("timestamp", time.Now()))

	return string(resultJSON), nil
}

// 展示消息构造
type imageSearchShowBuilder struct{}

func (imageSearchShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params imageSearchParams
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建图片查询请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	return "图片搜索：" + params.Words, nil
}

func (imageSearchShowBuilder) BuildResponse(response string) (string, error) {
	// 判断是否被限流了
	if strings.Contains(response, "调用频次过快") {
		return "图片搜索请求过于频繁，请稍后再试", nil
	}

	// 响应示例为 URL 列表 JSON： ["https://...", "https://..."]
	var urls []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(response)), &urls); err != nil {
		// 解析失败时直接原样返回，避免前端无法显示错误信息
		log.Error("构建图片搜索响应显示信息失败", zap.Error(err))
		return response, nil
	}
	if len(urls) == 0 {
		return "未找到相关图片", nil
	}
	var b strings.Builder
	// 使用 flex 容器让图片从左到右排列，自动换行
	b.WriteString("<div style=\"display:flex;flex-wrap:wrap;gap:8px;align-items:flex-start;\">")
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		// 去除可能包裹的反引号或引号
		u = strings.Trim(u, " `\"")
		if u == "" {
			continue
		}
		b.WriteString("<img src=\"" + html.EscapeString(u) + "\" alt=\"image\" style=\"width:120px;height:100px;object-fit:cover;border-radius:6px;flex:0 0 auto;\"/>")
	}
	b.WriteString("</div>")
	return b.String(), nil
}

func init() {
	RegisterShowMsgBuilder(imageSearchToolName, imageSearchShowBuilder{})
}
