package backend

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewSandboxBackend(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	if backend == nil {
		t.Fatal("Expected backend to be non-nil")
	}

	if backend.config.RootDir != tmpDir {
		t.Errorf("Expected root dir %s, got %s", tmpDir, backend.config.RootDir)
	}
}

func TestNewSandboxBackend_NilConfig(t *testing.T) {
	_, err := NewSandboxBackend(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestSandboxBackend_WriteAndRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 写入文件
	content := "Hello, Sandbox!"
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

func TestSandboxBackend_ReadOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.ReadOnly = true
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 尝试写入应该失败
	_, err = backend.WriteFile(ctx, "/test.txt", "content")
	if err == nil {
		t.Error("Expected error for write in read-only mode")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Errorf("Expected read-only error, got: %v", err)
	}

	// 尝试编辑应该失败
	_, err = backend.EditFile(ctx, "/test.txt", "old", "new", false)
	if err == nil {
		t.Error("Expected error for edit in read-only mode")
	}

	// 尝试执行应该失败
	_, err = backend.Execute(ctx, "ls", 1000)
	if err == nil {
		t.Error("Expected error for execute in read-only mode")
	}
}

func TestSandboxBackend_MaxFileSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.MaxFileSize = 100 // 100 bytes
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 写入小文件应该成功
	smallContent := "Small content"
	_, err = backend.WriteFile(ctx, "/small.txt", smallContent)
	if err != nil {
		t.Errorf("Expected success for small file, got error: %v", err)
	}

	// 写入大文件应该失败
	largeContent := strings.Repeat("x", 200)
	_, err = backend.WriteFile(ctx, "/large.txt", largeContent)
	if err == nil {
		t.Error("Expected error for large file")
	}
	if !strings.Contains(err.Error(), "exceeds limit") {
		t.Errorf("Expected size limit error, got: %v", err)
	}
}

func TestSandboxBackend_OperationLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.MaxOperations = 3
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 前3个操作应该成功
	_, err = backend.WriteFile(ctx, "/test1.txt", "content1")
	if err != nil {
		t.Errorf("Operation 1 should succeed, got error: %v", err)
	}

	_, err = backend.WriteFile(ctx, "/test2.txt", "content2")
	if err != nil {
		t.Errorf("Operation 2 should succeed, got error: %v", err)
	}

	_, err = backend.WriteFile(ctx, "/test3.txt", "content3")
	if err != nil {
		t.Errorf("Operation 3 should succeed, got error: %v", err)
	}

	// 第4个操作应该失败
	_, err = backend.WriteFile(ctx, "/test4.txt", "content4")
	if err == nil {
		t.Error("Expected error for operation limit exceeded")
	}
	if !strings.Contains(err.Error(), "limit exceeded") {
		t.Errorf("Expected operation limit error, got: %v", err)
	}

	// 重置计数器
	backend.ResetOperationCount()

	// 重置后应该可以继续操作
	_, err = backend.WriteFile(ctx, "/test5.txt", "content5")
	if err != nil {
		t.Errorf("Operation after reset should succeed, got error: %v", err)
	}
}

func TestSandboxBackend_BlockedPaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.BlockedPaths = []string{"/secret", "/private"}
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 访问被阻止的路径应该失败
	_, err = backend.WriteFile(ctx, "/secret/data.txt", "content")
	if err == nil {
		t.Error("Expected error for blocked path")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("Expected blocked path error, got: %v", err)
	}

	// 访问正常路径应该成功
	_, err = backend.WriteFile(ctx, "/public/data.txt", "content")
	if err != nil {
		t.Errorf("Expected success for allowed path, got error: %v", err)
	}
}

func TestSandboxBackend_AllowedPaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.AllowedPaths = []string{"/public", "/tmp"}
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 访问允许的路径应该成功
	_, err = backend.WriteFile(ctx, "/public/data.txt", "content")
	if err != nil {
		t.Errorf("Expected success for allowed path, got error: %v", err)
	}

	// 访问不在白名单的路径应该失败
	_, err = backend.WriteFile(ctx, "/secret/data.txt", "content")
	if err == nil {
		t.Error("Expected error for non-allowed path")
	}
	if !strings.Contains(err.Error(), "not in allowed list") {
		t.Errorf("Expected not allowed error, got: %v", err)
	}
}

func TestSandboxBackend_EditFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 创建测试文件
	content := "Hello World"
	_, err = backend.WriteFile(ctx, "/test.txt", content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 编辑文件
	result, err := backend.EditFile(ctx, "/test.txt", "World", "Sandbox", false)
	if err != nil {
		t.Fatalf("EditFile failed: %v", err)
	}

	if result.Replacements != 1 {
		t.Errorf("Expected 1 replacement, got %d", result.Replacements)
	}

	// 读取并验证编辑后的内容
	readContent, err := backend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	expected := "Hello Sandbox"
	if readContent != expected {
		t.Errorf("Expected %q, got %q", expected, readContent)
	}
}

