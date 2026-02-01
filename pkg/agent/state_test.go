package agent

import (
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func TestState_AddAndGetMessage(t *testing.T) {
	state := NewState()

	msg := llm.Message{
		Role:    llm.RoleUser,
		Content: "Hello",
	}

	state.AddMessage(msg)

	messages := state.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %q", messages[0].Content)
	}
}

func TestState_SetAndGetFile(t *testing.T) {
	state := NewState()

	state.SetFile("/test.txt", "Hello World")

	content, ok := state.GetFile("/test.txt")
	if !ok {
		t.Error("Expected file to exist")
	}

	if content != "Hello World" {
		t.Errorf("Expected content 'Hello World', got %q", content)
	}
}

func TestState_GetNonexistentFile(t *testing.T) {
	state := NewState()

	_, ok := state.GetFile("/nonexistent.txt")
	if ok {
		t.Error("Expected file to not exist")
	}
}

func TestState_SetAndGetMetadata(t *testing.T) {
	state := NewState()

	state.SetMetadata("key1", "value1")
	state.SetMetadata("key2", 42)

	value1, ok := state.GetMetadata("key1")
	if !ok {
		t.Error("Expected metadata key1 to exist")
	}

	if value1.(string) != "value1" {
		t.Errorf("Expected value 'value1', got %v", value1)
	}

	value2, ok := state.GetMetadata("key2")
	if !ok {
		t.Error("Expected metadata key2 to exist")
	}

	if value2.(int) != 42 {
		t.Errorf("Expected value 42, got %v", value2)
	}
}

func TestState_GetNonexistentMetadata(t *testing.T) {
	state := NewState()

	_, ok := state.GetMetadata("nonexistent")
	if ok {
		t.Error("Expected metadata to not exist")
	}
}

func TestState_ConcurrentAccess(t *testing.T) {
	state := NewState()

	// 并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			state.AddMessage(llm.Message{
				Role:    llm.RoleUser,
				Content: "Message",
			})
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	messages := state.GetMessages()
	if len(messages) != 10 {
		t.Errorf("Expected 10 messages, got %d", len(messages))
	}
}

func TestNewState(t *testing.T) {
	state := NewState()

	if state == nil {
		t.Fatal("Expected non-nil state")
	}

	if state.Messages == nil {
		t.Error("Expected Messages to be initialized")
	}

	if state.Files == nil {
		t.Error("Expected Files to be initialized")
	}

	if state.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}

	if len(state.Messages) != 0 {
		t.Error("Expected empty Messages")
	}

	if len(state.Files) != 0 {
		t.Error("Expected empty Files")
	}

	if len(state.Metadata) != 0 {
		t.Error("Expected empty Metadata")
	}
}
