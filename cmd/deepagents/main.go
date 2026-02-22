package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/zhoucx/deepagents-go/internal/config"
	"github.com/zhoucx/deepagents-go/internal/logger"
	"github.com/zhoucx/deepagents-go/internal/repl"
	"github.com/zhoucx/deepagents-go/pkg/agentkit"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func main() {
	configPath := flag.String("config", "", "配置文件路径（默认 ~/.deepagents/config.yaml）")
	logLevel := flag.String("log-level", "", "日志级别（debug, info, warn, error）")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	if err := initLogger(cfg); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	logger.Debug("启动 Agents CLI")
	logger.Debug("配置: model=%s, work_dir=%s, max_iter=%d", cfg.Model, cfg.WorkDir, cfg.MaxIterations)

	// 生成会话 ID
	sessionID := uuid.New().String()
	logger.Debug("会话 ID: %s", sessionID)

	_ = os.MkdirAll(cfg.WorkDir, 0755)
	sp, err := loadSystemPrompt(cfg)
	if err != nil {
		log.Fatalf("加载系统提示词失败, 请确认system_prompt.txt文件是否存在: %v", err)
		return
	}

	builder := agentkit.New(
		agentkit.WithLLM(llm.NewAnthropicClient(cfg.APIKey, cfg.Model, cfg.BaseUrl)),
		agentkit.WithConfig(agentkit.AgentConfig{
			SystemPrompt:  sp,
			MaxIterations: cfg.MaxIterations,
			MaxTokens:     cfg.MaxTokens,
			Temperature:   cfg.Temperature,
		}),
		agentkit.WithFilesystem(cfg.WorkDir),
		agentkit.WithSkillsDirs("skills"),
		agentkit.WithSessionID(sessionID),
		agentkit.EnableSummarization(),
		agentkit.EnableVerboseLogging(),
	)
	if err := builder.Build(); err != nil {
		log.Fatalf("构建 AgentBuilder 失败: %v", err)
		return
	}

	// 获取 backend 并创建 REPL
	r := repl.New(builder, sessionID, &repl.BannerInfo{
		Model:   cfg.Model,
		WorkDir: cfg.WorkDir,
		Version: "v1.0",
	})
	err = r.Run()
	if err != nil {
		log.Fatalf("运行失败: %v", err)
	}
}

func loadSystemPrompt(cfg *config.Config) (string, error) {
	prompt, err := cfg.LoadSystemPrompt()
	if err != nil {
		logger.Warn("加载系统提示词失败，使用默认提示词: %v", err)
		return "", nil
	}
	logger.Debug("系统提示词已加载，长度: %d 字符", len(prompt))
	return prompt, nil
}

// initLogger 初始化日志系统
func initLogger(cfg *config.Config) error {
	level := logger.ParseLevel(cfg.LogLevel)

	var output *os.File
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %w", err)
		}
		output = f
		logger.Info("日志输出到文件: %s", cfg.LogFile)
	} else {
		output = os.Stdout
	}

	l := logger.New(level, output, cfg.LogFormat)
	logger.SetDefault(l)

	logger.Debug("日志系统初始化完成")
	logger.Debug("配置: level=%s, format=%s, file=%s", cfg.LogLevel, cfg.LogFormat, cfg.LogFile)

	return nil
}
