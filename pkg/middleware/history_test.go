package middleware

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSessionRecord_Serialization(t *testing.T) {
	// 创建测试数据
	now := time.Now().UTC().Format(time.RFC3339)
	record := SessionRecord{
		SessionID: "test-session-123",
		StartTime: now,
		EndTime:   now,
		Messages: []MessageLog{
			{
				Role:      "user",
				Content:   "请帮我创建一个 README.md 文件",
				Timestamp: now,
			},
			{
				Role:    "assistant",
				Content: "好的，我来帮你创建 README.md 文件。",
				ToolCalls: []ToolCallLog{
					{
						ID:   "toolu_01ABC123",
						Name: "Write",
						Input: map[string]any{
							"file_path": "README.md",
							"content":   "# Project Title\n\n...",
						},
						Output:    "Successfully wrote 150 bytes to README.md",
						IsError:   false,
						Timestamp: now,
					},
				},
				StopReason: "tool_use",
				Timestamp:  now,
			},
		},
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded SessionRecord
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证数据
	if decoded.SessionID != record.SessionID {
		t.Errorf("SessionID 不匹配，期望 %q，实际 %q", record.SessionID, decoded.SessionID)
	}
	if decoded.StartTime != record.StartTime {
		t.Errorf("StartTime 不匹配")
	}
	if len(decoded.Messages) != 2 {
		t.Fatalf("Messages 长度不匹配，期望 2，实际 %d", len(decoded.Messages))
	}

	userMsg := decoded.Messages[0]
	if userMsg.Content != "请帮我创建一个 README.md 文件" {
		t.Errorf("UserMessage.Content 不匹配")
	}

	assistantMsg := decoded.Messages[1]
	if len(assistantMsg.ToolCalls) != 1 {
		t.Fatalf("ToolCalls 长度不匹配，期望 1，实际 %d", len(assistantMsg.ToolCalls))
	}

	toolCall := assistantMsg.ToolCalls[0]
	if toolCall.Name != "Write" {
		t.Errorf("ToolCall.Name 不匹配，期望 %q，实际 %q", "Write", toolCall.Name)
	}
	if toolCall.Input["file_path"] != "README.md" {
		t.Errorf("ToolCall.Input[file_path] 不匹配")
	}
}

func TestSessionRecord_EmptyIterations(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	record := SessionRecord{
		SessionID: "test-session-456",
		StartTime: now,
		EndTime:   now,
		Messages:  []MessageLog{},
	}

	// 序列化
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded SessionRecord
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if len(decoded.Messages) != 0 {
		t.Errorf("Messages 应为空，实际长度 %d", len(decoded.Messages))
	}
}

func TestIterationLog_NoToolCalls(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	messages := []MessageLog{
		{
			Role:      "user",
			Content:   "你好",
			Timestamp: now,
		},
		{
			Role:       "assistant",
			Content:    "你好！有什么我可以帮助你的吗？",
			ToolCalls:  []ToolCallLog{},
			StopReason: "end_turn",
			Timestamp:  now,
		},
	}

	// 序列化
	data, err := json.Marshal(messages)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded []MessageLog
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if len(decoded) != 2 {
		t.Fatalf("Messages 长度不匹配，期望 2，实际 %d", len(decoded))
	}

	assistantMsg := decoded[1]
	if len(assistantMsg.ToolCalls) != 0 {
		t.Errorf("ToolCalls 应为空，实际长度 %d", len(assistantMsg.ToolCalls))
	}
	if assistantMsg.StopReason != "end_turn" {
		t.Errorf("StopReason 不匹配，期望 %q，实际 %q", "end_turn", assistantMsg.StopReason)
	}
}

func TestToolCallLog_ErrorCase(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	toolCall := ToolCallLog{
		ID:   "toolu_error",
		Name: "Read",
		Input: map[string]any{
			"file_path": "/nonexistent/file.txt",
		},
		Output:    "Error: file not found",
		IsError:   true,
		Timestamp: now,
	}

	// 序列化
	data, err := json.Marshal(toolCall)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded ToolCallLog
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if !decoded.IsError {
		t.Error("IsError 应为 true")
	}
	if decoded.Output != "Error: file not found" {
		t.Errorf("Output 不匹配")
	}
}

func TestSessionRecord_MultipleIterations(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	record := SessionRecord{
		SessionID: "test-session-789",
		StartTime: now,
		EndTime:   now,
		Messages: []MessageLog{
			{
				Role:      "user",
				Content:   "第一条消息",
				Timestamp: now,
			},
			{
				Role:       "assistant",
				Content:    "第一条回复",
				ToolCalls:  []ToolCallLog{},
				StopReason: "end_turn",
				Timestamp:  now,
			},
			{
				Role:      "user",
				Content:   "第二条消息",
				Timestamp: now,
			},
			{
				Role:       "assistant",
				Content:    "第二条回复",
				ToolCalls:  []ToolCallLog{},
				StopReason: "end_turn",
				Timestamp:  now,
			},
		},
	}

	// 序列化
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded SessionRecord
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if len(decoded.Messages) != 4 {
		t.Fatalf("Messages 长度不匹配，期望 4，实际 %d", len(decoded.Messages))
	}
	if decoded.Messages[0].Content != "第一条消息" {
		t.Errorf("第一条消息内容不匹配")
	}
	if decoded.Messages[2].Content != "第二条消息" {
		t.Errorf("第二条消息内容不匹配")
	}
}

func TestSessionInfo_Duration(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)

	info := &SessionInfo{
		SessionID:    "test-123",
		StartTime:    startTime,
		EndTime:      endTime,
		MessageCount: 3,
		DateDir:      "2026-02-09",
	}

	duration := info.Duration()
	expected := 5 * time.Minute

	if duration != expected {
		t.Errorf("Expected duration %v, got %v", expected, duration)
	}
}

