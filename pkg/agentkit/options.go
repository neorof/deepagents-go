package agentkit

import (
	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
)

// Option 配置 AgentBuilder 的函数选项
type Option func(*AgentBuilder)

// AgentConfig 基础配置参数
type AgentConfig struct {
	SystemPrompt  string
	MaxIterations int
	MaxTokens     int
	Temperature   float64
}

// WithLLM 设置 LLM 客户端
func WithLLM(client llm.Client) Option {
	return func(a *AgentBuilder) {
		a.llmClient = client
	}
}

// WithConfig 注入基础配置参数
func WithConfig(cfg AgentConfig) Option {
	return func(a *AgentBuilder) {
		if cfg.SystemPrompt != "" {
			a.systemPrompt = cfg.SystemPrompt
		}
		if cfg.MaxIterations > 0 {
			a.maxIterations = cfg.MaxIterations
		}
		if cfg.MaxTokens > 0 {
			a.maxTokens = cfg.MaxTokens
		}
		if cfg.Temperature > 0 {
			a.temperature = cfg.Temperature
		}
	}
}

// WithFilesystem 设置工作目录（必须调用）
func WithFilesystem(workDir string) Option {
	return func(a *AgentBuilder) {
		a.fsWorkDir = workDir
	}
}

// WithWebConfig 自定义 Web 中间件配置
func WithWebConfig(cfg middleware.WebConfig) Option {
	return func(a *AgentBuilder) {
		a.webConfig = &cfg
	}
}

// WithSkillsDirs 添加技能目录
func WithSkillsDirs(dirs ...string) Option {
	return func(a *AgentBuilder) {
		a.skillsDirs = append(a.skillsDirs, dirs...)
	}
}

// WithMemoryPaths 添加记忆文件或目录路径（已废弃，保留用于向后兼容）
// 新的 MemoryMiddleware 自动记录对话历史，不再需要手动加载记忆文件
func WithMemoryPaths(paths ...string) Option {
	return func(a *AgentBuilder) {
		a.memoryDirs = append(a.memoryDirs, paths...)
	}
}

// WithSessionID 设置会话 ID
func WithSessionID(sessionID string) Option {
	return func(a *AgentBuilder) {
		a.sessionID = sessionID
	}
}

// EnableSummarization 启用摘要中间件（可选，自动复用 LLM 客户端）
func EnableSummarization(cfg ...*middleware.SummarizationConfig) Option {
	return func(a *AgentBuilder) {
		a.enableSummarize = true
		if len(cfg) > 0 {
			a.summarizeConfig = cfg[0]
		}
	}
}

// WithMiddleware 添加自定义中间件
func WithMiddleware(m agent.Middleware) Option {
	return func(a *AgentBuilder) {
		a.customMiddlewares = append(a.customMiddlewares, m)
	}
}

// WithOnToolCall 设置工具调用回调
func WithOnToolCall(fn func(string, map[string]any)) Option {
	return func(a *AgentBuilder) {
		a.onToolCall = fn
	}
}

// WithOnToolResult 设置工具结果回调
func WithOnToolResult(fn func(string, string, bool)) Option {
	return func(a *AgentBuilder) {
		a.onToolResult = fn
	}
}

// WithSystemPrompt 独立设置系统提示词（不需要通过 WithConfig）
func WithSystemPrompt(prompt string) Option {
	return func(a *AgentBuilder) {
		a.systemPrompt = prompt
	}
}

// EnableVerboseLogging 启用详细的工具调用日志输出
func EnableVerboseLogging() Option {
	return func(a *AgentBuilder) {
		a.onToolCall = DefaultToolCallHandler
		a.onToolResult = DefaultToolResultHandler
	}
}
