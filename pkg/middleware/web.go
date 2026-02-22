package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// WebConfig Web 工具配置
type WebConfig struct {
	SearchEngine      string `yaml:"search_engine" json:"search_engine"`           // 搜索引擎类型：duckduckgo, serpapi
	SerpAPIKey        string `yaml:"serpapi_key" json:"serpapi_key"`               // SerpAPI Key（可选）
	DefaultTimeout    int    `yaml:"default_timeout" json:"default_timeout"`       // 默认超时时间（秒）
	MaxContentLength  int    `yaml:"max_content_length" json:"max_content_length"` // 最大内容长度（字符）
	EnableReadability bool   `yaml:"enable_readability" json:"enable_readability"` // 是否启用智能内容提取
}

// DefaultWebConfig 返回默认 Web 配置
func DefaultWebConfig() WebConfig {
	return WebConfig{
		SearchEngine:      "duckduckgo",
		SerpAPIKey:        "",
		DefaultTimeout:    30,
		MaxContentLength:  100000,
		EnableReadability: true,
	}
}

// WebMiddleware Web 中间件
type WebMiddleware struct {
	*BaseMiddleware
	toolRegistry *tools.Registry
	config       WebConfig
}

// NewWebMiddleware 创建 Web 中间件
func NewWebMiddleware(toolRegistry *tools.Registry, cfg WebConfig) *WebMiddleware {
	m := &WebMiddleware{
		BaseMiddleware: NewBaseMiddleware("web"),
		toolRegistry:   toolRegistry,
		config:         cfg,
	}
	m.registerTools()
	return m
}

// registerTools 注册 Web 工具
func (m *WebMiddleware) registerTools() {
	engine := m.createSearchEngine()
	m.toolRegistry.Register(tools.NewWebSearchTool(engine))
	m.toolRegistry.Register(tools.NewWebFetchTool(m.config.EnableReadability, m.config.MaxContentLength))
}

// createSearchEngine 根据配置创建搜索引擎
func (m *WebMiddleware) createSearchEngine() tools.SearchEngine {
	if m.config.SearchEngine == "serpapi" {
		return tools.NewSerpAPIEngine(m.config.SerpAPIKey)
	}
	timeout := time.Duration(m.config.DefaultTimeout) * time.Second
	return tools.NewDuckDuckGoEngine(timeout)
}

// AfterTool 在工具执行后检查结果大小
func (m *WebMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	if len(result.Content) <= m.config.MaxContentLength {
		return nil
	}

	originalLen := len(result.Content)
	result.Content = fmt.Sprintf("%s\n\n... (内容已截断，原始长度: %d 字符，已显示: %d 字符)",
		result.Content[:m.config.MaxContentLength],
		originalLen,
		m.config.MaxContentLength,
	)
	return nil
}
