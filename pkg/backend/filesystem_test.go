package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemBackend_WriteAndRead(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 写入文件
	content := "Hello, Filesystem!"
	result, err := backend.WriteFile(ctx, "/test.txt", content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if result.BytesWritten != len(content) {
		t.Errorf("Expected %d bytes written, got %d", len(content), result.BytesWritten)
	}

	// 读取文件
	readContent, err := backend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}

	// 验证文件确实存在于磁盘上
	fullPath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("File should exist on disk")
	}
}

func TestFilesystemBackend_VirtualMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 尝试路径遍历（应该失败）
	_, err = backend.WriteFile(ctx, "/../etc/passwd", "malicious")
	if err == nil {
		t.Error("Expected error for path traversal, got nil")
	}

	// 尝试使用 .. （应该失败）
	_, err = backend.WriteFile(ctx, "/test/../../../etc/passwd", "malicious")
	if err == nil {
		t.Error("Expected error for path traversal with .., got nil")
	}
}

func TestFilesystemBackend_ListFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 创建多个文件
	backend.WriteFile(ctx, "/file1.txt", "content1")
	backend.WriteFile(ctx, "/file2.txt", "content2")
	backend.WriteFile(ctx, "/subdir/file3.txt", "content3")

	// 列出根目录
	files, err := backend.ListFiles(ctx, "/")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	// 应该有 2 个文件和 1 个目录
	if len(files) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(files))
	}
}

func TestFilesystemBackend_EditFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 写入初始内容
	content := "Hello, World!\nHello, Go!"
	backend.WriteFile(ctx, "/test.txt", content)

	// 编辑文件
	result, err := backend.EditFile(ctx, "/test.txt", "Hello", "Hi", false)
	if err != nil {
		t.Fatalf("EditFile failed: %v", err)
	}

	if result.Replacements != 1 {
		t.Errorf("Expected 1 replacement, got %d", result.Replacements)
	}

	// 验证内容
	newContent, _ := backend.ReadFile(ctx, "/test.txt", 0, 0)
	expected := "Hi, World!\nHello, Go!"
	if newContent != expected {
		t.Errorf("Expected content %q, got %q", expected, newContent)
	}
}

func TestFilesystemBackend_Grep(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 创建多个文件
	backend.WriteFile(ctx, "/file1.txt", "Hello World\nGoodbye World")
	backend.WriteFile(ctx, "/file2.txt", "Hello Go\nGoodbye Go")

	// 搜索 "Hello"
	matches, err := backend.Grep(ctx, "Hello", "", "")
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
}

func TestFilesystemBackend_Glob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 创建多个文件
	backend.WriteFile(ctx, "/file1.txt", "content1")
	backend.WriteFile(ctx, "/file2.go", "content2")
	backend.WriteFile(ctx, "/file3.txt", "content3")

	// 查找所有 .txt 文件
	files, err := backend.Glob(ctx, "*.txt", "/")
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 .txt files, got %d", len(files))
	}
}

func TestFilesystemBackend_Grep_Regex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 创建测试文件
	content := `func NewBackend() Backend {
	return &backend{}
}

func NewHandler() Handler {
	return &handler{}
}

var TODO = "something"
var FIXME = "another thing"
`
	backend.WriteFile(ctx, "/test.go", content)

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
		description   string
	}{
		{
			name:          "regex function definition",
			pattern:       `func\s+New\w+`,
			expectedCount: 2,
			description:   "应匹配 NewBackend 和 NewHandler",
		},
		{
			name:          "regex multiple keywords",
			pattern:       `TODO|FIXME`,
			expectedCount: 2,
			description:   "应匹配 TODO 和 FIXME",
		},
		{
			name:          "literal string",
			pattern:       "TODO",
			expectedCount: 1,
			description:   "字面字符串仍然有效",
		},
		{
			name:          "regex with escaped special chars",
			pattern:       `func.*\(\)`,
			expectedCount: 2,
			description:   "应匹配带括号的函数定义",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := backend.Grep(ctx, tt.pattern, "", "")
			if err != nil {
				t.Fatalf("Grep failed: %v", err)
			}

			if len(matches) != tt.expectedCount {
				t.Errorf("%s: expected %d matches, got %d\nPattern: %s",
					tt.description, tt.expectedCount, len(matches), tt.pattern)
				for _, m := range matches {
					t.Logf("  Match: line %d: %s", m.LineNumber, m.Line)
				}
			}
		})
	}
}

func TestFilesystemBackend_Grep_InvalidRegex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 创建测试文件
	backend.WriteFile(ctx, "/test.txt", "[unclosed bracket in content\nanother line")

	// 使用无效的正则表达式（未闭合的括号）
	// 应该降级到字面字符串匹配
	matches, err := backend.Grep(ctx, "[unclosed", "", "")
	if err != nil {
		t.Fatalf("Grep should not fail on invalid regex: %v", err)
	}

	// 降级后会搜索字面字符串 "[unclosed"
	if len(matches) != 1 {
		t.Errorf("Expected 1 match for literal string '[unclosed', got %d", len(matches))
		for _, m := range matches {
			t.Logf("  Match: line %d: %s", m.LineNumber, m.Line)
		}
	}

	// 测试正常的字面字符串（包含特殊字符）
	backend.WriteFile(ctx, "/test2.txt", "test[pattern]here\nanother line")
	matches, err = backend.Grep(ctx, `test\[pattern\]here`, "", "")
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match for escaped pattern, got %d", len(matches))
	}
}

func TestFilesystemBackend_Grep_Gitignore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 创建 .gitignore 文件
	gitignoreContent := `# 忽略规则
*.log
temp/
build/
node_modules/
`
	backend.WriteFile(ctx, "/.gitignore", gitignoreContent)

	// 创建各种文件
	backend.WriteFile(ctx, "/main.go", "package main // TODO: implement")
	backend.WriteFile(ctx, "/debug.log", "error log TODO: fix")
	backend.WriteFile(ctx, "/temp/cache.txt", "cached TODO: clear")
	backend.WriteFile(ctx, "/build/output.txt", "build TODO: optimize")
	backend.WriteFile(ctx, "/node_modules/pkg/index.js", "TODO: update")
	backend.WriteFile(ctx, "/src/app.go", "package app // TODO: refactor")

	// 搜索 "TODO"，应该只匹配未被忽略的文件
	matches, err := backend.Grep(ctx, "TODO", "", "")
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}

	// 只应该匹配 main.go 和 src/app.go
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
		for _, m := range matches {
			t.Logf("  Match: %s line %d", m.Path, m.LineNumber)
		}
	}

	// 验证匹配的是正确的文件
	foundMain := false
	foundApp := false
	for _, m := range matches {
		if m.Path == "/main.go" {
			foundMain = true
		}
		if m.Path == "/src/app.go" {
			foundApp = true
		}
	}

	if !foundMain {
		t.Error("Expected to find match in main.go")
	}
	if !foundApp {
		t.Error("Expected to find match in src/app.go")
	}
}

func TestFilesystemBackend_Grep_NoGitignore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deepagents-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewFilesystemBackend(tmpDir, true)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	ctx := context.Background()

	// 不创建 .gitignore，直接创建文件
	backend.WriteFile(ctx, "/file1.txt", "search term here")
	backend.WriteFile(ctx, "/temp/file2.txt", "search term here")

	// 应该能搜索到所有文件
	matches, err := backend.Grep(ctx, "search term", "", "")
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches without gitignore, got %d", len(matches))
	}
}
