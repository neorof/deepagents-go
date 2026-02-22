package integration

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func TestWebMiddleware_Integration(t *testing.T) {
	// 创建工具注册表
	toolRegistry := tools.NewRegistry()

	// 创建 Web 配置
	webConfig := middleware.DefaultWebConfig()

	// 创建 Web 中间件
	webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

	// 验证中间件名称
	if webMiddleware.Name() != "web" {
		t.Errorf("Expected middleware name 'web', got %q", webMiddleware.Name())
	}

	// 验证工具已注册
	webSearchTool, ok := toolRegistry.Get("web_search")
	if !ok {
		t.Fatal("web_search tool not registered")
	}

	webFetchTool, ok := toolRegistry.Get("web_fetch")
	if !ok {
		t.Fatal("web_fetch tool not registered")
	}

	// 验证工具属性
	if webSearchTool.Name() != "web_search" {
		t.Errorf("Expected tool name 'web_search', got %q", webSearchTool.Name())
	}

	if webFetchTool.Name() != "web_fetch" {
		t.Errorf("Expected tool name 'web_fetch', got %q", webFetchTool.Name())
	}

	// 验证工具描述
	if webSearchTool.Description() == "" {
		t.Error("web_search tool should have a description")
	}

	if webFetchTool.Description() == "" {
		t.Error("web_fetch tool should have a description")
	}

	// 验证工具参数
	searchParams := webSearchTool.Parameters()
	if searchParams == nil {
		t.Fatal("web_search tool should have parameters")
	}

	fetchParams := webFetchTool.Parameters()
	if fetchParams == nil {
		t.Fatal("web_fetch tool should have parameters")
	}

	_ = context.Background() // 避免未使用变量警告
}

func TestWebConfig_Default(t *testing.T) {
	cfg := middleware.DefaultWebConfig()

	if cfg.SearchEngine != "duckduckgo" {
		t.Errorf("Expected default search engine 'duckduckgo', got %q", cfg.SearchEngine)
	}

	if cfg.DefaultTimeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", cfg.DefaultTimeout)
	}

	if cfg.MaxContentLength != 100000 {
		t.Errorf("Expected max content length 100000, got %d", cfg.MaxContentLength)
	}

	if !cfg.EnableReadability {
		t.Error("Expected readability to be enabled by default")
	}
}

func TestWebConfig_Custom(t *testing.T) {
	cfg := middleware.WebConfig{
		SearchEngine:      "serpapi",
		SerpAPIKey:        "test-key",
		DefaultTimeout:    60,
		MaxContentLength:  50000,
		EnableReadability: false,
	}

	toolRegistry := tools.NewRegistry()
	webMiddleware := middleware.NewWebMiddleware(toolRegistry, cfg)

	if webMiddleware.Name() != "web" {
		t.Errorf("Expected middleware name 'web', got %q", webMiddleware.Name())
	}

	// 验证工具已注册
	if _, ok := toolRegistry.Get("web_search"); !ok {
		t.Error("web_search tool should be registered")
	}

	if _, ok := toolRegistry.Get("web_fetch"); !ok {
		t.Error("web_fetch tool should be registered")
	}
}
