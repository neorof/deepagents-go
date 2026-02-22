package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// TodoMiddleware Todo 列表中间件
type TodoMiddleware struct {
	*BaseMiddleware
	backend         backend.Backend
	toolRegistry    *tools.Registry
	contextInjector *ContextInjectionMiddleware // 上下文注入器
	roundCounter    *RoundCounter               // 轮次计数器
	sessionID       string                      // 会话 ID，用于隔离 todo 文件
}

// NewTodoMiddleware 创建 Todo 中间件
func NewTodoMiddleware(backend backend.Backend, toolRegistry *tools.Registry) *TodoMiddleware {
	return NewTodoMiddlewareWithInjector(backend, toolRegistry, nil)
}

// NewTodoMiddlewareWithInjector 创建带上下文注入器的 Todo 中间件
func NewTodoMiddlewareWithInjector(backend backend.Backend, toolRegistry *tools.Registry, contextInjector *ContextInjectionMiddleware) *TodoMiddleware {
	m := &TodoMiddleware{
		BaseMiddleware:  NewBaseMiddleware("todo"),
		backend:         backend,
		toolRegistry:    toolRegistry,
		contextInjector: contextInjector,
		roundCounter: NewRoundCounter(
			3, // 默认 3 轮未使用 Todo 后提醒
			contextInjector,
			"提醒：你已经连续多轮未使用 write_todos 工具。对于复杂任务（3步以上），建议使用 Todo 列表进行规划和跟踪进度。",
		),
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
				"goal": map[string]any{
					"type":        "string",
					"description": "用户的核心需求描述。首次创建 Todo 时必须填写，后续更新时可省略以保留原值。",
				},
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

			// 提取 goal 字段（可选）
			goal := ""
			if g, ok := args["goal"].(string); ok && g != "" {
				goal = g
			} else {
				// 没有传入新 goal，尝试从现有文件中保留
				goal = m.readGoalFromFile(ctx)
			}

			// 格式化 Todo 列表
			var content strings.Builder
			if goal != "" {
				fmt.Fprintf(&content, "# Goal\n\n%s\n\n", goal)
			}
			content.WriteString("# Todo List\n\n")

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

				fmt.Fprintf(&content, "## %s %s [%s]\n\n", statusIcon, title, id)
				if description != "" {
					fmt.Fprintf(&content, "%s\n\n", description)
				}
			}

			// 检查是否全部完成
			allCompleted := len(todosRaw) > 0
			for _, todoRaw := range todosRaw {
				todo, ok := todoRaw.(map[string]any)
				if !ok {
					continue
				}
				if todo["status"].(string) != "completed" {
					allCompleted = false
					break
				}
			}

			// 全部完成：删除 todo 文件
			if allCompleted {
				_ = m.backend.DeleteFile(ctx, m.todoPath())
				m.roundCounter.Reset()
				return fmt.Sprintf("All %d todo items completed, todo list cleaned up", len(todosRaw)), nil
			}

			// 保存到会话级 todo 文件
			_, err := m.backend.WriteFile(ctx, m.todoPath(), content.String())
			if err != nil {
				return "", fmt.Errorf("failed to write todos: %w", err)
			}

			// 重置轮次计数器
			m.roundCounter.Reset()

			return fmt.Sprintf("Successfully updated %d todo items", len(todosRaw)), nil
		},
	))
}

// readGoalFromFile 从现有 todo 文件中提取 goal
func (m *TodoMiddleware) readGoalFromFile(ctx context.Context) string {
	content, err := m.backend.ReadFile(ctx, m.todoPath(), 0, 0)
	if err != nil {
		return ""
	}
	// 解析 "# Goal\n\n{goal}\n\n" 格式
	const prefix = "# Goal\n\n"
	if !strings.HasPrefix(content, prefix) {
		return ""
	}
	rest := content[len(prefix):]
	if goal, _, ok := strings.Cut(rest, "\n\n"); ok {
		return goal
	}
	return ""
}

// todoPath 返回当前会话的 todo 文件路径
func (m *TodoMiddleware) todoPath() string {
	if m.sessionID != "" {
		return fmt.Sprintf("memory/todos/%s.md", m.sessionID)
	}
	return "todos.md"
}

