package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogRequest(t *testing.T) {
	// 清理测试日志目录
	logDir := filepath.Join(".", "logs", "llm_requests")
	os.RemoveAll(filepath.Join(".", "logs"))
	defer os.RemoveAll(filepath.Join(".", "logs"))

	// 创建测试请求
	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Hello"},
			{Role: RoleAssistant, Content: "Hi there!"},
		},
		SystemPrompt: "You are a helpful assistant",
		Tools: []ToolSchema{
			{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: map[string]any{"type": "object"},
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// 记录请求
	logRequest("anthropic", "claude-sonnet-4-5", req, false)

	// 等待文件写入
	time.Sleep(100 * time.Millisecond)

	// 验证日志文件是否创建
	date := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, "requests_"+date+".jsonl")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Fatalf("日志文件未创建: %s", logFile)
	}

	// 读取日志文件
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 解析日志
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		t.Fatal("日志文件为空")
	}

	// 使用最后一行（最新的日志）
	var log RequestLog
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &log); err != nil {
		t.Fatalf("解析日志失败: %v", err)
	}

	// 验证日志内容
	if log.Provider != "anthropic" {
		t.Errorf("Provider 错误: got %s, want anthropic", log.Provider)
	}
	if log.Model != "claude-sonnet-4-5" {
		t.Errorf("Model 错误: got %s, want claude-sonnet-4-5", log.Model)
	}
	if len(log.Messages) != 2 {
		t.Errorf("Messages 数量错误: got %d, want 2", len(log.Messages))
	}
	if log.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("SystemPrompt 错误: got %s", log.SystemPrompt)
	}
	if len(log.Tools) != 1 {
		t.Errorf("Tools 数量错误: got %d, want 1", len(log.Tools))
	}
	if log.MaxTokens != 1000 {
		t.Errorf("MaxTokens 错误: got %d, want 1000", log.MaxTokens)
	}
	if log.Temperature != 0.7 {
		t.Errorf("Temperature 错误: got %f, want 0.7", log.Temperature)
	}
	if log.IsStreaming {
		t.Error("IsStreaming 应该为 false")
	}
}

func TestLogRequest_Streaming(t *testing.T) {
	// 清理测试日志目录
	defer os.RemoveAll(filepath.Join(".", "logs"))

	// 创建测试请求
	req := &ModelRequest{
		Messages: []Message{
			{Role: RoleUser, Content: "Test streaming"},
		},
	}

	// 记录流式请求
	logRequest("openai", "gpt-4o", req, true)

	// 等待文件写入
	time.Sleep(100 * time.Millisecond)

	// 读取日志文件
	date := time.Now().Format("2006-01-02")
	logFile := filepath.Join(".", "logs", "llm_requests", "requests_"+date+".jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 解析日志
	var log RequestLog
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &log); err != nil {
		t.Fatalf("解析日志失败: %v", err)
	}

	// 验证流式标记
	if !log.IsStreaming {
		t.Error("IsStreaming 应该为 true")
	}
	if log.Provider != "openai" {
		t.Errorf("Provider 错误: got %s, want openai", log.Provider)
	}
}

func TestLogRequest_MultipleRequests(t *testing.T) {
	// 清理测试日志目录
	defer os.RemoveAll(filepath.Join(".", "logs"))

	// 记录多个请求
	for i := 0; i < 3; i++ {
		req := &ModelRequest{
			Messages: []Message{
				{Role: RoleUser, Content: "Request " + string(rune('A'+i))},
			},
		}
		logRequest("anthropic", "claude-sonnet-4-5", req, false)
	}

	// 等待文件写入
	time.Sleep(100 * time.Millisecond)

	// 读取日志文件
	date := time.Now().Format("2006-01-02")
	logFile := filepath.Join(".", "logs", "llm_requests", "requests_"+date+".jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	// 验证有3条日志
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("日志行数错误: got %d, want 3", len(lines))
	}

	// 验证每条日志都能正确解析
	for i, line := range lines {
		var log RequestLog
		if err := json.Unmarshal([]byte(line), &log); err != nil {
			t.Errorf("解析第 %d 条日志失败: %v", i+1, err)
		}
	}
}
