package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/zhoucx/deepagents-go/internal/color"
	"github.com/zhoucx/deepagents-go/internal/logger"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// REPL 交互式命令行
type REPL struct {
	executor        *agent.Executor
	messages        []llm.Message
	reader          *bufio.Reader
	writer          io.Writer
	enableStreaming bool // 是否启用流式响应
}

// New 创建新的 REPL
func New(executor *agent.Executor) *REPL {
	return &REPL{
		executor:        executor,
		messages:        make([]llm.Message, 0),
		reader:          bufio.NewReader(os.Stdin),
		writer:          os.Stdout,
		enableStreaming: true, // 默认启用流式
	}
}

// SetStreaming 设置是否启用流式响应
func (r *REPL) SetStreaming(enabled bool) {
	r.enableStreaming = enabled
}

// Run 运行 REPL
func (r *REPL) Run() error {
	r.printWelcome()

	for {
		// 显示提示符
		fmt.Fprint(r.writer, "\n"+color.Cyan("❯")+" ")

		// 读取用户输入
		input, err := r.reader.ReadString('\n')
		if err != nil {
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

// handleCommand 处理特殊命令
func (r *REPL) handleCommand(input string) bool {
	switch strings.ToLower(input) {
	case "exit", "quit", "q":
		fmt.Fprintln(r.writer, "再见！")
		os.Exit(0)
		return true

	case "help", "h", "?":
		r.printHelp()
		return true

	case "clear", "cls":
		r.clearHistory()
		return true

	case "history":
		r.printHistory()
		return true

	case "stream":
		r.toggleStreaming()
		return true

	default:
		return false
	}
}

// toggleStreaming 切换流式响应模式
func (r *REPL) toggleStreaming() {
	r.enableStreaming = !r.enableStreaming
	status := "禁用"
	if r.enableStreaming {
		status = "启用"
	}
	fmt.Fprintf(r.writer, color.Yellow("流式响应已%s\n"), status)
	logger.Info("流式响应已%s", status)
}

// execute 执行用户输入
func (r *REPL) execute(input string) error {
	logger.Debug("执行用户输入: %s", input)

	// 记录当前消息数量，用于后续只显示新消息
	previousMessageCount := len(r.messages)

	// 添加用户消息到历史
	r.messages = append(r.messages, llm.Message{
		Role:    llm.RoleUser,
		Content: input,
	})

	ctx := context.Background()

	// 根据配置选择流式或非流式执行
	if r.enableStreaming {
		return r.executeStream(ctx)
	}
	return r.executeNonStream(ctx, previousMessageCount)
}

// executeStream 执行流式调用
func (r *REPL) executeStream(ctx context.Context) error {
	// 调用流式 Agent
	stream, err := r.executor.InvokeStream(ctx, &agent.InvokeInput{
		Messages: r.messages,
	})
	if err != nil {
		return err
	}

	fmt.Fprint(r.writer, "\n")

	var assistantContent string

	// 处理流式事件
	for event := range stream {
		switch event.Type {
		case agent.AgentEventTypeLLMStart:
			// 显示 Assistant 前缀
			fmt.Fprint(r.writer, color.Green("🤖 "))

		case agent.AgentEventTypeLLMText:
			// 实时显示 LLM 输出
			fmt.Fprint(r.writer, event.Content)
			assistantContent += event.Content

		case agent.AgentEventTypeLLMEnd:
			// LLM 生成结束，换行
			fmt.Fprintln(r.writer)

		case agent.AgentEventTypeEnd:
			// Agent 执行完成
			// 从元数据中获取最终消息
			if event.Metadata != nil {
				if messages, ok := event.Metadata["messages"].([]llm.Message); ok {
					r.messages = messages
				}
			}

		case agent.AgentEventTypeError:
			return event.Error
		}
	}

	logger.Debug("流式执行完成，消息数量: %d", len(r.messages))
	return nil
}

// executeNonStream 执行非流式调用
func (r *REPL) executeNonStream(ctx context.Context, previousMessageCount int) error {
	// 调用 Agent
	output, err := r.executor.Invoke(ctx, &agent.InvokeInput{
		Messages: r.messages,
	})

	if err != nil {
		return err
	}

	// 更新消息历史
	r.messages = output.Messages

	// 只打印新的 Agent 响应（从 previousMessageCount 之后的消息）
	for i := previousMessageCount; i < len(output.Messages); i++ {
		msg := output.Messages[i]
		if msg.Role == llm.RoleAssistant && msg.Content != "" {
			fmt.Fprintf(r.writer, "\n%s %s\n", color.Green("🤖"), msg.Content)
		}
	}

	logger.Debug("非流式执行完成，消息数量: %d", len(r.messages))
	return nil
}

// printWelcome 打印欢迎信息
func (r *REPL) printWelcome() {
	fmt.Fprintln(r.writer, color.Bold("=== Deep Agents 交互模式 ==="))
	fmt.Fprintln(r.writer, color.Gray("输入 'help' 查看帮助，'exit' 退出"))
	fmt.Fprintln(r.writer, "")
}

// printHelp 打印帮助信息
func (r *REPL) printHelp() {
	fmt.Fprintln(r.writer, color.Gray("\n可用命令:"))
	fmt.Fprintln(r.writer, color.Gray("  help, h, ?    - 显示此帮助信息"))
	fmt.Fprintln(r.writer, color.Gray("  exit, quit, q - 退出程序"))
	fmt.Fprintln(r.writer, color.Gray("  clear, cls    - 清除对话历史"))
	fmt.Fprintln(r.writer, color.Gray("  history       - 显示对话历史"))
	fmt.Fprintln(r.writer, color.Gray("  stream        - 切换流式响应模式"))
	fmt.Fprintln(r.writer, "")
	streamStatus := "禁用"
	if r.enableStreaming {
		streamStatus = "启用"
	}
	fmt.Fprintf(r.writer, color.Gray("当前流式响应: %s\n"), streamStatus)
	fmt.Fprintln(r.writer, "")
	fmt.Fprintln(r.writer, color.Gray("直接输入文本即可与 AI 对话"))
}

// clearHistory 清除对话历史
func (r *REPL) clearHistory() {
	r.messages = make([]llm.Message, 0)
	fmt.Fprintln(r.writer, color.Yellow("对话历史已清除"))
	logger.Info("对话历史已清除")
}

// printHistory 打印对话历史
func (r *REPL) printHistory() {
	if len(r.messages) == 0 {
		fmt.Fprintln(r.writer, "暂无对话历史")
		return
	}

	fmt.Fprintln(r.writer, "\n=== 对话历史 ===")
	for i, msg := range r.messages {
		role := "用户"
		if msg.Role == llm.RoleAssistant {
			role = "助手"
		}
		fmt.Fprintf(r.writer, "[%d] %s: %s\n", i+1, role, truncate(msg.Content, 100))
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