// BeforeAgent 在 Agent 执行前加载 Todo 列表，并捕获用户原始请求
func (m *TodoMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	// 从 state 获取 session_id
	if sid, ok := state.GetMetadata("session_id"); ok {
		if s, ok := sid.(string); ok && s != "" {
			m.sessionID = s
		}
	}

	// 捕获用户的第一条消息作为原始请求（兜底机制）
	if _, exists := state.GetMetadata("original_request"); !exists {
		for _, msg := range state.GetMessages() {
			if msg.Role == llm.RoleUser && strings.TrimSpace(msg.Content) != "" {
				state.SetMetadata("original_request", msg.Content)
				break
			}
		}
	}

	// 尝试读取现有的 Todo 列表
	content, err := m.backend.ReadFile(ctx, m.todoPath(), 0, 0)
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
	content, err := m.backend.ReadFile(ctx, m.todoPath(), 0, 0)
	if err != nil {
		// 没有 Todo 列表，跳过
		return nil
	}

	// 提取原始需求：优先从 todo 文件的 goal 字段，其次从请求消息中提取
	originalRequest := m.extractOriginalRequest(content, req.Messages)

	// 统计进度
	completed, total := m.countProgress(content)

	// 分层注入：goal → SystemPrompt（持续锚定），进度+指引 → 消息流（靠近当前上下文）

	// 1. 原始需求注入到 SystemPrompt
	if originalRequest != "" {
		req.SystemPrompt += fmt.Sprintf("\n\n## 任务目标\n%s\n", originalRequest)
	}

	// 2. 进度和行动指引注入到最后一条消息
	var progress strings.Builder
	fmt.Fprintf(&progress, "\n\n<system-reminder>\n## 任务进度（已完成 %d/%d）\n%s\n", completed, total, content)
	progress.WriteString("### 行动指引\n请围绕任务目标工作。完成当前 in_progress 的任务后，使用 write_todos 更新 Todo 列表。\n</system-reminder>\n")

	appendToLastMessage(req, progress.String())

	return nil
}

// extractOriginalRequest 提取原始需求：优先 goal，其次第一条 user 消息
func (m *TodoMiddleware) extractOriginalRequest(todoContent string, messages []llm.Message) string {
	// 优先从 todo 文件中提取 goal
	const prefix = "# Goal\n\n"
	if strings.HasPrefix(todoContent, prefix) {
		rest := todoContent[len(prefix):]
		if goal, _, ok := strings.Cut(rest, "\n\n"); ok && goal != "" {
			return goal
		}
	}

	// 兜底：从消息中提取第一条 user 消息
	for _, msg := range messages {
		if msg.Role == llm.RoleUser && strings.TrimSpace(msg.Content) != "" {
			return msg.Content
		}
	}
	return ""
}

// countProgress 统计 todo 完成进度
func (m *TodoMiddleware) countProgress(content string) (completed, total int) {
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "## ") {
			continue
		}
		total++
		if strings.HasPrefix(line, "## ✅") {
			completed++
		}
	}
	return
}

// AfterModel 在模型响应后检查是否使用了 Todo 工具
func (m *TodoMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
	// 检查是否调用了 write_todos 工具
	usedTodo := false
	for _, tc := range resp.ToolCalls {
		if tc.Name == "write_todos" {
			usedTodo = true
			break
		}
	}

	// 动态更新警告消息，包含原始需求
	if !usedTodo {
		originalRequest := ""
		if req, ok := state.GetMetadata("original_request"); ok {
			if s, ok := req.(string); ok {
				originalRequest = s
			}
		}
		if originalRequest == "" {
			originalRequest = m.readGoalFromFile(ctx)
		}

		if originalRequest != "" {
			m.roundCounter.SetWarningMessage(fmt.Sprintf(
				"提醒：你已经连续多轮未使用 write_todos 工具。用户的原始需求是：「%s」。请围绕该需求使用 Todo 列表规划和跟踪进度。",
				originalRequest,
			))
		}
	}

	// 使用轮次计数器跟踪
	m.roundCounter.Track(usedTodo)

	return nil
}

// SetMaxRoundsWarning 设置触发警告的阈值
func (m *TodoMiddleware) SetMaxRoundsWarning(rounds int) {
	m.roundCounter.SetMaxWarning(rounds)
}

// GetRoundsWithoutTodo 获取未使用 Todo 的轮次（用于测试）
func (m *TodoMiddleware) GetRoundsWithoutTodo() int {
	return m.roundCounter.GetCount()
}

// ResetRoundsCounter 重置轮次计数器
func (m *TodoMiddleware) ResetRoundsCounter() {
	m.roundCounter.Reset()
}
