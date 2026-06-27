package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	adaptercommon "txing-ai/internal/adapter/common"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils/fileextract"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type ChatClient struct {
	Endpoint string
	ApiKey   string
	client   *openai.Client
}

func (c ChatClient) StreamChat(ctx context.Context, conf *adaptercommon.ChatConfig, callback global.Hook) error {
	// 构建消息
	messages := make([]openai.ChatCompletionMessage, 0, len(conf.Message))
	for _, msg := range conf.Message {
		// 检查是否包含图片内容
		hasImageContent := len(msg.Images) > 0
		hasImageAttachment := false
		for _, att := range msg.Attachments {
			if isImageType(att.FileType) {
				hasImageAttachment = true
				break
			}
		}

		// 如果消息包含图片，使用多模态格式
		if (hasImageContent || hasImageAttachment) && msg.Role == global.User {
			parts := []openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: msg.Content},
			}
			// 添加图片 URL
			for _, imgURL := range msg.Images {
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: imgURL,
					},
				})
			}
			// 添加图片类型的附件
			for _, att := range msg.Attachments {
				if isImageType(att.FileType) {
					parts = append(parts, openai.ChatMessagePart{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: att.FileURL,
						},
					})
				}
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:         msg.Role,
				MultiContent: parts,
			})
		} else {
			// 对于非图片内容，将文档附件信息添加到文本中
			content := msg.Content
			if len(msg.Attachments) > 0 && msg.Role == global.User {
				attachmentInfo := buildAttachmentInfo(ctx, msg.Attachments)
				if attachmentInfo != "" {
					if content == "" {
						content = attachmentInfo
					} else {
						content = content + "\n\n" + attachmentInfo
					}
				}
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: content,
			})
		}
	}

	// 构建请求
	req := openai.ChatCompletionRequest{
		Model:    conf.Model,
		Messages: messages,
		Stream:   true,
	}

	// 设置可选参数
	if conf.MaxTokens != nil {
		req.MaxTokens = *conf.MaxTokens
	}
	if conf.Temperature != nil {
		req.Temperature = *conf.Temperature
	}
	if conf.TopP != nil {
		req.TopP = *conf.TopP
	}
	if conf.PresencePenalty != nil {
		req.PresencePenalty = *conf.PresencePenalty
	}
	if conf.FrequencyPenalty != nil {
		req.FrequencyPenalty = *conf.FrequencyPenalty
	}

	// 创建流式请求
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Error("create stream error", zap.Error(err))
		return fmt.Errorf("create stream error: %v", err)
	}
	defer stream.Close()

	// 处理流式响应
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				log.Error("stream recv error", zap.Error(err))
				return fmt.Errorf("stream recv error: %v", err)
			}

			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				content := choice.Delta.Content
				reasoningContent := choice.Delta.ReasoningContent
				if content != "" || reasoningContent != "" {
					chunk := &global.Chunk{
						Content:          choice.Delta.Content,
						ReasoningContent: choice.Delta.ReasoningContent,
					}
					if err := callback(chunk); err != nil {
						log.Error("callback error", zap.Error(err))
						return fmt.Errorf("callback error: %v", err)
					}
				}
			}
		}
	}
}

func NewChatClient(endpoint, apiKey string) *ChatClient {
	config := openai.DefaultConfig(apiKey)
	if endpoint != "" {
		config.BaseURL = endpoint
	}
	client := openai.NewClientWithConfig(config)
	return &ChatClient{
		Endpoint: endpoint,
		ApiKey:   apiKey,
		client:   client,
	}
}

// isImageType 判断文件类型是否为图片
func isImageType(fileType string) bool {
	imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"}
	for _, t := range imageTypes {
		if fileType == t {
			return true
		}
	}
	return false
}

// buildAttachmentInfo 构建附件信息文本
// 会尝试提取文件文本内容，如果提取失败则只返回元数据
func buildAttachmentInfo(ctx context.Context, attachments []global.Attachment) string {
	if len(attachments) == 0 {
		return ""
	}

	var info string
	for _, att := range attachments {
		// 只处理非图片类型的附件
		if isImageType(att.FileType) {
			continue
		}

		// 尝试提取文件文本内容
		extractedText, err := extractFileText(ctx, att)
		if err != nil {
			// 提取失败，只返回元数据
			log.Warn("提取文件文本失败，使用元数据",
				zap.String("file", att.FileName),
				zap.String("type", att.FileType),
				zap.Error(err))
			info += fmt.Sprintf("[附件: %s (类型: %s, 大小: %d bytes)]\n", att.FileName, att.FileType, att.FileSize)
			continue
		}

		// 提取成功，返回文件内容
		if extractedText != "" {
			// 限制内容长度，避免超出 LLM 上下文限制
			const maxContentLength = 50000 // 约 50KB 文本
			if len(extractedText) > maxContentLength {
				extractedText = extractedText[:maxContentLength] + "\n...(内容已截断)"
			}
			info += fmt.Sprintf("[文件: %s]\n%s\n\n", att.FileName, extractedText)
		} else {
			info += fmt.Sprintf("[附件: %s (类型: %s, 大小: %d bytes, 内容为空)]\n", att.FileName, att.FileType, att.FileSize)
		}
	}
	return info
}

// extractFileText 提取文件文本内容
func extractFileText(ctx context.Context, att global.Attachment) (string, error) {
	// 检查是否支持该文件类型
	if !fileextract.Default.IsSupported(att.FileType) {
		return "", fmt.Errorf("unsupported file type: %s", att.FileType)
	}

	// 获取提取器
	extractor, err := fileextract.Default.GetExtractor(att.FileType)
	if err != nil {
		return "", err
	}

	// 下载文件到临时目录
	tmpFile, err := downloadToTemp(ctx, att.FileURL, att.FileName)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer os.Remove(tmpFile) // 清理临时文件

	// 提取文本
	return extractor.Extract(ctx, tmpFile)
}

// downloadToTemp 下载文件到临时目录
func downloadToTemp(ctx context.Context, url, fileName string) (string, error) {
	// 创建临时目录
	tmpDir := filepath.Join(os.TempDir(), "txing-ai-fileextract")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}

	// 创建临时文件
	tmpFile := filepath.Join(tmpDir, fileName)

	// 下载文件
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 写入文件
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return "", err
	}

	return tmpFile, nil
}

var _ adaptercommon.ChatRequester = (*ChatClient)(nil)
