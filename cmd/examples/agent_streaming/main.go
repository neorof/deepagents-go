package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zhoucx/deepagents-go/internal/progress"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	fmt.Println("=== Agent 流式执行示例 ===")
	fmt.Println()

	// 获取 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	// 创建临时工作目录
	workDir := "./agent_stream_workspace"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		log.Fatalf("Failed to create workspace: %v", err)
	}
	defer os.RemoveAll(workDir)

	// 演示1: 基本流式执行
	fmt.Println("演示1: Agent 基本流式执行")
	fmt.Println("-----------------------------------")
	demoBasicStream(apiKey, workDir)

	// 演示2: 带工具调用的流式执行
	fmt.Println("\n演示2: 带工具调用的流式执行")
	fmt.Println("-----------------------------------")
	demoStreamWithTools(apiKey, workDir)

	fmt.Println("\n=== 演示完成 ===")
}

// demoBasicStream 演示基本流式执行
func demoBasicStream(apiKey, _ /* workDir */ string) {
	// 创建 LLM 客户端
	llmClient := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", "")

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建 Agent
	config := &agent.Config{
		LLMClient:     llmClient,
		ToolRegistry:  toolRegistry,
		SystemPrompt:  "你是一个有用的助手。",
		MaxIterations: 5,
	}
	executor := agent.NewRunnable(config)

	// 准备输入
	input := &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "用一句话介绍什么是 Go 语言，然后列出3个优势。",
			},
		},
	}

	// 调用流式 API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := executor.InvokeStream(ctx, input)
	if err != nil {
		log.Fatalf("InvokeStream failed: %v", err)
	}

	fmt.Print("Agent: ")

	// 处理流式事件
	for event := range stream {
		switch event.Type {
		case agent.AgentEventTypeStart:
			fmt.Print("[开始执行]\n")

		case agent.AgentEventTypeLLMStart:
			fmt.Printf("[迭代 %d: LLM 开始]\n", event.Iteration)

		case agent.AgentEventTypeLLMText:
			// 实时显示 LLM 输出
			fmt.Print(event.Content)

		case agent.AgentEventTypeLLMEnd:
			fmt.Print("\n[LLM 完成]\n")

		case agent.AgentEventTypeEnd:
			fmt.Println("\n[Agent 执行完成]")

		case agent.AgentEventTypeError:
			fmt.Printf("\n[错误: %v]\n", event.Error)
		}
	}
}

// demoStreamWithTools 演示带工具调用的流式执行
func demoStreamWithTools(apiKey, workDir string) {
	// 创建 LLM 客户端
	llmClient := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", "")

	// 创建工具注册表和后端
	toolRegistry := tools.NewRegistry()
	fsBackend, err := backend.NewFilesystemBackend(workDir, true)
	if err != nil {
		log.Fatalf("Failed to create backend: %v", err)
	}

	// 创建文件系统中间件
	fsMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)

	// 创建 Agent
	config := &agent.Config{
		LLMClient:     llmClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []agent.Middleware{fsMiddleware},
		SystemPrompt:  `你是一个文件管理助手。你可以创建、读取和编辑文件。`,
		MaxIterations: 10,
	}
	executor := agent.NewRunnable(config)

	// 准备输入
	input := &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "创建一个文件 /hello.txt，内容为 'Hello from streaming agent!'，然后读取它的内容。",
			},
		},
	}

	// 调用流式 API
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := executor.InvokeStream(ctx, input)
	if err != nil {
		log.Fatalf("InvokeStream failed: %v", err)
	}

	fmt.Println("Agent 执行过程:")
	fmt.Println()

	// 创建进度跟踪器
	tracker := progress.NewTracker(true)
	tracker.Start()
	defer tracker.Stop()

	var isFirstText bool

	// 处理流式事件
	for event := range stream {
		switch event.Type {
		case agent.AgentEventTypeStart:
			fmt.Println("✓ Agent 开始执行")

		case agent.AgentEventTypeLLMStart:
			tracker.StartIteration(event.Iteration)
			fmt.Printf("\n→ 迭代 %d: LLM 开始生成...\n", event.Iteration)
			isFirstText = true

		case agent.AgentEventTypeLLMText:
			// 第一次输出文本时，停止进度指示器
			if isFirstText {
				tracker.Stop()
				isFirstText = false
			}
			// 实时显示 LLM 输出
			fmt.Print(event.Content)

		case agent.AgentEventTypeLLMToolCall:
			// 显示工具调用
			if event.ToolCall != nil {
				fmt.Printf("\n\n→ 调用工具: %s\n", event.ToolCall.Name)
				fmt.Printf("  参数: %v\n", event.ToolCall.Input)
			}

		case agent.AgentEventTypeLLMEnd:
			fmt.Println()

		case agent.AgentEventTypeToolStart:
			if event.ToolCall != nil {
				tracker.RecordToolCall(event.ToolCall.Name)
				fmt.Printf("→ 执行工具: %s...\n", event.ToolCall.Name)
			}

		case agent.AgentEventTypeToolResult:
			if event.ToolResult != nil {
				if event.ToolResult.IsError {
					fmt.Printf("✗ 工具执行失败: %s\n", event.ToolResult.Content)
				} else {
					fmt.Printf("✓ 工具执行成功\n")
					fmt.Printf("  结果: %s\n", event.ToolResult.Content)
				}
			}

		case agent.AgentEventTypeIterationEnd:
			fmt.Printf("\n✓ 迭代 %d 完成\n", event.Iteration)
			// 重新启动进度指示器（如果还有下一轮）
			tracker.Start()

		case agent.AgentEventTypeEnd:
			tracker.Stop()
			fmt.Println("\n✓ Agent 执行完成")
			// 显示统计信息
			tracker.PrintStats()

		case agent.AgentEventTypeError:
			tracker.Stop()
			fmt.Printf("\n✗ 错误: %v\n", event.Error)
		}
	}

	fmt.Println()
	fmt.Println("优势:")
	fmt.Println("  - 实时显示执行进度")
	fmt.Println("  - 可见每个步骤的详细信息")
	fmt.Println("  - 更好的用户体验")
	fmt.Println("  - 可以提前终止执行")
	fmt.Println("  - 自动跟踪迭代次数和工具调用")
}
