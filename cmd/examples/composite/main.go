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

	// 创建工作目录
	workDir := "./workspace_composite"
	dataDir := "./workspace_composite/data"
	configDir := "./workspace_composite/config"

	for _, dir := range []string{workDir, dataDir, configDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}

	// 创建 LLM 客户端
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	llmClient := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", baseURL)

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建组合后端
	memoryBackend := backend.NewStateBackend()
	composite := backend.NewCompositeBackend(memoryBackend)

	// 添加文件系统后端
	dataBackend, _ := backend.NewFilesystemBackend(dataDir, true)
	configBackend, _ := backend.NewFilesystemBackend(configDir, true)

	composite.AddRoute("/data", dataBackend)
	composite.AddRoute("/config", configBackend)

	// 创建中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(composite, toolRegistry)
	todoMiddleware := middleware.NewTodoMiddleware(composite, toolRegistry)

	// 创建 Agent 配置
	config := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			filesystemMiddleware,
			todoMiddleware,
		},
		SystemPrompt: `你是一个有用的 AI 助手，可以管理多个存储后端。

存储后端：
- /data - 数据文件（磁盘存储）
- /config - 配置文件（磁盘存储）
- / - 临时文件（内存存储）

可用工具：
- write_todos: 创建任务计划
- ls, read_file, write_file, edit_file, grep, glob: 文件操作

请根据文件类型选择合适的存储位置。`,
		MaxIterations: 30,
		MaxTokens:     4096,
		Temperature:   0.7,
	}

	// 创建 Agent
	executor := agent.NewRunnable(config)

	// 执行示例任务
	ctx := context.Background()

	fmt.Println("=== Deep Agents Go - 组合后端示例 ===")
	fmt.Println()

	task := `创建一个简单的应用配置系统：

1. 在 /config 目录创建 app.yaml 配置文件，包含：
   - 应用名称
   - 端口号
   - 数据库连接信息

2. 在 /data 目录创建 users.json 数据文件，包含 3 个示例用户

3. 在根目录（内存）创建 /session.txt 临时会话文件

4. 使用 grep 搜索所有包含 "port" 的文件

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

	// 打印最后的响应
	fmt.Println("=== 执行结果 ===")
	for _, msg := range output.Messages {
		if msg.Role == llm.RoleAssistant && msg.Content != "" {
			fmt.Printf("\n%s\n", msg.Content)
		}
	}

	// 验证文件位置
	fmt.Println("\n=== 验证文件位置 ===")

	// 检查配置文件（应该在磁盘）
	if _, err := os.Stat(configDir + "/app.yaml"); err == nil {
		fmt.Println("✓ /config/app.yaml 存在于磁盘")
	}

	// 检查数据文件（应该在磁盘）
	if _, err := os.Stat(dataDir + "/users.json"); err == nil {
		fmt.Println("✓ /data/users.json 存在于磁盘")
	}

	// 检查会话文件（应该在内存）
	if _, err := memoryBackend.ReadFile(ctx, "/session.txt", 0, 0); err == nil {
		fmt.Println("✓ /session.txt 存在于内存")
	}

	fmt.Printf("\n工作目录: %s\n", workDir)
	fmt.Println("你可以查看生成的文件。")
}
