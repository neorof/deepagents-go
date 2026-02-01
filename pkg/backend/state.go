package backend

import (
	"context"
	"fmt"
	"sync"
)

// StateBackend 实现基于内存的状态存储
type StateBackend struct {
	mu    sync.RWMutex
	files map[string]string // path -> content
}

// NewStateBackend 创建状态后端
func NewStateBackend() *StateBackend {
	return &StateBackend{
		files: make(map[string]string),
	}
}

// ListFiles 列出目录下的文件
func (b *StateBackend) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var files []FileInfo
	for p := range b.files {
		// 简化实现：返回所有文件
		files = append(files, FileInfo{
			Path:  p,
			IsDir: false,
		})
	}
	return files, nil
}

// ReadFile 读取文件内容
func (b *StateBackend) ReadFile(ctx context.Context, path string, offset, limit int) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	content, ok := b.files[path]
	if !ok {
		return "", fmt.Errorf("file not found: %s", path)
	}

	// 处理偏移和限制
	if offset > 0 || limit > 0 {
		lines := splitLines(content)
		if offset >= len(lines) {
			return "", nil
		}
		end := len(lines)
		if limit > 0 && offset+limit < end {
			end = offset + limit
		}
		return joinLines(lines[offset:end]), nil
	}

	return content, nil
}

// WriteFile 写入文件
func (b *StateBackend) WriteFile(ctx context.Context, path, content string) (*WriteResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.files[path] = content
	return &WriteResult{
		Path:         path,
		BytesWritten: len(content),
	}, nil
}

// EditFile 编辑文件
func (b *StateBackend) EditFile(ctx context.Context, path, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	content, ok := b.files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	oldContent := content
	var replacements int

	if replaceAll {
		// 替换所有匹配
		newContent := ""
		for {
			idx := indexOf(content, oldStr)
			if idx == -1 {
				newContent += content
				break
			}
			newContent += content[:idx] + newStr
			content = content[idx+len(oldStr):]
			replacements++
		}
		content = newContent
	} else {
		// 只替换第一个匹配
		idx := indexOf(content, oldStr)
		if idx == -1 {
			return nil, fmt.Errorf("string not found: %s", oldStr)
		}
		content = content[:idx] + newStr + content[idx+len(oldStr):]
		replacements = 1
	}

	b.files[path] = content
	return &EditResult{
		Path:         path,
		Replacements: replacements,
		OldContent:   oldContent,
		NewContent:   content,
	}, nil
}

// Grep 搜索文件内容
func (b *StateBackend) Grep(ctx context.Context, pattern, path, glob string) ([]GrepMatch, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var matches []GrepMatch
	for p, content := range b.files {
		lines := splitLines(content)
		for i, line := range lines {
			if contains(line, pattern) {
				matches = append(matches, GrepMatch{
					Path:       p,
					LineNumber: i + 1,
					Line:       line,
					Match:      pattern,
				})
			}
		}
	}
	return matches, nil
}

// Glob 查找匹配的文件
func (b *StateBackend) Glob(ctx context.Context, pattern, path string) ([]FileInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var files []FileInfo
	for p := range b.files {
		// 简化实现：返回所有文件
		files = append(files, FileInfo{
			Path:  p,
			IsDir: false,
		})
	}
	return files, nil
}

// Execute 不支持命令执行
func (b *StateBackend) Execute(ctx context.Context, command string, timeout int) (*ExecuteResult, error) {
	return nil, fmt.Errorf("execute not supported by StateBackend")
}

// 辅助函数
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	return indexOf(s, substr) != -1
}
