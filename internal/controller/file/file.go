package file

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 文件上传响应
type UploadResponse struct {
	FileName string `json:"fileName"` // 文件名
	FileSize int64  `json:"fileSize"` // 文件大小
	FileURL  string `json:"fileUrl"`  // 文件下载URL
}

// Upload 上传文件
// @Summary 上传文件
// @Description 上传文件到服务器本地
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "用户令牌"
// @Param file formData file true "文件"
// @Success 200 {object} utils.Response{data=UploadResponse} "成功"
// @Failure 400 {object} utils.Response "请求错误"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/file/upload [post]
func Upload(c *gin.Context) {

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("获取上传文件失败", zap.Error(err))
		utils.ErrorWithMsg(c, "获取上传文件失败", err)
		return
	}
	defer file.Close()

	localUploadConfig := global.LoadConfig().LocalUploadConfig

	// 检查文件大小
	if header.Size > int64(localUploadConfig.MaxSize) {
		log.Error("文件大小超出限制", zap.Int64("size", header.Size), zap.Int("max_size", localUploadConfig.MaxSize))
		utils.ErrorWithMsg(c, "文件大小超出限制: "+strconv.FormatInt(int64(localUploadConfig.MaxSize/1024/1024), 10)+" MB", fmt.Errorf("文件大小超出限制"))
		return
	}

	userId := utils.GetUIDFromContext(c)

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("获取当前工作目录失败", zap.Error(err))
		utils.ErrorWithMsg(c, "获取当前工作目录失败", err)
		return
	}

	// 获取当前日期作为目录名
	currentDate := time.Now().Format("2006-01-02")

	fileDir := filepath.Join(localUploadConfig.Dir, strconv.FormatInt(userId, 10), currentDate)

	// 生成唯一文件名（时间戳+原始文件名）
	fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	filePath := filepath.Join(fileDir, fileName)

	// 确保上传目录存在（用户ID/日期）
	absUploadDir := filepath.Join(currentDir, fileDir)
	if err := os.MkdirAll(absUploadDir, 0755); err != nil {
		log.Error("创建上传目录失败", zap.String("dir", absUploadDir), zap.Error(err))
		utils.ErrorWithMsg(c, "创建上传目录失败", err)
		return
	}

	absUploadPath := filepath.Join(absUploadDir, fileName)

	// 创建目标文件
	dst, err := os.Create(absUploadPath)
	if err != nil {
		log.Error("创建文件失败", zap.String("path", absUploadPath), zap.Error(err))
		utils.ErrorWithMsg(c, "创建文件失败", err)
		return
	}
	defer dst.Close()

	// 将上传的文件内容复制到目标文件
	if _, err = io.Copy(dst, file); err != nil {
		log.Error("保存文件失败", zap.String("path", absUploadPath), zap.Error(err))
		utils.ErrorWithMsg(c, "保存文件失败", err)
		return
	}

	// 获取文件大小
	fileInfo, err := os.Stat(absUploadPath)
	if err != nil {
		log.Error("获取文件信息失败", zap.String("path", filePath), zap.Error(err))
		utils.ErrorWithMsg(c, "获取文件信息失败", err)
		return
	}

	// 记录成功日志
	log.Info("文件上传成功",
		zap.String("filename", header.Filename),
		zap.String("path", filePath),
		zap.Int64("size", fileInfo.Size()),
		zap.Time("timestamp", time.Now()))

	// 返回响应
	utils.OkWithData(c, UploadResponse{
		FileName: header.Filename,
		FileSize: fileInfo.Size(),
		FileURL:  fmt.Sprintf("%s/%s/%s", strconv.FormatInt(userId, 10), currentDate, fileName),
	})
}

// Download 下载文件
// @Summary 下载文件
// @Description 从服务器下载文件
// @Tags 文件
// @Produce octet-stream
// @Param filePath query string true "文件相对路径"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} utils.Response "请求错误"
// @Failure 404 {object} utils.Response "文件不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/file/download [get]
func Download(c *gin.Context) {
	// 获取文件名
	filePath := c.Query("filePath")
	if filePath == "" {
		utils.ErrorWithCode(c, global.CodeInvalidParams, nil)
		return
	}

	config := global.LoadConfig().LocalUploadConfig

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current working directory", zap.Error(err))
		utils.ErrorWithMsg(c, "获取当前工作目录失败", err)
		return
	}

	// 构建文件路径
	absFilePath := filepath.Join(currentDir, config.Dir, filePath)

	// 检查文件是否存在
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		log.Error("File does not exist", zap.String("path", absFilePath))
		utils.ErrorWithMsg(c, "文件不存在", err)
		return
	}

	// 获取原始文件名
	originalFileName := filepath.Base(filePath)

	// URL编码文件名，确保特殊字符正确处理
	encodedFileName := url.QueryEscape(originalFileName)

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", originalFileName, encodedFileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Cache-Control", "no-cache")

	// 使用绝对路径提供文件
	c.File(absFilePath)

	// 记录下载日志
	log.Info("File downloaded",
		zap.String("filename", originalFileName),
		zap.String("path", absFilePath),
		zap.Time("timestamp", time.Now()))

	utils.Ok(c)
}
