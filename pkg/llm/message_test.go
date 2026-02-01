package llm

import (
	"testing"
)

func TestAnthropicClient_CountTokens(t *testing.T) {
	client := &AnthropicClient{
		model: "claude-3-5-sonnet-20241022",
	}

	messages := []Message{
		{Role: RoleUser, Content: "Hello, how are you?"},
		{Role: RoleAssistant, Content: "I'm doing well, thank you!"},
	}

	tokens := client.CountTokens(messages)
	if tokens <= 0 {
		t.Error("Expected positive token count")
	}

	// 验证更长的消息有更多 tokens
	longMessages := []Message{
		{Role: RoleUser, Content: "This is a much longer message that should have more tokens than the previous one. Let's add even more text to make sure the token count increases proportionally."},
	}

	longTokens := client.CountTokens(longMessages)
	if longTokens <= tokens {
		t.Error("Expected longer message to have more tokens")
	}
}

func TestMessage_Roles(t *testing.T) {
	tests := []struct {
		role     Role
		expected string
	}{
		{RoleUser, "user"},
		{RoleAssistant, "assistant"},
		{RoleSystem, "system"},
	}

	for _, tt := range tests {
		if string(tt.role) != tt.expected {
			t.Errorf("Expected role %q, got %q", tt.expected, string(tt.role))
		}
	}
}

func TestToolCall_Structure(t *testing.T) {
	toolCall := ToolCall{
		ID:   "call_123",
		Name: "test_tool",
		Input: map[string]any{
			"param1": "value1",
			"param2": 42,
		},
	}

	if toolCall.ID != "call_123" {
		t.Errorf("Expected ID 'call_123', got %q", toolCall.ID)
	}

	if toolCall.Name != "test_tool" {
		t.Errorf("Expected Name 'test_tool', got %q", toolCall.Name)
	}

	if len(toolCall.Input) != 2 {
		t.Errorf("Expected 2 input parameters, got %d", len(toolCall.Input))
	}
}

func TestToolResult_Structure(t *testing.T) {
	result := ToolResult{
		ToolCallID: "call_123",
		Content:    "Success",
		IsError:    false,
	}

	if result.ToolCallID != "call_123" {
		t.Errorf("Expected ToolCallID 'call_123', got %q", result.ToolCallID)
	}

	if result.IsError {
		t.Error("Expected IsError to be false")
	}
}

func TestModelRequest_Structure(t *testing.T) {
	req := ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Hello"},
		},
		SystemPrompt: "You are a helpful assistant",
		MaxTokens:    1000,
		Temperature:  0.7,
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.MaxTokens != 1000 {
		t.Errorf("Expected MaxTokens 1000, got %d", req.MaxTokens)
	}

	if req.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", req.Temperature)
	}
}

func TestModelResponse_Structure(t *testing.T) {
	resp := ModelResponse{
		Content: "Hello!",
		ToolCalls: []ToolCall{
			{ID: "call_1", Name: "tool1"},
		},
		StopReason: "end_turn",
	}

	if resp.Content != "Hello!" {
		t.Errorf("Expected Content 'Hello!', got %q", resp.Content)
	}

	if len(resp.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(resp.ToolCalls))
	}

	if resp.StopReason != "end_turn" {
		t.Errorf("Expected StopReason 'end_turn', got %q", resp.StopReason)
	}
}

func TestToolSchema_Structure(t *testing.T) {
	schema := ToolSchema{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"param1": map[string]any{"type": "string"},
			},
		},
	}

	if schema.Name != "test_tool" {
		t.Errorf("Expected Name 'test_tool', got %q", schema.Name)
	}

	if schema.Description != "A test tool" {
		t.Errorf("Expected Description 'A test tool', got %q", schema.Description)
	}

	if schema.InputSchema["type"] != "object" {
		t.Error("Expected InputSchema type to be 'object'")
	}
}
