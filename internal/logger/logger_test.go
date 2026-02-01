package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo}, // 默认返回 Info
	}

	for _, tt := range tests {
		result := ParseLevel(tt.input)
		if result != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("Level(%d).String() = %q, expected %q", tt.level, result, tt.expected)
		}
	}
}

func TestLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug, &buf, "text")

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Expected output to contain 'INFO'")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain 'test message'")
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug, &buf, "json")

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("Expected JSON output to contain level field")
	}
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Error("Expected JSON output to contain msg field")
	}
	if !strings.Contains(output, `"time"`) {
		t.Error("Expected JSON output to contain time field")
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelWarn, &buf, "text")

	// Debug 和 Info 不应该输出
	logger.Debug("debug message")
	logger.Info("info message")

	if buf.Len() > 0 {
		t.Error("Expected no output for debug and info when level is warn")
	}

	// Warn 和 Error 应该输出
	logger.Warn("warn message")
	output := buf.String()
	if !strings.Contains(output, "warn message") {
		t.Error("Expected warn message to be logged")
	}

	buf.Reset()
	logger.Error("error message")
	output = buf.String()
	if !strings.Contains(output, "error message") {
		t.Error("Expected error message to be logged")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo, &buf, "text")

	logger.Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("Expected no output for debug when level is info")
	}

	// 修改日志级别
	logger.SetLevel(LevelDebug)
	logger.Debug("should appear")
	if buf.Len() == 0 {
		t.Error("Expected output for debug after setting level to debug")
	}
}

func TestLogger_AllLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelDebug, &buf, "text")

	logger.Debug("debug msg")
	logger.Info("info msg")
	logger.Warn("warn msg")
	logger.Error("error msg")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected DEBUG in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO in output")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR in output")
	}
}

func TestLogger_FormatArgs(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo, &buf, "text")

	logger.Info("test %s %d", "message", 42)

	output := buf.String()
	if !strings.Contains(output, "test message 42") {
		t.Error("Expected formatted message in output")
	}
}

func TestEscapeJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`test"quote`, `test\"quote`},
		{"test\nline", `test\nline`},
		{"test\\slash", `test\\slash`},
		{"test\ttab", `test\ttab`},
	}

	for _, tt := range tests {
		result := escapeJSON(tt.input)
		if result != tt.expected {
			t.Errorf("escapeJSON(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New(LevelInfo, &buf, "text")
	SetDefault(logger)

	Info("test default logger")

	output := buf.String()
	if !strings.Contains(output, "test default logger") {
		t.Error("Expected message from default logger")
	}
}

func TestNew_DefaultValues(t *testing.T) {
	logger := New(LevelInfo, nil, "")

	if logger.level != LevelInfo {
		t.Error("Expected level to be set")
	}
	if logger.format != "text" {
		t.Error("Expected default format to be 'text'")
	}
	if logger.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}
