package middleware

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// MemoryMiddleware 管理 Agent 的记忆
type MemoryMiddleware struct {
	*BaseMiddleware
	memories []Memory
}

// Memory 表示一条记忆
type Memory struct {
	Source  string // 记忆来源（文件路径或标识）
	Content string // 记忆内容
	Type    string // 记忆类型（system, user, context）
}

// NewMemoryMiddleware 创建记忆中间件
func NewMemoryMiddleware() *MemoryMiddleware {
	return &MemoryMiddleware{
		BaseMiddleware: NewBaseMiddleware("memory"),
		memories:       make([]Memory, 0),
	}
}

// LoadFromFile 从文件加载记忆
func (m *MemoryMiddleware) LoadFromFile(path string) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("记忆文件不存在: %s", path)
	}

	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取记忆文件失败: %w", err)
	}

	// 添加记忆
	memory := Memory{
		Source:  path,
		Content: string(content),
		Type:    "system",
	}

	m.memories = append(m.memories, memory)
	return nil
}

// LoadFromDirectory 从目录加载所有记忆文件
func (m *MemoryMiddleware) LoadFromDirectory(dir string) error {
	// 支持的文件扩展名
	extensions := []string{".md", ".txt"}

	// 遍历目录
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	loadedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 检查文件扩展名
		ext := filepath.Ext(entry.Name())
		supported := false
		for _, validExt := range extensions {
			if ext == validExt {
				supported = true
				break
			}
		}

		if !supported {
			continue
		}

		// 加载文件
		path := filepath.Join(dir, entry.Name())
		if err := m.LoadFromFile(path); err != nil {
			// 记录错误但继续加载其他文件
			continue
		}
		loadedCount++
	}

	if loadedCount == 0 {
		return fmt.Errorf("目录中没有找到记忆文件: %s", dir)
	}

	return nil
}

// AddMemory 添加记忆
func (m *MemoryMiddleware) AddMemory(source, content, memoryType string) {
	memory := Memory{
		Source:  source,
		Content: content,
		Type:    memoryType,
	}
	m.memories = append(m.memories, memory)
}

// GetMemories 获取所有记忆
func (m *MemoryMiddleware) GetMemories() []Memory {
	return m.memories
}

// GetMemoriesByType 按类型获取记忆
func (m *MemoryMiddleware) GetMemoriesByType(memoryType string) []Memory {
	result := make([]Memory, 0)
	for _, memory := range m.memories {
		if memory.Type == memoryType {
			result = append(result, memory)
		}
	}
	return result
}

// ClearMemories 清除所有记忆
func (m *MemoryMiddleware) ClearMemories() {
	m.memories = make([]Memory, 0)
}

// RemoveMemory 删除指定来源的记忆
func (m *MemoryMiddleware) RemoveMemory(source string) bool {
	for i, memory := range m.memories {
		if memory.Source == source {
			m.memories = append(m.memories[:i], m.memories[i+1:]...)
			return true
		}
	}
	return false
}

// BeforeModel 在调用模型前注入记忆到系统提示
func (m *MemoryMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	if len(m.memories) == 0 {
		return nil
	}

	// 构建记忆内容
	var memoryContent strings.Builder
	memoryContent.WriteString("\n\n=== Agent 记忆 ===\n")

	// 按类型组织记忆
	systemMemories := m.GetMemoriesByType("system")
	if len(systemMemories) > 0 {
		memoryContent.WriteString("\n## 系统记忆\n")
		for _, memory := range systemMemories {
			memoryContent.WriteString(memory.Content)
			memoryContent.WriteString("\n\n")
		}
	}

	contextMemories := m.GetMemoriesByType("context")
	if len(contextMemories) > 0 {
		memoryContent.WriteString("\n## 上下文记忆\n")
		for _, memory := range contextMemories {
			memoryContent.WriteString(memory.Content)
			memoryContent.WriteString("\n\n")
		}
	}

	// 注入到系统提示
	if req.SystemPrompt == "" {
		req.SystemPrompt = memoryContent.String()
	} else {
		req.SystemPrompt += memoryContent.String()
	}

	return nil
}

// BeforeAgent 在 Agent 开始前加载默认记忆
func (m *MemoryMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	// 尝试加载默认的 AGENTS.md 文件
	defaultPaths := []string{
		"AGENTS.md",
		".agents/AGENTS.md",
		"docs/AGENTS.md",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			// 检查是否已经加载过
			alreadyLoaded := false
			for _, memory := range m.memories {
				if memory.Source == path {
					alreadyLoaded = true
					break
				}
			}

			if !alreadyLoaded {
				m.LoadFromFile(path)
			}
			break
		}
	}

	return nil
}
