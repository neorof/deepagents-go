package middleware

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// ContextInjectionMiddleware 上下文注入中间件
// 用于在调用模型前隐形注入提醒和上下文信息
type ContextInjectionMiddleware struct {
	*BaseMiddleware
	mu            sync.Mutex
	pendingBlocks []string
}

// NewContextInjectionMiddleware 创建上下文注入中间件
func NewContextInjectionMiddleware() *ContextInjectionMiddleware {
	return &ContextInjectionMiddleware{
		BaseMiddleware: NewBaseMiddleware("context_injection"),
		pendingBlocks:  make([]string, 0),
	}
}

// EnsureContextBlock 添加上下文块（去重）
// 添加的内容会在下次调用模型时注入到系统提示词中
func (m *ContextInjectionMiddleware) EnsureContextBlock(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在相同内容
	if slices.Contains(m.pendingBlocks, text) {
		return
	}
	m.pendingBlocks = append(m.pendingBlocks, text)
}

// RemoveContextBlock 移除指定的上下文块
func (m *ContextInjectionMiddleware) RemoveContextBlock(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, block := range m.pendingBlocks {
		if block == text {
			m.pendingBlocks = append(m.pendingBlocks[:i], m.pendingBlocks[i+1:]...)
			return
		}
	}
}

// ClearPendingBlocks 清空所有待注入的上下文块
func (m *ContextInjectionMiddleware) ClearPendingBlocks() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pendingBlocks = make([]string, 0)
}

// GetPendingBlocks 获取当前待注入的上下文块（用于调试）
func (m *ContextInjectionMiddleware) GetPendingBlocks() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]string, len(m.pendingBlocks))
	copy(result, m.pendingBlocks)
	return result
}

// BeforeModel 在调用模型前注入上下文
func (m *ContextInjectionMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.pendingBlocks) == 0 {
		return nil
	}

	// 构建注入内容，使用 <system-reminder> 标签包装
	var injection strings.Builder
	injection.WriteString("\n\n")

	for _, block := range m.pendingBlocks {
		fmt.Fprintf(&injection, "<system-reminder>\n%s\n</system-reminder>\n\n", block)
	}

	// 注入到最后一条消息末尾（靠近模型当前注意力焦点）
	appendToLastMessage(req, injection.String())

	// 清空 pending blocks（一次性注入）
	m.pendingBlocks = make([]string, 0)

	return nil
}
