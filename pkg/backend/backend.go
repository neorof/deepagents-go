package backend

import "context"

// Backend 定义存储后端接口
type Backend interface {
	// ListFiles 列出目录下的文件
	ListFiles(ctx context.Context, path string) ([]FileInfo, error)

	// ReadFile 读取文件内容
	ReadFile(ctx context.Context, path string, offset, limit int) (string, error)

	// WriteFile 写入文件
	WriteFile(ctx context.Context, path, content string) (*WriteResult, error)

	// EditFile 编辑文件（字符串替换）
	EditFile(ctx context.Context, path, oldStr, newStr string, replaceAll bool) (*EditResult, error)

	// Grep 搜索文件内容
	Grep(ctx context.Context, pattern, path, glob string) ([]GrepMatch, error)

	// Glob 查找匹配的文件
	Glob(ctx context.Context, pattern, path string) ([]FileInfo, error)

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, path string) error

	// Execute 执行命令（可选，某些后端不支持）
	Execute(ctx context.Context, command string, timeout int) (*ExecuteResult, error)
}

// FileInfo 文件信息
type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime int64  `json:"mod_time"`
}

// WriteResult 写入结果
type WriteResult struct {
	Path         string `json:"path"`
	BytesWritten int    `json:"bytes_written"`
}

// EditResult 编辑结果
type EditResult struct {
	Path         string `json:"path"`
	Replacements int    `json:"replacements"`
	OldContent   string `json:"old_content,omitempty"`
	NewContent   string `json:"new_content,omitempty"`
}

// GrepMatch 搜索匹配结果
type GrepMatch struct {
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
	Match      string `json:"match"`
}

// ExecuteResult 命令执行结果
type ExecuteResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}
