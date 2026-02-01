package backend

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// SandboxConfig 沙箱配置
type SandboxConfig struct {
	// RootDir 根目录（沙箱工作目录）
	RootDir string

	// ReadOnly 只读模式
	ReadOnly bool

	// AllowedPaths 允许访问的路径（白名单）
	AllowedPaths []string

	// BlockedPaths 禁止访问的路径（黑名单）
	BlockedPaths []string

	// MaxFileSize 最大文件大小（字节）
	MaxFileSize int64

	// MaxOperations 最大操作次数（0 表示无限制）
	MaxOperations int

	// OperationTimeout 单个操作超时时间
	OperationTimeout time.Duration

	// AllowedCommands 允许执行的命令（白名单）
	AllowedCommands []string

	// EnableAuditLog 启用审计日志
	EnableAuditLog bool
}

// DefaultSandboxConfig 返回默认配置
func DefaultSandboxConfig(rootDir string) *SandboxConfig {
	return &SandboxConfig{
		RootDir:          rootDir,
		ReadOnly:         false,
		AllowedPaths:     []string{},       // 空表示允许所有（在 rootDir 内）
		BlockedPaths:     []string{},       // 默认没有黑名单
		MaxFileSize:      10 * 1024 * 1024, // 10MB
		MaxOperations:    1000,
		OperationTimeout: 30 * time.Second,
		AllowedCommands:  []string{"ls", "cat", "echo", "pwd"}, // 安全的命令
		EnableAuditLog:   true,
	}
}

// SandboxBackend 沙箱后端，提供安全隔离
type SandboxBackend struct {
	config     *SandboxConfig
	backend    *FilesystemBackend
	mu         sync.RWMutex
	operations int // 操作计数器
	auditLog   []AuditEntry
}

// AuditEntry 审计日志条目
type AuditEntry struct {
	Timestamp time.Time
	Operation string
	Path      string
	Success   bool
	Error     string
}

// NewSandboxBackend 创建沙箱后端
func NewSandboxBackend(config *SandboxConfig) (*SandboxBackend, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 创建底层文件系统后端（使用虚拟模式）
	backend, err := NewFilesystemBackend(config.RootDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem backend: %w", err)
	}

	return &SandboxBackend{
		config:   config,
		backend:  backend,
		auditLog: make([]AuditEntry, 0),
	}, nil
}

// checkOperation 检查操作是否允许
func (b *SandboxBackend) checkOperation(ctx context.Context, _ /* operation */, path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查操作次数限制
	if b.config.MaxOperations > 0 && b.operations >= b.config.MaxOperations {
		return fmt.Errorf("operation limit exceeded: %d", b.config.MaxOperations)
	}

	// 检查超时
	if b.config.OperationTimeout > 0 {
		deadline, ok := ctx.Deadline()
		if !ok {
			// 设置默认超时
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, b.config.OperationTimeout)
			defer cancel()
		} else if time.Until(deadline) > b.config.OperationTimeout {
			// 缩短超时时间
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, b.config.OperationTimeout)
			defer cancel()
		}
	}

	// 检查路径是否在黑名单中
	for _, blocked := range b.config.BlockedPaths {
		if strings.HasPrefix(path, blocked) {
			return fmt.Errorf("path is blocked: %s", path)
		}
	}

	// 检查路径是否在白名单中（如果有白名单）
	if len(b.config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range b.config.AllowedPaths {
			if strings.HasPrefix(path, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path not in allowed list: %s", path)
		}
	}

	// 增加操作计数
	b.operations++

	return nil
}

// audit 记录审计日志
func (b *SandboxBackend) audit(operation, path string, success bool, err error) {
	if !b.config.EnableAuditLog {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	entry := AuditEntry{
		Timestamp: time.Now(),
		Operation: operation,
		Path:      path,
		Success:   success,
	}
	if err != nil {
		entry.Error = err.Error()
	}

	b.auditLog = append(b.auditLog, entry)
}

// GetAuditLog 获取审计日志
func (b *SandboxBackend) GetAuditLog() []AuditEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 返回副本
	log := make([]AuditEntry, len(b.auditLog))
	copy(log, b.auditLog)
	return log
}

// GetOperationCount 获取操作计数
func (b *SandboxBackend) GetOperationCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.operations
}

// ResetOperationCount 重置操作计数
func (b *SandboxBackend) ResetOperationCount() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.operations = 0
}

// ListFiles 列出目录下的文件
func (b *SandboxBackend) ListFiles(ctx context.Context, path string) ([]FileInfo, error) {
	if err := b.checkOperation(ctx, "ListFiles", path); err != nil {
		b.audit("ListFiles", path, false, err)
		return nil, err
	}

	files, err := b.backend.ListFiles(ctx, path)
	b.audit("ListFiles", path, err == nil, err)
	return files, err
}

