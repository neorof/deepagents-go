package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Indicator 进度指示器
type Indicator struct {
	writer  io.Writer
	enabled bool
	mu      sync.Mutex
	done    chan struct{}
	message string
	spinner []string
	index   int
}

// New 创建新的进度指示器
func New(enabled bool) *Indicator {
	return &Indicator{
		writer:  os.Stderr,
		enabled: enabled,
		done:    make(chan struct{}),
		spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Start 开始显示进度
func (i *Indicator) Start(message string) {
	if !i.enabled {
		return
	}

	i.mu.Lock()
	i.message = message
	i.index = 0
	i.mu.Unlock()

	go i.spin()
}

// Update 更新进度消息
func (i *Indicator) Update(message string) {
	if !i.enabled {
		return
	}

	i.mu.Lock()
	i.message = message
	i.mu.Unlock()
}

// Stop 停止显示进度
func (i *Indicator) Stop() {
	if !i.enabled {
		return
	}

	select {
	case <-i.done:
		// 已经停止
		return
	default:
		close(i.done)
		i.clear()
	}
}

// spin 旋转动画
func (i *Indicator) spin() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-i.done:
			return
		case <-ticker.C:
			i.render()
		}
	}
}

// render 渲染进度
func (i *Indicator) render() {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 清除当前行
	i.clear()

	// 显示旋转指示器和消息
	frame := i.spinner[i.index%len(i.spinner)]
	fmt.Fprintf(i.writer, "\r%s %s", frame, i.message)

	i.index++
}

// clear 清除当前行
func (i *Indicator) clear() {
	// 使用 ANSI 转义序列清除行
	fmt.Fprint(i.writer, "\r"+strings.Repeat(" ", 80)+"\r")
}

// Bar 进度条
type Bar struct {
	writer  io.Writer
	enabled bool
	total   int
	current int
	width   int
	mu      sync.Mutex
}

// NewBar 创建新的进度条
func NewBar(total int, enabled bool) *Bar {
	return &Bar{
		writer:  os.Stderr,
		enabled: enabled,
		total:   total,
		current: 0,
		width:   40,
	}
}

// Increment 增加进度
func (b *Bar) Increment() {
	if !b.enabled {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.current++
	b.render()
}

// Set 设置当前进度
func (b *Bar) Set(current int) {
	if !b.enabled {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.current = current
	b.render()
}

// Finish 完成进度条
func (b *Bar) Finish() {
	if !b.enabled {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.current = b.total
	b.render()
	fmt.Fprintln(b.writer)
}

// render 渲染进度条
func (b *Bar) render() {
	percent := float64(b.current) / float64(b.total)
	filled := int(percent * float64(b.width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", b.width-filled)
	fmt.Fprintf(b.writer, "\r[%s] %d/%d (%.1f%%)", bar, b.current, b.total, percent*100)
}

// Stats 统计信息
type Stats struct {
	Iterations int
	ToolCalls  int
	Tokens     int
}

// Tracker 进度跟踪器
type Tracker struct {
	indicator *Indicator
	stats     Stats
	mu        sync.Mutex
}

// NewTracker 创建新的进度跟踪器
func NewTracker(enabled bool) *Tracker {
	return &Tracker{
		indicator: New(enabled),
	}
}

// StartIteration 开始新的迭代
func (t *Tracker) StartIteration(n int) {
	t.mu.Lock()
	t.stats.Iterations = n
	t.mu.Unlock()

	t.indicator.Update(fmt.Sprintf("迭代 %d", n))
}

// RecordToolCall 记录工具调用
func (t *Tracker) RecordToolCall(toolName string) {
	t.mu.Lock()
	t.stats.ToolCalls++
	t.mu.Unlock()

	t.indicator.Update(fmt.Sprintf("迭代 %d - 调用工具: %s", t.stats.Iterations, toolName))
}

// RecordTokens 记录 token 使用
func (t *Tracker) RecordTokens(tokens int) {
	t.mu.Lock()
	t.stats.Tokens += tokens
	t.mu.Unlock()
}

// Start 开始跟踪
func (t *Tracker) Start() {
	t.indicator.Start("开始执行...")
}

// Stop 停止跟踪
func (t *Tracker) Stop() {
	t.indicator.Stop()
}

// GetStats 获取统计信息
func (t *Tracker) GetStats() Stats {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.stats
}

// PrintStats 打印统计信息
func (t *Tracker) PrintStats() {
	stats := t.GetStats()
	fmt.Printf("\n统计信息:\n")
	fmt.Printf("  迭代次数: %d\n", stats.Iterations)
	fmt.Printf("  工具调用: %d\n", stats.ToolCalls)
	fmt.Printf("  Token 使用: %d\n", stats.Tokens)
}
