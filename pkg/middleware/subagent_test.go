package middleware

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// subAgentMockLLMClient 用于测试的模拟LLM客户端
type subAgentMockLLMClient struct {
	responses []*llm.ModelResponse
	callCount int
}

func (m *subAgentMockLLMClient) Generate(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
	if m.callCount >= len(m.responses) {
		// 默认返回一个简单的响应
		return &llm.ModelResponse{
			Content:    "Task completed",
			StopReason: "end_turn",
		}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *subAgentMockLLMClient) StreamGenerate(ctx context.Context, req *llm.ModelRequest) (<-chan llm.StreamEvent, error) {
	eventChan := make(chan llm.StreamEvent, 10)

	go func() {
		defer close(eventChan)

		resp, err := m.Generate(ctx, req)
		if err != nil {
			eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeError, Error: err, Done: true}
			return
		}

		eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeStart, Done: false}
		if resp.Content != "" {
			eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeText, Content: resp.Content, Done: false}
		}
		for i := range resp.ToolCalls {
			eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeToolUse, ToolCall: &resp.ToolCalls[i], Done: false}
		}
		eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeEnd, StopReason: resp.StopReason, Done: true}
	}()

	return eventChan, nil
}

func (m *subAgentMockLLMClient) CountTokens(messages []llm.Message) int {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 3
	}
	return total
}

func TestNewSubAgentMiddleware(t *testing.T) {
	llmClient := &subAgentMockLLMClient{}
	toolRegistry := tools.NewRegistry()

	config := &SubAgentConfig{
		MaxDepth: 3,
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{},
		"Test system prompt",
		4096,
		0.7,
	)

	if m == nil {
		t.Fatal("Expected non-nil middleware")
	}

	if m.Name() != "subagent" {
		t.Errorf("Expected name 'subagent', got %s", m.Name())
	}

	if m.config.MaxDepth != 3 {
		t.Errorf("Expected MaxDepth 3, got %d", m.config.MaxDepth)
	}

	// 验证工具已注册
	tool, ok := toolRegistry.Get("delegate_to_subagent")
	if !ok {
		t.Error("Expected delegate_to_subagent tool to be registered")
	}
	if tool == nil {
		t.Error("Expected non-nil tool")
	}
}

func TestNewSubAgentMiddleware_DefaultConfig(t *testing.T) {
	llmClient := &subAgentMockLLMClient{}
	toolRegistry := tools.NewRegistry()

	// 传入nil配置，应该使用默认值
	m := NewSubAgentMiddleware(
		nil,
		llmClient,
		toolRegistry,
		[]agent.Middleware{},
		"Test system prompt",
		4096,
		0.7,
	)

	if m.config.MaxDepth != 3 {
		t.Errorf("Expected default MaxDepth 3, got %d", m.config.MaxDepth)
	}
}