// ReadFile 读取文件内容
func (b *SandboxBackend) ReadFile(ctx context.Context, path string, offset, limit int) (string, error) {
	if err := b.checkOperation(ctx, "ReadFile", path); err != nil {
		b.audit("ReadFile", path, false, err)
		return "", err
	}

	content, err := b.backend.ReadFile(ctx, path, offset, limit)
	b.audit("ReadFile", path, err == nil, err)
	return content, err
}

// WriteFile 写入文件
func (b *SandboxBackend) WriteFile(ctx context.Context, path, content string) (*WriteResult, error) {
	if b.config.ReadOnly {
		err := fmt.Errorf("write operation not allowed in read-only mode")
		b.audit("WriteFile", path, false, err)
		return nil, err
	}

	if err := b.checkOperation(ctx, "WriteFile", path); err != nil {
		b.audit("WriteFile", path, false, err)
		return nil, err
	}

	// 检查文件大小
	if b.config.MaxFileSize > 0 && int64(len(content)) > b.config.MaxFileSize {
		err := fmt.Errorf("file size exceeds limit: %d > %d", len(content), b.config.MaxFileSize)
		b.audit("WriteFile", path, false, err)
		return nil, err
	}

	result, err := b.backend.WriteFile(ctx, path, content)
	b.audit("WriteFile", path, err == nil, err)
	return result, err
}

// EditFile 编辑文件
func (b *SandboxBackend) EditFile(ctx context.Context, path, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	if b.config.ReadOnly {
		err := fmt.Errorf("edit operation not allowed in read-only mode")
		b.audit("EditFile", path, false, err)
		return nil, err
	}

	if err := b.checkOperation(ctx, "EditFile", path); err != nil {
		b.audit("EditFile", path, false, err)
		return nil, err
	}

	result, err := b.backend.EditFile(ctx, path, oldStr, newStr, replaceAll)

	// 检查编辑后的文件大小
	if err == nil && b.config.MaxFileSize > 0 && int64(len(result.NewContent)) > b.config.MaxFileSize {
		// 回滚编辑
		_, _ = b.backend.WriteFile(ctx, path, result.OldContent)
		err = fmt.Errorf("edited file size exceeds limit: %d > %d", len(result.NewContent), b.config.MaxFileSize)
		b.audit("EditFile", path, false, err)
		return nil, err
	}

	b.audit("EditFile", path, err == nil, err)
	return result, err
}

// Grep 搜索文件内容
func (b *SandboxBackend) Grep(ctx context.Context, pattern, path, glob string) ([]GrepMatch, error) {
	if err := b.checkOperation(ctx, "Grep", path); err != nil {
		b.audit("Grep", path, false, err)
		return nil, err
	}

	matches, err := b.backend.Grep(ctx, pattern, path, glob)
	b.audit("Grep", path, err == nil, err)
	return matches, err
}

// Glob 查找匹配的文件
func (b *SandboxBackend) Glob(ctx context.Context, pattern, path string) ([]FileInfo, error) {
	if err := b.checkOperation(ctx, "Glob", path); err != nil {
		b.audit("Glob", path, false, err)
		return nil, err
	}

	files, err := b.backend.Glob(ctx, pattern, path)
	b.audit("Glob", path, err == nil, err)
	return files, err
}

// Execute 执行命令
func (b *SandboxBackend) Execute(ctx context.Context, command string, timeout int) (*ExecuteResult, error) {
	if b.config.ReadOnly {
		err := fmt.Errorf("execute operation not allowed in read-only mode")
		b.audit("Execute", command, false, err)
		return nil, err
	}

	if err := b.checkOperation(ctx, "Execute", command); err != nil {
		b.audit("Execute", command, false, err)
		return nil, err
	}

	// 检查命令是否在白名单中
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		err := fmt.Errorf("empty command")
		b.audit("Execute", command, false, err)
		return nil, err
	}

	cmdName := filepath.Base(cmdParts[0])
	if !slices.Contains(b.config.AllowedCommands, cmdName) {
		err := fmt.Errorf("command not allowed: %s", cmdName)
		b.audit("Execute", command, false, err)
		return nil, err
	}

	// 设置超时
	execTimeout := time.Duration(timeout) * time.Millisecond
	if execTimeout == 0 || execTimeout > b.config.OperationTimeout {
		execTimeout = b.config.OperationTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, execTimeout)
	defer cancel()

	// 执行命令
	cmd := exec.CommandContext(execCtx, cmdParts[0], cmdParts[1:]...)
	cmd.Dir = b.config.RootDir

	output, err := cmd.CombinedOutput()

	result := &ExecuteResult{
		Stdout:   string(output),
		Stderr:   "",
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Stderr = string(output)
		} else {
			b.audit("Execute", command, false, err)
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	b.audit("Execute", command, err == nil, err)
	return result, nil
}
