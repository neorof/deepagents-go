package integration

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// TestFileOperations 测试文件操作的完整流程
func TestFileOperations(t *testing.T) {
	// 创建 Mock LLM 客户端
	mockClient := NewMockLLMClient([]llm.ModelResponse{
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Name: "write_file",
					Input: map[string]any{
						"path":    "/test.txt",
						"content": "Hello, World!",
					},
				},
			},
		},
		{
			Content:    "文件已创建成功",
			StopReason: "end_turn",
		},
	})

	// 创建后端和工具
	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()

	// 创建中间件
	fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

	// 创建 Agent
	config := &agent.Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []agent.Middleware{fsMiddleware},
		SystemPrompt:  "你是一个文件管理助手",
		MaxIterations: 5,
		MaxTokens:     1000,
	}

	executor := agent.NewRunnable(config)

	// 执行任务
	ctx := context.Background()
	output, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "创建文件 /test.txt，内容为 'Hello, World!'"},
		},
	})

	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 验证结果
	if len(output.Messages) == 0 {
		t.Error("期望有消息输出")
	}

	// 验证文件已创建
	content, err := backend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	if content != "Hello, World!" {
		t.Errorf("期望文件内容为 'Hello, World!'，实际为 %q", content)
	}
}

// TestMultipleToolCalls 测试多次工具调用
func TestMultipleToolCalls(t *testing.T) {
	mockClient := NewMockLLMClient([]llm.ModelResponse{
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Name: "write_file",
					Input: map[string]any{
						"path":    "/file1.txt",
						"content": "Content 1",
					},
				},
			},
		},
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_2",
					Name: "write_file",
					Input: map[string]any{
						"path":    "/file2.txt",
						"content": "Content 2",
					},
				},
			},
		},
		{
			Content:    "两个文件都已创建",
			StopReason: "end_turn",
		},
	})

	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

	config := &agent.Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []agent.Middleware{fsMiddleware},
		SystemPrompt:  "你是一个文件管理助手",
		MaxIterations: 10,
		MaxTokens:     1000,
	}

	executor := agent.NewRunnable(config)
	ctx := context.Background()

	output, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "创建两个文件"},
		},
	})

	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 验证两个文件都已创建
	content1, err := backend.ReadFile(ctx, "/file1.txt", 0, 0)
	if err != nil {
		t.Errorf("读取 file1.txt 失败: %v", err)
	}
	if content1 != "Content 1" {
		t.Errorf("file1.txt 内容不正确")
	}

	content2, err := backend.ReadFile(ctx, "/file2.txt", 0, 0)
	if err != nil {
		t.Errorf("读取 file2.txt 失败: %v", err)
	}
	if content2 != "Content 2" {
		t.Errorf("file2.txt 内容不正确")
	}

	// 验证消息数量
	if len(output.Messages) < 3 {
		t.Errorf("期望至少 3 条消息，实际 %d 条", len(output.Messages))
	}
}

// TestReadAndEdit 测试读取和编辑文件
func TestReadAndEdit(t *testing.T) {
	mockClient := NewMockLLMClient([]llm.ModelResponse{
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Name: "write_file",
					Input: map[string]any{
						"path":    "/test.txt",
						"content": "Hello World",
					},
				},
			},
		},
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_2",
					Name: "edit_file",
					Input: map[string]any{
						"path":       "/test.txt",
						"old_string": "World",
						"new_string": "Go",
					},
				},
			},
		},
		{
			Content:    "文件已编辑",
			StopReason: "end_turn",
		},
	})

	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

	config := &agent.Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []agent.Middleware{fsMiddleware},
		SystemPrompt:  "你是一个文件管理助手",
		MaxIterations: 10,
		MaxTokens:     1000,
	}

	executor := agent.NewRunnable(config)
	ctx := context.Background()

	_, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "创建并编辑文件"},
		},
	})

	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 验证文件内容已修改
	content, err := backend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	if content != "Hello Go" {
		t.Errorf("期望文件内容为 'Hello Go'，实际为 %q", content)
	}
}

// TestSearchFiles 测试文件搜索
func TestSearchFiles(t *testing.T) {
	mockClient := NewMockLLMClient([]llm.ModelResponse{
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_1",
					Name: "write_file",
					Input: map[string]any{
						"path":    "/test1.txt",
						"content": "Hello World",
					},
				},
			},
		},
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_2",
					Name: "write_file",
					Input: map[string]any{
						"path":    "/test2.txt",
						"content": "Hello Go",
					},
				},
			},
		},
		{
			Content:    "",
			StopReason: "tool_use",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_3",
					Name: "grep",
					Input: map[string]any{
						"pattern": "Hello",
					},
				},
			},
		},
		{
			Content:    "找到了包含 Hello 的文件",
			StopReason: "end_turn",
		},
	})

	backend := backend.NewStateBackend()
	toolRegistry := tools.NewRegistry()
	fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

	config := &agent.Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []agent.Middleware{fsMiddleware},
		SystemPrompt:  "你是一个文件管理助手",
		MaxIterations: 10,
		MaxTokens:     1000,
	}

	executor := agent.NewRunnable(config)
	ctx := context.Background()

	output, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "创建文件并搜索"},
		},
	})

	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 验证执行成功
	if len(output.Messages) < 4 {
		t.Errorf("期望至少 4 条消息，实际 %d 条", len(output.Messages))
	}

	// 验证最后一条消息是助手的回复
	lastMsg := output.Messages[len(output.Messages)-1]
	if lastMsg.Role != llm.RoleAssistant {
		t.Error("期望最后一条消息来自助手")
	}
	if lastMsg.Content == "" {
		t.Error("期望助手有回复内容")
	}
}
