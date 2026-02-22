package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	fmt.Println("=== SandboxBackend 示例 ===")

	// 获取 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	// 创建临时工作目录
	workDir := "./sandbox_workspace"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		log.Fatalf("Failed to create workspace: %v", err)
	}
	defer os.RemoveAll(workDir)

	// 演示1: 基本沙箱配置
	fmt.Println("演示1: 基本沙箱配置")
	demo1BasicSandbox(workDir)

	// 演示2: 只读模式
	fmt.Println("\n演示2: 只读模式")
	demo2ReadOnlyMode(workDir)

	// 演示3: 资源限制
	fmt.Println("\n演示3: 资源限制")
	demo3ResourceLimits(workDir)

	// 演示4: 路径控制
	fmt.Println("\n演示4: 路径控制（白名单/黑名单）")
	demo4PathControl(workDir)

	// 演示5: 审计日志
	fmt.Println("\n演示5: 审计日志")
	demo5AuditLog(workDir)

	// 演示6: 命令执行
	fmt.Println("\n演示6: 命令执行")
	demo6CommandExecution(workDir)

	// 演示7: 与 Agent 集成
	fmt.Println("\n演示7: 与 Agent 集成")
	demo7AgentIntegration(apiKey, workDir)

	fmt.Println("\n=== 演示完成 ===")
}

// demo1BasicSandbox 演示基本沙箱配置
func demo1BasicSandbox(workDir string) {
	config := backend.DefaultSandboxConfig(workDir)
	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	ctx := context.Background()

	// 写入文件
	content := "Hello from Sandbox!"
	result, err := sandboxBackend.WriteFile(ctx, "/test.txt", content)
	if err != nil {
		log.Printf("WriteFile failed: %v", err)
		return
	}
	fmt.Printf("✓ 写入文件: %s (%d bytes)\n", result.Path, result.BytesWritten)

	// 读取文件
	readContent, err := sandboxBackend.ReadFile(ctx, "/test.txt", 0, 0)
	if err != nil {
		log.Printf("ReadFile failed: %v", err)
		return
	}
	fmt.Printf("✓ 读取文件: %s\n", readContent)
}

// demo2ReadOnlyMode 演示只读模式
func demo2ReadOnlyMode(workDir string) {
	config := backend.DefaultSandboxConfig(workDir)
	config.ReadOnly = true
	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	ctx := context.Background()

	// 尝试写入（应该失败）
	_, err = sandboxBackend.WriteFile(ctx, "/test.txt", "content")
	if err != nil {
		fmt.Printf("✓ 只读模式阻止写入: %v\n", err)
	} else {
		fmt.Println("✗ 只读模式未生效")
	}
}

// demo3ResourceLimits 演示资源限制
func demo3ResourceLimits(workDir string) {
	config := backend.DefaultSandboxConfig(workDir)
	config.MaxFileSize = 100 // 限制文件大小为 100 bytes
	config.MaxOperations = 3 // 限制操作次数为 3
	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	ctx := context.Background()

	// 测试文件大小限制
	largeContent := string(make([]byte, 200))
	_, err = sandboxBackend.WriteFile(ctx, "/large.txt", largeContent)
	if err != nil {
		fmt.Printf("✓ 文件大小限制生效: %v\n", err)
	}

	// 测试操作次数限制
	for i := 1; i <= 5; i++ {
		_, err := sandboxBackend.WriteFile(ctx, fmt.Sprintf("/file%d.txt", i), "content")
		if err != nil {
			fmt.Printf("✓ 操作次数限制生效（第%d次）: %v\n", i, err)
			break
		} else {
			fmt.Printf("  操作 %d 成功\n", i)
		}
	}
}

