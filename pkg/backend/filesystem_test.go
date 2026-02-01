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
