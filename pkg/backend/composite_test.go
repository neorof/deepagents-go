package backend

import (
	"context"
	"os"
	"testing"
)

func TestCompositeBackend_AddRoute(t *testing.T) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	backend1 := NewStateBackend()
	backend2 := NewStateBackend()

	composite.AddRoute("/data", backend1)
	composite.AddRoute("/config", backend2)

	if len(composite.routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(composite.routes))
	}
}

func TestCompositeBackend_GetBackendAndKey(t *testing.T) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	backend1 := NewStateBackend()
	composite.AddRoute("/data", backend1)

	// 测试匹配路由
	backend, key := composite.getBackendAndKey("/data/file.txt")
	if backend != backend1 {
		t.Error("Expected backend1 for /data path")
	}
	if key != "/file.txt" {
		t.Errorf("Expected key '/file.txt', got %q", key)
	}

	// 测试默认后端
	backend, key = composite.getBackendAndKey("/other/file.txt")
	if backend != defaultBackend {
		t.Error("Expected defaultBackend for /other path")
	}
	if key != "/other/file.txt" {
		t.Errorf("Expected key '/other/file.txt', got %q", key)
	}
}

func TestCompositeBackend_WriteAndRead(t *testing.T) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	dataBackend := NewStateBackend()
	composite.AddRoute("/data", dataBackend)

	ctx := context.Background()

	// 写入到路由后端
	content := "Hello from data backend"
	_, err := composite.WriteFile(ctx, "/data/test.txt", content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 读取
	readContent, err := composite.ReadFile(ctx, "/data/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if readContent != content {
		t.Errorf("Expected content %q, got %q", content, readContent)
	}

	// 验证文件在正确的后端
	_, err = dataBackend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Error("File should exist in dataBackend")
	}

	// 验证文件不在默认后端
	_, err = defaultBackend.ReadFile(ctx, "/data/test.txt", 0, 0)
	if err == nil {
		t.Error("File should not exist in defaultBackend")
	}
}

func TestCompositeBackend_LongestPrefixMatch(t *testing.T) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	backend1 := NewStateBackend()
	backend2 := NewStateBackend()

	composite.AddRoute("/data", backend1)
	composite.AddRoute("/data/special", backend2)

	ctx := context.Background()

	// 写入到更具体的路由
	composite.WriteFile(ctx, "/data/special/file.txt", "special")
	composite.WriteFile(ctx, "/data/normal/file.txt", "normal")

	// 验证路由到正确的后端
	_, err := backend2.ReadFile(ctx, "/file.txt", 0, 0)
	if err != nil {
		t.Error("File should be in backend2 (special)")
	}

	_, err = backend1.ReadFile(ctx, "/normal/file.txt", 0, 0)
	if err != nil {
		t.Error("File should be in backend1 (normal)")
	}
}

func TestCompositeBackend_GrepAcrossBackends(t *testing.T) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	backend1 := NewStateBackend()
	backend2 := NewStateBackend()

	composite.AddRoute("/data", backend1)
	composite.AddRoute("/config", backend2)

	ctx := context.Background()

	// 在不同后端写入文件
	composite.WriteFile(ctx, "/data/file1.txt", "Hello World")
	composite.WriteFile(ctx, "/config/file2.txt", "Hello Config")
	composite.WriteFile(ctx, "/default.txt", "Hello Default")

	// 搜索所有后端
	matches, err := composite.Grep(ctx, "Hello", "", "")
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}

	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

func TestCompositeBackend_GlobAcrossBackends(t *testing.T) {
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	backend1 := NewStateBackend()
	composite.AddRoute("/data", backend1)

	ctx := context.Background()

	// 在不同后端写入文件
	composite.WriteFile(ctx, "/data/file1.txt", "content1")
	composite.WriteFile(ctx, "/data/file2.go", "content2")
	composite.WriteFile(ctx, "/file3.txt", "content3")

	// 搜索所有后端
	files, err := composite.Glob(ctx, "*.txt", "")
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}

	// 应该找到 2 个 .txt 文件
	if len(files) < 2 {
		t.Errorf("Expected at least 2 .txt files, got %d", len(files))
	}
}

func TestCompositeBackend_WithFilesystemBackend(t *testing.T) {
	// 创建临时目录
	tmpDir1, err := os.MkdirTemp("", "composite-test-1-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	tmpDir2, err := os.MkdirTemp("", "composite-test-2-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	// 创建组合后端
	defaultBackend := NewStateBackend()
	composite := NewCompositeBackend(defaultBackend)

	fsBackend1, _ := NewFilesystemBackend(tmpDir1, true)
	fsBackend2, _ := NewFilesystemBackend(tmpDir2, true)

	composite.AddRoute("/disk1", fsBackend1)
	composite.AddRoute("/disk2", fsBackend2)

	ctx := context.Background()

	// 写入到不同的文件系统后端
	composite.WriteFile(ctx, "/disk1/file1.txt", "content1")
	composite.WriteFile(ctx, "/disk2/file2.txt", "content2")
	composite.WriteFile(ctx, "/memory.txt", "content3")

	// 验证文件在正确的位置
	content1, err := composite.ReadFile(ctx, "/disk1/file1.txt", 0, 0)
	if err != nil || content1 != "content1" {
		t.Error("File should exist in disk1")
	}

	content2, err := composite.ReadFile(ctx, "/disk2/file2.txt", 0, 0)
	if err != nil || content2 != "content2" {
		t.Error("File should exist in disk2")
	}

	content3, err := composite.ReadFile(ctx, "/memory.txt", 0, 0)
	if err != nil || content3 != "content3" {
		t.Error("File should exist in memory")
	}
}
