package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	fmt.Println("=== Deep Agents Go - OpenAI 客户端示例 ===")
	fmt.Println()

	// 检查 API Key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 OPENAI_API_KEY 环境变量")
	}

	// 创建 OpenAI 客户端
	// 可以使用不同的模型：
	// - openai.GPT4o (默认)
	// - openai.GPT4oMini
	// - openai.GPT4TurboPreview
	// - openai.GPT35Turbo
	baseURL := os.Getenv("OPENAI_BASE_URL") // 可选：自定义 API 端点
	llmClient := llm.NewOpenAIClient(apiKey, openai.GPT4oMini, baseURL)

	fmt.Printf("✓ 使用模型: %s\n", openai.GPT4oMini)
	if baseURL != "" {
		fmt.Printf("✓ 自定义端点: %s\n", baseURL)
	}
	fmt.Println()

	// 创建工具注册表和后端
	toolRegistry := tools.NewRegistry()
	workDir := "./workspace_openai"
	fsBackend, err := backend.NewFilesystemBackend(workDir, true)
	if err != nil {
		log.Fatalf("创建文件系统后端失败: %v", err)
	}

	// 创建中间件
	fsMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)
	todoMiddleware := middleware.NewTodoMiddleware(fsBackend, toolRegistry)

	// 创建 Agent 配置
	agentConfig := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			fsMiddleware,
			todoMiddleware,
		},
		SystemPrompt:  "你是一个有用的AI助手，可以帮助用户完成各种任务。",
		MaxIterations: 25,
		MaxTokens:     4096,
		Temperature:   0.7,
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

	// 示例 1: 简单对话
	fmt.Println("=== 示例 1: 简单对话 ===")
	fmt.Println("任务: 打个招呼")
	fmt.Println()

	output1, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "你好！请简单介绍一下你自己。",
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output1.Messages) > 0 {
			lastMsg := output1.Messages[len(output1.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("✅ Agent 响应:\n%s\n", lastMsg.Content)
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 示例 2: 使用工具
	fmt.Println("=== 示例 2: 使用文件系统工具 ===")
	fmt.Println("任务: 创建文件并读取")
	fmt.Println()

	output2, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role: llm.RoleUser,
				Content: `请执行以下任务：
1. 创建一个文件 /greeting.txt，内容为 "Hello from OpenAI!"
2. 读取这个文件的内容
3. 告诉我文件的内容`,
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output2.Messages) > 0 {
			lastMsg := output2.Messages[len(output2.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("✅ Agent 响应:\n%s\n", truncate(lastMsg.Content, 300))
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 示例 3: 复杂任务
	fmt.Println("=== 示例 3: 复杂任务（创建项目结构） ===")
	fmt.Println("任务: 创建一个简单的 Python 项目")
	fmt.Println()

	output3, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role: llm.RoleUser,
				Content: `请创建一个简单的 Python 项目结构：
1. 创建 /python_project/main.py 文件，包含一个简单的 Hello World 程序
2. 创建 /python_project/README.md 文件，包含项目说明
3. 创建 /python_project/requirements.txt 文件（空文件即可）
4. 列出创建的所有文件`,
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output3.Messages) > 0 {
			lastMsg := output3.Messages[len(output3.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("✅ Agent 响应:\n%s\n", truncate(lastMsg.Content, 400))
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 示例 4: 使用 Todo 中间件
	fmt.Println("=== 示例 4: 任务规划 ===")
	fmt.Println("任务: 规划一个开发任务")
	fmt.Println()

	output4, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role: llm.RoleUser,
				Content: `请帮我规划一个"开发博客系统"的任务，创建一个 Todo 列表，包含以下步骤：
1. 设计数据库结构
2. 实现用户认证
3. 实现文章 CRUD
4. 实现评论功能
5. 部署上线

使用 write_todos 工具创建这个任务列表。`,
			},
		},
	})

	if err != nil {
		log.Printf("❌ 任务执行失败: %v\n", err)
	} else {
		if len(output4.Messages) > 0 {
			lastMsg := output4.Messages[len(output4.Messages)-1]
			if lastMsg.Role == llm.RoleAssistant {
				fmt.Printf("✅ Agent 响应:\n%s\n", truncate(lastMsg.Content, 300))
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// 打印工作目录
	fmt.Printf("工作目录: %s\n", workDir)
	fmt.Println()
	fmt.Println("=== 示例完成 ===")
	fmt.Println()
	fmt.Println("💡 提示:")
	fmt.Println("- OpenAI 客户端与 Anthropic 客户端使用相同的接口")
	fmt.Println("- 可以通过修改 model 参数使用不同的 OpenAI 模型")
	fmt.Println("- 支持自定义 API 端点（通过 OPENAI_BASE_URL 环境变量）")
	fmt.Println("- 所有工具和中间件都可以正常使用")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
