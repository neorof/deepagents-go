package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/backend"
)

func TestNewListFilesTool(t *testing.T) {
	// 创建测试后端
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	b.WriteFile(ctx, "/test1.txt", "content1")
	b.WriteFile(ctx, "/test2.txt", "content2")

	// 创建工具
	tool := NewListFilesTool(b)

	// 验证工具属性
	if tool.Name() != "ls" {
		t.Errorf("Expected name 'ls', got %q", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// 测试执行
	result, err := tool.Execute(ctx, map[string]any{
		"path": "/",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "test1.txt") {
		t.Error("Expected result to contain test1.txt")
	}

	if !strings.Contains(result, "test2.txt") {
		t.Error("Expected result to contain test2.txt")
	}
}

func TestNewListFilesTool_InvalidPath(t *testing.T) {
	b := backend.NewStateBackend()
	tool := NewListFilesTool(b)
	ctx := context.Background()

	// 测试无效参数
	_, err := tool.Execute(ctx, map[string]any{
		"path": 123, // 应该是字符串
	})

	if err == nil {
		t.Error("Expected error for invalid path type")
	}
}

func TestNewReadFileTool(t *testing.T) {
	// 创建测试后端
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	content := "line1\nline2\nline3\nline4\nline5"
	b.WriteFile(ctx, "/test.txt", content)

	// 创建工具
	tool := NewReadFileTool(b)

	// 验证工具属性
	if tool.Name() != "read_file" {
		t.Errorf("Expected name 'read_file', got %q", tool.Name())
	}

	// 测试读取整个文件
	result, err := tool.Execute(ctx, map[string]any{
		"path": "/test.txt",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != content {
		t.Errorf("Expected content %q, got %q", content, result)
	}
}

func TestNewReadFileTool_WithOffsetAndLimit(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	content := "line1\nline2\nline3\nline4\nline5"
	b.WriteFile(ctx, "/test.txt", content)

	tool := NewReadFileTool(b)

	// 测试带 offset 和 limit
	result, err := tool.Execute(ctx, map[string]any{
		"path":   "/test.txt",
		"offset": float64(1), // JSON 数字是 float64
		"limit":  float64(2),
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "line2") {
		t.Error("Expected result to contain line2")
	}
}

func TestNewReadFileTool_FileNotFound(t *testing.T) {
	b := backend.NewStateBackend()
	tool := NewReadFileTool(b)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]any{
		"path": "/nonexistent.txt",
	})

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestNewWriteFileTool(t *testing.T) {
	// 创建测试后端
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 创建工具
	tool := NewWriteFileTool(b)

	// 验证工具属性
	if tool.Name() != "write_file" {
		t.Errorf("Expected name 'write_file', got %q", tool.Name())
	}

	// 测试写入文件
	content := "Hello, World!"
	result, err := tool.Execute(ctx, map[string]any{
		"path":    "/test.txt",
		"content": content,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Successfully wrote") {
		t.Error("Expected success message")
	}

	// 验证文件已写入
	readContent, err := b.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}
}

func TestNewWriteFileTool_InvalidArgs(t *testing.T) {
	b := backend.NewStateBackend()
	tool := NewWriteFileTool(b)
	ctx := context.Background()

	// 测试缺少 content 参数
	_, err := tool.Execute(ctx, map[string]any{
		"path": "/test.txt",
	})

	if err == nil {
		t.Error("Expected error for missing content")
	}

	// 测试无效的 path 类型
	_, err = tool.Execute(ctx, map[string]any{
		"path":    123,
		"content": "test",
	})

	if err == nil {
		t.Error("Expected error for invalid path type")
	}
}

func TestNewEditFileTool(t *testing.T) {
	// 创建测试后端
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	content := "Hello World\nHello Go\nHello Test"
	b.WriteFile(ctx, "/test.txt", content)

	// 创建工具
	tool := NewEditFileTool(b)

	// 验证工具属性
	if tool.Name() != "edit_file" {
		t.Errorf("Expected name 'edit_file', got %q", tool.Name())
	}

	// 测试替换（只替换第一个）
	result, err := tool.Execute(ctx, map[string]any{
		"path":       "/test.txt",
		"old_string": "Hello",
		"new_string": "Hi",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Successfully replaced") {
		t.Error("Expected success message")
	}

	// 验证文件已修改
	newContent, _ := b.ReadFile(ctx, "/test.txt", 0, 0)
	if !strings.Contains(newContent, "Hi") {
		t.Error("Expected content to contain 'Hi'")
	}
}

func TestNewEditFileTool_ReplaceAll(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	content := "Hello World\nHello Go\nHello Test"
	b.WriteFile(ctx, "/test.txt", content)

	tool := NewEditFileTool(b)

	// 测试替换所有
	result, err := tool.Execute(ctx, map[string]any{
		"path":        "/test.txt",
		"old_string":  "Hello",
		"new_string":  "Hi",
		"replace_all": true,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "3 occurrence") {
		t.Errorf("Expected 3 replacements, got %q", result)
	}
}

func TestNewEditFileTool_InvalidArgs(t *testing.T) {
	b := backend.NewStateBackend()
	tool := NewEditFileTool(b)
	ctx := context.Background()

	// 测试缺少参数
	_, err := tool.Execute(ctx, map[string]any{
		"path":       "/test.txt",
		"old_string": "old",
	})

	if err == nil {
		t.Error("Expected error for missing new_string")
	}
}

func TestNewGrepTool(t *testing.T) {
	// 创建测试后端
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	b.WriteFile(ctx, "/test1.txt", "Hello World\nGoodbye World")
	b.WriteFile(ctx, "/test2.txt", "Hello Go\nGoodbye Go")

	// 创建工具
	tool := NewGrepTool(b)

	// 验证工具属性
	if tool.Name() != "grep" {
		t.Errorf("Expected name 'grep', got %q", tool.Name())
	}

	// 测试搜索
	result, err := tool.Execute(ctx, map[string]any{
		"pattern": "Hello",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Hello") {
		t.Error("Expected result to contain 'Hello'")
	}

	if !strings.Contains(result, "Found") {
		t.Error("Expected result to contain match count")
	}
}

func TestNewGrepTool_NoMatches(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	b.WriteFile(ctx, "/test.txt", "Hello World")

	tool := NewGrepTool(b)

	result, err := tool.Execute(ctx, map[string]any{
		"pattern": "NonExistent",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "No matches found") {
		t.Error("Expected 'No matches found' message")
	}
}

func TestNewGrepTool_WithPathAndGlob(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	b.WriteFile(ctx, "/dir/test.go", "package main")
	b.WriteFile(ctx, "/dir/test.txt", "Hello World")

	tool := NewGrepTool(b)

	// 测试带 path 和 glob
	result, err := tool.Execute(ctx, map[string]any{
		"pattern": "package",
		"path":    "/dir",
		"glob":    "*.go",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "package") {
		t.Error("Expected result to contain 'package'")
	}
}

func TestNewGrepTool_InvalidPattern(t *testing.T) {
	b := backend.NewStateBackend()
	tool := NewGrepTool(b)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]any{
		"pattern": 123, // 应该是字符串
	})

	if err == nil {
		t.Error("Expected error for invalid pattern type")
	}
}

func TestNewGlobTool(t *testing.T) {
	// 创建测试后端
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	b.WriteFile(ctx, "/test1.go", "package main")
	b.WriteFile(ctx, "/test2.go", "package test")
	b.WriteFile(ctx, "/test.txt", "text file")

	// 创建工具
	tool := NewGlobTool(b)

	// 验证工具属性
	if tool.Name() != "glob" {
		t.Errorf("Expected name 'glob', got %q", tool.Name())
	}

	// 测试文件匹配
	// 注意：StateBackend 的 Glob 是简化实现，返回所有文件
	result, err := tool.Execute(ctx, map[string]any{
		"pattern": "*.go",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "test1.go") {
		t.Error("Expected result to contain test1.go")
	}

	if !strings.Contains(result, "test2.go") {
		t.Error("Expected result to contain test2.go")
	}

	// StateBackend 返回所有文件，所以这个检查不适用
	// 真正的 glob 匹配由 FilesystemBackend 实现
}

func TestNewGlobTool_NoMatches(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 不写入任何文件，测试空后端
	tool := NewGlobTool(b)

	result, err := tool.Execute(ctx, map[string]any{
		"pattern": "*.go",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "No files found") {
		t.Error("Expected 'No files found' message")
	}
}

func TestNewGlobTool_WithPath(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	b.WriteFile(ctx, "/dir/test.go", "package main")
	b.WriteFile(ctx, "/other/test.go", "package other")

	tool := NewGlobTool(b)

	// 测试带 path
	result, err := tool.Execute(ctx, map[string]any{
		"pattern": "*.go",
		"path":    "/dir",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Found") {
		t.Error("Expected result to contain match count")
	}
}

func TestNewGlobTool_InvalidPattern(t *testing.T) {
	b := backend.NewStateBackend()
	tool := NewGlobTool(b)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]any{
		"pattern": 123, // 应该是字符串
	})

	if err == nil {
		t.Error("Expected error for invalid pattern type")
	}
}
