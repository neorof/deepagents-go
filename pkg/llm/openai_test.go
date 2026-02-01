package llm

import (
	"context"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewOpenAIClient(t *testing.T) {
	// 测试使用默认 model
	client := NewOpenAIClient("test-api-key", "", "")

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.model != openai.GPT4o {
		t.Errorf("Expected default model %q, got %q", openai.GPT4o, client.model)
	}

	if client.client == nil {
		t.Error("Expected non-nil internal client")
	}
}

func TestNewOpenAIClient_CustomModel(t *testing.T) {
	// 测试使用自定义 model
	customModel := openai.GPT4TurboPreview
	client := NewOpenAIClient("test-api-key", customModel, "")

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.model != customModel {
		t.Errorf("Expected model %q, got %q", customModel, client.model)
	}
}

func TestNewOpenAIClient_CustomBaseURL(t *testing.T) {
	// 测试使用自定义 base URL
	customURL := "https://api.example.com/v1"
	client := NewOpenAIClient("test-api-key", "", customURL)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// 无法直接验证 baseURL，但至少确保客户端创建成功
	if client.client == nil {
		t.Error("Expected non-nil internal client")
	}
}

func TestOpenAIClient_Generate_Integration(t *testing.T) {
	// 这是一个集成测试，需要真实的 API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set OPENAI_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewOpenAIClient(apiKey, openai.GPT4oMini, "")
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

func TestOpenAIClient_Generate_WithSystemPrompt(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set OPENAI_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewOpenAIClient(apiKey, openai.GPT4oMini, "")
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

func TestOpenAIClient_Generate_WithTools(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set OPENAI_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewOpenAIClient(apiKey, openai.GPT4oMini, "")
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

func TestOpenAIClient_Generate_WithTemperature(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: set OPENAI_API_KEY and RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewOpenAIClient(apiKey, openai.GPT4oMini, "")
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

func TestOpenAIClient_Generate_InvalidAPIKey(t *testing.T) {
	// 测试无效的 API key
	client := NewOpenAIClient("invalid-key", openai.GPT4oMini, "")
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

func TestOpenAIClient_CountTokens(t *testing.T) {
	client := NewOpenAIClient("test-key", "", "")

	messages := []Message{
		{Role: RoleUser, Content: "Hello, how are you?"},
		{Role: RoleAssistant, Content: "I'm doing well, thank you!"},
	}

	count := client.CountTokens(messages)
	if count <= 0 {
		t.Error("Expected positive token count")
	}

	// 验证计数大致正确（简化算法：字符数/3）
	expectedMin := (len(messages[0].Content) + len(messages[1].Content)) / 4
	expectedMax := (len(messages[0].Content) + len(messages[1].Content)) / 2

	if count < expectedMin || count > expectedMax {
		t.Errorf("Token count %d outside expected range [%d, %d]", count, expectedMin, expectedMax)
	}
}

func TestOpenAIClient_Generate_MultipleMessages(t *testing.T) {
	// 测试多条消息（不需要真实 API）
	client := NewOpenAIClient("test-key", openai.GPT4oMini, "")
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

func TestOpenAIClient_Generate_AllParameters(t *testing.T) {
	// 测试所有参数组合（不需要真实 API）
	client := NewOpenAIClient("test-key", openai.GPT4oMini, "")
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
