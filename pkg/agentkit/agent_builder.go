package agentkit

import (
	"fmt"
	"log"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// AgentBuilder 通过 Option 模式构建 AgentBuilder
type AgentBuilder struct {
	// 基础配置
	llmClient     llm.Client
	systemPrompt  string
	maxIterations int
	maxTokens     int
	temperature   float64

	// 回调
	onToolCall   func(string, map[string]any)
	onToolResult func(string, string, bool)

	// 中间件配置参数
	fsWorkDir       string
	webConfig       *middleware.WebConfig
	skillsDirs      []string
	memoryDirs      []string
	enableSummarize bool
	summarizeConfig *middleware.SummarizationConfig
	sessionID       string // 会话 ID

	// 自定义中间件
	customMiddlewares []agent.Middleware

	// 构建后的内部组件
	toolRegistry *tools.Registry
	Backend      backend.Backend
	SessionStore *middleware.SessionStore
	middlewares  []agent.Middleware

	Runnable *agent.Runnable
}

// New 创建 AgentBuilder 并应用所有 Option
func New(opts ...Option) *AgentBuilder {
	a := &AgentBuilder{
		maxIterations: 25,
		maxTokens:     4096,
		temperature:   0.8,
	}

	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Build 根据配置构建内部组件（工具注册表、中间件、Runnable）
// 默认启用所有内置中间件：Filesystem、ContextInjection、Todo、Web、Skills、Memory、SubAgent
func (a *AgentBuilder) Build() error {
	if a.fsWorkDir == "" {
		return fmt.Errorf("WithFilesystem 是必须的，请指定工作目录")
	}

	a.toolRegistry = tools.NewRegistry()
	a.middlewares = make([]agent.Middleware, 0)

	// 1. 文件系统中间件
	fsBackend, err := backend.NewFilesystemBackend(a.fsWorkDir, true)
	if err != nil {
		return fmt.Errorf("创建文件系统后端失败: %w", err)
	}
	a.Backend = fsBackend
	a.middlewares = append(a.middlewares, middleware.NewFilesystemMiddleware(fsBackend, a.toolRegistry))

	// 2. Agent 配置中间件
	agentMemMiddleware := middleware.NewAgentMemMiddleware(fsBackend)
	a.middlewares = append(a.middlewares, agentMemMiddleware)

	// 3. 上下文注入中间件
	contextInjector := middleware.NewContextInjectionMiddleware()
	a.middlewares = append(a.middlewares, contextInjector)

	// 4. 会话记录中间件
	a.SessionStore = middleware.NewSessionStore(a.Backend)
	sessionRecorder := middleware.NewHistoryMiddleware(a.SessionStore)
	a.middlewares = append(a.middlewares, sessionRecorder)

	// 5. Todo 中间件
	a.middlewares = append(a.middlewares, middleware.NewTodoMiddlewareWithInjector(a.Backend, a.toolRegistry, contextInjector))

	// 6. Web 中间件
	webCfg := middleware.DefaultWebConfig()
	if a.webConfig != nil {
		webCfg = *a.webConfig
	}
	a.middlewares = append(a.middlewares, middleware.NewWebMiddleware(a.toolRegistry, webCfg))

	// 7. Skills 中间件
	skillsMiddleware := middleware.NewSkillsMiddleware(a.toolRegistry)
	for _, dir := range a.skillsDirs {
		if err := skillsMiddleware.LoadFromDirectory(dir); err != nil {
			log.Printf("警告: 无法加载技能目录 %s: %v\n", dir, err)
		}
	}
	a.middlewares = append(a.middlewares, skillsMiddleware)

	// 8. SubAgent 中间件
	a.middlewares = append(a.middlewares, middleware.NewSubAgentMiddleware(
		nil, a.llmClient, a.toolRegistry, a.middlewares,
		a.systemPrompt, a.maxTokens, a.temperature,
	))

	// 9. Summarization 中间件（可选）
	if a.enableSummarize {
		cfg := a.summarizeConfig
		if cfg == nil {
			cfg = &middleware.SummarizationConfig{
				Enabled: true,
			}
		}
		if cfg.LLMClient == nil {
			cfg.LLMClient = a.llmClient
		}
		a.middlewares = append(a.middlewares, middleware.NewSummarizationMiddleware(cfg))
	}

	// 10. 自定义中间件
	a.middlewares = append(a.middlewares, a.customMiddlewares...)

	// 构建 Runnable
	a.Runnable = agent.NewRunnable(&agent.Config{
		LLMClient:     a.llmClient,
		ToolRegistry:  a.toolRegistry,
		Middlewares:   a.middlewares,
		SystemPrompt:  a.systemPrompt,
		MaxIterations: a.maxIterations,
		MaxTokens:     a.maxTokens,
		Temperature:   a.temperature,
		OnToolCall:    a.onToolCall,
		OnToolResult:  a.onToolResult,
	})
	return nil
}
