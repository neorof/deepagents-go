package middleware

// RoundCounter 轮次计数器组件，用于跟踪连续未使用某功能的轮次
type RoundCounter struct {
	count           int                         // 当前计数
	maxWarningRound int                         // 触发警告的阈值
	injector        *ContextInjectionMiddleware // 上下文注入器
	warningMessage  string                      // 警告消息
}

// NewRoundCounter 创建轮次计数器
func NewRoundCounter(maxWarning int, injector *ContextInjectionMiddleware, msg string) *RoundCounter {
	return &RoundCounter{
		count:           0,
		maxWarningRound: maxWarning,
		injector:        injector,
		warningMessage:  msg,
	}
}

// Track 跟踪功能使用情况
// used: true 表示本轮使用了功能，false 表示未使用
func (rc *RoundCounter) Track(used bool) {
	if used {
		// 使用了功能，重置计数器
		rc.count = 0
	} else {
		// 未使用功能，增加计数器
		rc.count++

		// 超过阈值时通过上下文注入器注入提醒
		if rc.count >= rc.maxWarningRound && rc.injector != nil && rc.warningMessage != "" {
			rc.injector.EnsureContextBlock(rc.warningMessage)
		}
	}
}

// GetCount 获取当前计数
func (rc *RoundCounter) GetCount() int {
	return rc.count
}

// Reset 重置计数器
func (rc *RoundCounter) Reset() {
	rc.count = 0
}

// SetMaxWarning 设置触发警告的阈值
func (rc *RoundCounter) SetMaxWarning(max int) {
	rc.maxWarningRound = max
}

// SetWarningMessage 设置警告消息
func (rc *RoundCounter) SetWarningMessage(msg string) {
	rc.warningMessage = msg
}
