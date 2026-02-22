package middleware

import (
	"context"
	"strings"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func TestContextInjection_BeforeModel_InjectsToLastMessage(t *testing.T) {
	m := NewContextInjectionMiddleware()
	m.EnsureContextBlock("reminder one")
	m.EnsureContextBlock("reminder two")

	req := &llm.ModelRequest{
		SystemPrompt: "You are helpful.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "hello"},
			{Role: llm.RoleAssistant, Content: "hi there"},
		},
	}

	if err := m.BeforeModel(context.Background(), req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// SystemPrompt 不应被修改
	if req.SystemPrompt != "You are helpful." {
		t.Errorf("SystemPrompt should not change, got: %s", req.SystemPrompt)
	}

	// 最后一条消息应包含注入内容
	last := req.Messages[len(req.Messages)-1].Content
	if !strings.Contains(last, "<system-reminder>") {
		t.Error("Last message should contain <system-reminder> tag")
	}
	if !strings.Contains(last, "reminder one") {
		t.Error("Last message should contain 'reminder one'")
	}
	if !strings.Contains(last, "reminder two") {
		t.Error("Last message should contain 'reminder two'")
	}

	// pending blocks 应被清空
	if len(m.GetPendingBlocks()) != 0 {
		t.Error("Pending blocks should be cleared after injection")
	}
}

func TestContextInjection_BeforeModel_FallbackToSystemPrompt(t *testing.T) {
	m := NewContextInjectionMiddleware()
	m.EnsureContextBlock("fallback reminder")

	req := &llm.ModelRequest{
		SystemPrompt: "You are helpful.",
		Messages:     []llm.Message{},
	}

	if err := m.BeforeModel(context.Background(), req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 没有消息时应 fallback 到 SystemPrompt
	if !strings.Contains(req.SystemPrompt, "fallback reminder") {
		t.Errorf("Should fallback to SystemPrompt when no messages, got: %s", req.SystemPrompt)
	}
}

func TestContextInjection_BeforeModel_NoPendingBlocks(t *testing.T) {
	m := NewContextInjectionMiddleware()

	req := &llm.ModelRequest{
		SystemPrompt: "You are helpful.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "hello"},
		},
	}

	if err := m.BeforeModel(context.Background(), req); err != nil {
		t.Fatalf("BeforeModel failed: %v", err)
	}

	// 无 pending blocks 时不应修改任何内容
	if req.SystemPrompt != "You are helpful." {
		t.Error("SystemPrompt should not change when no pending blocks")
	}
	if req.Messages[0].Content != "hello" {
		t.Error("Message should not change when no pending blocks")
	}
}
