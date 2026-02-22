package middleware

import (
	"context"
	"strings"
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
	content, err := backend.ReadFile(ctx, "todos.md", 0, 0)
	if err != nil {
		t.Fatalf("Failed to read todos.md: %v", err)
	}

	// 验证内容包含标题
	if len(content) == 0 {
		t.Error("todos.md should not be empty")
	}
}

func TestTodoMiddleware_WriteTodosWithGoal(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	_ = NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	tool, ok := toolRegistry.Get("write_todos")
	if !ok {
		t.Fatal("write_todos tool not found")
	}

	// 带 goal 字段写入
	args := map[string]any{
		"goal": "实现一个完整的 Agent 框架",
		"todos": []any{
			map[string]any{
				"id":     "1",
				"title":  "设计架构",
				"status": "completed",
			},
			map[string]any{
				"id":     "2",
				"title":  "实现核心",
				"status": "in_progress",
			},
		},
	}

	_, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	content, err := be.ReadFile(ctx, "todos.md", 0, 0)
	if err != nil {
		t.Fatalf("Failed to read todos.md: %v", err)
	}

	// 验证 goal 写入文件头部
	if !strings.HasPrefix(content, "# Goal\n\n实现一个完整的 Agent 框架\n\n") {
		t.Errorf("Expected goal header, got: %s", content)
	}

	// 不传 goal 再次更新（保留一个非 completed），应保留原 goal
	args2 := map[string]any{
		"todos": []any{
			map[string]any{
				"id":     "1",
				"title":  "设计架构",
				"status": "completed",
			},
			map[string]any{
				"id":     "2",
				"title":  "实现核心",
				"status": "in_progress",
			},
		},
	}

	_, err = tool.Execute(ctx, args2)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	content2, err := be.ReadFile(ctx, "todos.md", 0, 0)
	if err != nil {
		t.Fatalf("Failed to read todos.md: %v", err)
	}

	if !strings.HasPrefix(content2, "# Goal\n\n实现一个完整的 Agent 框架\n\n") {
		t.Errorf("Goal should be preserved on update without goal field, got: %s", content2)
	}
}

func TestTodoMiddleware_BeforeModel(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(backend, toolRegistry)

	ctx := context.Background()

	// 先创建一个 Todo 列表（无 goal，无消息 → 进度 fallback 到 SystemPrompt）
	backend.WriteFile(ctx, "todos.md", "# Todo List\n\n## ✅ Task 1 [1]\n\nCompleted task\n\n")

	req := &llm.ModelRequest{
		SystemPrompt: "You are a helpful assistant.",
		Messages:     []llm.Message{},
	}

	err := middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 无消息时进度 fallback 到 SystemPrompt
	if !strings.Contains(req.SystemPrompt, "任务进度") {
		t.Error("Progress should fallback to SystemPrompt when no messages")
	}
}

func TestTodoMiddleware_BeforeModel_InjectsToLastMessage(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	be.WriteFile(ctx, "todos.md", "# Todo List\n\n## ✅ Task 1 [1]\n\nCompleted task\n\n")

	req := &llm.ModelRequest{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "hello"},
			{Role: llm.RoleAssistant, Content: "hi"},
		},
	}

	err := middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 有消息时进度应注入到最后一条消息
	last := req.Messages[len(req.Messages)-1].Content
	if !strings.Contains(last, "任务进度") {
		t.Error("Progress should be injected into last message")
	}
	if !strings.Contains(last, "<system-reminder>") {
		t.Error("Progress should be wrapped in <system-reminder> tag")
	}
}

func TestTodoMiddleware_BeforeModelStructuredInjection(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	// 创建带 goal 的 todo 文件
	be.WriteFile(ctx, "todos.md", "# Goal\n\n实现用户认证\n\n# Todo List\n\n## ✅ 设计 API [1]\n\n## 🔄 实现登录 [2]\n\n## ⬜ 编写测试 [3]\n\n")

	req := &llm.ModelRequest{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "开始工作"},
			{Role: llm.RoleAssistant, Content: "好的"},
		},
	}

	err := middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// goal 应注入到 SystemPrompt
	if !strings.Contains(req.SystemPrompt, "## 任务目标\n实现用户认证") {
		t.Errorf("SystemPrompt should contain goal, got: %s", req.SystemPrompt)
	}

	// 进度和行动指引应注入到最后一条消息
	last := req.Messages[len(req.Messages)-1].Content
	if !strings.Contains(last, "已完成 1/3") {
		t.Errorf("Last message should contain progress stats 1/3, got: %s", last)
	}
	if !strings.Contains(last, "### 行动指引") {
		t.Error("Last message should contain action guidance section")
	}
	if !strings.Contains(last, "<system-reminder>") {
		t.Error("Last message should be wrapped in <system-reminder> tag")
	}
}

