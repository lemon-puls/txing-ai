package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SaveUploadedFile 保存上传的文件到本地并返回相对路径
// 参数:
// - file: 上传的文件
// - userId: 用户ID，用于创建用户专属目录
// - customDir: 自定义子目录名，如果为空则使用当前日期
// - customFileName: 自定义文件名，如果为空则使用时间戳+原始文件名
// 返回:
// - string: 文件相对路径
// - int64: 文件大小
// - error: 错误信息
func SaveUploadedFile(file io.ReadCloser, fileName string, userId int64, customDir string, customFileName string) (string, int64, error) {
	localUploadConfig := global.LoadConfig().LocalUploadConfig

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current working directory", zap.Error(err))
		return "", 0, fmt.Errorf("获取当前工作目录失败: %v", err)
	}

	// 确定目录名
	dirName := customDir
	if dirName == "" {
		// 获取当前日期作为目录名
		dirName = time.Now().Format("2006-01-02")
	}

	// 构建文件目录路径
	fileDir := filepath.Join(localUploadConfig.Dir, strconv.FormatInt(userId, 10), dirName)

	// 确定文件名
	finalFileName := customFileName
	if finalFileName == "" {
		// 生成唯一文件名（时间戳+原始文件名）
		finalFileName = fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName)
	}

	// 构建相对路径和绝对路径
	relativePath := filepath.Join(".", fileDir, finalFileName)
	absUploadDir := filepath.Join(currentDir, fileDir)
	absUploadPath := filepath.Join(absUploadDir, finalFileName)

	// 确保上传目录存在
	if err := os.MkdirAll(absUploadDir, 0755); err != nil {
		log.Error("Failed to create upload directory", zap.String("dir", absUploadDir), zap.Error(err))
		return "", 0, fmt.Errorf("创建上传目录失败: %v", err)
	}

	// 创建目标文件
	dst, err := os.Create(absUploadPath)
	if err != nil {
		log.Error("Failed to create file", zap.String("path", absUploadPath), zap.Error(err))
		return "", 0, fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 将上传的文件内容复制到目标文件
	if _, err = io.Copy(dst, file); err != nil {
		log.Error("Failed to save file", zap.String("path", absUploadPath), zap.Error(err))
		return "", 0, fmt.Errorf("保存文件失败: %v", err)
	}

	// 获取文件大小
	fileInfo, err := os.Stat(absUploadPath)
	if err != nil {
		log.Error("Failed to get file info", zap.String("path", absUploadPath), zap.Error(err))
		return "", 0, fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 记录成功日志
	log.Info("File uploaded successfully",
		zap.String("filename", fileName),
		zap.String("path", relativePath),
		zap.Int64("size", fileInfo.Size()),
		zap.Time("timestamp", time.Now()))

	return relativePath, fileInfo.Size(), nil
}

// SaveUploadedFileFromRequest 从请求中获取文件并保存到本地
// 参数:
// - c: gin上下文
// - formName: 表单字段名
// - userId: 用户ID
// - customDir: 自定义子目录名，如果为空则使用当前日期
// - customFileName: 自定义文件名，如果为空则使用时间戳+原始文件名
// 返回:
// - string: 文件相对路径
// - int64: 文件大小
// - error: 错误信息
func SaveUploadedFileFromRequest(c *gin.Context, formName string, userId int64, customDir string, customFileName string) (string, int64, error) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile(formName)
	if err != nil {
		log.Error("Failed to get uploaded file", zap.Error(err))
		return "", 0, fmt.Errorf("获取上传文件失败: %v", err)
	}
	defer file.Close()

	// 检查文件大小
	localUploadConfig := global.LoadConfig().LocalUploadConfig
	if header.Size > int64(localUploadConfig.MaxSize) {
		log.Error("File size exceeds limit", zap.Int64("size", header.Size), zap.Int("max_size", localUploadConfig.MaxSize))
		return "", 0, fmt.Errorf("文件大小超出限制: %d MB", localUploadConfig.MaxSize/1024/1024)
	}

	return SaveUploadedFile(file, header.Filename, userId, customDir, customFileName)
}
