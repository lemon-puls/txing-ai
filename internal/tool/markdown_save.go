package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

// markdownSaveParams 保存Markdown文件参数
type markdownSaveParams struct {
	Content  string `json:"content" jsonschema:"description=要保存的Markdown内容"`
	Filename string `json:"filename" jsonschema:"description=文件名(不含扩展名)"`
	// IsAppend 追加模式：内容过长时可分多次保存到同一文件（第一次 false 新建，后续 true 追加）
	IsAppend bool `json:"is_append,omitempty" jsonschema:"description=是否追加到已存在的文件（内容过长时可分多次保存：第一次传 false，后续传 true 追加到同一文件）"`
}

// saveMarkdown 将Markdown内容保存到本地文件
func saveMarkdown(ctx context.Context, params *markdownSaveParams) (string, error) {
	// 参数校验：文件名与内容均不能为空，避免生成 0 字节空文件浪费轮次
	filename := strings.TrimSpace(params.Filename)
	if filename == "" {
		return "保存失败：文件名为空，请提供文件名（不含扩展名）", nil
	}
	if strings.ContainsAny(filename, `/\`) || strings.Contains(filename, "..") {
		return fmt.Sprintf("保存失败：文件名不合法: %q，请只提供文件名（不含扩展名），不要包含路径", params.Filename), nil
	}
	if strings.TrimSpace(params.Content) == "" {
		return "保存失败：内容为空，请提供要保存的 Markdown 内容", nil
	}

	savePath, err := buildSaveDir(ctx)
	if err != nil {
		return fmt.Sprintf("构建保存目录失败: %v", err), nil
	}

	// 确保文件名有.md扩展名
	if filepath.Ext(filename) != ".md" {
		filename = filename + ".md"
	}

	// 构建完整的文件路径
	fullPath := filepath.Join(savePath, filename)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error("创建目录失败", zap.String("dir", dir), zap.Error(err))
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	if params.IsAppend {
		// 追加模式：打开已存在文件追加内容（不存在则创建）
		f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Error("追加Markdown文件失败", zap.String("path", fullPath), zap.Error(err))
			return "", fmt.Errorf("追加Markdown文件失败: %v", err)
		}
		if _, err := f.WriteString(params.Content); err != nil {
			f.Close()
			log.Error("追加Markdown文件失败", zap.String("path", fullPath), zap.Error(err))
			return "", fmt.Errorf("追加Markdown文件失败: %v", err)
		}
		f.Close()
	} else {
		// 写入文件（覆盖）
		err = os.WriteFile(fullPath, []byte(params.Content), 0644)
		if err != nil {
			log.Error("保存Markdown文件失败", zap.String("path", fullPath), zap.Error(err))
			return "", fmt.Errorf("保存Markdown文件失败: %v", err)
		}
	}

	// 记录成功日志
	log.Info("Markdown文件保存成功",
		zap.String("path", fullPath),
		zap.Int("contentLength", len(params.Content)),
		zap.Bool("append", params.IsAppend),
		zap.Time("timestamp", time.Now()))

	if params.IsAppend {
		return fmt.Sprintf("Markdown内容已成功追加到: ./%s", filename), nil
	}
	return fmt.Sprintf("Markdown文件已成功保存到: ./%s", filename), nil
}

// 展示消息构造
type markdownSaveShowBuilder struct{}

func (markdownSaveShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params markdownSaveParams
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建 Markdown 保存请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	return "保存 Markdown 文件：" + params.Filename, nil
}

func (markdownSaveShowBuilder) BuildResponse(response string) (string, error) {
	return response, nil
}

func init() {
	RegisterShowMsgBuilder(markdownSaveToolName, markdownSaveShowBuilder{})
}
