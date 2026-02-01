package backend

import (
	"context"
	"testing"
)

// BenchmarkStateBackend_WriteFile 测试写文件性能
func BenchmarkStateBackend_WriteFile(b *testing.B) {
	backend := NewStateBackend()
	ctx := context.Background()
	content := "Test content for benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.WriteFile(ctx, "/test.txt", content)
	}
}

// BenchmarkStateBackend_ReadFile 测试读文件性能
func BenchmarkStateBackend_ReadFile(b *testing.B) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 准备数据
	content := ""
	for i := 0; i < 1000; i++ {
		content += "Line " + string(rune(i)) + "\n"
	}
	backend.WriteFile(ctx, "/test.txt", content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.ReadFile(ctx, "/test.txt", 0, 0)
	}
}

// BenchmarkStateBackend_EditFile 测试编辑文件性能
func BenchmarkStateBackend_EditFile(b *testing.B) {
	backend := NewStateBackend()
	ctx := context.Background()

	content := "Hello World\nHello Go\nHello Test"
	backend.WriteFile(ctx, "/test.txt", content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.EditFile(ctx, "/test.txt", "Hello", "Hi", false)
		// 恢复内容
		backend.WriteFile(ctx, "/test.txt", content)
	}
}

// BenchmarkStateBackend_Grep 测试搜索性能
func BenchmarkStateBackend_Grep(b *testing.B) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 准备大量数据
	for i := 0; i < 50; i++ {
		content := ""
		for j := 0; j < 100; j++ {
			content += "Line " + string(rune(j)) + " with some text\n"
		}
		backend.WriteFile(ctx, "/file"+string(rune(i))+".txt", content)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Grep(ctx, "text", "", "")
	}
}

// BenchmarkStateBackend_ListFiles 测试列出文件性能
func BenchmarkStateBackend_ListFiles(b *testing.B) {
	backend := NewStateBackend()
	ctx := context.Background()

	// 准备数据
	for i := 0; i < 100; i++ {
		backend.WriteFile(ctx, "/test"+string(rune(i))+".txt", "content")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.ListFiles(ctx, "/")
	}
}

// BenchmarkCompositeBackend_WriteAndRead 测试组合后端性能
func BenchmarkCompositeBackend_WriteAndRead(b *testing.B) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)
	state1 := NewStateBackend()
	state2 := NewStateBackend()

	composite.AddRoute("/state1", state1)
	composite.AddRoute("/state2", state2)

	ctx := context.Background()
	content := "Test content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		composite.WriteFile(ctx, "/state1/test.txt", content)
		composite.ReadFile(ctx, "/state1/test.txt", 0, 0)
	}
}

// BenchmarkFilesystemBackend_WriteFile 测试文件系统后端写入性能
func BenchmarkFilesystemBackend_WriteFile(b *testing.B) {
	tmpDir := b.TempDir()
	backend, err := NewFilesystemBackend(tmpDir, false)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	content := "Test content for benchmark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.WriteFile(ctx, "/test.txt", content)
	}
}

// BenchmarkFilesystemBackend_ReadFile 测试文件系统后端读取性能
func BenchmarkFilesystemBackend_ReadFile(b *testing.B) {
	tmpDir := b.TempDir()
	backend, err := NewFilesystemBackend(tmpDir, false)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	content := ""
	for i := 0; i < 1000; i++ {
		content += "Line " + string(rune(i)) + "\n"
	}
	backend.WriteFile(ctx, "/test.txt", content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.ReadFile(ctx, "/test.txt", 0, 0)
	}
}
