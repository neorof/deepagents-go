package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// mockLLMClient 用于测试的 Mock LLM 客户端
type mockLLMClient struct {
	generateFunc       func(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error)
	streamGenerateFunc func(ctx context.Context, req *llm.ModelRequest) (<-chan llm.StreamEvent, error)
	countTokensFunc    func(messages []llm.Message) int
}

func (m *mockLLMClient) Generate(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}
	return &llm.ModelResponse{
		Content:    "Mock summary",
		StopReason: "end_turn",
	}, nil
}

func (m *mockLLMClient) StreamGenerate(ctx context.Context, req *llm.ModelRequest) (<-chan llm.StreamEvent, error) {
	if m.streamGenerateFunc != nil {
		return m.streamGenerateFunc(ctx, req)
	}

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

func (m *mockLLMClient) CountTokens(messages []llm.Message) int {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(messages)
	}
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 3
	}
	return total
}

func TestNewSummarizationMiddleware(t *testing.T) {
	mockClient := &mockLLMClient{}

	config := &SummarizationConfig{
		MaxTokens:    5000,
		TargetTokens: 1000,
		LLMClient:    mockClient,
		Enabled:      true,
	}

	m := NewSummarizationMiddleware(config)

	if m == nil {
		t.Fatal("Expected non-nil middleware")
	}

	if m.Name() != "summarization" {
		t.Errorf("Expected name 'summarization', got %q", m.Name())
	}

	if m.maxTokens != 5000 {
		t.Errorf("Expected maxTokens 5000, got %d", m.maxTokens)
	}

	if m.targetTokens != 1000 {
		t.Errorf("Expected targetTokens 1000, got %d", m.targetTokens)
	}

	if !m.enabled {
		t.Error("Expected middleware to be enabled")
	}
}

func TestNewSummarizationMiddleware_DefaultConfig(t *testing.T) {
	m := NewSummarizationMiddleware(nil)

	if m.maxTokens != 8000 {
		t.Errorf("Expected default maxTokens 8000, got %d", m.maxTokens)
	}

	if m.targetTokens != 2000 {
		t.Errorf("Expected default targetTokens 2000, got %d", m.targetTokens)
	}

	if !m.enabled {
		t.Error("Expected middleware to be enabled by default")
	}
}

func TestSummarizationMiddleware_SetEnabled(t *testing.T) {
	m := NewSummarizationMiddleware(nil)

	m.SetEnabled(false)
	if m.IsEnabled() {
		t.Error("Expected middleware to be disabled")
	}

	m.SetEnabled(true)
	if !m.IsEnabled() {
		t.Error("Expected middleware to be enabled")
	}
}

func TestSummarizationMiddleware_ShouldSummarize(t *testing.T) {
	mockClient := &mockLLMClient{
		countTokensFunc: func(messages []llm.Message) int {
			return 10000 // 超过阈值
		},
	}

	config := &SummarizationConfig{
		MaxTokens: 5000,
		LLMClient: mockClient,
		Enabled:   true,
	}

	m := NewSummarizationMiddleware(config)

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "Test message"},
	}

	if !m.ShouldSummarize(messages) {
		t.Error("Expected ShouldSummarize to return true")
	}
}

func TestSummarizationMiddleware_ShouldNotSummarize(t *testing.T) {
	mockClient := &mockLLMClient{
		countTokensFunc: func(messages []llm.Message) int {
			return 1000 // 未超过阈值
		},
	}

	config := &SummarizationConfig{
		MaxTokens: 5000,
		LLMClient: mockClient,
		Enabled:   true,
	}

	m := NewSummarizationMiddleware(config)

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "Test message"},
	}

	if m.ShouldSummarize(messages) {
		t.Error("Expected ShouldSummarize to return false")
	}
}

func TestSummarizationMiddleware_BeforeModel_NoSummary(t *testing.T) {
	mockClient := &mockLLMClient{
		countTokensFunc: func(messages []llm.Message) int {
			return 1000 // 未超过阈值
		},
	}

	config := &SummarizationConfig{
		MaxTokens: 5000,
		LLMClient: mockClient,
		Enabled:   true,
	}

	m := NewSummarizationMiddleware(config)

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Test message"},
		},
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证消息未被修改
	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if m.GetSummaryCount() != 0 {
		t.Error("Expected summary count to be 0")
	}
}

