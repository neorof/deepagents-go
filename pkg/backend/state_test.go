package backend

import (
	"context"
	"testing"
)

func TestStateBackend_WriteAndRead(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 写入文件
	content := "Hello, World!"
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
}

func TestStateBackend_EditFile(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 写入初始内容
	content := "Hello, World!\nHello, Go!"
	backend.WriteFile(ctx, "/test.txt", content)

	// 编辑文件（替换第一个 Hello）
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

func TestStateBackend_EditFileReplaceAll(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 写入初始内容
	content := "Hello, World!\nHello, Go!"
	backend.WriteFile(ctx, "/test.txt", content)

	// 编辑文件（替换所有 Hello）
	result, err := backend.EditFile(ctx, "/test.txt", "Hello", "Hi", true)
	if err != nil {
		t.Fatalf("EditFile failed: %v", err)
	}

	if result.Replacements != 2 {
		t.Errorf("Expected 2 replacements, got %d", result.Replacements)
	}

	// 验证内容
	newContent, _ := backend.ReadFile(ctx, "/test.txt", 0, 0)
	expected := "Hi, World!\nHi, Go!"
	if newContent != expected {
		t.Errorf("Expected content %q, got %q", expected, newContent)
	}
}

func TestStateBackend_Grep(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 写入多个文件
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

func TestStateBackend_ListFiles(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 写入多个文件
	backend.WriteFile(ctx, "/file1.txt", "content1")
	backend.WriteFile(ctx, "/file2.txt", "content2")
	backend.WriteFile(ctx, "/file3.txt", "content3")

	// 列出文件
	files, err := backend.ListFiles(ctx, "/")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestStateBackend_ReadFileNotFound(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 读取不存在的文件
	_, err := backend.ReadFile(ctx, "/nonexistent.txt", 0, 0)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestStateBackend_ReadFileWithOffset(t *testing.T) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 写入多行内容
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	backend.WriteFile(ctx, "/test.txt", content)

	// 读取从第 2 行开始的 2 行
	result, err := backend.ReadFile(ctx, "/test.txt", 1, 2)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	expected := "Line 2\nLine 3"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
