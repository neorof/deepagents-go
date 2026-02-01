package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zhoucx/deepagents-go/internal/color"
	"github.com/zhoucx/deepagents-go/internal/config"
	"github.com/zhoucx/deepagents-go/internal/logger"
	"github.com/zhoucx/deepagents-go/internal/repl"
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "", "配置文件路径（默认 ~/.deepagents/config.yaml）")
	logLevel := flag.String("log-level", "", "日志级别（debug, info, warn, error）")
	flag.Parse()

	// 加载配置文件
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 命令行参数覆盖配置文件
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	// 初始化日志系统
	if err := initLogger(cfg); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	logger.Info("启动 Deep Agents CLI")
	logger.Debug("配置: model=%s, work_dir=%s, max_iter=%d", cfg.Model, cfg.WorkDir, cfg.MaxIterations)

	// 从环境变量获取 API Key（如果配置文件没有指定）
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	if cfg.BaseUrl == "" {
		cfg.BaseUrl = os.Getenv("ANTHROPIC_BASE_URL")
	}

	// 启动交互模式
	runInteractive(cfg)
}

func runInteractive(cfg *config.Config) {
	logger.Info("启动交互模式")

	// 创建工作目录
	if err := os.MkdirAll(cfg.WorkDir, 0755); err != nil {
		log.Fatalf("创建工作目录失败: %v", err)
	}

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

	// 加载系统提示词
	systemPrompt, err := cfg.LoadSystemPrompt()
	if err != nil {
		logger.Warn("加载系统提示词失败，使用默认提示词: %v", err)
		systemPrompt = `你是一个有用的 AI 助手，可以帮助用户管理文件和执行任务。

可用工具：
- write_todos: 创建任务计划
- ls, read_file, write_file, edit_file, grep, glob: 文件操作
- bash: 执行 bash 命令（支持任意系统命令）

请根据用户的需求使用这些工具完成任务。`
	}

	logger.Debug("系统提示词已加载，长度: %d 字符", len(systemPrompt))

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
		// 设置工具调用回调
		OnToolCall: func(toolName string, input map[string]any) {
			fmt.Printf("\n%s %s", color.Cyan("🔧 Tool:"), color.Bold(toolName))
			// 打印关键参数
			var paramStr string
			if command, ok := input["command"].(string); ok && len(command) > 0 {
				if len(command) > 50 {
					paramStr = fmt.Sprintf("command=%s...", command[:50])
				} else {
					paramStr = fmt.Sprintf("command=%s", command)
				}
			} else if path, ok := input["path"].(string); ok && len(path) > 0 {
				paramStr = fmt.Sprintf("path=%s", path)
			} else if pattern, ok := input["pattern"].(string); ok && len(pattern) > 0 {
				paramStr = fmt.Sprintf("pattern=%s", pattern)
			}
			if paramStr != "" {
				fmt.Printf(" %s", color.Gray(fmt.Sprintf("(%s)", paramStr)))
			}
			fmt.Println()
		},
		OnToolResult: func(toolName string, result string, isError bool) {
			if isError {
				// 截断错误信息
				if len(result) > 100 {
					result = result[:100] + "..."
				}
				fmt.Printf("   %s %s\n", color.Red("❌ Error:"), result)
			} else {
				// 截断结果显示
				displayResult := result
				if len(displayResult) > 100 {
					displayResult = displayResult[:100] + "..."
				}
				fmt.Printf("   %s %s\n", color.Green("✓ Done:"), color.Gray(displayResult))
			}
		},
	}

	// 创建 Agent
	executor := agent.NewExecutor(agentConfig)

	// 创建并运行 REPL
	r := repl.New(executor)
	r.SetStreaming(cfg.EnableStreaming)
	if err := r.Run(); err != nil {
		logger.Error("REPL 运行失败: %v", err)
		log.Fatalf("REPL 运行失败: %v", err)
	}
}

// initLogger 初始化日志系统
func initLogger(cfg *config.Config) error {
	// 解析日志级别
	level := logger.ParseLevel(cfg.LogLevel)

	// 确定日志输出目标
	var output *os.File
	if cfg.LogFile != "" {
		// 输出到文件
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		output = f
		logger.Info("日志输出到文件: %s", cfg.LogFile)
	} else {
		// 输出到标准输出
		output = os.Stdout
	}

	// 创建日志记录器
	l := logger.New(level, output, cfg.LogFormat)
	logger.SetDefault(l)

	logger.Debug("日志系统初始化完成")
	logger.Debug("配置: level=%s, format=%s, file=%s", cfg.LogLevel, cfg.LogFormat, cfg.LogFile)

	return nil
}