func TestSummarizationMiddleware_BeforeModel_WithSummary(t *testing.T) {
	mockClient := &mockLLMClient{
		countTokensFunc: func(messages []llm.Message) int {
			return 10000 // 超过阈值
		},
		generateFunc: func(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
			return &llm.ModelResponse{
				Content:    "This is a summary of the conversation",
				StopReason: "end_turn",
			}, nil
		},
	}

	config := &SummarizationConfig{
		MaxTokens: 5000,
		LLMClient: mockClient,
		Enabled:   true,
	}

	m := NewSummarizationMiddleware(config)

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Message 1"},
			{Role: llm.RoleAssistant, Content: "Response 1"},
			{Role: llm.RoleUser, Content: "Message 2"},
			{Role: llm.RoleAssistant, Content: "Response 2"},
			{Role: llm.RoleUser, Content: "Message 3"},
		},
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证消息已被摘要
	if len(req.Messages) != 4 { // 摘要 + 最后 3 条消息
		t.Errorf("Expected 4 messages after summarization, got %d", len(req.Messages))
	}

	// 验证第一条消息是摘要
	if !strings.Contains(req.Messages[0].Content, "对话摘要") {
		t.Error("Expected first message to be a summary")
	}

	if m.GetSummaryCount() != 1 {
		t.Errorf("Expected summary count to be 1, got %d", m.GetSummaryCount())
	}

	if m.GetLastSummarySize() == 0 {
		t.Error("Expected last summary size to be non-zero")
	}
}

func TestSummarizationMiddleware_BeforeModel_Disabled(t *testing.T) {
	mockClient := &mockLLMClient{
		countTokensFunc: func(messages []llm.Message) int {
			return 10000 // 超过阈值
		},
	}

	config := &SummarizationConfig{
		MaxTokens: 5000,
		LLMClient: mockClient,
		Enabled:   false, // 禁用
	}

	m := NewSummarizationMiddleware(config)

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Test message"},
		},
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证消息未被修改
	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if m.GetSummaryCount() != 0 {
		t.Error("Expected summary count to be 0")
	}
}

func TestSummarizationMiddleware_GetTokenCount(t *testing.T) {
	mockClient := &mockLLMClient{
		countTokensFunc: func(messages []llm.Message) int {
			return 1234
		},
	}

	config := &SummarizationConfig{
		LLMClient: mockClient,
	}

	m := NewSummarizationMiddleware(config)

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "Test message"},
	}

	count := m.GetTokenCount(messages)
	if count != 1234 {
		t.Errorf("Expected token count 1234, got %d", count)
	}
}

func TestSummarizationMiddleware_GetTokenCount_NoClient(t *testing.T) {
	m := NewSummarizationMiddleware(&SummarizationConfig{
		LLMClient: nil,
	})

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "123456789"}, // 9 字符 / 3 = 3 tokens
	}

	count := m.GetTokenCount(messages)
	if count != 3 {
		t.Errorf("Expected token count 3, got %d", count)
	}
}

func TestSummarizationMiddleware_SetMaxTokens(t *testing.T) {
	m := NewSummarizationMiddleware(nil)

	m.SetMaxTokens(10000)
	if m.GetMaxTokens() != 10000 {
		t.Errorf("Expected maxTokens 10000, got %d", m.GetMaxTokens())
	}
}

func TestSummarizationMiddleware_SetTargetTokens(t *testing.T) {
	m := NewSummarizationMiddleware(nil)

	m.SetTargetTokens(3000)
	if m.GetTargetTokens() != 3000 {
		t.Errorf("Expected targetTokens 3000, got %d", m.GetTargetTokens())
	}
}

func TestSummarizationMiddleware_BeforeModel_NoClient(t *testing.T) {
	m := NewSummarizationMiddleware(&SummarizationConfig{
		LLMClient: nil,
		Enabled:   true,
	})

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Test message"},
		},
	}

	ctx := context.Background()
	if err := m.BeforeModel(ctx, req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证消息未被修改（因为没有客户端）
	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}
}
