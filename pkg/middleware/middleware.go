package middleware

import (
	"context"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// BaseMiddleware 提供中间件的默认实现
type BaseMiddleware struct {
	name string
}

// NewBaseMiddleware 创建基础中间件
func NewBaseMiddleware(name string) *BaseMiddleware {
	return &BaseMiddleware{name: name}
}

func (m *BaseMiddleware) Name() string {
	return m.name
}

func (m *BaseMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	return nil
}

func (m *BaseMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	return nil
}

func (m *BaseMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
	return nil
}

func (m *BaseMiddleware) BeforeTool(ctx context.Context, toolCall *llm.ToolCall, state *agent.State) error {
	return nil
}

func (m *BaseMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	return nil
}
