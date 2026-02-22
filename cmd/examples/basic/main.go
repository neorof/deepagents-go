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

	// 创建 LLM 客户端
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	llmClient := llm.NewAnthropicClient(apiKey, "", baseURL)

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建后端
	stateBackend := backend.NewStateBackend()

	// 创建中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(stateBackend, toolRegistry)

	// 创建 Agent 配置
	config := &agent.Config{
		LLMClient:     llmClient,
		ToolRegistry:  toolRegistry,
		Middlewares:   []agent.Middleware{filesystemMiddleware},
		SystemPrompt:  "你是一个有用的 AI 助手，可以帮助用户管理文件和执行任务。",
		MaxIterations: 25,
		MaxTokens:     4096,
		Temperature:   0.7,
	}

	// 创建 Agent
	executor := agent.NewRunnable(config)

	// 执行示例任务
	ctx := context.Background()
	input := &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "创建一个文件 /test.txt，内容为 'Hello, Deep Agents!'",
			},
		},
	}

	fmt.Println("执行任务：创建文件 /test.txt")
	output, err := executor.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}

	// 打印结果
	fmt.Println("\n=== 执行结果 ===")
	for _, msg := range output.Messages {
		fmt.Printf("[%s] %s\n", msg.Role, msg.Content)
	}

	// 验证文件是否创建
	fmt.Println("\n=== 验证文件 ===")
	content, err := stateBackend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}
	fmt.Printf("文件内容: %s\n", content)
}
