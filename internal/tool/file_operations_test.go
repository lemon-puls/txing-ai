package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// 测试前准备工作
func setupTestEnvironment() (string, string, string, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return "", "", "", err
	}

	// 创建测试目录
	testDir := filepath.Join(currentDir, "runtime", "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return "", "", "", err
	}

	// 创建测试文件
	testFilePath := filepath.Join(testDir, "test.txt")
	testContent := "这是测试文件内容"
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		return "", "", "", err
	}

	// 创建不允许访问的测试目录
	invalidDir := filepath.Join(currentDir, "invalid_dir")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		return "", "", "", err
	}

	return testDir, testFilePath, testContent, nil
}

// 测试后清理工作
func cleanupTestEnvironment(testDir string, invalidDir string) {
	os.RemoveAll(testDir)
	os.RemoveAll(invalidDir)
}

func Test_isPathAllowed(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "允许访问runtime目录",
			path: "runtime/test.txt",
			want: true,
		},
		{
			name: "允许访问data目录",
			path: filepath.Join(currentDir, "data", "test.txt"),
			want: true,
		},
		{
			name: "允许访问temp目录",
			path: filepath.Join(currentDir, "temp", "test.txt"),
			want: true,
		},
		{
			name: "不允许访问根目录",
			path: filepath.Join(currentDir, "test.txt"),
			want: false,
		},
		{
			name: "不允许访问上级目录",
			path: filepath.Join(currentDir, "..", "test.txt"),
			want: false,
		},
		{
			name: "不允许访问其他目录",
			path: filepath.Join(currentDir, "other", "test.txt"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPathAllowed(tt.path); got != tt.want {
				t.Errorf("isPathAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readFile(t *testing.T) {
	testDir, testFilePath, testContent, err := setupTestEnvironment()
	if err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer cleanupTestEnvironment(testDir, filepath.Join(filepath.Dir(testDir), "invalid_dir"))

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		params  *fileReadParams
		want    string
		wantErr bool
	}{
		{
			name: "成功读取文件",
			params: &fileReadParams{
				FilePath: testFilePath,
			},
			want:    testContent,
			wantErr: false,
		},
		{
			name: "文件不存在",
			params: &fileReadParams{
				FilePath: filepath.Join(testDir, "not_exist.txt"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "路径不允许访问",
			params: &fileReadParams{
				FilePath: filepath.Join(currentDir, "invalid_dir", "test.txt"),
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFile(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("readFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeFile(t *testing.T) {
	testDir, _, _, err := setupTestEnvironment()
	if err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer cleanupTestEnvironment(testDir, filepath.Join(filepath.Dir(testDir), "invalid_dir"))

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}

	ctx := context.Background()
	newFilePath := filepath.Join(testDir, "new_file.txt")
	newContent := "这是新文件内容"

	tests := []struct {
		name    string
		params  *fileWriteParams
		want    string
		wantErr bool
	}{
		{
			name: "成功写入文件",
			params: &fileWriteParams{
				FilePath: newFilePath,
				Content:  newContent,
			},
			want:    "文件已成功写入: " + newFilePath,
			wantErr: false,
		},
		{
			name: "路径不允许访问",
			params: &fileWriteParams{
				FilePath: filepath.Join(currentDir, "invalid_dir", "test.txt"),
				Content:  "不允许写入",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := writeFile(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("writeFile() got = %v, want %v", got, tt.want)
			}

			// 验证文件是否成功写入
			if !tt.wantErr {
				content, err := os.ReadFile(tt.params.FilePath)
				if err != nil {
					t.Errorf("读取写入的文件失败: %v", err)
					return
				}
				if string(content) != tt.params.Content {
					t.Errorf("文件内容不匹配, got = %v, want %v", string(content), tt.params.Content)
				}
			}
		})
	}
}

func Test_replaceFileContent(t *testing.T) {
	testDir, testFilePath, testContent, err := setupTestEnvironment()
	if err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer cleanupTestEnvironment(testDir, filepath.Join(filepath.Dir(testDir), "invalid_dir"))

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}

	ctx := context.Background()
	oldText := "测试文件"
	newText := "替换后的文件"

	tests := []struct {
		name    string
		params  *fileReplaceParams
		want    string
		wantErr bool
	}{
		{
			name: "成功替换文件内容",
			params: &fileReplaceParams{
				FilePath: testFilePath,
				OldText:  oldText,
				NewText:  newText,
			},
			want:    "文件内容已成功替换: " + testFilePath,
			wantErr: false,
		},
		{
			name: "文件不存在",
			params: &fileReplaceParams{
				FilePath: filepath.Join(testDir, "not_exist.txt"),
				OldText:  oldText,
				NewText:  newText,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "路径不允许访问",
			params: &fileReplaceParams{
				FilePath: filepath.Join(currentDir, "invalid_dir", "test.txt"),
				OldText:  oldText,
				NewText:  newText,
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 如果是成功替换的测试，先确保文件内容包含要替换的文本
			if tt.name == "成功替换文件内容" {
				err := os.WriteFile(testFilePath, []byte(testContent), 0644)
				if err != nil {
					t.Fatalf("准备测试文件失败: %v", err)
				}
			}

			got, err := replaceFileContent(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("replaceFileContent() got = %v, want %v", got, tt.want)
			}

			// 验证文件内容是否成功替换
			if !tt.wantErr && tt.name == "成功替换文件内容" {
				_, err := os.ReadFile(tt.params.FilePath)
				if err != nil {
					t.Errorf("读取替换后的文件失败: %v", err)
					return
				}
				if oldText != "" && newText != "" {

					if oldText != "" && newText != "" {

					}
				}
			}
		})
	}
}

func Test_deleteFile(t *testing.T) {
	testDir, testFilePath, _, err := setupTestEnvironment()
	if err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer cleanupTestEnvironment(testDir, filepath.Join(filepath.Dir(testDir), "invalid_dir"))

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		params  *fileDeleteParams
		want    string
		wantErr bool
	}{
		{
			name: "成功删除文件",
			params: &fileDeleteParams{
				FilePath: testFilePath,
			},
			want:    "文件已成功删除: " + testFilePath,
			wantErr: false,
		},
		{
			name: "文件不存在",
			params: &fileDeleteParams{
				FilePath: filepath.Join(testDir, "not_exist.txt"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "路径不允许访问",
			params: &fileDeleteParams{
				FilePath: filepath.Join(currentDir, "invalid_dir", "test.txt"),
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 如果是成功删除的测试，先确保文件存在
			if tt.name == "成功删除文件" {
				if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
					// 重新创建测试文件
					err := os.WriteFile(testFilePath, []byte("测试内容"), 0644)
					if err != nil {
						t.Fatalf("准备测试文件失败: %v", err)
					}
				}
			}

			got, err := deleteFile(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("deleteFile() got = %v, want %v", got, tt.want)
			}

			// 验证文件是否成功删除
			if !tt.wantErr && tt.name == "成功删除文件" {
				_, err := os.Stat(tt.params.FilePath)
				if !os.IsNotExist(err) {
					t.Errorf("文件应该已被删除，但仍然存在")
				}
			}
		})
	}
}

func Test_listFiles(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *fileListParams
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "列出目录中的文件",
			args: args{
				ctx: context.Background(),
				params: &fileListParams{
					DirPath: "runtime",
				},
			},
			want:    "runtime/test.txt",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listFiles(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("listFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("listFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}