func TestSandboxBackend_EditFile_SizeLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.MaxFileSize = 50 // 50 bytes
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 创建小文件
	content := "Hello"
	_, err = backend.WriteFile(ctx, "/test.txt", content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 编辑后大小超过限制应该失败
	largeReplacement := strings.Repeat("x", 100)
	_, err = backend.EditFile(ctx, "/test.txt", "Hello", largeReplacement, false)
	if err == nil {
		t.Error("Expected error for edited file size exceeds limit")
	}

	// 验证文件内容未被修改（回滚成功）
	readContent, err := backend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if readContent != content {
		t.Error("File content should be unchanged after failed edit")
	}
}

func TestSandboxBackend_ListFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 创建测试文件
	_, err = backend.WriteFile(ctx, "/test1.txt", "content1")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	_, err = backend.WriteFile(ctx, "/test2.txt", "content2")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 列出文件
	files, err := backend.ListFiles(ctx, "/")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestSandboxBackend_Grep(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 创建测试文件
	content := "Hello World\nGoodbye World"
	_, err = backend.WriteFile(ctx, "/test.txt", content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 搜索
	matches, err := backend.Grep(ctx, "World", "", "")
	if err != nil {
		t.Fatalf("Grep failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
}

func TestSandboxBackend_Glob(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 创建测试文件
	_, err = backend.WriteFile(ctx, "/test.txt", "content")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	_, err = backend.WriteFile(ctx, "/test.md", "content")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// 搜索 txt 文件
	files, err := backend.Glob(ctx, "*.txt", "")
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

func TestSandboxBackend_Execute(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 执行允许的命令
	result, err := backend.Execute(ctx, "echo hello", 1000)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("Expected output to contain 'hello', got: %s", result.Stdout)
	}
}

func TestSandboxBackend_Execute_NotAllowed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 执行不允许的命令应该失败
	_, err = backend.Execute(ctx, "rm -rf /", 1000)
	if err == nil {
		t.Error("Expected error for non-allowed command")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("Expected not allowed error, got: %v", err)
	}
}

func TestSandboxBackend_Execute_EmptyCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 空命令应该失败
	_, err = backend.Execute(ctx, "", 1000)
	if err == nil {
		t.Error("Expected error for empty command")
	}
}

func TestSandboxBackend_AuditLog(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.EnableAuditLog = true
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 执行一些操作
	_, _ = backend.WriteFile(ctx, "/test.txt", "content")
	_, _ = backend.ReadFile(ctx, "/test.txt", 0, 0)

	// 检查审计日志
	log := backend.GetAuditLog()
	if len(log) != 2 {
		t.Errorf("Expected 2 audit entries, got %d", len(log))
	}

	if log[0].Operation != "WriteFile" {
		t.Errorf("Expected first operation to be WriteFile, got %s", log[0].Operation)
	}

	if log[1].Operation != "ReadFile" {
		t.Errorf("Expected second operation to be ReadFile, got %s", log[1].Operation)
	}
}

func TestSandboxBackend_AuditLog_Disabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	config.EnableAuditLog = false
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 执行操作
	_, _ = backend.WriteFile(ctx, "/test.txt", "content")

	// 审计日志应该为空
	log := backend.GetAuditLog()
	if len(log) != 0 {
		t.Errorf("Expected 0 audit entries when disabled, got %d", len(log))
	}
}

func TestSandboxBackend_GetOperationCount(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := DefaultSandboxConfig(tmpDir)
	backend, err := NewSandboxBackend(config)
	if err != nil {
		t.Fatalf("Failed to create sandbox backend: %v", err)
	}

	ctx := context.Background()

	// 初始计数应该为0
	if count := backend.GetOperationCount(); count != 0 {
		t.Errorf("Expected initial count 0, got %d", count)
	}

	// 执行操作
	_, _ = backend.WriteFile(ctx, "/test.txt", "content")
	_, _ = backend.ReadFile(ctx, "/test.txt", 0, 0)

	// 计数应该增加
	if count := backend.GetOperationCount(); count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestDefaultSandboxConfig(t *testing.T) {
	tmpDir := "/test/dir"
	config := DefaultSandboxConfig(tmpDir)

	if config.RootDir != tmpDir {
		t.Errorf("Expected root dir %s, got %s", tmpDir, config.RootDir)
	}

	if config.ReadOnly {
		t.Error("Expected read-only to be false by default")
	}

	if config.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected max file size 10MB, got %d", config.MaxFileSize)
	}

	if config.MaxOperations != 1000 {
		t.Errorf("Expected max operations 1000, got %d", config.MaxOperations)
	}

	if config.OperationTimeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.OperationTimeout)
	}

	if !config.EnableAuditLog {
		t.Error("Expected audit log to be enabled by default")
	}
}
