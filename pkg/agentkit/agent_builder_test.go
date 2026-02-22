package agentkit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
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
		Content:    "Mock response",
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
		eventChan <- llm.StreamEvent{Type: llm.StreamEventTypeEnd, Done: true}
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

func TestBuild_NoFilesystem(t *testing.T) {
	builder := New(
		WithLLM(&mockLLMClient{}),
	)

	err := builder.Build()
	if err == nil {
		t.Fatal("Expected error when WithFilesystem is not set")
	}
}

func TestBuild_DefaultMiddlewares(t *testing.T) {
	tmpDir := t.TempDir()

	builder := New(
		WithLLM(&mockLLMClient{}),
		WithFilesystem(tmpDir),
	)

	if err := builder.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// 验证所有默认中间件都已注册
	expected := map[string]bool{
		"filesystem":        false,
		"context_injection": false,
		"todo":              false,
		"web":               false,
		"skills":            false,
		"memory":            false,
		"subagent":          false,
	}

	for _, mw := range builder.middlewares {
		if _, ok := expected[mw.Name()]; ok {
			expected[mw.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected %s middleware to be registered", name)
		}
	}
}

func TestBuild_WithMemoryPaths(t *testing.T) {
	tmpDir := t.TempDir()
	memFile := filepath.Join(tmpDir, "memory.md")
	if err := os.WriteFile(memFile, []byte("# Test Memory"), 0644); err != nil {
		t.Fatalf("Failed to create memory file: %v", err)
	}

	builder := New(
		WithLLM(&mockLLMClient{}),
		WithFilesystem(tmpDir),
		WithMemoryPaths(memFile),
	)

	if err := builder.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	found := false
	for _, mw := range builder.middlewares {
		if mw.Name() == "memory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected memory middleware to be registered")
	}
}

func TestBuild_WithMemoryPaths_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "mem1.md"), []byte("# Memory 1"), 0644); err != nil {
		t.Fatalf("Failed to create memory file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "mem2.txt"), []byte("Memory 2"), 0644); err != nil {
		t.Fatalf("Failed to create memory file: %v", err)
	}

	builder := New(
		WithLLM(&mockLLMClient{}),
		WithFilesystem(tmpDir),
		WithMemoryPaths(tmpDir),
	)

	if err := builder.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	found := false
	for _, mw := range builder.middlewares {
		if mw.Name() == "memory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected memory middleware to be registered")
	}
}

func TestBuild_WithSummarization(t *testing.T) {
	tmpDir := t.TempDir()
	mockClient := &mockLLMClient{}

	builder := New(
		WithLLM(mockClient),
		WithFilesystem(tmpDir),
		EnableSummarization(),
	)

	if err := builder.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	found := false
	for _, mw := range builder.middlewares {
		if mw.Name() == "summarization" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected summarization middleware to be registered")
	}
}

func TestBuild_WithSummarization_CustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mockClient := &mockLLMClient{}

	builder := New(
		WithLLM(mockClient),
		WithFilesystem(tmpDir),
		EnableSummarization(&middleware.SummarizationConfig{
			MaxTokens:    10000,
			TargetTokens: 3000,
			Enabled:      true,
		}),
	)

	if err := builder.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	found := false
	for _, mw := range builder.middlewares {
		if mw.Name() == "summarization" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected summarization middleware to be registered")
	}
}

func TestBuild_WithSkillsDirs(t *testing.T) {
	tmpDir := t.TempDir()

	builder := New(
		WithLLM(&mockLLMClient{}),
		WithFilesystem(tmpDir),
		WithSkillsDirs("skills"),
	)

	if err := builder.Build(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	found := false
	for _, mw := range builder.middlewares {
		if mw.Name() == "skills" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected skills middleware to be registered")
	}
}
