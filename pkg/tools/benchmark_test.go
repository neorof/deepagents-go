package tools

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/backend"
)

// BenchmarkListFilesTool 测试 ls 工具性能
func BenchmarkListFilesTool(b *testing.B) {
	backend := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 100; i++ {
		backend.WriteFile(ctx, "/test"+string(rune(i))+".txt", "content")
	}

	tool := NewListFilesTool(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]any{
			"path": "/",
		})
	}
}

// BenchmarkReadFileTool 测试 read_file 工具性能
func BenchmarkReadFileTool(b *testing.B) {
	backend := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	content := ""
	for i := 0; i < 1000; i++ {
		content += "This is line " + string(rune(i)) + "\n"
	}
	backend.WriteFile(ctx, "/test.txt", content)

	tool := NewReadFileTool(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]any{
			"path": "/test.txt",
		})
	}
}

// BenchmarkWriteFileTool 测试 write_file 工具性能
func BenchmarkWriteFileTool(b *testing.B) {
	backend := backend.NewStateBackend()
	ctx := context.Background()
	tool := NewWriteFileTool(backend)

	content := "Test content for benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]any{
			"path":    "/test.txt",
			"content": content,
		})
	}
}

// BenchmarkEditFileTool 测试 edit_file 工具性能
func BenchmarkEditFileTool(b *testing.B) {
	backend := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	content := "Hello World\nHello Go\nHello Test"
	backend.WriteFile(ctx, "/test.txt", content)

	tool := NewEditFileTool(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]any{
			"path":       "/test.txt",
			"old_string": "Hello",
			"new_string": "Hi",
		})
		// 恢复内容
		backend.WriteFile(ctx, "/test.txt", content)
	}
}

// BenchmarkGrepTool 测试 grep 工具性能
func BenchmarkGrepTool(b *testing.B) {
	backend := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 50; i++ {
		content := ""
		for j := 0; j < 100; j++ {
			content += "Line " + string(rune(j)) + " with some text\n"
		}
		backend.WriteFile(ctx, "/file"+string(rune(i))+".txt", content)
	}

	tool := NewGrepTool(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]any{
			"pattern": "text",
		})
	}
}

// BenchmarkGlobTool 测试 glob 工具性能
func BenchmarkGlobTool(b *testing.B) {
	backend := backend.NewStateBackend()
	ctx := context.Background()

	// 准备测试数据
	for i := 0; i < 100; i++ {
		backend.WriteFile(ctx, "/test"+string(rune(i))+".txt", "content")
		backend.WriteFile(ctx, "/test"+string(rune(i))+".go", "code")
	}

	tool := NewGlobTool(backend)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]any{
			"pattern": "*.txt",
		})
	}
}

// BenchmarkRegistry_RegisterAndGet 测试注册表性能
func BenchmarkRegistry_RegisterAndGet(b *testing.B) {
	registry := NewRegistry()
	backend := backend.NewStateBackend()

	// 注册一些工具
	registry.Register(NewListFilesTool(backend))
	registry.Register(NewReadFileTool(backend))
	registry.Register(NewWriteFileTool(backend))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Get("ls")
		registry.Get("read_file")
		registry.Get("write_file")
	}
}

// BenchmarkRegistry_List 测试列出所有工具的性能
func BenchmarkRegistry_List(b *testing.B) {
	registry := NewRegistry()
	backend := backend.NewStateBackend()

	// 注册多个工具
	registry.Register(NewListFilesTool(backend))
	registry.Register(NewReadFileTool(backend))
	registry.Register(NewWriteFileTool(backend))
	registry.Register(NewEditFileTool(backend))
	registry.Register(NewGrepTool(backend))
	registry.Register(NewGlobTool(backend))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.List()
	}
}
