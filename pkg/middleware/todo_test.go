package middleware

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func TestTodoMiddleware_WriteTodos(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	_ = NewTodoMiddleware(backend, toolRegistry)

	ctx := context.Background()

	// 获取 write_todos 工具
	tool, ok := toolRegistry.Get("write_todos")
	if !ok {
		t.Fatal("write_todos tool not found")
	}

	// 执行工具
	args := map[string]any{
		"todos": []any{
			map[string]any{
				"id":          "1",
				"title":       "实现 Agent 核心",
				"status":      "completed",
				"description": "完成 Agent 执行器和状态管理",
			},
			map[string]any{
				"id":     "2",
				"title":  "实现文件系统中间件",
				"status": "in_progress",
			},
			map[string]any{
				"id":     "3",
				"title":  "实现 Todo 中间件",
				"status": "pending",
			},
		},
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != "Successfully updated 3 todo items" {
		t.Errorf("Unexpected result: %s", result)
	}

	// 验证文件已创建
	content, err := backend.ReadFile(ctx, "/todos.md", 0, 0)
	if err != nil {
		t.Fatalf("Failed to read todos.md: %v", err)
	}

	// 验证内容包含标题
	if len(content) == 0 {
		t.Error("todos.md should not be empty")
	}
}

func TestTodoMiddleware_BeforeModel(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(backend, toolRegistry)

	ctx := context.Background()

	// 先创建一个 Todo 列表
	backend.WriteFile(ctx, "/todos.md", "# Todo List\n\n## ✅ Task 1 [1]\n\nCompleted task\n\n")

	// 创建请求
	req := &llm.ModelRequest{
		SystemPrompt: "You are a helpful assistant.",
		Messages:     []llm.Message{},
	}

	// 执行 BeforeModel 钩子
	err := middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证系统提示已更新
	if len(req.SystemPrompt) <= len("You are a helpful assistant.") {
		t.Error("SystemPrompt should be updated with todo list")
	}
}

func TestTodoMiddleware_BeforeAgent(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(backend, toolRegistry)

	ctx := context.Background()

	// 先创建一个 Todo 列表
	todoContent := "# Todo List\n\n## ✅ Task 1 [1]\n\n"
	backend.WriteFile(ctx, "/todos.md", todoContent)

	// 创建状态
	state := agent.NewState()

	// 执行 BeforeAgent 钩子
	err := middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	// 验证状态已更新
	todos, ok := state.GetMetadata("todos")
	if !ok {
		t.Error("todos metadata should be set")
	}

	if todos.(string) != todoContent {
		t.Errorf("Expected todos content %q, got %q", todoContent, todos)
	}
}
