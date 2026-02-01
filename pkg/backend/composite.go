package backend

import (
	"context"
	"fmt"
	"strings"
)

// CompositeBackend 组合多个后端，根据路径前缀路由
type CompositeBackend struct {
	routes         []route
	defaultBackend Backend
}

type route struct {
	prefix  string
	backend Backend
}

// NewCompositeBackend 创建组合后端
func NewCompositeBackend(defaultBackend Backend) *CompositeBackend {
	return &CompositeBackend{
		routes:         []route{},
		defaultBackend: defaultBackend,
	}
}

// AddRoute 添加路由规则
func (b *CompositeBackend) AddRoute(prefix string, backend Backend) {
	b.routes = append(b.routes, route{
		prefix:  prefix,
		backend: backend,
	})

	// 按前缀长度排序（最长的优先）
	for i := len(b.routes) - 1; i > 0; i-- {
		if len(b.routes[i].prefix) > len(b.routes[i-1].prefix) {
			b.routes[i], b.routes[i-1] = b.routes[i-1], b.routes[i]
		}
	}
}

// getBackendAndKey 根据路径获取对应的后端和相对路径
func (b *CompositeBackend) getBackendAndKey(key string) (Backend, string) {
	// 最长前缀匹配
	for _, route := range b.routes {
		if strings.HasPrefix(key, route.prefix) {
			suffix := strings.TrimPrefix(key, route.prefix)
			if suffix == "" {
				suffix = "/"
			}
			return route.backend, suffix
		}
	}
	return b.defaultBackend, key
}

// ListFiles 列出目录下的文件
func (b *CompositeBackend) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	backend, relPath := b.getBackendAndKey(path)
	files, err := backend.ListFiles(ctx, relPath)
	if err != nil {
		return nil, err
	}

	// 调整路径前缀
	for i := range files {
		if !strings.HasPrefix(files[i].Path, path) {
			files[i].Path = path + strings.TrimPrefix(files[i].Path, "/")
		}
	}

	return files, nil
}

// ReadFile 读取文件内容
func (b *CompositeBackend) ReadFile(ctx context.Context, path string, offset, limit int) (string, error) {
	backend, relPath := b.getBackendAndKey(path)
	return backend.ReadFile(ctx, relPath, offset, limit)
}

// WriteFile 写入文件
func (b *CompositeBackend) WriteFile(ctx context.Context, path, content string) (*WriteResult, error) {
	backend, relPath := b.getBackendAndKey(path)
	result, err := backend.WriteFile(ctx, relPath, content)
	if err != nil {
		return nil, err
	}

	// 调整路径
	result.Path = path
	return result, nil
}

// EditFile 编辑文件
func (b *CompositeBackend) EditFile(ctx context.Context, path, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	backend, relPath := b.getBackendAndKey(path)
	result, err := backend.EditFile(ctx, relPath, oldStr, newStr, replaceAll)
	if err != nil {
		return nil, err
	}

	// 调整路径
	result.Path = path
	return result, nil
}

// Grep 搜索文件内容
func (b *CompositeBackend) Grep(ctx context.Context, pattern, path, glob string) ([]GrepMatch, error) {
	if path == "" {
		// 搜索所有后端
		var allMatches []GrepMatch

		// 搜索默认后端
		matches, err := b.defaultBackend.Grep(ctx, pattern, "", glob)
		if err == nil {
			allMatches = append(allMatches, matches...)
		}

		// 搜索所有路由后端
		for _, route := range b.routes {
			matches, err := route.backend.Grep(ctx, pattern, "", glob)
			if err == nil {
				// 调整路径前缀
				for i := range matches {
					matches[i].Path = route.prefix + strings.TrimPrefix(matches[i].Path, "/")
				}
				allMatches = append(allMatches, matches...)
			}
		}

		return allMatches, nil
	}

	// 搜索特定路径
	backend, relPath := b.getBackendAndKey(path)
	matches, err := backend.Grep(ctx, pattern, relPath, glob)
	if err != nil {
		return nil, err
	}

	// 调整路径前缀
	for i := range matches {
		if !strings.HasPrefix(matches[i].Path, path) {
			matches[i].Path = path + strings.TrimPrefix(matches[i].Path, "/")
		}
	}

	return matches, nil
}

// Glob 查找匹配的文件
func (b *CompositeBackend) Glob(ctx context.Context, pattern, path string) ([]FileInfo, error) {
	if path == "" {
		// 搜索所有后端
		var allFiles []FileInfo

		// 搜索默认后端
		files, err := b.defaultBackend.Glob(ctx, pattern, "")
		if err == nil {
			allFiles = append(allFiles, files...)
		}

		// 搜索所有路由后端
		for _, route := range b.routes {
			files, err := route.backend.Glob(ctx, pattern, "")
			if err == nil {
				// 调整路径前缀
				for i := range files {
					files[i].Path = route.prefix + strings.TrimPrefix(files[i].Path, "/")
				}
				allFiles = append(allFiles, files...)
			}
		}

		return allFiles, nil
	}

	// 搜索特定路径
	backend, relPath := b.getBackendAndKey(path)
	files, err := backend.Glob(ctx, pattern, relPath)
	if err != nil {
		return nil, err
	}

	// 调整路径前缀
	for i := range files {
		if !strings.HasPrefix(files[i].Path, path) {
			files[i].Path = path + strings.TrimPrefix(files[i].Path, "/")
		}
	}

	return files, nil
}

// Execute 执行命令
func (b *CompositeBackend) Execute(ctx context.Context, command string, timeout int) (*ExecuteResult, error) {
	return nil, fmt.Errorf("execute not supported by CompositeBackend")
}
