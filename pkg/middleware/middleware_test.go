package middleware

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func TestFilesystemMiddleware_Name(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewFilesystemMiddleware(backend, toolRegistry)

	if middleware.Name() != "filesystem" {
		t.Errorf("Expected name 'filesystem', got %q", middleware.Name())
	}
}

func TestFilesystemMiddleware_RegistersTools(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	NewFilesystemMiddleware(backend, toolRegistry)

	// 验证工具已注册
	expectedTools := []string{"ls", "read_file", "write_file", "edit_file", "grep", "glob"}

	for _, toolName := range expectedTools {
		_, ok := toolRegistry.Get(toolName)
		if !ok {
			t.Errorf("Expected tool %q to be registered", toolName)
		}
	}
}

func TestFilesystemMiddleware_AfterTool_SmallResult(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewFilesystemMiddleware(backend, toolRegistry)

	ctx := context.Background()
	state := agent.NewState()

	// 小结果不应该被驱逐
	result := &llm.ToolResult{
		ToolCallID: "call_123",
		Content:    "Small result",
		IsError:    false,
	}

	err := middleware.AfterTool(ctx, result, state)
	if err != nil {
		t.Fatalf("AfterTool failed: %v", err)
	}

	// 内容应该保持不变
	if result.Content != "Small result" {
		t.Errorf("Expected content to remain unchanged, got %q", result.Content)
	}
}

func TestFilesystemMiddleware_AfterTool_LargeResult(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewFilesystemMiddleware(backend, toolRegistry)

	ctx := context.Background()
	state := agent.NewState()

	// 创建一个大结果（超过 80,000 字符）
	largeContent := ""
	for i := 0; i < 10000; i++ {
		largeContent += "This is line " + string(rune(i)) + " of the large result.\n"
	}

	result := &llm.ToolResult{
		ToolCallID: "call_123",
		Content:    largeContent,
		IsError:    false,
	}

	err := middleware.AfterTool(ctx, result, state)
	if err != nil {
		t.Fatalf("AfterTool failed: %v", err)
	}

	// 内容应该被替换为预览
	if len(result.Content) >= len(largeContent) {
		t.Error("Expected content to be truncated")
	}

	// 应该包含文件路径
	if len(result.Content) < 20 {
		t.Error("Expected result to contain file path and preview")
	}
}

func TestBaseMiddleware_DefaultImplementations(t *testing.T) {
	middleware := NewBaseMiddleware("test")

	ctx := context.Background()
	state := agent.NewState()

	// 所有默认实现应该返回 nil
	if err := middleware.BeforeAgent(ctx, state); err != nil {
		t.Errorf("BeforeAgent should return nil, got %v", err)
	}

	req := &llm.ModelRequest{}
	if err := middleware.BeforeModel(ctx, req); err != nil {
		t.Errorf("BeforeModel should return nil, got %v", err)
	}

	resp := &llm.ModelResponse{}
	if err := middleware.AfterModel(ctx, resp, state); err != nil {
		t.Errorf("AfterModel should return nil, got %v", err)
	}

	toolCall := &llm.ToolCall{}
	if err := middleware.BeforeTool(ctx, toolCall, state); err != nil {
		t.Errorf("BeforeTool should return nil, got %v", err)
	}

	toolResult := &llm.ToolResult{}
	if err := middleware.AfterTool(ctx, toolResult, state); err != nil {
		t.Errorf("AfterTool should return nil, got %v", err)
	}
}
