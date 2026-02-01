package agent

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// MockLLMClient 模拟 LLM 客户端
type MockLLMClient struct {
	responses []*llm.ModelResponse
	callCount int
}

func (m *MockLLMClient) Generate(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
	if m.callCount >= len(m.responses) {
		return &llm.ModelResponse{
			Content:    "Done",
			StopReason: "end_turn",
		}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *MockLLMClient) StreamGenerate(ctx context.Context, req *llm.ModelRequest) (<-chan llm.StreamEvent, error) {
	eventChan := make(chan llm.StreamEvent, 10)

	go func() {
		defer close(eventChan)

		// 获取响应
		resp, err := m.Generate(ctx, req)
		if err != nil {
			eventChan <- llm.StreamEvent{
				Type:  llm.StreamEventTypeError,
				Error: err,
				Done:  true,
			}
			return
		}

		// 发送开始事件
		eventChan <- llm.StreamEvent{
			Type: llm.StreamEventTypeStart,
			Done: false,
		}

		// 发送文本内容
		if resp.Content != "" {
			eventChan <- llm.StreamEvent{
				Type:    llm.StreamEventTypeText,
				Content: resp.Content,
				Done:    false,
			}
		}

		// 发送工具调用
		for i := range resp.ToolCalls {
			eventChan <- llm.StreamEvent{
				Type:     llm.StreamEventTypeToolUse,
				ToolCall: &resp.ToolCalls[i],
				Done:     false,
			}
		}

		// 发送结束事件
		eventChan <- llm.StreamEvent{
			Type:       llm.StreamEventTypeEnd,
			StopReason: resp.StopReason,
			Done:       true,
		}
	}()

	return eventChan, nil
}

func (m *MockLLMClient) CountTokens(messages []llm.Message) int {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4
	}
	return total
}

func TestExecutor_BasicInvoke(t *testing.T) {
	// 创建 mock LLM 客户端
	mockClient := &MockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content: "好的，我来帮你创建文件。",
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
				StopReason: "tool_use",
			},
			{
				Content:    "文件已创建成功！",
				StopReason: "end_turn",
			},
		},
	}

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 注册一个简单的 write_file 工具
	toolRegistry.Register(tools.NewBaseTool(
		"write_file",
		"写入文件",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":    map[string]any{"type": "string"},
				"content": map[string]any{"type": "string"},
			},
			"required": []string{"path", "content"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			return "File written successfully", nil
		},
	))

	// 创建 Agent 配置
	config := &Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []Middleware{},
		SystemPrompt:  "你是一个有用的助手",
		MaxIterations: 10,
	}

	// 创建执行器
	executor := NewExecutor(config)

	// 执行任务
	ctx := context.Background()
	input := &InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "创建文件 /test.txt"},
		},
	}

	output, err := executor.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// 验证结果
	if len(output.Messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(output.Messages))
	}

	// 验证 LLM 被调用了 2 次
	if mockClient.callCount != 2 {
		t.Errorf("Expected 2 LLM calls, got %d", mockClient.callCount)
	}
}

func TestExecutor_MaxIterations(t *testing.T) {
	// 创建一个总是返回工具调用的 mock 客户端
	mockClient := &MockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content: "调用工具",
				ToolCalls: []llm.ToolCall{
					{
						ID:    "call_1",
						Name:  "test_tool",
						Input: map[string]any{},
					},
				},
				StopReason: "tool_use",
			},
		},
	}

	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(tools.NewBaseTool(
		"test_tool",
		"测试工具",
		map[string]any{"type": "object"},
		func(ctx context.Context, args map[string]any) (string, error) {
			return "OK", nil
		},
	))

	config := &Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []Middleware{},
		MaxIterations: 3, // 限制最多 3 次迭代
	}

	executor := NewExecutor(config)

	ctx := context.Background()
	input := &InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "测试"},
		},
	}

	_, err := executor.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// 验证最多执行了 3 次迭代
	if mockClient.callCount > 3 {
		t.Errorf("Expected at most 3 iterations, got %d", mockClient.callCount)
	}
}

func TestExecutor_ToolNotFound(t *testing.T) {
	mockClient := &MockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content: "调用不存在的工具",
				ToolCalls: []llm.ToolCall{
					{
						ID:    "call_1",
						Name:  "nonexistent_tool",
						Input: map[string]any{},
					},
				},
				StopReason: "tool_use",
			},
			{
				Content:    "工具不存在",
				StopReason: "end_turn",
			},
		},
	}

	toolRegistry := tools.NewRegistry()

	config := &Config{
		LLMClient:     mockClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []Middleware{},
		MaxIterations: 10,
	}

	executor := NewExecutor(config)

	ctx := context.Background()
	input := &InvokeInput{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "测试"},
		},
	}

	output, err := executor.Invoke(ctx, input)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// 验证有错误消息
	found := false
	for _, msg := range output.Messages {
		if msg.Role == llm.RoleUser && len(msg.Content) > 0 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected error message for tool not found")
	}
}
