package middleware

import (
	"context"
	"fmt"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// SubAgentConfig 子Agent配置
type SubAgentConfig struct {
	MaxDepth int // 最大递归深度，0表示不限制
}

// SubAgentMiddleware 子Agent中间件
// 支持创建子Agent来处理复杂任务
type SubAgentMiddleware struct {
	*BaseMiddleware
	config       *SubAgentConfig
	toolRegistry *tools.Registry
	llmClient    llm.Client
	middlewares  []agent.Middleware
	systemPrompt string
	maxTokens    int
	temperature  float64
}

// NewSubAgentMiddleware 创建子Agent中间件
func NewSubAgentMiddleware(
	config *SubAgentConfig,
	llmClient llm.Client,
	toolRegistry *tools.Registry,
	middlewares []agent.Middleware,
	systemPrompt string,
	maxTokens int,
	temperature float64,
) *SubAgentMiddleware {
	if config == nil {
		config = &SubAgentConfig{
			MaxDepth: 3, // 默认最大深度为3
		}
	}

	m := &SubAgentMiddleware{
		BaseMiddleware: NewBaseMiddleware("subagent"),
		config:         config,
		toolRegistry:   toolRegistry,
		llmClient:      llmClient,
		middlewares:    middlewares,
		systemPrompt:   systemPrompt,
		maxTokens:      maxTokens,
		temperature:    temperature,
	}

	// 注册 delegate_to_subagent 工具
	m.registerTools()

	return m
}

// registerTools 注册工具
func (m *SubAgentMiddleware) registerTools() {
	m.toolRegistry.Register(tools.NewBaseTool(
		"delegate_to_subagent",
		"将复杂任务委派给子Agent处理。子Agent有独立的状态和上下文，适合处理需要多步骤的子任务。",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task": map[string]any{
					"type":        "string",
					"description": "要委派给子Agent的任务描述",
				},
				"context": map[string]any{
					"type":        "string",
					"description": "传递给子Agent的上下文信息（可选）",
				},
			},
			"required": []string{"task"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			return m.executeSubAgent(ctx, args)
		},
	))
}

// executeSubAgent 执行子Agent
func (m *SubAgentMiddleware) executeSubAgent(ctx context.Context, args map[string]any) (string, error) {
	// 获取任务描述
	task, ok := args["task"].(string)
	if !ok {
		return "", fmt.Errorf("task must be a string")
	}

	// 获取上下文（可选）
	contextInfo := ""
	if ctx, ok := args["context"].(string); ok {
		contextInfo = ctx
	}

	// 检查递归深度
	depth := m.getDepth(ctx)
	if m.config.MaxDepth > 0 && depth >= m.config.MaxDepth {
		return "", fmt.Errorf("maximum recursion depth (%d) exceeded", m.config.MaxDepth)
	}

	// 创建子Agent的消息
	messages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: task,
		},
	}

	// 如果有上下文信息，添加到消息前面
	if contextInfo != "" {
		messages = []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("Context: %s\n\nTask: %s", contextInfo, task),
			},
		}
	}

	// 创建子Agent配置（不包含SubAgentMiddleware，避免无限递归）
	subMiddlewares := make([]agent.Middleware, 0)
	for _, mw := range m.middlewares {
		// 跳过SubAgentMiddleware，避免无限递归
		if mw.Name() != "subagent" {
			subMiddlewares = append(subMiddlewares, mw)
		}
	}

	subAgentConfig := &agent.Config{
		LLMClient:     m.llmClient,
		ToolRegistry:  m.toolRegistry,
		Middlewares:   subMiddlewares,
		SystemPrompt:  m.systemPrompt,
		MaxIterations: 10, // 子Agent的迭代次数限制
		MaxTokens:     m.maxTokens,
		Temperature:   m.temperature,
	}

	// 创建子Agent执行器
	subExecutor := agent.NewRunnable(subAgentConfig)

	// 创建带深度信息的上下文
	subCtx := m.withDepth(ctx, depth+1)

	// 执行子Agent
	output, err := subExecutor.Invoke(subCtx, &agent.InvokeInput{
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("subagent execution failed: %w", err)
	}

	// 提取子Agent的最终响应
	if len(output.Messages) == 0 {
		return "SubAgent completed but returned no messages", nil
	}

	// 获取最后一条助手消息
	var lastResponse string
	for i := len(output.Messages) - 1; i >= 0; i-- {
		if output.Messages[i].Role == llm.RoleAssistant {
			lastResponse = output.Messages[i].Content
			break
		}
	}

	if lastResponse == "" {
		return "SubAgent completed but returned no assistant message", nil
	}

	return fmt.Sprintf("SubAgent completed successfully.\n\nResult:\n%s", lastResponse), nil
}

// getDepth 从上下文中获取当前递归深度
func (m *SubAgentMiddleware) getDepth(ctx context.Context) int {
	if depth, ok := ctx.Value(subAgentDepthKey{}).(int); ok {
		return depth
	}
	return 0
}

// withDepth 创建带深度信息的上下文
func (m *SubAgentMiddleware) withDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, subAgentDepthKey{}, depth)
}

// subAgentDepthKey 用于在上下文中存储递归深度
type subAgentDepthKey struct{}