// demo4PathControl 演示路径控制
func demo4PathControl(workDir string) {
	config := backend.DefaultSandboxConfig(workDir)
	config.AllowedPaths = []string{"/public"}
	config.BlockedPaths = []string{"/secret"}
	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	ctx := context.Background()

	// 访问允许的路径
	_, err = sandboxBackend.WriteFile(ctx, "/public/data.txt", "content")
	if err == nil {
		fmt.Println("✓ 允许访问白名单路径: /public")
	}

	// 访问被阻止的路径
	_, err = sandboxBackend.WriteFile(ctx, "/secret/data.txt", "content")
	if err != nil {
		fmt.Printf("✓ 阻止访问黑名单路径: %v\n", err)
	}

	// 访问不在白名单的路径
	_, err = sandboxBackend.WriteFile(ctx, "/other/data.txt", "content")
	if err != nil {
		fmt.Printf("✓ 阻止访问非白名单路径: %v\n", err)
	}
}

// demo5AuditLog 演示审计日志
func demo5AuditLog(workDir string) {
	config := backend.DefaultSandboxConfig(workDir)
	config.EnableAuditLog = true
	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	ctx := context.Background()

	// 执行一些操作
	_, _ = sandboxBackend.WriteFile(ctx, "/log_test.txt", "content")
	_, _ = sandboxBackend.ReadFile(ctx, "/log_test.txt", 0, 0)
	_, _ = sandboxBackend.ListFiles(ctx, "/")

	// 获取审计日志
	auditLog := sandboxBackend.GetAuditLog()
	fmt.Printf("✓ 记录了 %d 条审计日志:\n", len(auditLog))
	for i, entry := range auditLog {
		status := "成功"
		if !entry.Success {
			status = "失败: " + entry.Error
		}
		fmt.Printf("  %d. [%s] %s %s - %s\n",
			i+1,
			entry.Timestamp.Format("15:04:05"),
			entry.Operation,
			entry.Path,
			status,
		)
	}
}

// demo6CommandExecution 演示命令执行
func demo6CommandExecution(workDir string) {
	config := backend.DefaultSandboxConfig(workDir)
	config.AllowedCommands = []string{"echo", "ls", "pwd"}
	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	ctx := context.Background()

	// 执行允许的命令
	result, err := sandboxBackend.Execute(ctx, "echo Hello Sandbox", 1000)
	if err == nil {
		fmt.Printf("✓ 执行允许的命令: %s\n", result.Stdout)
	}

	// 执行不允许的命令
	_, err = sandboxBackend.Execute(ctx, "rm -rf /", 1000)
	if err != nil {
		fmt.Printf("✓ 阻止危险命令: %v\n", err)
	}
}

// demo7AgentIntegration 演示与 Agent 集成
func demo7AgentIntegration(apiKey, workDir string) {
	// 创建沙箱配置
	config := backend.DefaultSandboxConfig(workDir)
	config.MaxFileSize = 1024 * 1024 // 1MB
	config.MaxOperations = 50
	config.EnableAuditLog = true
	config.AllowedCommands = []string{"ls", "cat", "echo", "pwd"}

	sandboxBackend, err := backend.NewSandboxBackend(config)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	// 创建 LLM 客户端
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	llmClient := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", baseURL)

	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建文件系统中间件（使用沙箱后端）
	filesystemMiddleware := middleware.NewFilesystemMiddleware(sandboxBackend, toolRegistry)

	// 创建 Agent 配置
	agentConfig := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares:  []agent.Middleware{filesystemMiddleware},
		SystemPrompt: `你是一个运行在沙箱环境中的 AI 助手。
你可以安全地执行文件操作和命令，所有操作都受到严格的资源限制和权限控制。`,
	}

	// 创建 Agent
	executor := agent.NewRunnable(agentConfig)

	// 执行任务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err := executor.Invoke(ctx, &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "创建一个文件 /demo.txt，内容为 'Hello from Sandbox Agent!'，然后读取它的内容。",
			},
		},
	})

	if err != nil {
		log.Printf("Agent execution failed: %v", err)
		return
	}

	fmt.Printf("✓ Agent 响应:\n%s\n", output.Messages[len(output.Messages)-1].Content)

	// 显示审计日志
	auditLog := sandboxBackend.GetAuditLog()
	fmt.Printf("\n✓ Agent 执行了 %d 个操作\n", len(auditLog))
	fmt.Printf("✓ 操作计数: %d\n", sandboxBackend.GetOperationCount())
}
