package repl

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// mockLLMClient 用于测试的 mock LLM 客户端
type mockLLMClient struct{}

// mockMemoryMiddleware 模拟记忆中间件
type mockMemoryMiddleware struct {
	memories []string
}

func (m *mockMemoryMiddleware) Name() string                                              { return "memory" }
func (m *mockMemoryMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error { return nil }
func (m *mockMemoryMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	return nil
}
func (m *mockMemoryMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
	return nil
}
func (m *mockMemoryMiddleware) BeforeTool(ctx context.Context, toolCall *llm.ToolCall, state *agent.State) error {
	return nil
}
func (m *mockMemoryMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	return nil
}
func (m *mockMemoryMiddleware) ClearMemories() { m.memories = nil }

func (m *mockLLMClient) Generate(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
	return &llm.ModelResponse{
		Content:    "Mock response",
		StopReason: "end_turn",
	}, nil
}

func (m *mockLLMClient) StreamGenerate(ctx context.Context, req *llm.ModelRequest) (<-chan llm.StreamEvent, error) {
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
		eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeEnd, StopReason: resp.StopReason, Done: true}
	}()

	return eventChan, nil
}

func (m *mockLLMClient) CountTokens(messages []llm.Message) int {
	return len(messages) * 10
}

func createTestExecutor() *agent.Runnable {
	llmClient := &mockLLMClient{}
	toolRegistry := tools.NewRegistry()

	config := &agent.Config{
		LLMClient:     llmClient,
		ToolRegistry:  toolRegistry,
		SystemPrompt:  "Test system prompt",
		MaxIterations: 5,
		MaxTokens:     100,
	}

	return agent.NewExecutor(config)
}

func TestNew(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	if repl == nil {
		t.Fatal("Expected non-nil REPL")
	}

	if repl.executor == nil {
		t.Error("Expected executor to be set")
	}

	if repl.messages == nil {
		t.Error("Expected messages to be initialized")
	}

	if len(repl.messages) != 0 {
		t.Error("Expected empty messages initially")
	}
}

func TestHandleCommand_Help(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	tests := []string{"/help", "/h", "/?", "/HELP"}
	for _, cmd := range tests {
		buf.Reset()
		result := repl.handleCommand(cmd)
		if !result {
			t.Errorf("Expected handleCommand(%q) to return true", cmd)
		}
		if !strings.Contains(buf.String(), "可用命令") {
			t.Errorf("Expected help output for command %q", cmd)
		}
	}
}

func TestHandleCommand_Clear(t *testing.T) {
	// 带记忆中间件：/clear 同时清除对话历史和记忆
	mockMem := &mockMemoryMiddleware{memories: []string{"a", "b"}}
	config := &agent.Config{
		LLMClient:    &mockLLMClient{},
		ToolRegistry: tools.NewRegistry(),
		Middlewares:  []agent.Middleware{mockMem},
	}
	executor := agent.NewExecutor(config)
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	repl.messages = []llm.Message{
		{Role: llm.RoleUser, Content: "test"},
	}

	result := repl.handleCommand("/clear")
	if !result {
		t.Error("Expected handleCommand('/clear') to return true")
	}

	if len(repl.messages) != 0 {
		t.Error("Expected messages to be cleared")
	}

	if len(mockMem.memories) != 0 {
		t.Errorf("Expected memories to be cleared, got %d", len(mockMem.memories))
	}

	output := buf.String()
	if !strings.Contains(output, "对话历史已清除") {
		t.Error("Expected history clear confirmation")
	}
	if !strings.Contains(output, "记忆已清除") {
		t.Error("Expected memory clear confirmation")
	}
}

func TestHandleCommand_Clear_NoMemoryMiddleware(t *testing.T) {
	// 无记忆中间件：/clear 只清除对话历史，不 panic
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	repl.messages = []llm.Message{
		{Role: llm.RoleUser, Content: "test"},
	}

	result := repl.handleCommand("/clear")
	if !result {
		t.Error("Expected handleCommand('/clear') to return true")
	}

	if len(repl.messages) != 0 {
		t.Error("Expected messages to be cleared")
	}

	output := buf.String()
	if !strings.Contains(output, "对话历史已清除") {
		t.Error("Expected history clear confirmation")
	}
}

func TestHandleCommand_History(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	// 测试空历史
	repl.handleCommand("/history")
	if !strings.Contains(buf.String(), "暂无") {
		t.Error("Expected empty history message")
	}

	// 添加消息
	buf.Reset()
	repl.messages = []llm.Message{
		{Role: llm.RoleUser, Content: "test message"},
		{Role: llm.RoleAssistant, Content: "response"},
	}

	repl.handleCommand("/history")
	output := buf.String()
	if !strings.Contains(output, "对话历史") {
		t.Error("Expected history header")
	}
	if !strings.Contains(output, "用户") {
		t.Error("Expected user role in history")
	}
	if !strings.Contains(output, "助手") {
		t.Error("Expected assistant role in history")
	}
}

func TestHandleCommand_Unknown(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	// 不带 / 前缀的输入不是命令，返回 false
	result := repl.handleCommand("unknown_command")
	if result {
		t.Error("Expected handleCommand to return false for input without / prefix")
	}

	// 带 / 前缀的未知命令，返回 true 并提示错误
	buf.Reset()
	result = repl.handleCommand("/unknown")
	if !result {
		t.Error("Expected handleCommand to return true for unknown / command")
	}
	if !strings.Contains(buf.String(), "未知命令") {
		t.Error("Expected unknown command error message")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is a ..."},
		{"exact", 5, "exact"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestPrintWelcome(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	repl.printWelcome()

	output := buf.String()
	if !strings.Contains(output, "Deep Agents") {
		t.Error("Expected welcome message to contain 'Deep Agents'")
	}
	if !strings.Contains(output, "/help") {
		t.Error("Expected welcome message to contain '/help'")
	}
}

func TestPrintHelp(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	repl.printHelp()

	output := buf.String()
	if !strings.Contains(output, "/help") {
		t.Error("Expected help to mention '/help' command")
	}
	if !strings.Contains(output, "/exit") {
		t.Error("Expected help to mention '/exit' command")
	}
	if !strings.Contains(output, "/clear") {
		t.Error("Expected help to mention '/clear' command")
	}
	if !strings.Contains(output, "/history") {
		t.Error("Expected help to mention '/history' command")
	}
}

func TestPrintHistory_Empty(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	repl.printHistory()

	if !strings.Contains(buf.String(), "暂无") {
		t.Error("Expected empty history message")
	}
}

func TestPrintHistory_WithMessages(t *testing.T) {
	executor := createTestExecutor()
	repl := New(executor)

	var buf bytes.Buffer
	repl.writer = &buf

	repl.messages = []llm.Message{
		{Role: llm.RoleUser, Content: "Hello"},
		{Role: llm.RoleAssistant, Content: "Hi there!"},
	}

	repl.printHistory()

	output := buf.String()
	if !strings.Contains(output, "[1]") {
		t.Error("Expected message numbering")
	}
	if !strings.Contains(output, "用户") {
		t.Error("Expected user role")
	}
	if !strings.Contains(output, "助手") {
		t.Error("Expected assistant role")
	}
}
