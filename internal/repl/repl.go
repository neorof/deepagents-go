package repl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/zhoucx/deepagents-go/internal/color"
	"github.com/zhoucx/deepagents-go/internal/logger"
	"github.com/zhoucx/deepagents-go/internal/progress"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/agentkit"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"golang.org/x/term"
)

// BannerInfo 启动 Banner 显示的配置信息
type BannerInfo struct {
	Model   string
	WorkDir string
	Version string
}

// REPL 交互式命令行
type REPL struct {
	executor     *agent.Runnable
	messages     []llm.Message
	rl           *readline.Instance
	writer       io.Writer
	sessionID    string                   // 会话 ID
	sessionStore *middleware.SessionStore // 会话存储
	bannerInfo   *BannerInfo              // 启动 Banner 信息
}

// New 创建新的 REPL
func New(builder *agentkit.AgentBuilder, sessionId string, banner *BannerInfo) *REPL {
	// 构建历史文件路径 ~/.deepagents/history
	historyFile := ""
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, ".deepagents")
		_ = os.MkdirAll(dir, 0755)
		historyFile = filepath.Join(dir, "history")
	}

	prompt := color.Cyan("❯") + " "
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            prompt,
		HistoryFile:       historyFile,
		HistoryLimit:      1000,
		InterruptPrompt:   "\n",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		// fallback: 如果 readline 初始化失败，使用无历史文件的配置
		rl, _ = readline.NewEx(&readline.Config{
			Prompt:          prompt,
			InterruptPrompt: "\n",
			EOFPrompt:       "exit",
		})
	}

	return &REPL{
		executor:     builder.Runnable,
		messages:     make([]llm.Message, 0),
		rl:           rl,
		writer:       os.Stdout,
		sessionID:    sessionId,
		sessionStore: builder.SessionStore,
		bannerInfo:   banner,
	}
}

// Run 运行 REPL
func (r *REPL) Run() error {
	defer r.rl.Close()

	r.printWelcome()

	for {
		// 上方横线
		fmt.Fprintf(r.writer, "\n%s\n", hrLine())
	readInput:
		input, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl+C：回到上一行并清除，重新显示提示符
				fmt.Fprint(r.writer, "\033[A\r\033[K")
				goto readInput
			}
			if err == io.EOF {
				fmt.Fprintln(r.writer, "\n再见！")
				return nil
			}
			return fmt.Errorf("读取输入失败: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 下方横线
		fmt.Fprintln(r.writer, hrLine())

		// 处理特殊命令
		if r.handleCommand(input) {
			continue
		}

		// 执行用户输入
		if err := r.execute(input); err != nil {
			logger.Error("执行失败: %v", err)
			fmt.Fprintf(r.writer, color.Red("❌ 错误: %v\n"), err)
		}
	}
}

// handleCommand 处理特殊命令（以 / 开头）
func (r *REPL) handleCommand(input string) bool {
	cmd := strings.ToLower(input)
	if !strings.HasPrefix(cmd, "/") {
		return false
	}

	// 解析命令和参数
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	switch cmdName {
	case "exit", "quit", "q":
		fmt.Fprintln(r.writer, "再见！")
		os.Exit(0)
		return true

	case "help", "h", "?":
		r.printHelp()
		return true

	case "clear":
		r.clear()
		return true

	case "history":
		r.printHistory()
		return true

	case "sessions":
		r.handleSessions()
		return true

	case "resume":
		if len(args) == 0 {
			fmt.Fprintln(r.writer, color.Red("❌ 用法: /resume <session_id>"))
			return true
		}
		r.handleResume(args[0])
		return true

	default:
		fmt.Fprintf(r.writer, color.Red("未知命令: %s\n"), input)
		fmt.Fprintln(r.writer, color.Gray("输入 /help 查看可用命令"))
		return true
	}
}

