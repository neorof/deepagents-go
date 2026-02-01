package middleware

import (
	"context"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// Chain 中间件链
type Chain struct {
	middlewares []agent.Middleware
}

// NewChain 创建中间件链
func NewChain(middlewares ...agent.Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

// Add 添加中间件
func (c *Chain) Add(middleware agent.Middleware) {
	c.middlewares = append(c.middlewares, middleware)
}

// BeforeAgent 执行所有中间件的 BeforeAgent 钩子
func (c *Chain) BeforeAgent(ctx context.Context, state *agent.State) error {
	for _, m := range c.middlewares {
		if err := m.BeforeAgent(ctx, state); err != nil {
			return err
		}
	}
	return nil
}

// BeforeModel 执行所有中间件的 BeforeModel 钩子
func (c *Chain) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	for _, m := range c.middlewares {
		if err := m.BeforeModel(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// AfterModel 执行所有中间件的 AfterModel 钩子
func (c *Chain) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
	for _, m := range c.middlewares {
		if err := m.AfterModel(ctx, resp, state); err != nil {
			return err
		}
	}
	return nil
}

// BeforeTool 执行所有中间件的 BeforeTool 钩子
func (c *Chain) BeforeTool(ctx context.Context, toolCall *llm.ToolCall, state *agent.State) error {
	for _, m := range c.middlewares {
		if err := m.BeforeTool(ctx, toolCall, state); err != nil {
			return err
		}
	}
	return nil
}

// AfterTool 执行所有中间件的 AfterTool 钩子
func (c *Chain) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	for _, m := range c.middlewares {
		if err := m.AfterTool(ctx, result, state); err != nil {
			return err
		}
	}
	return nil
}
