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

	// 创建文件系统后端
	fsBackend, err := backend.NewFilesystemBackend(workDir, true)
	if err != nil {
		log.Fatalf("Failed to create filesystem backend: %v", err)
	}

	// 创建文件系统中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)

	// 创建 Web 中间件
	webConfig := middleware.DefaultWebConfig()
	webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

	// 创建 Agent 配置
	agentConfig := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares:  []agent.Middleware{filesystemMiddleware, webMiddleware},
		SystemPrompt: `你是一个有用的 AI 助手，可以帮助用户搜索网络内容和管理文件。

可用工具：
- web_search: 搜索网络内容并返回结果摘要
- web_fetch: 获取指定 URL 的内容并转换为 Markdown
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
	executor := agent.NewRunnable(agentConfig)

	// 执行示例任务
	ctx := context.Background()

	fmt.Println("=== Deep Agents Go - Web 工具示例 ===")

	// 任务 1：搜索网络内容
	fmt.Println("\n任务 1：搜索 'Go 语言 2026 新特性'")
	output1, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "搜索 'Go 语言 2026 新特性'，并总结前 3 条结果。",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 1 失败: %v", err)
	}
	printMessages(output1.Messages)

	// 任务 2：获取网页内容
	fmt.Println("\n任务 2：获取 https://go.dev 首页的内容摘要")
	output2, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "获取 https://go.dev 首页的内容，并提取主要信息（标题、简介等）。",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 2 失败: %v", err)
	}
	printMessages(output2.Messages)

	// 任务 3：搜索并保存结果
	fmt.Println("\n任务 3：搜索最新的 AI 新闻，并保存结果到文件")
	output3, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "搜索 '2026 AI 新闻'，获取前 5 条结果，并将结果保存到 /ai_news.md 文件中。文件格式为 Markdown，包含标题、链接和摘要。",
			},
		},
	})
	if err != nil {
		log.Fatalf("任务 3 失败: %v", err)
	}
	printMessages(output3.Messages)

	// 任务 4：获取网页并保存
	fmt.Println("\n任务 4：获取一篇技术文章并保存")
	output4, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "搜索 'Go 并发编程最佳实践'，选择第一个结果，获取该网页的内容，并保存到 /go_concurrency.md 文件中。",
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
