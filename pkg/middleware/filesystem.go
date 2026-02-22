package middleware

import (
	"context"
	"fmt"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/internal/stringutil"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// FilesystemMiddleware 文件系统中间件（简化版）
type FilesystemMiddleware struct {
	*BaseMiddleware
	backend      backend.Backend
	toolRegistry *tools.Registry
}

// NewFilesystemMiddleware 创建文件系统中间件
func NewFilesystemMiddleware(backend backend.Backend, toolRegistry *tools.Registry) *FilesystemMiddleware {
	m := &FilesystemMiddleware{
		BaseMiddleware: NewBaseMiddleware("filesystem"),
		backend:        backend,
		toolRegistry:   toolRegistry,
	}

	// 注册文件系统工具
	m.registerTools()

	return m
}

// registerTools 注册文件系统工具
func (m *FilesystemMiddleware) registerTools() {
	// 注册基础工具
	m.toolRegistry.Register(tools.NewListFilesTool(m.backend))
	m.toolRegistry.Register(tools.NewReadFileTool(m.backend))
	m.toolRegistry.Register(tools.NewWriteFileTool(m.backend))
	m.toolRegistry.Register(tools.NewEditFileTool(m.backend))
	m.toolRegistry.Register(tools.NewGrepTool(m.backend))
	m.toolRegistry.Register(tools.NewGlobTool(m.backend))
	m.toolRegistry.Register(tools.NewBashTool())
}

// AfterTool 在工具执行后检查结果大小
func (m *FilesystemMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	// 检查结果大小（20,000 tokens ≈ 80,000 字符）
	if len(result.Content) > 80000 {
		// 保存到大结果文件
		filePath := fmt.Sprintf("/large_tool_results/%s", result.ToolCallID)
		_, err := m.backend.WriteFile(ctx, filePath, result.Content)
		if err != nil {
			return fmt.Errorf("failed to save large result: %w", err)
		}

		// 创建预览（前5行 + 后5行）
		preview := createPreview(result.Content, 5, 5)
		result.Content = fmt.Sprintf("Result too large, saved to %s\n\nPreview:\n%s", filePath, preview)
	}

	return nil
}

// createPreview 创建内容预览
func createPreview(content string, headLines, tailLines int) string {
	lines := stringutil.SplitLines(content)
	if len(lines) <= headLines+tailLines {
		return content
	}

	var preview string
	// 前几行
	for i := 0; i < headLines && i < len(lines); i++ {
		preview += lines[i] + "\n"
	}

	preview += fmt.Sprintf("\n... (%d lines omitted) ...\n\n", len(lines)-headLines-tailLines)

	// 后几行
	start := len(lines) - tailLines
	for i := start; i < len(lines); i++ {
		preview += lines[i] + "\n"
	}

	return preview
}
