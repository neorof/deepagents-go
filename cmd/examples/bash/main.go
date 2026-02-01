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
		log.Fatal("请设置 ANTHROPIC_API_KEY 环境变量")
	}

	// 创建 LLM 客户端
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	llmClient := llm.NewAnthropicClient(apiKey, "", baseURL)

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建后端
	fsBackend, err := backend.NewFilesystemBackend("./workspace_bash", true)
	if err != nil {
		log.Fatalf("创建文件系统后端失败: %v", err)
	}

	// 创建中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)

	// 创建 Agent 配置
	agentConfig := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			filesystemMiddleware,
		},
		SystemPrompt: `你是一个有用的 AI 助手，可以帮助用户执行系统命令和管理文件。

可用工具：
- bash: 执行 bash 命令（支持任意系统命令）
- ls, read_file, write_file, edit_file, grep, glob: 文件操作

请根据用户的需求使用这些工具完成任务。`,
		MaxIterations: 10,
	}

	// 创建 Agent
	executor := agent.NewExecutor(agentConfig)

	// 示例任务
	tasks := []string{
		"使用 bash 命令查看当前目录的内容（使用 ls -la）",
		"使用 bash 命令创建一个名为 test.txt 的文件，内容为当前日期和时间",
		"使用 bash 命令统计当前目录下有多少个 .go 文件",
		"使用 bash 命令查看系统信息（uname -a）",
	}

	ctx := context.Background()

	for i, task := range tasks {
		fmt.Printf("\n=== 任务 %d ===\n", i+1)
		fmt.Printf("任务: %s\n\n", task)

		output, err := executor.Invoke(ctx, &agent.InvokeInput{
			Messages: []llm.Message{
				{
					Role:    llm.RoleUser,
					Content: task,
				},
			},
		})

		if err != nil {
			log.Printf("任务执行失败: %v", err)
			continue
		}

		// 打印结果
		for _, msg := range output.Messages {
			if msg.Role == llm.RoleAssistant && msg.Content != "" {
				fmt.Printf("结果: %s\n", msg.Content)
			}
		}
	}

	fmt.Printf("\n工作目录: ./workspace_bash\n")
}
