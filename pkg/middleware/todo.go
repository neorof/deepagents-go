package middleware

import (
	"context"
	"fmt"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// TodoMiddleware Todo 列表中间件
type TodoMiddleware struct {
	*BaseMiddleware
	backend      backend.Backend
	toolRegistry *tools.Registry
}

// NewTodoMiddleware 创建 Todo 中间件
func NewTodoMiddleware(backend backend.Backend, toolRegistry *tools.Registry) *TodoMiddleware {
	m := &TodoMiddleware{
		BaseMiddleware: NewBaseMiddleware("todo"),
		backend:        backend,
		toolRegistry:   toolRegistry,
	}

	// 注册 write_todos 工具
	m.registerTools()

	return m
}

// registerTools 注册工具
func (m *TodoMiddleware) registerTools() {
	m.toolRegistry.Register(tools.NewBaseTool(
		"write_todos",
		"写入或更新 Todo 列表。用于任务规划和跟踪。",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"todos": map[string]any{
					"type":        "array",
					"description": "Todo 项列表",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{
								"type":        "string",
								"description": "Todo 项的唯一标识",
							},
							"title": map[string]any{
								"type":        "string",
								"description": "Todo 项标题",
							},
							"status": map[string]any{
								"type":        "string",
								"description": "状态：pending, in_progress, completed",
								"enum":        []string{"pending", "in_progress", "completed"},
							},
							"description": map[string]any{
								"type":        "string",
								"description": "详细描述",
							},
						},
						"required": []string{"id", "title", "status"},
					},
				},
			},
			"required": []string{"todos"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			todosRaw, ok := args["todos"].([]any)
			if !ok {
				return "", fmt.Errorf("todos must be an array")
			}

			// 格式化 Todo 列表
			var content string
			content += "# Todo List\n\n"

			for _, todoRaw := range todosRaw {
				todo, ok := todoRaw.(map[string]any)
				if !ok {
					continue
				}

				id := todo["id"].(string)
				title := todo["title"].(string)
				status := todo["status"].(string)
				description := ""
				if desc, ok := todo["description"].(string); ok {
					description = desc
				}

				statusIcon := "⬜"
				switch status {
				case "in_progress":
					statusIcon = "🔄"
				case "completed":
					statusIcon = "✅"
				}

				content += fmt.Sprintf("## %s %s [%s]\n\n", statusIcon, title, id)
				if description != "" {
					content += fmt.Sprintf("%s\n\n", description)
				}
			}

			// 保存到 /todos.md
			_, err := m.backend.WriteFile(ctx, "/todos.md", content)
			if err != nil {
				return "", fmt.Errorf("failed to write todos: %w", err)
			}

			return fmt.Sprintf("Successfully updated %d todo items", len(todosRaw)), nil
		},
	))
}

// BeforeAgent 在 Agent 执行前加载 Todo 列表
func (m *TodoMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	// 尝试读取现有的 Todo 列表
	content, err := m.backend.ReadFile(ctx, "/todos.md", 0, 0)
	if err != nil {
		// 文件不存在，忽略错误
		return nil
	}

	// 将 Todo 列表添加到状态
	state.SetMetadata("todos", content)

	return nil
}

// BeforeModel 在调用 LLM 前注入 Todo 列表到系统提示
func (m *TodoMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	// 从最后一条消息的状态中获取 Todo 列表
	// 这里简化实现，实际应该从 state 中获取
	content, err := m.backend.ReadFile(ctx, "/todos.md", 0, 0)
	if err != nil {
		// 没有 Todo 列表，跳过
		return nil
	}

	// 注入到系统提示
	todoPrompt := fmt.Sprintf("\n\n## Current Todo List\n\n%s\n\n", content)
	todoPrompt += "You can use the `write_todos` tool to update the todo list as you make progress."

	req.SystemPrompt += todoPrompt

	return nil
}