func TestTodoMiddleware_BeforeModelFallbackToUserMessage(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	// 创建不带 goal 的 todo 文件
	be.WriteFile(ctx, "todos.md", "# Todo List\n\n## ⬜ Task 1 [1]\n\n")

	req := &llm.ModelRequest{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "帮我实现一个 HTTP 服务器"},
			{Role: llm.RoleAssistant, Content: "好的，我来帮你实现"},
		},
	}

	err := middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 从第一条 user 消息提取的原始需求应注入到 SystemPrompt
	if !strings.Contains(req.SystemPrompt, "帮我实现一个 HTTP 服务器") {
		t.Error("Should fallback to first user message as original request in SystemPrompt")
	}

	// 进度应注入到最后一条消息
	last := req.Messages[len(req.Messages)-1].Content
	if !strings.Contains(last, "任务进度") {
		t.Error("Progress should be injected into last message")
	}
}

func TestTodoMiddleware_BeforeAgent(t *testing.T) {
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(backend, toolRegistry)

	ctx := context.Background()

	// 先创建一个 Todo 列表
	todoContent := "# Todo List\n\n## ✅ Task 1 [1]\n\n"
	backend.WriteFile(ctx, "todos.md", todoContent)

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

func TestTodoMiddleware_BeforeAgentCapturesOriginalRequest(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	state := agent.NewState()
	state.AddMessage(llm.Message{Role: llm.RoleUser, Content: "帮我重构认证模块"})
	state.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: "好的"})

	err := middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	req, ok := state.GetMetadata("original_request")
	if !ok {
		t.Fatal("original_request metadata should be set")
	}
	if req.(string) != "帮我重构认证模块" {
		t.Errorf("Expected original request, got: %s", req)
	}

	// 再次调用不应覆盖
	state.AddMessage(llm.Message{Role: llm.RoleUser, Content: "另一个请求"})
	err = middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	req2, _ := state.GetMetadata("original_request")
	if req2.(string) != "帮我重构认证模块" {
		t.Error("original_request should not be overwritten on subsequent calls")
	}
}

func TestTodoMiddleware_AfterModelWarningWithOriginalRequest(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	injector := NewContextInjectionMiddleware()
	middleware := NewTodoMiddlewareWithInjector(be, toolRegistry, injector)
	middleware.SetMaxRoundsWarning(2) // 2 轮后触发

	ctx := context.Background()

	state := agent.NewState()
	state.SetMetadata("original_request", "实现分布式缓存")

	// 模拟 2 轮未使用 todo
	resp := &llm.ModelResponse{ToolCalls: []llm.ToolCall{}}
	for i := 0; i < 2; i++ {
		middleware.AfterModel(ctx, resp, state)
	}

	// 检查注入的警告消息包含原始需求
	blocks := injector.GetPendingBlocks()
	found := false
	for _, block := range blocks {
		if strings.Contains(block, "实现分布式缓存") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Warning message should contain original request, got blocks: %v", blocks)
	}
}

func TestTodoMiddleware_PerSessionPath(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	// 模拟 BeforeAgent 设置 session_id
	state := agent.NewState()
	state.SetMetadata("session_id", "test-session-123")
	middleware.BeforeAgent(ctx, state)

	// 通过工具写入 todo
	tool, _ := toolRegistry.Get("write_todos")
	_, err := tool.Execute(ctx, map[string]any{
		"goal": "测试 per-session",
		"todos": []any{
			map[string]any{"id": "1", "title": "Task 1", "status": "pending"},
		},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 文件应写入 session 级路径
	_, err = be.ReadFile(ctx, "memory/todos/test-session-123.md", 0, 0)
	if err != nil {
		t.Fatal("Todo file should be at per-session path")
	}

	// 旧的全局路径不应存在
	_, err = be.ReadFile(ctx, "todos.md", 0, 0)
	if err == nil {
		t.Error("Todo file should NOT exist at global /todos.md path")
	}
}

func TestTodoMiddleware_AutoDeleteOnAllCompleted(t *testing.T) {
	be := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	middleware := NewTodoMiddleware(be, toolRegistry)

	ctx := context.Background()

	// 设置 session
	state := agent.NewState()
	state.SetMetadata("session_id", "session-auto-delete")
	middleware.BeforeAgent(ctx, state)

	tool, _ := toolRegistry.Get("write_todos")

	// 先写入未完成的 todo
	_, err := tool.Execute(ctx, map[string]any{
		"todos": []any{
			map[string]any{"id": "1", "title": "Task 1", "status": "in_progress"},
			map[string]any{"id": "2", "title": "Task 2", "status": "pending"},
		},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 文件应存在
	_, err = be.ReadFile(ctx, "memory/todos/session-auto-delete.md", 0, 0)
	if err != nil {
		t.Fatal("Todo file should exist when not all completed")
	}

	// 全部标记为 completed
	result, err := tool.Execute(ctx, map[string]any{
		"todos": []any{
			map[string]any{"id": "1", "title": "Task 1", "status": "completed"},
			map[string]any{"id": "2", "title": "Task 2", "status": "completed"},
		},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 返回消息应提示清理
	if !strings.Contains(result, "completed") || !strings.Contains(result, "cleaned up") {
		t.Errorf("Expected cleanup message, got: %s", result)
	}

	// 文件应已被删除
	_, err = be.ReadFile(ctx, "memory/todos/session-auto-delete.md", 0, 0)
	if err == nil {
		t.Error("Todo file should be deleted after all items completed")
	}
}
