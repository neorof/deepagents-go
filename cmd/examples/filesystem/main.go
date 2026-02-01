package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	// 获取 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	// 创建临时工作目录
	workDir := "./workspace"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		log.Fatalf("Failed to create workspace: %v", err)
	}

	// 创建 LLM 客户端
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	llmClient := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", baseURL)

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建文件系统后端（使用真实文件系统）
	fsBackend, err := backend.NewFilesystemBackend(workDir, true)
	if err != nil {
		log.Fatalf("Failed to create filesystem backend: %v", err)
	}

	// 创建文件系统中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)

	// 创建 Agent 配置
	config := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares:  []agent.Middleware{filesystemMiddleware},
		SystemPrompt: `你是一个有用的 AI 助手，可以帮助用户管理文件和执行任务。

可用工具：
- ls: 列出目录下的文件
- read_file: 读取文件内容
- write_file: 写入文件
- edit_file: 编辑文件（字符串替换）
- grep: 搜索文件内容
- glob: 查找匹配的文件

请根据用户的需求使用这些工具完成任务。`,
		MaxIterations: 25,
		MaxTokens:     4096,
		Temperature:   0.7,
	}

	// 创建 Agent
	executor := agent.NewExecutor(config)

	// 执行示例任务
	ctx := context.Background()

	fmt.Println("=== Deep Agents Go - 文件系统示例 ===")

	// 任务 1：创建文件
	fmt.Println("任务 1：创建一个 README.md 文件")
	output1, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "创建一个 README.md 文件，内容包括项目标题 'My Project' 和简短描述 'This is a test project.'",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 1 失败: %v", err)
	}
	printMessages(output1.Messages)

	// 任务 2：读取并修改文件
	fmt.Println("\n任务 2：读取 README.md 并添加一个新章节")
	output2, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "读取 README.md 文件，然后在末尾添加一个 '## Features' 章节，列出 3 个功能特性。",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 2 失败: %v", err)
	}
	printMessages(output2.Messages)

	// 任务 3：创建多个文件
	fmt.Println("\n任务 3：创建项目结构")
	output3, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "创建以下文件结构：\n- src/main.go (包含 Hello World 程序)\n- src/utils.go (包含一个简单的工具函数)\n- tests/main_test.go (包含一个测试用例)",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 3 失败: %v", err)
	}
	printMessages(output3.Messages)

	// 任务 4：搜索文件
	fmt.Println("\n任务 4：搜索包含 'Hello' 的文件")
	output4, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "使用 grep 搜索所有包含 'Hello' 的文件，并列出结果。",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 4 失败: %v", err)
	}
	printMessages(output4.Messages)

	fmt.Println("\n=== 所有任务完成 ===")
	fmt.Printf("\n工作目录: %s\n", workDir)
	fmt.Println("你可以查看生成的文件。")
}

func printMessages(messages []llm.Message) {
	for _, msg := range messages {
		if msg.Role == llm.RoleAssistant && msg.Content != "" {
			fmt.Printf("\n[助手] %s\n", msg.Content)
		}
	}
}
