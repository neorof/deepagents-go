package tools

import (
	"context"
	"strings"
	"testing"
)

func TestBashTool(t *testing.T) {
	tool := NewBashTool()

	// 测试工具基本信息
	if tool.Name() != "bash" {
		t.Errorf("Expected name 'bash', got '%s'", tool.Name())
	}

	// 测试简单命令
	t.Run("simple command", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{
			"command": "echo 'Hello World'",
		})
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !strings.Contains(result, "Hello World") {
			t.Errorf("Expected output to contain 'Hello World', got: %s", result)
		}
	})

	// 测试带错误的命令
	t.Run("command with error", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{
			"command": "ls /nonexistent",
		})
		// bash 工具不应该返回 error，而是在结果中包含错误信息
		if err != nil {
			t.Fatalf("Execute should not return error: %v", err)
		}
		if !strings.Contains(result, "ERROR") && !strings.Contains(result, "STDERR") {
			t.Errorf("Expected error information in result, got: %s", result)
		}
	})

	// 测试超时
	t.Run("timeout", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{
			"command": "sleep 5",
			"timeout": float64(1), // JSON 数字会被解析为 float64
		})
		if err != nil {
			t.Fatalf("Execute should not return error: %v", err)
		}
		if !strings.Contains(result, "timed out") && !strings.Contains(result, "killed") {
			t.Errorf("Expected timeout message, got: %s", result)
		}
	})

	// 测试多行输出
	t.Run("multiline output", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{
			"command": "echo 'line1'; echo 'line2'; echo 'line3'",
		})
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !strings.Contains(result, "line1") || !strings.Contains(result, "line2") || !strings.Contains(result, "line3") {
			t.Errorf("Expected all lines in output, got: %s", result)
		}
	})

	// 测试 stderr 输出
	t.Run("stderr output", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]any{
			"command": "echo 'error message' >&2",
		})
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !strings.Contains(result, "STDERR") || !strings.Contains(result, "error message") {
			t.Errorf("Expected stderr output, got: %s", result)
		}
	})
}