// execute 执行用户输入
func (r *REPL) execute(input string) error {
	logger.Debug("执行用户输入: %s", input)

	// 添加用户消息到历史
	r.messages = append(r.messages, llm.Message{
		Role:    llm.RoleUser,
		Content: input,
	})

	return r.executeStream(context.Background())
}

// executeStream 执行流式调用
func (r *REPL) executeStream(ctx context.Context) error {
	// 创建进度跟踪器
	tracker := progress.NewTracker(true)
	tracker.Start()
	defer tracker.Stop()

	// 创建输入，包含 session_id
	input := &agent.InvokeInput{
		Messages: r.messages,
	}
	if r.sessionID != "" {
		input.Metadata = map[string]any{
			"session_id": r.sessionID,
		}
	}

	// 调用流式 Agent
	stream, err := r.executor.InvokeStream(ctx, input)
	if err != nil {
		return err
	}

	fmt.Fprint(r.writer, "\n")

	var assistantContent string
	var isFirstText bool

	// 处理流式事件
	for event := range stream {
		switch event.Type {
		case agent.AgentEventTypeLLMStart:
			// 更新进度：开始新的迭代
			tracker.StartIteration(event.Iteration)
			isFirstText = true

		case agent.AgentEventTypeLLMText:
			// 第一次输出文本时，停止进度指示器并显示 Assistant 前缀
			if isFirstText {
				tracker.Stop()
				fmt.Fprint(r.writer, color.Green("Assistant: "))
				isFirstText = false
			}
			// 实时显示 LLM 输出
			fmt.Fprint(r.writer, event.Content)
			assistantContent += event.Content

		case agent.AgentEventTypeLLMEnd:
			// LLM 生成结束，换行
			fmt.Fprintln(r.writer)

		case agent.AgentEventTypeToolStart:
			// 记录工具调用
			if event.ToolCall != nil {
				tracker.RecordToolCall(event.ToolCall.Name)
			}

		case agent.AgentEventTypeIterationEnd:
			// 迭代结束，重新启动进度指示器（如果还有下一轮）
			tracker.Start()

		case agent.AgentEventTypeEnd:
			// Agent 执行完成
			tracker.Stop()
			// 从元数据中获取最终消息
			if event.Metadata != nil {
				if messages, ok := event.Metadata["messages"].([]llm.Message); ok {
					r.messages = messages
				}
			}
			// 显示统计信息
			tracker.PrintStats()

		case agent.AgentEventTypeError:
			tracker.Stop()
			return event.Error
		}
	}

	logger.Debug("流式执行完成，消息数量: %d", len(r.messages))
	return nil
}

