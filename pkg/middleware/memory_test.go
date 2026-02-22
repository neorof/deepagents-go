package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func TestNewMemoryMiddleware(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	if m == nil {
		t.Fatal("Expected non-nil middleware")
	}

	if m.Name() != "memory" {
		t.Errorf("Expected name 'memory', got %q", m.Name())
	}
}

func TestMemoryMiddleware_BeforeAgent_InitializesSession(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 调用 BeforeAgent
	err := m.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	// 验证 session_id 已设置
	if m.sessionID == "" {
		t.Error("Expected sessionID to be set")
	}

	// 验证 sessionRecord 已初始化
	if m.sessionRecord == nil {
		t.Fatal("Expected sessionRecord to be initialized")
	}

	if m.sessionRecord.SessionID != m.sessionID {
		t.Errorf("SessionRecord.SessionID mismatch")
	}

	if m.sessionRecord.StartTime == "" {
		t.Error("Expected StartTime to be set")
	}

	if m.sessionRecord.Iterations == nil {
		t.Error("Expected Iterations to be initialized")
	}
}

func TestMemoryMiddleware_BeforeAgent_UsesProvidedSessionID(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()
	state.SetMetadata("session_id", "custom-session-123")

	// 调用 BeforeAgent
	err := m.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	// 验证使用了提供的 session_id
	if m.sessionID != "custom-session-123" {
		t.Errorf("Expected sessionID 'custom-session-123', got %q", m.sessionID)
	}
}

func TestMemoryMiddleware_BeforeModel_RecordsUserMessage(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话
	m.BeforeAgent(ctx, state)

	// 创建请求
	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "你好"},
		},
	}

	// 调用 BeforeModel
	err := m.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 验证迭代日志已创建
	if m.iterationLog == nil {
		t.Fatal("Expected iterationLog to be created")
	}

	if m.iterationLog.Iteration != 1 {
		t.Errorf("Expected iteration 1, got %d", m.iterationLog.Iteration)
	}

	if m.iterationLog.UserMessage == nil {
		t.Fatal("Expected UserMessage to be set")
	}

	if m.iterationLog.UserMessage.Content != "你好" {
		t.Errorf("Expected content '你好', got %q", m.iterationLog.UserMessage.Content)
	}
}

func TestMemoryMiddleware_AfterModel_RecordsAssistantResponse(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话和迭代
	m.BeforeAgent(ctx, state)
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "你好"}},
	})

	// 创建响应（无工具调用）
	resp := &llm.ModelResponse{
		Content:    "你好！有什么我可以帮助你的吗？",
		StopReason: "end_turn",
		ToolCalls:  []llm.ToolCall{},
	}

	// 调用 AfterModel
	err := m.AfterModel(ctx, resp, state)
	if err != nil {
		t.Fatalf("AfterModel failed: %v", err)
	}

	// 验证迭代已持久化
	if len(m.sessionRecord.Iterations) != 1 {
		t.Fatalf("Expected 1 iteration, got %d", len(m.sessionRecord.Iterations))
	}

	iter := m.sessionRecord.Iterations[0]
	if iter.Assistant == nil {
		t.Fatal("Expected Assistant to be set")
	}

	if iter.Assistant.Content != "你好！有什么我可以帮助你的吗？" {
		t.Errorf("Assistant content mismatch")
	}

	if iter.Assistant.StopReason != "end_turn" {
		t.Errorf("StopReason mismatch")
	}

	// 验证 iterationLog 已清空
	if m.iterationLog != nil {
		t.Error("Expected iterationLog to be cleared")
	}
}

func TestMemoryMiddleware_AfterModel_RecordsToolCalls(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话和迭代
	m.BeforeAgent(ctx, state)
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "创建文件"}},
	})

	// 创建响应（有工具调用）
	resp := &llm.ModelResponse{
		Content:    "好的，我来创建文件。",
		StopReason: "tool_use",
		ToolCalls: []llm.ToolCall{
			{
				ID:   "toolu_123",
				Name: "Write",
				Input: map[string]any{
					"file_path": "test.txt",
					"content":   "Hello",
				},
			},
		},
	}

	// 调用 AfterModel
	err := m.AfterModel(ctx, resp, state)
	if err != nil {
		t.Fatalf("AfterModel failed: %v", err)
	}

	// 验证工具调用已记录，但迭代未持久化（等待工具执行）
	if m.iterationLog == nil {
		t.Fatal("Expected iterationLog to be set")
	}

	if m.iterationLog.Assistant == nil {
		t.Fatal("Expected Assistant to be set")
	}

	if len(m.iterationLog.Assistant.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(m.iterationLog.Assistant.ToolCalls))
	}

	toolCall := m.iterationLog.Assistant.ToolCalls[0]
	if toolCall.ID != "toolu_123" {
		t.Errorf("ToolCall ID mismatch")
	}

	if toolCall.Name != "Write" {
		t.Errorf("ToolCall Name mismatch")
	}

	// 验证迭代未持久化
	if len(m.sessionRecord.Iterations) != 0 {
		t.Error("Expected iteration not to be persisted yet")
	}
}

