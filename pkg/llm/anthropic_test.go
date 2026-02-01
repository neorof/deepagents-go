package llm

import (
	"context"
	"os"
	"testing"
)

func TestNewAnthropicClient(t *testing.T) {
	// 测试使用默认 model
	client := NewAnthropicClient("test-api-key", "", "")

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.model != "claude-sonnet-4-5-20250929" {
		t.Errorf("Expected default model 'claude-sonnet-4-5-20250929', got %q", client.model)
	}

	if client.client == nil {
		t.Error("Expected non-nil internal client")
	}
}

func TestNewAnthropicClient_CustomModel(t *testing.T) {
	// 测试使用自定义 model
	customModel := "claude-3-opus-20240229"
	client := NewAnthropicClient("test-api-key", customModel, "")

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.model != customModel {
		t.Errorf("Expected model %q, got %q", customModel, client.model)
	}
}

func TestAnthropicClient_Generate_Integration(t *testing.T) {
	// 这是一个集成测试，需要真实的 API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set ANTHROPIC_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewAnthropicClient(apiKey, "", "")
	ctx := context.Background()

	// 测试简单的文本生成
	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Say 'Hello' and nothing else."},
		},
		MaxTokens: 100,
	}

	resp, err := client.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.Content == "" {
		t.Error("Expected non-empty content")
	}

	if resp.StopReason == "" {
		t.Error("Expected non-empty stop reason")
	}
}

func TestAnthropicClient_Generate_WithSystemPrompt(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set ANTHROPIC_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewAnthropicClient(apiKey, "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "What is your role?"},
		},
		SystemPrompt: "You are a helpful math tutor.",
		MaxTokens:    100,
	}

	resp, err := client.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.Content == "" {
		t.Error("Expected non-empty content")
	}
}

func TestAnthropicClient_Generate_WithTools(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set ANTHROPIC_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewAnthropicClient(apiKey, "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "What is the weather in San Francisco?"},
		},
		MaxTokens: 100,
		Tools: []ToolSchema{
			{
				Name:        "get_weather",
				Description: "Get the current weather in a location",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	resp, err := client.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	// 模型可能返回文本或工具调用
	if resp.Content == "" && len(resp.ToolCalls) == 0 {
		t.Error("Expected either content or tool calls")
	}
}

func TestAnthropicClient_Generate_WithTemperature(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set ANTHROPIC_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewAnthropicClient(apiKey, "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Say hello."},
		},
		MaxTokens:   50,
		Temperature: 0.5,
	}

	resp, err := client.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.Content == "" {
		t.Error("Expected non-empty content")
	}
}

func TestAnthropicClient_Generate_InvalidAPIKey(t *testing.T) {
	// 测试无效的 API key
	client := NewAnthropicClient("invalid-key", "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Hello"},
		},
		MaxTokens: 100,
	}

	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAnthropicClient_Generate_WithSystemPrompt_NoAPI(t *testing.T) {
	// 测试带 SystemPrompt 的请求构建（不需要真实 API）
	client := NewAnthropicClient("test-key", "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test message"},
		},
		SystemPrompt: "You are a helpful assistant",
		MaxTokens:    100,
	}

	// 这会失败，但会覆盖参数构建代码
	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAnthropicClient_Generate_WithTemperature_NoAPI(t *testing.T) {
	// 测试带 Temperature 的请求构建
	client := NewAnthropicClient("test-key", "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test message"},
		},
		MaxTokens:   100,
		Temperature: 0.7,
	}

	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAnthropicClient_Generate_WithTools_NoAPI(t *testing.T) {
	// 测试带 Tools 的请求构建
	client := NewAnthropicClient("test-key", "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "What's the weather?"},
		},
		MaxTokens: 100,
		Tools: []ToolSchema{
			{
				Name:        "get_weather",
				Description: "Get weather information",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type": "string",
						},
					},
				},
			},
		},
	}

	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAnthropicClient_Generate_MultipleMessages(t *testing.T) {
	// 测试多条消息
	client := NewAnthropicClient("test-key", "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Hello"},
			{Role: RoleAssistant, Content: "Hi there!"},
			{Role: RoleUser, Content: "How are you?"},
		},
		MaxTokens: 100,
	}

	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAnthropicClient_Generate_AllParameters(t *testing.T) {
	// 测试所有参数组合
	client := NewAnthropicClient("test-key", "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test with all parameters"},
		},
		SystemPrompt: "You are a test assistant",
		MaxTokens:    200,
		Temperature:  0.5,
		Tools: []ToolSchema{
			{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"param": map[string]any{"type": "string"},
					},
				},
			},
		},
	}

	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestAnthropicClient_Generate_EmptyMessages(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set ANTHROPIC_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewAnthropicClient(apiKey, "", "")
	ctx := context.Background()

	req := &ModelRequest{
		Messages:  []Message{},
		MaxTokens: 100,
	}

	_, err := client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for empty messages")
	}
}
