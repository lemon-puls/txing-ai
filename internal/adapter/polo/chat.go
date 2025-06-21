package polo

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

type ChatClient struct {
	Endpoint string
	ApiKey   string
}

func (c ChatClient) StreamChat(ctx context.Context, conf *adaptercommon.ChatConfig, callback global.Hook) error {
	// 构建请求体
	payload := map[string]interface{}{
		"model":    conf.Model,
		"stream":   true,
		"messages": conf.Message,
	}

	// 添加可选参数
	if conf.MaxTokens != nil {
		payload["max_tokens"] = *conf.MaxTokens
	}
	if conf.Temperature != nil {
		payload["temperature"] = *conf.Temperature
	}
	if conf.TopP != nil {
		payload["top_p"] = *conf.TopP
	}
	if conf.TopK != nil {
		payload["top_k"] = *conf.TopK
	}
	if conf.PresencePenalty != nil {
		payload["presence_penalty"] = *conf.PresencePenalty
	}
	if conf.FrequencyPenalty != nil {
		payload["frequency_penalty"] = *conf.FrequencyPenalty
	}
	if conf.RepetitionPenalty != nil {
		payload["repetition_penalty"] = *conf.RepetitionPenalty
	}

	// 将请求体转换为 JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request body failed: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "sk-"+c.ApiKey)
	req.Header.Set("User-Agent", "PoloAPI/1.0.0 (https://poloai.top)")
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 2 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 读取响应流
	buffer := ""
	reader := bufio.NewReader(resp.Body)
	for {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read response failed: %v", err)
		}

		// 处理空行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理数据行
		if strings.HasPrefix(line, "data: ") {
			dataLine := line[6:] // 去掉 "data: " 前缀
			if dataLine == "[DONE]" {
				break
			}

			// 解析 JSON 数据
			var data struct {
				Choices []struct {
					Delta struct {
						Content           string `json:"content"`
						Reasoning_content string `json:"reasoning_content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(dataLine), &data); err != nil {
				// 如果 JSON 解析失败，可能是数据不完整，继续累积 buffer
				buffer = line + "\n" + buffer
				continue
			}

			// 处理消息块
			if len(data.Choices) > 0 {
				content := data.Choices[0].Delta.Content
				reasoning_content := data.Choices[0].Delta.Reasoning_content
				if content != "" || reasoning_content != "" {
					chunk := &global.Chunk{
						Content:          content,
						ReasoningContent: reasoning_content,
					}
					if err := callback(chunk); err != nil {
						log.Error("callback error", zap.Error(err))
						return err
					}
				}
			}
		}
	}

	return nil
}

func NewChatClient(endpoint, apiKey string) *ChatClient {
	return &ChatClient{
		Endpoint: endpoint,
		ApiKey:   apiKey,
	}
}

var _ adaptercommon.ChatRequester = (*ChatClient)(nil)
