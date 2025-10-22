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
	"txing-ai/internal/utils/httputils"

	"go.uber.org/zap"
)

// imageDownloadParams 下载图片参数
type imageDownloadParams struct {
	ImageURL string `json:"image_url" jsonschema:"description=图片URL地址"`
	Filename string `json:"filename,omitempty" jsonschema:"description=保存的文件名(可选,默认从URL获取)"`
}

// downloadImage 下载图片到本地
func downloadImage(ctx context.Context, params *imageDownloadParams) (string, error) {
	// 验证图片URL
	if params.ImageURL == "" {
		return "保存失败，图片URL不能为空", nil
	}

	savePath, err := buildSaveDir(ctx)
	if err != nil {
		return fmt.Sprintf("构建保存目录失败: %v", err), nil
	}

	// 确定文件名
	filename := params.Filename
	if filename == "" {
		// 从URL中提取文件名
		urlParts := strings.Split(params.ImageURL, "/")
		if len(urlParts) > 0 {
			filename = urlParts[len(urlParts)-1]
			// 移除URL参数
			if idx := strings.Index(filename, "?"); idx > 0 {
				filename = filename[:idx]
			}
		}

		// 如果仍然没有文件名，使用时间戳
		if filename == "" {
			filename = fmt.Sprintf("image_%d", time.Now().Unix())
		}
	}

	// 确保文件名有扩展名
	if filepath.Ext(filename) == "" {
		// 默认添加.jpg扩展名
		filename = filename + ".jpg"
	}

	// 构建完整的文件路径
	fullPath := filepath.Join(savePath, filename)

	// 使用项目的HTTP工具下载图片
	httpClient := httputils.DefaultHTTPClient()

	// 记录开始下载
	log.Debug("开始下载图片", zap.String("url", params.ImageURL))

	// 直接使用Get方法下载图片内容
	imageData, err := httpClient.Get(ctx, params.ImageURL, nil, nil)
	if err != nil {
		log.Error("下载图片失败", zap.String("url", params.ImageURL), zap.Error(err))
		return "下载图片失败，" + err.Error(), nil
	}

	// 检查是否成功获取到图片数据
	if len(imageData) == 0 {
		log.Error("下载的图片数据为空", zap.String("url", params.ImageURL))
		return "下载的图片数据为空", nil
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		log.Error("创建文件失败", zap.String("path", fullPath), zap.Error(err))
		return "创建文件失败", nil
	}
	defer file.Close()

	// 将图片数据写入文件
	_, err = file.Write(imageData)
	if err != nil {
		log.Error("写入文件失败", zap.String("path", fullPath), zap.Error(err))
		return "写入文件失败", nil
	}

	// 记录成功日志
	log.Info("图片下载成功",
		zap.String("url", params.ImageURL),
		zap.String("path", fullPath),
		zap.Time("timestamp", time.Now()))

	return fmt.Sprintf("下载完成: ./%s", filename), nil
}

// 展示消息构造
type imageDownloadShowBuilder struct{}

func (imageDownloadShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params imageDownloadParams
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建图片下载请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	return "图片保存到本地：" + params.Filename + "(" + params.ImageURL + ")", nil
}

func (imageDownloadShowBuilder) BuildResponse(response string) (string, error) {
	return response, nil
}

func init() {
	RegisterShowMsgBuilder(imageDownloadToolName, imageDownloadShowBuilder{})
}
