package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// AgentMemMiddleware 负责加载和注入 Agent.md 配置文件
type AgentMemMiddleware struct {
	*BaseMiddleware
	backend       backend.Backend
	configContent string // 缓存的配置内容
	configPath    string // 找到的配置文件路径
	loaded        bool   // 是否已加载
}

// NewAgentMemMiddleware 创建新的 AgentMemMiddleware
func NewAgentMemMiddleware(backend backend.Backend) *AgentMemMiddleware {
	return &AgentMemMiddleware{
		BaseMiddleware: &BaseMiddleware{name: "AgentConfig"},
		backend:        backend,
	}
}

// BeforeAgent 在 Agent 开始执行前加载配置文件
func (m *AgentMemMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	if m.loaded {
		return nil // 已加载，跳过
	}

	// 只查找 AGENT.md 文件
	content, err := m.backend.ReadFile(ctx, "AGENT.md", 0, 0)
	if err == nil && content != "" {
		m.configContent = content
		m.configPath = "AGENT.md"
		m.loaded = true
		return nil
	}

	// 文件不存在，静默跳过
	m.loaded = true
	return nil
}

// BeforeModel 在调用 LLM 前注入配置到系统提示词
func (m *AgentMemMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	if m.configContent == "" {
		return nil // 没有配置内容，跳过
	}

	// 构建配置注入内容
	configInjection := fmt.Sprintf("\n\n=== Agent 指示 ===\n%s\n", strings.TrimSpace(m.configContent))

	// 注入到系统提示词
	if req.SystemPrompt != "" {
		req.SystemPrompt += configInjection
	} else {
		req.SystemPrompt = configInjection
	}

	return nil
}
