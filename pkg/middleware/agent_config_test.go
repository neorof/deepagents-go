package middleware

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func TestAgentConfigMiddleware_BeforeAgent_LoadsConfig(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	// 创建 Agent.md 文件
	configContent := "# Agent 配置\n\n你是一个专业的 Go 开发助手。"
	_, err := backend.WriteFile(ctx, "/memory/Agent.md", configContent)
	if err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 调用 BeforeAgent
	err = middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent 失败: %v", err)
	}

	// 验证配置已加载
	if middleware.configContent != configContent {
		t.Errorf("配置内容不匹配，期望 %q，实际 %q", configContent, middleware.configContent)
	}
	if middleware.configPath != "/memory/Agent.md" {
		t.Errorf("配置路径不匹配，期望 %q，实际 %q", "/memory/Agent.md", middleware.configPath)
	}
	if !middleware.loaded {
		t.Error("loaded 标志应为 true")
	}
}

func TestAgentConfigMiddleware_BeforeAgent_FileNotExists(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 调用 BeforeAgent（文件不存在）
	err := middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent 应该静默跳过，但返回错误: %v", err)
	}

	// 验证没有加载配置
	if middleware.configContent != "" {
		t.Error("configContent 应为空")
	}
	if middleware.configPath != "" {
		t.Error("configPath 应为空")
	}
	if !middleware.loaded {
		t.Error("loaded 标志应为 true（表示已尝试加载）")
	}
}

func TestAgentConfigMiddleware_BeforeAgent_PathPriority(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	// 创建多个配置文件
	_, _ = backend.WriteFile(ctx, "/.claude/Agent.md", "claude config")
	_, _ = backend.WriteFile(ctx, "/memories/Agent.md", "memories config")
	_, _ = backend.WriteFile(ctx, "/memory/Agent.md", "memory config")

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 调用 BeforeAgent
	err := middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent 失败: %v", err)
	}

	// 验证优先级：/memory/Agent.md 应该被加载
	if middleware.configPath != "/memory/Agent.md" {
		t.Errorf("应该加载 /memory/Agent.md，实际加载 %q", middleware.configPath)
	}
	if middleware.configContent != "memory config" {
		t.Errorf("配置内容不匹配，期望 %q，实际 %q", "memory config", middleware.configContent)
	}
}

func TestAgentConfigMiddleware_BeforeAgent_CachesConfig(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	// 创建配置文件
	configContent := "# Agent 配置"
	_, err := backend.WriteFile(ctx, "/memory/Agent.md", configContent)
	if err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 第一次调用
	err = middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("第一次 BeforeAgent 失败: %v", err)
	}

	// 验证配置已加载
	if middleware.configContent != configContent {
		t.Error("配置应该被加载")
	}

	// 第二次调用（应该使用缓存，不重新读取）
	err = middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("第二次 BeforeAgent 失败: %v", err)
	}

	// 验证配置仍然存在（使用缓存）
	if middleware.configContent != configContent {
		t.Error("配置应该被缓存")
	}
	if !middleware.loaded {
		t.Error("loaded 标志应为 true")
	}
}

func TestAgentConfigMiddleware_BeforeModel_InjectsConfig(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	// 创建配置文件
	configContent := "你是一个专业的 Go 开发助手。"
	_, err := backend.WriteFile(ctx, "/memory/Agent.md", configContent)
	if err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 加载配置
	err = middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent 失败: %v", err)
	}

	// 创建请求
	req := &llm.ModelRequest{
		SystemPrompt: "原始系统提示词",
		Messages:     []llm.Message{{Role: "user", Content: "你好"}},
	}

	// 调用 BeforeModel
	err = middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel 失败: %v", err)
	}

	// 验证配置已注入
	expectedPrompt := "原始系统提示词\n\n=== Agent 配置 ===\n你是一个专业的 Go 开发助手。\n"
	if req.SystemPrompt != expectedPrompt {
		t.Errorf("系统提示词不匹配\n期望: %q\n实际: %q", expectedPrompt, req.SystemPrompt)
	}
}

func TestAgentConfigMiddleware_BeforeModel_NoConfig(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 加载配置（文件不存在）
	err := middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent 失败: %v", err)
	}

	// 创建请求
	originalPrompt := "原始系统提示词"
	req := &llm.ModelRequest{
		SystemPrompt: originalPrompt,
		Messages:     []llm.Message{{Role: "user", Content: "你好"}},
	}

	// 调用 BeforeModel
	err = middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel 失败: %v", err)
	}

	// 验证系统提示词未改变
	if req.SystemPrompt != originalPrompt {
		t.Errorf("系统提示词不应改变，期望 %q，实际 %q", originalPrompt, req.SystemPrompt)
	}
}

func TestAgentConfigMiddleware_BeforeModel_EmptySystemPrompt(t *testing.T) {
	ctx := context.Background()
	backend := backend.NewStateBackend()

	// 创建配置文件
	configContent := "你是一个专业的 Go 开发助手。"
	_, err := backend.WriteFile(ctx, "/memory/Agent.md", configContent)
	if err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	middleware := NewAgentMemMiddleware(backend)
	state := agent.NewState()

	// 加载配置
	err = middleware.BeforeAgent(ctx, state)
	if err != nil {
		t.Fatalf("BeforeAgent 失败: %v", err)
	}

	// 创建请求（空系统提示词）
	req := &llm.ModelRequest{
		SystemPrompt: "",
		Messages:     []llm.Message{{Role: "user", Content: "你好"}},
	}

	// 调用 BeforeModel
	err = middleware.BeforeModel(ctx, req)
	if err != nil {
		t.Fatalf("BeforeModel 失败: %v", err)
	}

	// 验证配置已注入
	expectedPrompt := "\n\n=== Agent 配置 ===\n你是一个专业的 Go 开发助手。\n"
	if req.SystemPrompt != expectedPrompt {
		t.Errorf("系统提示词不匹配\n期望: %q\n实际: %q", expectedPrompt, req.SystemPrompt)
	}
}

func TestAgentConfigMiddleware_Name(t *testing.T) {
	backend := backend.NewStateBackend()
	middleware := NewAgentMemMiddleware(backend)

	if middleware.Name() != "AgentConfig" {
		t.Errorf("中间件名称不匹配，期望 %q，实际 %q", "AgentConfig", middleware.Name())
	}
}
