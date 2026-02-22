package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/zhoucx/deepagents-go/internal/config"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	fmt.Println("=== System Prompt 测试 ===")

	// 获取 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 ANTHROPIC_API_KEY 环境变量")
	}

	// 加载配置
	cfg := config.DefaultConfig()
	cfg.APIKey = apiKey
	cfg.BaseUrl = os.Getenv("ANTHROPIC_BASE_URL")
	cfg.WorkDir = "./workspace_prompt_test"

	// 加载系统提示词
	systemPrompt, err := cfg.LoadSystemPrompt()
	if err != nil {
		log.Fatalf("加载系统提示词失败: %v", err)
	}

	fmt.Printf("✓ 系统提示词已加载，长度: %d 字符\n\n", len(systemPrompt))

	// 创建 LLM 客户端
	llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model, cfg.BaseUrl)

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建后端
	fsBackend, err := backend.NewFilesystemBackend(cfg.WorkDir, true)
	if err != nil {
		log.Fatalf("创建文件系统后端失败: %v", err)
	}

	// 创建中间件
	filesystemMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)
	todoMiddleware := middleware.NewTodoMiddleware(fsBackend, toolRegistry)

	// 记录工具调用
	var toolCalls []string

	// 创建 Agent 配置
	agentConfig := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			filesystemMiddleware,
			todoMiddleware,
		},
		SystemPrompt:  systemPrompt,
		MaxIterations: cfg.MaxIterations,
		MaxTokens:     cfg.MaxTokens,
		Temperature:   cfg.Temperature,
		OnToolCall: func(toolName string, input map[string]any) {
			toolCalls = append(toolCalls, toolName)
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

	// 测试用例：这些任务应该触发正确的工具使用
	testCases := []struct {
		name         string
		task         string
		shouldUse    []string // 应该使用的工具
		shouldNotUse []string // 不应该使用的工具
	}{
		{
			name:         "测试1: 读取文件应该用 read_file 而不是 bash cat",
			task:         "读取 system_prompt.txt 文件的前10行内容",
			shouldUse:    []string{"read_file"},
			shouldNotUse: []string{"bash"},
		},
		{
			name:         "测试2: 搜索文件应该用 glob 而不是 bash find",
			task:         "查找当前目录下所有的 .go 文件",
			shouldUse:    []string{"glob"},
			shouldNotUse: []string{"bash"},
		},
		{
			name:         "测试3: 搜索内容应该用 grep 而不是 bash grep",
			task:         "在所有 .go 文件中搜索包含 'func main' 的行",
			shouldUse:    []string{"grep"},
			shouldNotUse: []string{"bash"},
		},
		{
			name:         "测试4: 编辑文件应该先 read_file 再 edit_file",
			task:         "创建一个测试文件 test.txt，内容为 'hello'，然后把 'hello' 改为 'world'",
			shouldUse:    []string{"write_file", "read_file", "edit_file"},
			shouldNotUse: []string{},
		},
		{
			name:         "测试5: git 操作应该用 bash",
			task:         "查看 git 状态",
			shouldUse:    []string{"bash"},
			shouldNotUse: []string{},
		},
	}

	ctx := context.Background()
	passCount := 0
	failCount := 0

	for i, tc := range testCases {
		fmt.Printf("\n=== %s ===\n", tc.name)
		fmt.Printf("任务: %s\n\n", tc.task)

		// 重置工具调用记录
		toolCalls = []string{}

		output, err := executor.Invoke(ctx, &agent.InvokeInput{
			Messages: []llm.Message{
				{
					Role:    llm.RoleUser,
					Content: tc.task,
				},
			},
		})

		if err != nil {
			log.Printf("❌ 任务执行失败: %v\n", err)
			failCount++
			continue
		}

		// 验证工具使用
		passed := true

		// 检查应该使用的工具
		for _, tool := range tc.shouldUse {
			if !contains(toolCalls, tool) {
				fmt.Printf("  ❌ 应该使用 %s 但没有使用\n", tool)
				passed = false
			}
		}

		// 检查不应该使用的工具
		for _, tool := range tc.shouldNotUse {
			if contains(toolCalls, tool) {
				fmt.Printf("  ❌ 不应该使用 %s 但使用了\n", tool)
				passed = false
			}
		}

		if passed {
			fmt.Printf("  ✅ 测试通过\n")
			passCount++
		} else {
			fmt.Printf("  ❌ 测试失败\n")
			failCount++
		}

		// 打印实际使用的工具
		fmt.Printf("  实际调用的工具: %s\n", strings.Join(toolCalls, ", "))

		// 打印 agent 的响应
		for _, msg := range output.Messages {
			if msg.Role == llm.RoleAssistant && msg.Content != "" {
				fmt.Printf("\n  Agent 响应: %s\n", truncate(msg.Content, 200))
			}
		}

		// 避免请求过快
		if i < len(testCases)-1 {
			fmt.Println("\n等待 2 秒...")
			// time.Sleep(2 * time.Second)
		}
	}

	// 打印测试总结
	fmt.Printf("\n\n=== 测试总结 ===\n")
	fmt.Printf("通过: %d/%d\n", passCount, len(testCases))
	fmt.Printf("失败: %d/%d\n", failCount, len(testCases))

	if passCount == len(testCases) {
		fmt.Println("\n🎉 所有测试通过！system_prompt.txt 规则生效！")
	} else {
		fmt.Println("\n⚠️  部分测试失败，请检查 system_prompt.txt 配置")
	}

	fmt.Printf("\n工作目录: %s\n", cfg.WorkDir)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