func TestSubAgentMiddleware_ExecuteSubAgent(t *testing.T) {
	// 创建模拟LLM客户端
	llmClient := &subAgentMockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content:    "SubAgent completed the task successfully",
				StopReason: "end_turn",
			},
		},
	}

	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()

	// 创建文件系统中间件（子Agent需要）
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 3,
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	ctx := context.Background()

	// 执行子Agent
	args := map[string]any{
		"task": "Test task for subagent",
	}

	result, err := m.executeSubAgent(ctx, args)
	if err != nil {
		t.Fatalf("executeSubAgent failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	// 验证结果包含子Agent的响应
	if len(result) < 10 {
		t.Errorf("Expected longer result, got: %s", result)
	}
}

func TestSubAgentMiddleware_ExecuteSubAgent_WithContext(t *testing.T) {
	llmClient := &subAgentMockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content:    "Task completed with context",
				StopReason: "end_turn",
			},
		},
	}

	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 3,
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	ctx := context.Background()

	// 执行子Agent，带上下文
	args := map[string]any{
		"task":    "Test task",
		"context": "This is important context information",
	}

	result, err := m.executeSubAgent(ctx, args)
	if err != nil {
		t.Fatalf("executeSubAgent failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestSubAgentMiddleware_MaxDepthExceeded(t *testing.T) {
	llmClient := &subAgentMockLLMClient{}
	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 2, // 设置最大深度为2
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	// 创建一个深度为2的上下文
	ctx := context.Background()
	ctx = m.withDepth(ctx, 2)

	// 尝试执行子Agent，应该失败
	args := map[string]any{
		"task": "Test task",
	}

	_, err := m.executeSubAgent(ctx, args)
	if err == nil {
		t.Error("Expected error when max depth exceeded")
	}

	// 验证错误消息
	expectedMsg := "maximum recursion depth"
	if len(err.Error()) < len(expectedMsg) || err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestSubAgentMiddleware_DepthTracking(t *testing.T) {
	llmClient := &subAgentMockLLMClient{}
	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 5,
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	ctx := context.Background()

	// 测试初始深度
	depth := m.getDepth(ctx)
	if depth != 0 {
		t.Errorf("Expected initial depth 0, got %d", depth)
	}

	// 测试设置深度
	ctx = m.withDepth(ctx, 2)
	depth = m.getDepth(ctx)
	if depth != 2 {
		t.Errorf("Expected depth 2, got %d", depth)
	}

	// 测试增加深度
	ctx = m.withDepth(ctx, 3)
	depth = m.getDepth(ctx)
	if depth != 3 {
		t.Errorf("Expected depth 3, got %d", depth)
	}
}

func TestSubAgentMiddleware_InvalidArgs(t *testing.T) {
	llmClient := &subAgentMockLLMClient{}
	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 3,
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	ctx := context.Background()

	// 测试缺少task参数
	args := map[string]any{}
	_, err := m.executeSubAgent(ctx, args)
	if err == nil {
		t.Error("Expected error when task is missing")
	}

	// 测试task参数类型错误
	args = map[string]any{
		"task": 123, // 应该是string
	}
	_, err = m.executeSubAgent(ctx, args)
	if err == nil {
		t.Error("Expected error when task is not a string")
	}
}

func TestSubAgentMiddleware_StateIsolation(t *testing.T) {
	// 这个测试验证子Agent有独立的状态
	llmClient := &subAgentMockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content:    "SubAgent response",
				StopReason: "end_turn",
			},
		},
	}

	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 3,
	}

	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	ctx := context.Background()

	// 执行子Agent
	args := map[string]any{
		"task": "Test task",
	}

	result, err := m.executeSubAgent(ctx, args)
	if err != nil {
		t.Fatalf("executeSubAgent failed: %v", err)
	}

	// 验证结果不为空
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// 子Agent的状态应该是独立的，不会影响父Agent
	// 这里我们只是验证执行成功，实际的状态隔离在集成测试中验证
}

func TestSubAgentMiddleware_NoInfiniteRecursion(t *testing.T) {
	// 验证SubAgentMiddleware不会被包含在子Agent的中间件中
	llmClient := &subAgentMockLLMClient{
		responses: []*llm.ModelResponse{
			{
				Content:    "Task completed",
				StopReason: "end_turn",
			},
		},
	}

	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()
	fsMiddleware := NewFilesystemMiddleware(stateBackend, toolRegistry)

	config := &SubAgentConfig{
		MaxDepth: 3,
	}

	// 创建包含SubAgentMiddleware的中间件列表
	m := NewSubAgentMiddleware(
		config,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware},
		"You are a helpful assistant",
		4096,
		0.7,
	)

	// 将SubAgentMiddleware添加到中间件列表中
	m.middlewares = append(m.middlewares, m)

	ctx := context.Background()

	// 执行子Agent，应该不会导致无限递归
	args := map[string]any{
		"task": "Test task",
	}

	result, err := m.executeSubAgent(ctx, args)
	if err != nil {
		t.Fatalf("executeSubAgent failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}
