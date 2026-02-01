package middleware

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func TestNewMemoryMiddleware(t *testing.T) {
	m := NewMemoryMiddleware()

	if m == nil {
		t.Fatal("Expected non-nil middleware")
	}

	if m.Name() != "memory" {
		t.Errorf("Expected name 'memory', got %q", m.Name())
	}

	if len(m.memories) != 0 {
		t.Error("Expected empty memories initially")
	}
}

func TestMemoryMiddleware_AddMemory(t *testing.T) {
	m := NewMemoryMiddleware()

	m.AddMemory("test", "Test content", "system")

	memories := m.GetMemories()
	if len(memories) != 1 {
		t.Fatalf("Expected 1 memory, got %d", len(memories))
	}

	if memories[0].Source != "test" {
		t.Errorf("Expected source 'test', got %q", memories[0].Source)
	}

	if memories[0].Content != "Test content" {
		t.Errorf("Expected content 'Test content', got %q", memories[0].Content)
	}

	if memories[0].Type != "system" {
		t.Errorf("Expected type 'system', got %q", memories[0].Type)
	}
}

func TestMemoryMiddleware_GetMemoriesByType(t *testing.T) {
	m := NewMemoryMiddleware()

	m.AddMemory("test1", "System memory", "system")
	m.AddMemory("test2", "Context memory", "context")
	m.AddMemory("test3", "Another system memory", "system")

	systemMemories := m.GetMemoriesByType("system")
	if len(systemMemories) != 2 {
		t.Errorf("Expected 2 system memories, got %d", len(systemMemories))
	}

	contextMemories := m.GetMemoriesByType("context")
	if len(contextMemories) != 1 {
		t.Errorf("Expected 1 context memory, got %d", len(contextMemories))
	}
}

func TestMemoryMiddleware_ClearMemories(t *testing.T) {
	m := NewMemoryMiddleware()

	m.AddMemory("test1", "Content 1", "system")
	m.AddMemory("test2", "Content 2", "context")

	if len(m.GetMemories()) != 2 {
		t.Error("Expected 2 memories before clear")
	}

	m.ClearMemories()

	if len(m.GetMemories()) != 0 {
		t.Error("Expected 0 memories after clear")
	}
}

func TestMemoryMiddleware_RemoveMemory(t *testing.T) {
	m := NewMemoryMiddleware()

	m.AddMemory("test1", "Content 1", "system")
	m.AddMemory("test2", "Content 2", "context")

	// 删除存在的记忆
	removed := m.RemoveMemory("test1")
	if !removed {
		t.Error("Expected RemoveMemory to return true")
	}

	if len(m.GetMemories()) != 1 {
		t.Errorf("Expected 1 memory after removal, got %d", len(m.GetMemories()))
	}

	// 删除不存在的记忆
	removed = m.RemoveMemory("nonexistent")
	if removed {
		t.Error("Expected RemoveMemory to return false for nonexistent memory")
	}
}

func TestMemoryMiddleware_LoadFromFile(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	content := "# Test Memory\n\nThis is a test memory."
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	m := NewMemoryMiddleware()

	// 加载文件
	if err := m.LoadFromFile(testFile); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	memories := m.GetMemories()
	if len(memories) != 1 {
		t.Fatalf("Expected 1 memory, got %d", len(memories))
	}

	if memories[0].Source != testFile {
		t.Errorf("Expected source %q, got %q", testFile, memories[0].Source)
	}

	if memories[0].Content != content {
		t.Errorf("Expected content %q, got %q", content, memories[0].Content)
	}

	if memories[0].Type != "system" {
		t.Errorf("Expected type 'system', got %q", memories[0].Type)
	}
}

func TestMemoryMiddleware_LoadFromFile_NotExists(t *testing.T) {
	m := NewMemoryMiddleware()

	err := m.LoadFromFile("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestMemoryMiddleware_LoadFromDirectory(t *testing.T) {
	// 创建临时目录和文件
	tmpDir := t.TempDir()

	files := map[string]string{
		"memory1.md":  "# Memory 1",
		"memory2.txt": "Memory 2 content",
		"ignore.json": "Should be ignored",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	m := NewMemoryMiddleware()

	// 加载目录
	if err := m.LoadFromDirectory(tmpDir); err != nil {
		t.Fatalf("LoadFromDirectory failed: %v", err)
	}

	memories := m.GetMemories()
	if len(memories) != 2 {
		t.Errorf("Expected 2 memories, got %d", len(memories))
	}
}

func TestMemoryMiddleware_LoadFromDirectory_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	m := NewMemoryMiddleware()

	err := m.LoadFromDirectory(tmpDir)
	if err == nil {
		t.Error("Expected error for empty directory")
	}
}

func TestMemoryMiddleware_BeforeModel(t *testing.T) {
	m := NewMemoryMiddleware()

	m.AddMemory("test1", "System memory content", "system")
	m.AddMemory("test2", "Context memory content", "context")

	req := &llm.ModelRequest{
		SystemPrompt: "Original prompt",
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证系统提示已被修改
	if !strings.Contains(req.SystemPrompt, "Original prompt") {
		t.Error("Expected original prompt to be preserved")
	}

	if !strings.Contains(req.SystemPrompt, "System memory content") {
		t.Error("Expected system memory to be injected")
	}

	if !strings.Contains(req.SystemPrompt, "Context memory content") {
		t.Error("Expected context memory to be injected")
	}

	if !strings.Contains(req.SystemPrompt, "Agent 记忆") {
		t.Error("Expected memory header to be present")
	}
}

func TestMemoryMiddleware_BeforeModel_EmptyMemories(t *testing.T) {
	m := NewMemoryMiddleware()

	req := &llm.ModelRequest{
		SystemPrompt: "Original prompt",
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证系统提示未被修改
	if req.SystemPrompt != "Original prompt" {
		t.Error("Expected system prompt to remain unchanged")
	}
}

func TestMemoryMiddleware_BeforeModel_EmptySystemPrompt(t *testing.T) {
	m := NewMemoryMiddleware()

	m.AddMemory("test", "Memory content", "system")

	req := &llm.ModelRequest{
		SystemPrompt: "",
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证系统提示已被设置
	if req.SystemPrompt == "" {
		t.Error("Expected system prompt to be set")
	}

	if !strings.Contains(req.SystemPrompt, "Memory content") {
		t.Error("Expected memory content to be present")
	}
}

func TestMemoryMiddleware_BeforeAgent(t *testing.T) {
	m := NewMemoryMiddleware()

	// 创建临时 AGENTS.md 文件
	tmpDir := t.TempDir()
	agentsFile := filepath.Join(tmpDir, "AGENTS.md")
	content := "# Agent Memory\n\nTest content"

	if err := os.WriteFile(agentsFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create AGENTS.md: %v", err)
	}

	// 切换到临时目录
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	state := agent.NewState()
	ctx := context.Background()

	if err := m.BeforeAgent(ctx, state); err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	// 验证记忆已加载
	memories := m.GetMemories()
	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}
}

func TestMemoryMiddleware_BeforeAgent_NoDefaultFile(t *testing.T) {
	m := NewMemoryMiddleware()

	// 切换到临时目录（没有 AGENTS.md）
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	state := agent.NewState()
	ctx := context.Background()

	// 应该不会报错，只是不加载记忆
	if err := m.BeforeAgent(ctx, state); err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	// 验证没有加载记忆
	memories := m.GetMemories()
	if len(memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(memories))
	}
}