func TestSessionInfo_Duration_ZeroEndTime(t *testing.T) {
	info := &SessionInfo{
		SessionID:    "test-123",
		StartTime:    time.Now(),
		EndTime:      time.Time{}, // 零值
		MessageCount: 3,
		DateDir:      "2026-02-09",
	}

	duration := info.Duration()
	if duration != 0 {
		t.Errorf("Expected duration 0 for zero EndTime, got %v", duration)
	}
}

func TestSessionInfo_ShortID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		expected  string
	}{
		{
			name:      "长 ID",
			sessionID: "abcdef123456789",
			expected:  "abcdef12",
		},
		{
			name:      "正好 8 位",
			sessionID: "12345678",
			expected:  "12345678",
		},
		{
			name:      "短于 8 位",
			sessionID: "abc",
			expected:  "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &SessionInfo{
				SessionID: tt.sessionID,
			}

			result := info.ShortID()
			if result != tt.expected {
				t.Errorf("Expected ShortID %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseSessionRecord(t *testing.T) {
	jsonData := `{
		"session_id": "test-123",
		"start_time": "2026-02-09T10:00:00Z",
		"end_time": "2026-02-09T10:05:00Z",
		"messages": [
			{
				"role": "user",
				"content": "Hello",
				"timestamp": "2026-02-09T10:00:00Z"
			}
		]
	}`

	record, err := ParseSessionRecord(jsonData)
	if err != nil {
		t.Fatalf("ParseSessionRecord failed: %v", err)
	}

	if record.SessionID != "test-123" {
		t.Errorf("Expected SessionID 'test-123', got %q", record.SessionID)
	}

	if len(record.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(record.Messages))
	}

	if record.Messages[0].Content != "Hello" {
		t.Errorf("Expected user message 'Hello', got %q", record.Messages[0].Content)
	}
}

func TestParseSessionRecord_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json}`

	_, err := ParseSessionRecord(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestSessionRecord_ToSessionInfo(t *testing.T) {
	record := &SessionRecord{
		SessionID: "test-123",
		StartTime: "2026-02-09T10:00:00Z",
		EndTime:   "2026-02-09T10:05:00Z",
		Messages: []MessageLog{
			{Role: "user", Content: "msg1"},
			{Role: "assistant", Content: "reply1"},
			{Role: "user", Content: "msg2"},
		},
	}

	dateDir := "2026-02-09"
	info := record.ToSessionInfo(dateDir)

	if info.SessionID != "test-123" {
		t.Errorf("Expected SessionID 'test-123', got %q", info.SessionID)
	}

	if info.MessageCount != 3 {
		t.Errorf("Expected 3 messages, got %d", info.MessageCount)
	}

	if info.DateDir != dateDir {
		t.Errorf("Expected DateDir %q, got %q", dateDir, info.DateDir)
	}

	// 验证时间解析
	expectedStart, _ := time.Parse(time.RFC3339, "2026-02-09T10:00:00Z")
	if !info.StartTime.Equal(expectedStart) {
		t.Errorf("StartTime mismatch: expected %v, got %v", expectedStart, info.StartTime)
	}

	expectedEnd, _ := time.Parse(time.RFC3339, "2026-02-09T10:05:00Z")
	if !info.EndTime.Equal(expectedEnd) {
		t.Errorf("EndTime mismatch: expected %v, got %v", expectedEnd, info.EndTime)
	}
}

func TestSessionRecord_ToSessionInfo_InvalidTime(t *testing.T) {
	record := &SessionRecord{
		SessionID: "test-123",
		StartTime: "invalid-time",
		EndTime:   "invalid-time",
		Messages:  []MessageLog{},
	}

	info := record.ToSessionInfo("2026-02-09")

	// 解析失败时应该返回零值时间
	if !info.StartTime.IsZero() {
		t.Error("Expected zero StartTime for invalid time format")
	}

	if !info.EndTime.IsZero() {
		t.Error("Expected zero EndTime for invalid time format")
	}
}
