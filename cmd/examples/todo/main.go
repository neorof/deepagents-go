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
	workDir := "./workspace_todo"
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

	// 创建中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)
	todoMiddleware := middleware.NewTodoMiddleware(fsBackend, toolRegistry)

	// 创建 Agent 配置
	config := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			filesystemMiddleware,
			todoMiddleware,
		},
		SystemPrompt: `你是一个有用的 AI 助手，擅长任务规划和执行。

可用工具：
- write_todos: 创建或更新 Todo 列表
- ls, read_file, write_file, edit_file, grep, glob: 文件系统操作

工作流程：
1. 收到复杂任务时，先使用 write_todos 创建任务计划
2. 逐步执行每个任务
3. 完成任务后更新 Todo 状态
4. 所有任务完成后给出总结`,
		MaxIterations: 30,
		MaxTokens:     4096,
		Temperature:   0.7,
	}

	// 创建 Agent
	executor := agent.NewExecutor(config)

	// 执行示例任务
	ctx := context.Background()

	fmt.Println("=== Deep Agents Go - Todo 管理示例 ===")
	fmt.Println()

	// 复杂任务：创建一个 Go 项目
	task := `创建一个简单的 Go Web 服务器项目，包括：
1. main.go - HTTP 服务器主程序
2. handlers.go - 请求处理函数
3. README.md - 项目文档
4. go.mod - Go 模块文件

请先制定任务计划，然后逐步执行。`

	fmt.Printf("任务：%s\n\n", task)

	output, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: task,
			},
		},
	})
	if err != nil {
		log.Fatalf("任务执行失败: %v", err)
	}

	// 打印对话历史
	fmt.Println("=== 执行过程 ===")
	for i, msg := range output.Messages {
		if msg.Role == llm.RoleAssistant && msg.Content != "" {
			fmt.Printf("\n[步骤 %d] %s\n", i+1, msg.Content)
		}
	}

	// 读取最终的 Todo 列表
	fmt.Println("\n=== 最终 Todo 列表 ===")
	todoContent, err := fsBackend.ReadFile(ctx, "/todos.md", 0, 0)
	if err == nil {
		fmt.Println(todoContent)
	}

	fmt.Printf("\n工作目录: %s\n", workDir)
	fmt.Println("你可以查看生成的文件和 Todo 列表。")
}