// printWelcome 打印欢迎信息
func (r *REPL) printWelcome() {
	const boxWidth = 40 // 内容区宽度（不含左右边距）

	// 准备版本号
	version := "v1.0"
	if r.bannerInfo != nil && r.bannerInfo.Version != "" {
		version = r.bannerInfo.Version
	}

	// 品牌名和副标题（居中显示）
	title := "Deep Agents Go"
	subtitle := "AI Agent Framework"

	// title 居中计算（version 作为小标记固定在右侧）
	titleLen := len(title)
	versionLen := len(version)
	titleLeftPad := (boxWidth - titleLen) / 2
	titleRightPad := boxWidth - titleLen - titleLeftPad - versionLen - 1 // -1 是 version 前的空格

	// subtitle 居中计算
	subtitleLen := len(subtitle)
	subtitleLeftPad := (boxWidth - subtitleLen) / 2
	subtitleRightPad := boxWidth - subtitleLen - subtitleLeftPad

	// 绘制边框（统一格式：│ + 内容 + │，左右各有1个空格边距）
	fmt.Fprintf(r.writer, "  %s\n", color.Gray("╭"+strings.Repeat("─", boxWidth+2)+"╮"))
	fmt.Fprintf(r.writer, "  %s %s%s%s %s%s\n",
		color.Gray("│"),
		strings.Repeat(" ", titleLeftPad),
		color.BoldCyan(title),
		strings.Repeat(" ", titleRightPad),
		color.Gray(version),
		color.Gray(" │"),
	)
	fmt.Fprintf(r.writer, "  %s %s%s%s%s\n",
		color.Gray("│"),
		strings.Repeat(" ", subtitleLeftPad),
		color.Gray(subtitle),
		strings.Repeat(" ", subtitleRightPad),
		color.Gray(" │"),
	)
	fmt.Fprintf(r.writer, "  %s\n", color.Gray("╰"+strings.Repeat("─", boxWidth+2)+"╯"))

	// 配置信息面板
	if r.bannerInfo != nil {
		if r.bannerInfo.Model != "" {
			fmt.Fprintf(r.writer, "    %s  %s\n", color.Gray("Model  "), color.Cyan(r.bannerInfo.Model))
		}
		if r.bannerInfo.WorkDir != "" {
			fmt.Fprintf(r.writer, "    %s  %s\n", color.Gray("WorkDir"), color.Green(r.bannerInfo.WorkDir))
		}
	}
	if r.sessionID != "" {
		shortID := r.sessionID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		fmt.Fprintf(r.writer, "    %s  %s\n", color.Gray("Session"), color.Yellow(shortID))
	}

	fmt.Fprintln(r.writer)
	fmt.Fprintf(r.writer, "    %s\n", color.Gray("输入 /help 查看帮助，/exit 退出"))
	fmt.Fprintln(r.writer)
}

// printHelp 打印帮助信息
func (r *REPL) printHelp() {
	fmt.Fprintln(r.writer, color.Gray("\n可用命令:"))
	fmt.Fprintln(r.writer, color.Gray("  /help, /h, /?       - 显示此帮助信息"))
	fmt.Fprintln(r.writer, color.Gray("  /exit, /quit, /q    - 退出程序"))
	fmt.Fprintln(r.writer, color.Gray("  /clear              - 清除对话历史和记忆"))
	fmt.Fprintln(r.writer, color.Gray("  /history            - 显示对话历史"))
	fmt.Fprintln(r.writer, color.Gray("  /sessions           - 列出可用的历史会话"))
	fmt.Fprintln(r.writer, color.Gray("  /resume <id>        - 恢复指定的会话（支持前缀匹配）"))
	fmt.Fprintln(r.writer, "")
	fmt.Fprintln(r.writer, color.Gray("直接输入文本即可与 AI 对话"))
}

