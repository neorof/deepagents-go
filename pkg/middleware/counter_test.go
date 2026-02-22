package middleware

import (
	"testing"
)

func TestNewRoundCounter(t *testing.T) {
	counter := NewRoundCounter(5, nil, "test warning")

	if counter.GetCount() != 0 {
		t.Errorf("初始计数应为 0，实际为 %d", counter.GetCount())
	}

	if counter.maxWarningRound != 5 {
		t.Errorf("maxWarningRound 应为 5，实际为 %d", counter.maxWarningRound)
	}

	if counter.warningMessage != "test warning" {
		t.Errorf("warningMessage 应为 'test warning'，实际为 %q", counter.warningMessage)
	}
}

func TestRoundCounter_Track_NotUsed(t *testing.T) {
	counter := NewRoundCounter(5, nil, "")

	// 连续 3 轮未使用
	counter.Track(false)
	if counter.GetCount() != 1 {
		t.Errorf("第 1 轮未使用后计数应为 1，实际为 %d", counter.GetCount())
	}

	counter.Track(false)
	if counter.GetCount() != 2 {
		t.Errorf("第 2 轮未使用后计数应为 2，实际为 %d", counter.GetCount())
	}

	counter.Track(false)
	if counter.GetCount() != 3 {
		t.Errorf("第 3 轮未使用后计数应为 3，实际为 %d", counter.GetCount())
	}
}

func TestRoundCounter_Track_Used(t *testing.T) {
	counter := NewRoundCounter(5, nil, "")

	// 先增加计数
	counter.Track(false)
	counter.Track(false)
	if counter.GetCount() != 2 {
		t.Errorf("未使用 2 轮后计数应为 2，实际为 %d", counter.GetCount())
	}

	// 使用功能，计数器应重置
	counter.Track(true)
	if counter.GetCount() != 0 {
		t.Errorf("使用功能后计数应重置为 0，实际为 %d", counter.GetCount())
	}
}

func TestRoundCounter_Track_WithInjector(t *testing.T) {
	// 创建上下文注入器
	injector := NewContextInjectionMiddleware()
	counter := NewRoundCounter(3, injector, "警告：连续未使用功能")

	// 未达到阈值，不应注入警告
	counter.Track(false)
	counter.Track(false)

	blocks := injector.GetPendingBlocks()
	if len(blocks) != 0 {
		t.Errorf("未达到阈值时不应注入警告，实际注入了 %d 条", len(blocks))
	}

	// 达到阈值，应注入警告
	counter.Track(false)

	blocks = injector.GetPendingBlocks()
	if len(blocks) != 1 {
		t.Errorf("达到阈值时应注入 1 条警告，实际注入了 %d 条", len(blocks))
	}

	if len(blocks) > 0 && blocks[0] != "警告：连续未使用功能" {
		t.Errorf("警告消息不正确，期望 '警告：连续未使用功能'，实际为 %q", blocks[0])
	}
}

func TestRoundCounter_Track_WarningOnlyOnceAtThreshold(t *testing.T) {
	injector := NewContextInjectionMiddleware()
	counter := NewRoundCounter(2, injector, "警告消息")

	// 第 1 轮未使用
	counter.Track(false)
	if len(injector.GetPendingBlocks()) != 0 {
		t.Error("第 1 轮不应触发警告")
	}

	// 第 2 轮未使用，达到阈值
	counter.Track(false)
	if len(injector.GetPendingBlocks()) != 1 {
		t.Errorf("第 2 轮应触发警告，实际注入了 %d 条", len(injector.GetPendingBlocks()))
	}

	// 清空注入的警告（模拟下一轮）
	injector = NewContextInjectionMiddleware()
	counter.injector = injector

	// 第 3 轮未使用，继续超过阈值
	counter.Track(false)
	if len(injector.GetPendingBlocks()) != 1 {
		t.Errorf("第 3 轮应继续触发警告，实际注入了 %d 条", len(injector.GetPendingBlocks()))
	}
}

