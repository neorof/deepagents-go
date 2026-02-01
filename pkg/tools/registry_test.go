package tools

import (
	"context"
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()

	// 创建测试工具
	tool := NewBaseTool(
		"test_tool",
		"A test tool",
		map[string]any{"type": "object"},
		func(ctx context.Context, args map[string]any) (string, error) {
			return "OK", nil
		},
	)

	// 注册工具
	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 获取工具
	retrieved, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("Tool not found")
	}

	if retrieved.Name() != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got %q", retrieved.Name())
	}
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	registry := NewRegistry()

	tool1 := NewBaseTool("test_tool", "Tool 1", map[string]any{}, nil)
	tool2 := NewBaseTool("test_tool", "Tool 2", map[string]any{}, nil)

	// 第一次注册应该成功
	err := registry.Register(tool1)
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	// 第二次注册应该失败
	err = registry.Register(tool2)
	if err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// 注册多个工具
	tool1 := NewBaseTool("tool1", "Tool 1", map[string]any{}, nil)
	tool2 := NewBaseTool("tool2", "Tool 2", map[string]any{}, nil)
	tool3 := NewBaseTool("tool3", "Tool 3", map[string]any{}, nil)

	registry.Register(tool1)
	registry.Register(tool2)
	registry.Register(tool3)

	// 列出所有工具
	tools := registry.List()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}
}

func TestRegistry_Remove(t *testing.T) {
	registry := NewRegistry()

	tool := NewBaseTool("test_tool", "Test", map[string]any{}, nil)
	registry.Register(tool)

	// 验证工具存在
	_, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("Tool should exist before removal")
	}

	// 移除工具
	registry.Remove("test_tool")

	// 验证工具已被移除
	_, ok = registry.Get("test_tool")
	if ok {
		t.Error("Tool should not exist after removal")
	}
}

func TestRegistry_GetNonexistent(t *testing.T) {
	registry := NewRegistry()

	_, ok := registry.Get("nonexistent_tool")
	if ok {
		t.Error("Expected false for nonexistent tool, got true")
	}
}
