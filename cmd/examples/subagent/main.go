package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zhoucx/deepagents-go/internal/config"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	fmt.Println("=== Deep Agents Go - SubAgent Middleware 示例 ===")
	fmt.Println()

	// 检查 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 ANTHROPIC_API_KEY 环境变量")
	}

	// 加载配置
	cfg := config.DefaultConfig()
	cfg.APIKey = apiKey
	cfg.BaseUrl = os.Getenv("ANTHROPIC_BASE_URL")
	cfg.WorkDir = "./workspace_subagent"

	// 创建 LLM 客户端
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model, cfg.BaseUrl)

	// 创建工具注册表和后端
	toolRegistry := tools.NewRegistry()
	fsBackend, err := backend.NewFilesystemBackend(cfg.WorkDir, true)
	if err != nil {
		log.Fatalf("创建文件系统后端失败: %v", err)
	}

	// 创建基础中间件
	fsMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)
	todoMiddleware := middleware.NewTodoMiddleware(fsBackend, toolRegistry)

	// 创建 SubAgent 中间件
	subAgentConfig := &middleware.SubAgentConfig{
		MaxDepth: 3, // 最大递归深度为3
	}

	subAgentMiddleware := middleware.NewSubAgentMiddleware(
		subAgentConfig,
		llmClient,
		toolRegistry,
		[]agent.Middleware{fsMiddleware, todoMiddleware}, // 子Agent可用的中间件
		"你是一个有用的AI助手，可以帮助用户完成各种任务。",
		cfg.MaxTokens,
		cfg.Temperature,
	)

	// 创建 Agent 配置
	agentConfig := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			fsMiddleware,
			todoMiddleware,
			subAgentMiddleware, // 添加 SubAgent 中间件
		},
		SystemPrompt:  "你是一个有用的AI助手。当遇到复杂任务时，你可以使用 delegate_to_subagent 工具将子任务委派给子Agent处理。",
		MaxIterations: cfg.MaxIterations,
		MaxTokens:     cfg.MaxTokens,
		Temperature:   cfg.Temperature,
		OnToolCall: func(toolName string, input map[string]any) {
			fmt.Printf("  🔧 调用工具: %s\n", toolName)
		},
		OnToolResult: func(toolName string, result string, isError bool) {
			if isError {
				fmt.Printf("     ❌ 错误\n")
			} else {
				fmt.Printf("     ✓ 完成\n")
			}
		},
	}

	// 创建 Agent
	executor := agent.NewRunnable(agentConfig)

	// 示例 1: 简单任务（不需要子Agent）
	fmt.Println("=== 示例 1: 简单任务 ===")
	fmt.Println("任务: 创建一个简单的文件")
	fmt.Println()

	output1, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "请创建一个文件 /hello.txt，内容为 'Hello from Deep Agents!'",
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output1.Messages) > 0 {
			lastMsg := output1.Messages[len(output1.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("\n✅ Agent 响应: %s\n", truncate(lastMsg.Content, 200))
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 示例 2: 复杂任务（建议使用子Agent）
	fmt.Println("=== 示例 2: 复杂任务（使用子Agent） ===")
	fmt.Println("任务: 创建一个项目结构，包含多个文件和目录")
	fmt.Println()

	output2, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role: llm.RoleUser,
				Content: `请创建一个简单的Go项目结构：
1. 创建 /project/main.go 文件，包含一个简单的 Hello World 程序
2. 创建 /project/README.md 文件，包含项目说明
3. 创建 /project/go.mod 文件，包含模块定义

你可以使用 delegate_to_subagent 工具将每个文件的创建委派给子Agent处理。`,
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output2.Messages) > 0 {
			lastMsg := output2.Messages[len(output2.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("\n✅ Agent 响应: %s\n", truncate(lastMsg.Content, 300))
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 示例 3: 带上下文的子Agent
	fmt.Println("=== 示例 3: 带上下文的子Agent ===")
	fmt.Println("任务: 分析文件并生成报告")
	fmt.Println()

	output3, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role: llm.RoleUser,
				Content: `请执行以下任务：
1. 读取 /hello.txt 文件的内容
2. 使用 delegate_to_subagent 工具，将"分析文件内容并生成报告"的任务委派给子Agent
3. 将子Agent的分析结果保存到 /report.txt 文件

在委派给子Agent时，请将文件内容作为上下文传递。`,
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output3.Messages) > 0 {
			lastMsg := output3.Messages[len(output3.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("\n✅ Agent 响应: %s\n", truncate(lastMsg.Content, 300))
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 打印工作目录
	fmt.Printf("工作目录: %s\n", cfg.WorkDir)
	fmt.Println()
	fmt.Println("=== 示例完成 ===")
	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("- SubAgent 有独立的状态和上下文")
	fmt.Println("- 可以通过 context 参数传递上下文信息")
	fmt.Println("- 最大递归深度为 3，防止无限递归")
	fmt.Println("- 子Agent 不会再创建子Agent（避免无限递归）")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