func TestMemoryMiddleware_AfterTool_RecordsToolOutput(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话和迭代
	m.BeforeAgent(ctx, state)
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "创建文件"}},
	})

	// 记录工具调用
	m.AfterModel(ctx, &llm.ModelResponse{
		Content:    "好的",
		StopReason: "tool_use",
		ToolCalls: []llm.ToolCall{
			{ID: "toolu_123", Name: "Write", Input: map[string]any{"file_path": "test.txt"}},
		},
	}, state)

	// 记录工具输出
	result := &llm.ToolResult{
		ToolCallID: "toolu_123",
		Content:    "Successfully wrote file",
		IsError:    false,
	}

	err := m.AfterTool(ctx, result, state)
	if err != nil {
		t.Fatalf("AfterTool failed: %v", err)
	}

	// 验证迭代已持久化
	if len(m.sessionRecord.Iterations) != 1 {
		t.Fatalf("Expected 1 iteration, got %d", len(m.sessionRecord.Iterations))
	}

	// 验证工具输出已记录
	iter := m.sessionRecord.Iterations[0]
	if iter.Assistant.ToolCalls[0].Output != "Successfully wrote file" {
		t.Errorf("ToolCall output mismatch")
	}

	if iter.Assistant.ToolCalls[0].IsError {
		t.Error("Expected IsError to be false")
	}

	// 验证 iterationLog 已清空
	if m.iterationLog != nil {
		t.Error("Expected iterationLog to be cleared")
	}
}

func TestMemoryMiddleware_AfterTool_MultipleToolCalls(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话和迭代
	m.BeforeAgent(ctx, state)
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "创建两个文件"}},
	})

	// 记录两个工具调用
	m.AfterModel(ctx, &llm.ModelResponse{
		Content:    "好的",
		StopReason: "tool_use",
		ToolCalls: []llm.ToolCall{
			{ID: "toolu_1", Name: "Write", Input: map[string]any{"file_path": "file1.txt"}},
			{ID: "toolu_2", Name: "Write", Input: map[string]any{"file_path": "file2.txt"}},
		},
	}, state)

	// 记录第一个工具输出
	m.AfterTool(ctx, &llm.ToolResult{
		ToolCallID: "toolu_1",
		Content:    "Created file1.txt",
		IsError:    false,
	}, state)

	// 验证迭代未持久化（还有一个工具未执行）
	if len(m.sessionRecord.Iterations) != 0 {
		t.Error("Expected iteration not to be persisted yet")
	}

	// 记录第二个工具输出
	m.AfterTool(ctx, &llm.ToolResult{
		ToolCallID: "toolu_2",
		Content:    "Created file2.txt",
		IsError:    false,
	}, state)

	// 验证迭代已持久化
	if len(m.sessionRecord.Iterations) != 1 {
		t.Fatalf("Expected 1 iteration, got %d", len(m.sessionRecord.Iterations))
	}

	// 验证两个工具调用都已记录
	iter := m.sessionRecord.Iterations[0]
	if len(iter.Assistant.ToolCalls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(iter.Assistant.ToolCalls))
	}
}

func TestMemoryMiddleware_PersistSession_WritesToFile(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话
	m.BeforeAgent(ctx, state)

	// 模拟一次完整的对话
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "你好"}},
	})
	m.AfterModel(ctx, &llm.ModelResponse{
		Content:    "你好！",
		StopReason: "end_turn",
		ToolCalls:  []llm.ToolCall{},
	}, state)

	// 验证文件已写入
	dateDir := time.Now().Format("2006-01-02")
	filePath := "/memory/sessions/" + dateDir + "/" + m.sessionID + ".json"

	content, err := b.ReadFile(ctx, filePath, 0, 0)
	if err != nil {
		t.Fatalf("Failed to read session file: %v", err)
	}

	// 验证 JSON 格式
	var record SessionRecord
	if err := json.Unmarshal([]byte(content), &record); err != nil {
		t.Fatalf("Failed to parse session JSON: %v", err)
	}

	if record.SessionID != m.sessionID {
		t.Errorf("SessionID mismatch")
	}

	if len(record.Iterations) != 1 {
		t.Errorf("Expected 1 iteration in file, got %d", len(record.Iterations))
	}
}

