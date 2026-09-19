package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestResolveSavedFile 校验 resolveSavedFile 的路径解析：
// - 绝对路径原样返回
// - ./xxx.md 相对保存目录解析（markdown_save_tool 的返回格式）
// - runtime/... 相对工作目录解析（模型可能直接传完整相对路径）
// 不依赖 runtime/config.yaml（saveDirQuiet 在配置缺失时返回空串）
func TestResolveSavedFile(t *testing.T) {
	ctx := context.Background()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前工作目录失败: %v", err)
	}

	// 创建测试文件：runtime/resolve_test/a.md
	testRel := filepath.Join("runtime", "resolve_test", "a.md")
	if err := os.MkdirAll(filepath.Dir(testRel), 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(filepath.Join("runtime", "resolve_test"))
	if err := os.WriteFile(testRel, []byte("# hi"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		wantAbs  string // 期望的绝对路径；空表示只校验不报错
		wantErr  bool
	}{
		{
			name:     "绝对路径原样返回",
			filePath: filepath.Join(currentDir, testRel),
			wantAbs:  filepath.Join(currentDir, testRel),
		},
		{
			name:     "runtime 相对路径解析",
			filePath: testRel,
			wantAbs:  filepath.Join(currentDir, testRel),
		},
		{
			name:     "./ 前缀相对路径",
			filePath: "./" + filepath.ToSlash(testRel),
			wantAbs:  filepath.Join(currentDir, testRel),
		},
		{
			name:     "空路径报错",
			filePath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveSavedFile(ctx, tt.filePath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveSavedFile() 期望报错，实际返回 %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveSavedFile() 意外错误: %v", err)
			}
			if tt.wantAbs != "" && filepath.Clean(got) != filepath.Clean(tt.wantAbs) {
				t.Errorf("resolveSavedFile() = %q, want %q", got, tt.wantAbs)
			}
		})
	}
}
