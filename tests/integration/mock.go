package integration

import (
	"context"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// MockLLMClient 用于集成测试的 Mock LLM 客户端
type MockLLMClient struct {
	responses []llm.ModelResponse
	index     int
}

// NewMockLLMClient 创建 Mock LLM 客户端
func NewMockLLMClient(responses []llm.ModelResponse) *MockLLMClient {
	return &MockLLMClient{
		responses: responses,
		index:     0,
	}
}

// Generate 返回预设的响应
func (m *MockLLMClient) Generate(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
	if m.index >= len(m.responses) {
		// 如果没有更多响应，返回默认响应
		return &llm.ModelResponse{
			Content:    "Mock response",
			StopReason: "end_turn",
		}, nil
	}

	resp := m.responses[m.index]
	m.index++
	return &resp, nil
}

// StreamGenerate 返回流式响应
func (m *MockLLMClient) StreamGenerate(ctx context.Context, req *llm.ModelRequest) (<-chan llm.StreamEvent, error) {
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

// CountTokens 返回简单的 token 计数
func (m *MockLLMClient) CountTokens(messages []llm.Message) int {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 3
	}
	return total
}

// Reset 重置 Mock 客户端
func (m *MockLLMClient) Reset() {
	m.index = 0
}

// AddResponse 添加新的响应
func (m *MockLLMClient) AddResponse(resp llm.ModelResponse) {
	m.responses = append(m.responses, resp)
}