func TestMemoryMiddleware_LoadExistingSession(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 预先写入会话文件
	sessionID := "test-session-123"
	dateDir := time.Now().Format("2006-01-02")
	filePath := "/memory/sessions/" + dateDir + "/" + sessionID + ".json"

	existingRecord := SessionRecord{
		SessionID: sessionID,
		StartTime: time.Now().UTC().Format(time.RFC3339),
		EndTime:   time.Now().UTC().Format(time.RFC3339),
		Iterations: []IterationLog{
			{
				Iteration: 1,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				UserMessage: &MessageLog{
					Role:      "user",
					Content:   "之前的消息",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	}

	data, _ := json.Marshal(existingRecord)
	b.WriteFile(ctx, filePath, string(data))

	// 创建新的 middleware 并设置为恢复模式
	m := NewMemoryMiddleware(b)
	m.SetResumeMode(sessionID) // 设置恢复模式

	state := agent.NewState()
	state.SetMetadata("session_id", sessionID)

	err := m.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent failed: %v", err)
	}

	// 验证会话已恢复
	if len(m.sessionRecord.Iterations) != 1 {
		t.Fatalf("Expected 1 iteration from loaded session, got %d", len(m.sessionRecord.Iterations))
	}

	if m.currentIteration != 1 {
		t.Errorf("Expected currentIteration 1, got %d", m.currentIteration)
	}

	if m.sessionRecord.Iterations[0].UserMessage.Content != "之前的消息" {
		t.Errorf("Loaded iteration content mismatch")
	}
}

func TestMemoryMiddleware_ResumeMode_NotReloadOnSecondCall(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 预先写入会话文件（1 个迭代）
	sessionID := "test-resume-no-reload"
	dateDir := time.Now().Format("2006-01-02")
	filePath := "/memory/sessions/" + dateDir + "/" + sessionID + ".json"

	existingRecord := SessionRecord{
		SessionID: sessionID,
		StartTime: time.Now().UTC().Format(time.RFC3339),
		EndTime:   time.Now().UTC().Format(time.RFC3339),
		Iterations: []IterationLog{
			{
				Iteration: 1,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				UserMessage: &MessageLog{
					Role:      "user",
					Content:   "旧消息",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				},
			},
		},
	}
	data, _ := json.Marshal(existingRecord)
	b.WriteFile(ctx, filePath, string(data))

	m := NewMemoryMiddleware(b)
	m.SetResumeMode(sessionID)

	state := agent.NewState()
	state.SetMetadata("session_id", sessionID)

	// 第一次调用：应加载已有会话
	if err := m.BeforeAgent(ctx, state); err != nil {
		t.Fatalf("BeforeAgent (1st) failed: %v", err)
	}
	if len(m.sessionRecord.Iterations) != 1 {
		t.Fatalf("Expected 1 iteration after 1st call, got %d", len(m.sessionRecord.Iterations))
	}

	// 模拟新增一条迭代记录（正常对话流程会追加）
	m.sessionRecord.Iterations = append(m.sessionRecord.Iterations, IterationLog{
		Iteration: 2,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		UserMessage: &MessageLog{
			Role:    "user",
			Content: "新消息",
		},
	})

	// 第二次调用：不应重新加载，已有的 2 条记录应保留
	if err := m.BeforeAgent(ctx, state); err != nil {
		t.Fatalf("BeforeAgent (2nd) failed: %v", err)
	}
	if len(m.sessionRecord.Iterations) != 2 {
		t.Fatalf("Expected 2 iterations after 2nd call, got %d (session was incorrectly reloaded)", len(m.sessionRecord.Iterations))
	}
}

func TestMemoryMiddleware_LoadSessionMessages_BeforeBeforeAgent(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	// 预先写入会话文件
	sessionID := "test-load-before-agent"
	dateDir := time.Now().Format("2006-01-02")
	filePath := "/memory/sessions/" + dateDir + "/" + sessionID + ".json"

	existingRecord := SessionRecord{
		SessionID: sessionID,
		StartTime: time.Now().UTC().Format(time.RFC3339),
		EndTime:   time.Now().UTC().Format(time.RFC3339),
		Iterations: []IterationLog{
			{
				Iteration: 1,
				UserMessage: &MessageLog{
					Role:    "user",
					Content: "你好",
				},
				Assistant: &AssistantLog{
					Content: "你好！",
				},
			},
			{
				Iteration: 2,
				UserMessage: &MessageLog{
					Role:    "user",
					Content: "帮我写代码",
				},
				Assistant: &AssistantLog{
					Content: "好的",
				},
			},
		},
	}
	data, _ := json.Marshal(existingRecord)
	b.WriteFile(ctx, filePath, string(data))

	// 只调用 SetResumeMode，不调用 BeforeAgent
	m := NewMemoryMiddleware(b)
	m.SetResumeMode(sessionID)

	// 直接调用 LoadSessionMessages，应该能自动加载文件
	messages, err := m.LoadSessionMessages(ctx)
	if err != nil {
		t.Fatalf("LoadSessionMessages failed: %v", err)
	}

	if len(messages) != 4 {
		t.Fatalf("Expected 4 messages (2 user + 2 assistant), got %d", len(messages))
	}
	if messages[0].Content != "你好" {
		t.Errorf("Expected first message '你好', got %q", messages[0].Content)
	}
	if messages[1].Content != "你好！" {
		t.Errorf("Expected second message '你好！', got %q", messages[1].Content)
	}
}

func TestMemoryMiddleware_LoadSessionMessages_NoSessionID(t *testing.T) {
	b := backend.NewStateBackend()
	ctx := context.Background()

	m := NewMemoryMiddleware(b)

	// sessionID 为空，应返回错误
	_, err := m.LoadSessionMessages(ctx)
	if err == nil {
		t.Fatal("Expected error when sessionID is empty")
	}
}

func TestMemoryMiddleware_MultipleIterations(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话
	m.BeforeAgent(ctx, state)

	// 第一次对话
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "第一条消息"}},
	})
	m.AfterModel(ctx, &llm.ModelResponse{
		Content:    "第一条回复",
		StopReason: "end_turn",
		ToolCalls:  []llm.ToolCall{},
	}, state)

	// 第二次对话
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "第二条消息"}},
	})
	m.AfterModel(ctx, &llm.ModelResponse{
		Content:    "第二条回复",
		StopReason: "end_turn",
		ToolCalls:  []llm.ToolCall{},
	}, state)

	// 验证两次迭代都已记录
	if len(m.sessionRecord.Iterations) != 2 {
		t.Fatalf("Expected 2 iterations, got %d", len(m.sessionRecord.Iterations))
	}

	if m.sessionRecord.Iterations[0].Iteration != 1 {
		t.Errorf("First iteration number mismatch")
	}

	if m.sessionRecord.Iterations[1].Iteration != 2 {
		t.Errorf("Second iteration number mismatch")
	}

	if m.sessionRecord.Iterations[0].UserMessage.Content != "第一条消息" {
		t.Errorf("First iteration content mismatch")
	}

	if m.sessionRecord.Iterations[1].UserMessage.Content != "第二条消息" {
		t.Errorf("Second iteration content mismatch")
	}
}