func TestRoundCounter_Reset(t *testing.T) {
	counter := NewRoundCounter(5, nil, "")

	// 增加计数
	counter.Track(false)
	counter.Track(false)
	counter.Track(false)

	if counter.GetCount() != 3 {
		t.Errorf("重置前计数应为 3，实际为 %d", counter.GetCount())
	}

	// 重置
	counter.Reset()

	if counter.GetCount() != 0 {
		t.Errorf("重置后计数应为 0，实际为 %d", counter.GetCount())
	}
}

func TestRoundCounter_SetMaxWarning(t *testing.T) {
	counter := NewRoundCounter(5, nil, "")

	counter.SetMaxWarning(10)

	if counter.maxWarningRound != 10 {
		t.Errorf("SetMaxWarning 后应为 10，实际为 %d", counter.maxWarningRound)
	}
}

func TestRoundCounter_SetWarningMessage(t *testing.T) {
	counter := NewRoundCounter(5, nil, "old message")

	counter.SetWarningMessage("new message")

	if counter.warningMessage != "new message" {
		t.Errorf("SetWarningMessage 后应为 'new message'，实际为 %q", counter.warningMessage)
	}
}

func TestRoundCounter_NoInjectorNoWarning(t *testing.T) {
	// 没有注入器，不应 panic
	counter := NewRoundCounter(2, nil, "警告消息")

	counter.Track(false)
	counter.Track(false)
	counter.Track(false)

	// 应该正常运行，不会 panic
	if counter.GetCount() != 3 {
		t.Errorf("计数应为 3，实际为 %d", counter.GetCount())
	}
}

func TestRoundCounter_EmptyWarningMessage(t *testing.T) {
	injector := NewContextInjectionMiddleware()
	counter := NewRoundCounter(2, injector, "")

	counter.Track(false)
	counter.Track(false)

	// 警告消息为空，不应注入
	blocks := injector.GetPendingBlocks()
	if len(blocks) != 0 {
		t.Errorf("警告消息为空时不应注入，实际注入了 %d 条", len(blocks))
	}
}

func TestRoundCounter_AlternatingUsage(t *testing.T) {
	counter := NewRoundCounter(3, nil, "")

	// 交替使用和不使用
	counter.Track(false) // count = 1
	if counter.GetCount() != 1 {
		t.Errorf("第 1 次未使用后计数应为 1，实际为 %d", counter.GetCount())
	}

	counter.Track(true) // count = 0
	if counter.GetCount() != 0 {
		t.Errorf("使用后计数应为 0，实际为 %d", counter.GetCount())
	}

	counter.Track(false) // count = 1
	counter.Track(false) // count = 2
	if counter.GetCount() != 2 {
		t.Errorf("再次未使用 2 次后计数应为 2，实际为 %d", counter.GetCount())
	}

	counter.Track(true) // count = 0
	if counter.GetCount() != 0 {
		t.Errorf("再次使用后计数应为 0，实际为 %d", counter.GetCount())
	}
}

func TestRoundCounter_Integration(t *testing.T) {
	// 模拟真实场景：Todo 中间件使用计数器
	injector := NewContextInjectionMiddleware()
	counter := NewRoundCounter(10, injector, "提醒：你已经连续多轮未使用 write_todos 工具。")

	// 模拟 10 轮未使用 Todo
	for i := 0; i < 9; i++ {
		counter.Track(false)
		if len(injector.GetPendingBlocks()) != 0 {
			t.Errorf("第 %d 轮不应触发警告", i+1)
		}
	}

	// 第 10 轮，达到阈值
	counter.Track(false)
	blocks := injector.GetPendingBlocks()
	if len(blocks) != 1 {
		t.Errorf("第 10 轮应触发警告，实际注入了 %d 条", len(blocks))
	}

	// 使用 Todo，计数器重置
	counter.Track(true)
	if counter.GetCount() != 0 {
		t.Errorf("使用 Todo 后计数应重置为 0，实际为 %d", counter.GetCount())
	}
}