// handleSessions 列出可用的会话
func (r *REPL) handleSessions() {
	ctx := context.Background()

	fmt.Fprintln(r.writer, "\n=== 可用会话 ===")

	sessions, err := r.sessionStore.ListSessions(ctx, 30)
	if err != nil || len(sessions) == 0 {
		fmt.Fprintln(r.writer, "没有找到历史会话")
		return
	}

	// 按 DateDir 分组
	grouped := make(map[string][]*middleware.SessionInfo)
	for _, info := range sessions {
		grouped[info.DateDir] = append(grouped[info.DateDir], info)
	}

	// 按日期排序并显示
	dates := make([]string, 0, len(grouped))
	for date := range grouped {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	for _, date := range dates {
		fmt.Fprintf(r.writer, "%s:\n", date)
		for _, info := range grouped[date] {
			duration := info.Duration()
			durationStr := fmt.Sprintf("%dm", int(duration.Minutes()))
			if duration.Hours() >= 1 {
				durationStr = fmt.Sprintf("%.1fh", duration.Hours())
			}

			fmt.Fprintf(r.writer, "  %s... - %s (%d 条消息, %s)\n",
				info.ShortID(),
				info.StartTime.Format("15:04:05"),
				info.MessageCount,
				durationStr,
			)
		}
		fmt.Fprintln(r.writer)
	}

	fmt.Fprintln(r.writer, "使用 /resume <session_id> 恢复会话")
}

// handleResume 恢复指定的会话
func (r *REPL) handleResume(sessionIDPrefix string) {
	ctx := context.Background()

	// 通过 SessionStore 查找匹配的会话
	matchedInfo, err := r.sessionStore.FindSession(ctx, sessionIDPrefix, 30)
	if err != nil {
		fmt.Fprintf(r.writer, color.Red("❌ %v\n"), err)
		return
	}

	// 恢复会话
	if err := r.resumeSession(ctx, matchedInfo); err != nil {
		fmt.Fprintf(r.writer, color.Red("❌ 恢复会话失败: %v\n"), err)
		return
	}

	fmt.Fprintf(r.writer, color.Green("✅ 已恢复会话 %s (%d 条消息)\n"),
		matchedInfo.ShortID(),
		matchedInfo.MessageCount,
	)
}

// resumeSession 执行会话恢复
func (r *REPL) resumeSession(ctx context.Context, info *middleware.SessionInfo) error {
	// 1. 更新 REPL 的 sessionID
	r.sessionID = info.SessionID

	// 2. 通知 MemoryMiddleware 进入恢复模式
	for _, mw := range r.executor.GetMiddlewares() {
		if memMw, ok := mw.(interface {
			SetResumeMode(string)
			LoadSessionMessages(context.Context) ([]llm.Message, error)
		}); ok {
			memMw.SetResumeMode(info.SessionID)

			// 3. 加载历史消息
			messages, err := memMw.LoadSessionMessages(ctx)
			if err != nil {
				return fmt.Errorf("加载历史消息失败: %w", err)
			}

			// 4. 更新 REPL 的消息列表
			r.messages = messages

			return nil
		}
	}

	return fmt.Errorf("MemoryMiddleware not found")
}

// sessionClearer 会话清除接口，避免直接依赖 pkg/middleware
type sessionClearer interface {
	ClearSession()
}

// clear 清除对话历史和会话记录
func (r *REPL) clear() {
	// 清除对话历史
	r.messages = make([]llm.Message, 0)
	fmt.Fprintln(r.writer, color.Yellow("对话历史已清除"))

	// 清除会话记录
	for _, mw := range r.executor.GetMiddlewares() {
		if mw.Name() == "memory" {
			if sc, ok := mw.(sessionClearer); ok {
				sc.ClearSession()
				fmt.Fprintln(r.writer, color.Yellow("会话记录已清除"))
			}
		}
	}
}

// printHistory 打印对话历史
func (r *REPL) printHistory() {
	if len(r.messages) == 0 {
		fmt.Fprintln(r.writer, "暂无对话历史")
		return
	}

	fmt.Fprintln(r.writer, "\n=== 对话历史 ===")
	for i, msg := range r.messages {
		switch {
		case msg.Role == llm.RoleAssistant:
			if msg.Content != "" {
				fmt.Fprintf(r.writer, "[%d] 助手: %s\n", i+1, truncate(msg.Content, 100))
			}
			for _, tc := range msg.ToolCalls {
				inputJSON, _ := json.Marshal(tc.Input)
				fmt.Fprintf(r.writer, "[%d]   ↳ 工具调用 [%s]: %s\n", i+1, tc.Name, truncate(string(inputJSON), 80))
			}

		case msg.Role == llm.RoleUser && len(msg.ToolResults) > 0:
			for _, tr := range msg.ToolResults {
				status := "✓"
				if tr.IsError {
					status = "✗"
				}
				fmt.Fprintf(r.writer, "[%d] 工具结果 %s [%s]: %s\n", i+1, status, tr.ToolCallID[:min(8, len(tr.ToolCallID))], truncate(tr.Content, 80))
			}

		default:
			fmt.Fprintf(r.writer, "[%d] 用户: %s\n", i+1, truncate(msg.Content, 100))
		}
	}
	fmt.Fprintln(r.writer, "")
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// hrLine 返回终端宽度的灰色水平线
func hrLine() string {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		width = w
	}
	return color.Gray(strings.Repeat("─", width))
}
