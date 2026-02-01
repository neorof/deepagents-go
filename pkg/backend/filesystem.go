package backend

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FilesystemBackend 实现基于真实文件系统的后端
type FilesystemBackend struct {
	rootDir     string
	virtualMode bool // 虚拟模式：限制在 rootDir 内
}

// NewFilesystemBackend 创建文件系统后端
func NewFilesystemBackend(rootDir string, virtualMode bool) (*FilesystemBackend, error) {
	// 确保根目录存在
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &FilesystemBackend{
		rootDir:     absRoot,
		virtualMode: virtualMode,
	}, nil
}

// resolvePath 解析路径
func (b *FilesystemBackend) resolvePath(key string) (string, error) {
	if b.virtualMode {
		// 阻止路径遍历
		if strings.Contains(key, "..") || strings.HasPrefix(key, "~") {
			return "", fmt.Errorf("path traversal not allowed")
		}
		// 限制在 root_dir 内
		full := filepath.Join(b.rootDir, strings.TrimPrefix(key, "/"))
		rel, err := filepath.Rel(b.rootDir, full)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("path outside root directory")
		}
		return full, nil
	}
	// 普通模式：支持绝对路径
	if filepath.IsAbs(key) {
		return key, nil
	}
	return filepath.Join(b.rootDir, key), nil
}

// ListFiles 列出目录下的文件
func (b *FilesystemBackend) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	fullPath, err := b.resolvePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Path:    filepath.Join(path, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Unix(),
		})
	}

	return files, nil
}

// ReadFile 读取文件内容
func (b *FilesystemBackend) ReadFile(ctx context.Context, path string, offset, limit int) (string, error) {
	fullPath, err := b.resolvePath(path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)

	// 处理偏移和限制
	if offset > 0 || limit > 0 {
		lines := strings.Split(content, "\n")
		if offset >= len(lines) {
			return "", nil
		}
		end := len(lines)
		if limit > 0 && offset+limit < end {
			end = offset + limit
		}
		return strings.Join(lines[offset:end], "\n"), nil
	}

	return content, nil
}

// WriteFile 写入文件
func (b *FilesystemBackend) WriteFile(ctx context.Context, path, content string) (*WriteResult, error) {
	fullPath, err := b.resolvePath(path)
	if err != nil {
		return nil, err
	}

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &WriteResult{
		Path:         path,
		BytesWritten: len(content),
	}, nil
}

// EditFile 编辑文件
func (b *FilesystemBackend) EditFile(ctx context.Context, path, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	fullPath, err := b.resolvePath(path)
	if err != nil {
		return nil, err
	}

	// 读取文件
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(data)
	newContent := oldContent
	replacements := 0

	if replaceAll {
		// 替换所有匹配
		replacements = strings.Count(oldContent, oldStr)
		newContent = strings.ReplaceAll(oldContent, oldStr, newStr)
	} else {
		// 只替换第一个匹配
		idx := strings.Index(oldContent, oldStr)
		if idx == -1 {
			return nil, fmt.Errorf("string not found: %s", oldStr)
		}
		newContent = oldContent[:idx] + newStr + oldContent[idx+len(oldStr):]
		replacements = 1
	}

	// 写回文件
	if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &EditResult{
		Path:         path,
		Replacements: replacements,
		OldContent:   oldContent,
		NewContent:   newContent,
	}, nil
}

// Grep 搜索文件内容
func (b *FilesystemBackend) Grep(ctx context.Context, pattern, path, glob string) ([]GrepMatch, error) {
	searchPath := b.rootDir
	if path != "" {
		fullPath, err := b.resolvePath(path)
		if err != nil {
			return nil, err
		}
		searchPath = fullPath
	}

	var matches []GrepMatch

	err := filepath.WalkDir(searchPath, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}

		if d.IsDir() {
			return nil
		}

		// 读取文件
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil // 忽略读取错误
		}

		// 搜索匹配
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, pattern) {
				relPath, _ := filepath.Rel(b.rootDir, filePath)
				matches = append(matches, GrepMatch{
					Path:       "/" + relPath,
					LineNumber: i + 1,
					Line:       line,
					Match:      pattern,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return matches, nil
}

// Glob 查找匹配的文件
func (b *FilesystemBackend) Glob(ctx context.Context, pattern, path string) ([]FileInfo, error) {
	searchPath := b.rootDir
	if path != "" {
		fullPath, err := b.resolvePath(path)
		if err != nil {
			return nil, err
		}
		searchPath = fullPath
	}

	globPattern := filepath.Join(searchPath, pattern)
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob: %w", err)
	}

	var files []FileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(b.rootDir, match)
		files = append(files, FileInfo{
			Path:    "/" + relPath,
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime().Unix(),
		})
	}

	return files, nil
}

// Execute 不支持命令执行
func (b *FilesystemBackend) Execute(ctx context.Context, command string, timeout int) (*ExecuteResult, error) {
	return nil, fmt.Errorf("execute not supported by FilesystemBackend")
}
