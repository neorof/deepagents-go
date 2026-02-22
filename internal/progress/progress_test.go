package progress

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	indicator := New(true)

	if indicator == nil {
		t.Fatal("Expected non-nil indicator")
	}

	if !indicator.enabled {
		t.Error("Expected indicator to be enabled")
	}

	if indicator.writer == nil {
		t.Error("Expected writer to be set")
	}
}

func TestIndicator_Disabled(t *testing.T) {
	indicator := New(false)

	// 禁用时不应该有任何输出
	var buf bytes.Buffer
	indicator.writer = &buf

	indicator.Start("test")
	time.Sleep(50 * time.Millisecond)
	indicator.Stop()

	if buf.Len() > 0 {
		t.Error("Expected no output when disabled")
	}
}

func TestIndicator_StartStop(t *testing.T) {
	indicator := New(true)
	var buf bytes.Buffer
	indicator.writer = &buf

	indicator.Start("test message")
	time.Sleep(150 * time.Millisecond)
	indicator.Stop()

	// 应该有输出
	if buf.Len() == 0 {
		t.Error("Expected output when enabled")
	}
}

func TestIndicator_Update(t *testing.T) {
	indicator := New(true)
	var buf bytes.Buffer
	indicator.writer = &buf

	indicator.Start("initial")
	time.Sleep(150 * time.Millisecond)
	indicator.Update("updated")
	time.Sleep(150 * time.Millisecond)
	indicator.Stop()

	output := buf.String()
	if !strings.Contains(output, "updated") {
		t.Error("Expected updated message in output")
	}
}

func TestIndicator_StopTwice(t *testing.T) {
	indicator := New(true)
	var buf bytes.Buffer
	indicator.writer = &buf

	indicator.Start("test")
	time.Sleep(50 * time.Millisecond)
	indicator.Stop()
	indicator.Stop() // 第二次调用不应该 panic
}

func TestNewBar(t *testing.T) {
	bar := NewBar(100, true)

	if bar == nil {
		t.Fatal("Expected non-nil bar")
	}

	if bar.total != 100 {
		t.Errorf("Expected total 100, got %d", bar.total)
	}

	if bar.current != 0 {
		t.Errorf("Expected current 0, got %d", bar.current)
	}
}

func TestBar_Disabled(t *testing.T) {
	bar := NewBar(10, false)
	var buf bytes.Buffer
	bar.writer = &buf

	bar.Increment()
	bar.Finish()

	if buf.Len() > 0 {
		t.Error("Expected no output when disabled")
	}
}

func TestBar_Increment(t *testing.T) {
	bar := NewBar(10, true)
	var buf bytes.Buffer
	bar.writer = &buf

	bar.Increment()

	if bar.current != 1 {
		t.Errorf("Expected current 1, got %d", bar.current)
	}

	output := buf.String()
	if !strings.Contains(output, "1/10") {
		t.Error("Expected progress in output")
	}
}

func TestBar_Set(t *testing.T) {
	bar := NewBar(100, true)
	var buf bytes.Buffer
	bar.writer = &buf

	bar.Set(50)

	if bar.current != 50 {
		t.Errorf("Expected current 50, got %d", bar.current)
	}

	output := buf.String()
	if !strings.Contains(output, "50/100") {
		t.Error("Expected progress in output")
	}
}

func TestBar_Finish(t *testing.T) {
	bar := NewBar(10, true)
	var buf bytes.Buffer
	bar.writer = &buf

	bar.Set(5)
	buf.Reset()
	bar.Finish()

	if bar.current != bar.total {
		t.Error("Expected current to equal total after finish")
	}

	output := buf.String()
	if !strings.Contains(output, "10/10") {
		t.Error("Expected 100% progress in output")
	}
}

func TestNewTracker(t *testing.T) {
	tracker := NewTracker(true)

	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}

	if tracker.indicator == nil {
		t.Error("Expected indicator to be set")
	}
}

func TestTracker_StartIteration(t *testing.T) {
	tracker := NewTracker(true)
	var buf bytes.Buffer
	tracker.indicator.writer = &buf

	tracker.Start()
	tracker.StartIteration(1)
	time.Sleep(50 * time.Millisecond)
	tracker.Stop()

	stats := tracker.GetStats()
	if stats.Iterations != 1 {
		t.Errorf("Expected iterations 1, got %d", stats.Iterations)
	}
}

func TestTracker_RecordToolCall(t *testing.T) {
	tracker := NewTracker(true)
	var buf bytes.Buffer
	tracker.indicator.writer = &buf

	tracker.Start()
	tracker.RecordToolCall("test_tool")
	time.Sleep(50 * time.Millisecond)
	tracker.Stop()

	stats := tracker.GetStats()
	if stats.ToolCalls != 1 {
		t.Errorf("Expected tool calls 1, got %d", stats.ToolCalls)
	}
}

func TestTracker_RecordTokens(t *testing.T) {
	tracker := NewTracker(true)

	tracker.RecordTokens(100)
	tracker.RecordTokens(50)

	stats := tracker.GetStats()
	if stats.Tokens != 150 {
		t.Errorf("Expected tokens 150, got %d", stats.Tokens)
	}
}

func TestTracker_GetStats(t *testing.T) {
	tracker := NewTracker(true)

	tracker.StartIteration(3)
	tracker.RecordToolCall("tool1")
	tracker.RecordToolCall("tool2")
	tracker.RecordTokens(200)

	stats := tracker.GetStats()
	if stats.Iterations != 3 {
		t.Errorf("Expected iterations 3, got %d", stats.Iterations)
	}
	if stats.ToolCalls != 2 {
		t.Errorf("Expected tool calls 2, got %d", stats.ToolCalls)
	}
	if stats.Tokens != 200 {
		t.Errorf("Expected tokens 200, got %d", stats.Tokens)
	}
}

func TestTracker_PrintStats(t *testing.T) {
	tracker := NewTracker(true)

	tracker.StartIteration(5)
	tracker.RecordToolCall("tool1")
	tracker.RecordTokens(300)

	// PrintStats 会输出到 stdout，我们只测试它不会 panic
	tracker.PrintStats()
}