func TestMemoryMiddleware_ConcurrentSafety(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话
	m.BeforeAgent(ctx, state)

	// 并发调用（测试 mutex 保护）
	done := make(chan bool, 2)

	go func() {
		m.BeforeModel(ctx, &llm.ModelRequest{
			Messages: []llm.Message{{Role: "user", Content: "消息1"}},
		})
		done <- true
	}()

	go func() {
		m.BeforeModel(ctx, &llm.ModelRequest{
			Messages: []llm.Message{{Role: "user", Content: "消息2"}},
		})
		done <- true
	}()

	<-done
	<-done

	// 验证没有 panic 或数据竞争
	if m.currentIteration < 1 {
		t.Error("Expected at least 1 iteration")
	}
}

func TestMemoryMiddleware_ToolCallError(t *testing.T) {
	b := backend.NewStateBackend()
	m := NewMemoryMiddleware(b)

	ctx := context.Background()
	state := agent.NewState()

	// 初始化会话和迭代
	m.BeforeAgent(ctx, state)
	m.BeforeModel(ctx, &llm.ModelRequest{
		Messages: []llm.Message{{Role: "user", Content: "读取文件"}},
	})

	// 记录工具调用
	m.AfterModel(ctx, &llm.ModelResponse{
		Content:    "好的",
		StopReason: "tool_use",
		ToolCalls: []llm.ToolCall{
			{ID: "toolu_123", Name: "Read", Input: map[string]any{"file_path": "/nonexistent"}},
		},
	}, state)

	// 记录工具错误
	result := &llm.ToolResult{
		ToolCallID: "toolu_123",
		Content:    "Error: file not found",
		IsError:    true,
	}

	err := m.AfterTool(ctx, result, state)
	if err != nil {
		t.Fatalf("AfterTool failed: %v", err)
	}

	// 验证错误已记录
	iter := m.sessionRecord.Iterations[0]
	if !iter.Assistant.ToolCalls[0].IsError {
		t.Error("Expected IsError to be true")
	}

	if !strings.Contains(iter.Assistant.ToolCalls[0].Output, "Error") {
		t.Errorf("Expected error message in output")
	}
}
