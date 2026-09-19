package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

// 允许操作的目录列表（与工具描述约定一致：only files in runtime directory are allowed）
// 覆盖 runtime/temp、runtime/temp_files（markdown_save / markdown_to_pdf 保存目录）等
var allowedDirectories = []string{
	"runtime",
}

// 检查文件路径是否在允许的目录中
// 相对路径（./xxx.pdf、xxx.pdf）会先按当前保存目录（runtime/temp_files/<日期>）解析，
// 再按当前工作目录解析，任一命中允许目录即放行
func isPathAllowed(ctx context.Context, path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error("获取当前工作目录失败", zap.Error(err))
		return false
	}

	// 候选绝对路径：原样 + 按保存目录解析
	candidates := []string{path}
	if !filepath.IsAbs(path) {
		if savePath := saveDirQuiet(ctx); savePath != "" {
			rel := strings.TrimPrefix(strings.TrimPrefix(path, "./"), `.\\`)
			candidates = append(candidates, filepath.Join(savePath, filepath.Clean(rel)))
		}
	}

	for _, c := range candidates {
		absPath, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		// 检查路径是否在允许的目录中
		for _, dir := range allowedDirectories {
			allowedPath := filepath.Clean(filepath.Join(currentDir, dir))
			if absPath == allowedPath || strings.HasPrefix(absPath, allowedPath+string(filepath.Separator)) {
				return true
			}
		}
	}

	return false
}

// 文件读取参数
type fileReadParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要读取的文件路径"`
}

// 读取文件内容
func readFile(ctx context.Context, params *fileReadParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(ctx, params.FilePath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.FilePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(params.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", params.FilePath)
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(params.FilePath)
	if err != nil {
		log.Error("读取文件失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	return string(content), nil
}

// 文件写入参数
type fileWriteParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要写入的文件路径"`
	Content  string `json:"content" jsonschema:"description=要写入的文件内容"`
}

// 写入文件内容
func writeFile(ctx context.Context, params *fileWriteParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(ctx, params.FilePath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.FilePath)
	}

	// 确保目录存在
	dir := filepath.Dir(params.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error("创建目录失败", zap.String("dir", dir), zap.Error(err))
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	err := ioutil.WriteFile(params.FilePath, []byte(params.Content), 0644)
	if err != nil {
		log.Error("写入文件失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	return fmt.Sprintf("文件已成功写入: %s", params.FilePath), nil
}

// 文件内容替换参数
type fileReplaceParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要替换内容的文件路径"`
	OldText  string `json:"old_text" jsonschema:"description=要替换的文本"`
	NewText  string `json:"new_text" jsonschema:"description=替换后的文本"`
}

// 替换文件内容
func replaceFileContent(ctx context.Context, params *fileReplaceParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(ctx, params.FilePath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.FilePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(params.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", params.FilePath)
	}

	// 读取文件内容
	content, err := ioutil.ReadFile(params.FilePath)
	if err != nil {
		log.Error("读取文件失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 替换内容
	newContent := strings.Replace(string(content), params.OldText, params.NewText, -1)

	// 写入文件
	err = ioutil.WriteFile(params.FilePath, []byte(newContent), 0644)
	if err != nil {
		log.Error("写入文件失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	return fmt.Sprintf("文件内容已成功替换: %s", params.FilePath), nil
}

// 文件列表参数
type fileListParams struct {
	DirPath string `json:"dir_path" jsonschema:"description=要列出文件的目录路径"`
}

// 文件信息结构
type fileInfo struct {
	Name      string `json:"name"`
	IsDir     bool   `json:"is_dir"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
}

// 列出目录中的文件
func listFiles(ctx context.Context, params *fileListParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(ctx, params.DirPath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.DirPath)
	}

	// 检查目录是否存在
	file, err := os.Stat(params.DirPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("目录不存在: %s", params.DirPath)
	}

	// 确保是目录
	if !file.IsDir() {
		return "", fmt.Errorf("指定路径不是目录: %s", params.DirPath)
	}

	// 读取目录内容
	files, err := ioutil.ReadDir(params.DirPath)
	if err != nil {
		log.Error("读取目录失败", zap.String("path", params.DirPath), zap.Error(err))
		return "", fmt.Errorf("读取目录失败: %v", err)
	}

	// 构建文件信息列表
	fileInfoList := make([]fileInfo, 0, len(files))
	for _, file := range files {
		fileInfoList = append(fileInfoList, fileInfo{
			Name:      file.Name(),
			IsDir:     file.IsDir(),
			Size:      file.Size(),
			UpdatedAt: file.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	// 转换为JSON
	result, err := json.MarshalIndent(fileInfoList, "", "  ")
	if err != nil {
		log.Error("转换文件列表为JSON失败", zap.Error(err))
		return "", fmt.Errorf("转换文件列表为JSON失败: %v", err)
	}

	return string(result), nil
}

// 文件删除参数
type fileDeleteParams struct {
	FilePath string `json:"file_path" jsonschema:"description=要删除的文件路径"`
}

// 删除文件
func deleteFile(ctx context.Context, params *fileDeleteParams) (string, error) {
	// 检查路径是否允许
	if !isPathAllowed(ctx, params.FilePath) {
		return "", fmt.Errorf("不允许访问该路径: %s", params.FilePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(params.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", params.FilePath)
	}

	// 删除文件
	err := os.Remove(params.FilePath)
	if err != nil {
		log.Error("删除文件失败", zap.String("path", params.FilePath), zap.Error(err))
		return "", fmt.Errorf("删除文件失败: %v", err)
	}

	return fmt.Sprintf("文件已成功删除: %s", params.FilePath), nil
}
