package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/backend"
)

// NewListFilesTool 创建 ls 工具
func NewListFilesTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"ls",
		"列出目录下的文件和子目录",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "目录路径",
				},
			},
			"required": []string{"path"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("path must be a string")
			}

			files, err := backend.ListFiles(ctx, path)
			if err != nil {
				return "", err
			}

			result := fmt.Sprintf("Files in %s:\n", path)
			for _, file := range files {
				if file.IsDir {
					result += fmt.Sprintf("  [DIR]  %s\n", file.Path)
				} else {
					result += fmt.Sprintf("  [FILE] %s (%d bytes)\n", file.Path, file.Size)
				}
			}
			return result, nil
		},
	)
}

// NewReadFileTool 创建 read_file 工具
func NewReadFileTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"read_file",
		"读取文件内容",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "文件路径",
				},
				"offset": map[string]any{
					"type":        "integer",
					"description": "起始行号（可选）",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "读取行数（可选）",
				},
			},
			"required": []string{"path"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("path must be a string")
			}

			offset := 0
			if v, ok := args["offset"].(float64); ok {
				offset = int(v)
			}

			limit := 0
			if v, ok := args["limit"].(float64); ok {
				limit = int(v)
			}

			content, err := backend.ReadFile(ctx, path, offset, limit)
			if err != nil {
				return "", err
			}

			return content, nil
		},
	)
}

// NewWriteFileTool 创建 write_file 工具
func NewWriteFileTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"write_file",
		"写入文件内容",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "文件路径",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "文件内容",
				},
			},
			"required": []string{"path", "content"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("path must be a string")
			}

			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("content must be a string")
			}

			result, err := backend.WriteFile(ctx, path, content)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("Successfully wrote %d bytes to %s", result.BytesWritten, result.Path), nil
		},
	)
}

// NewEditFileTool 创建 edit_file 工具
func NewEditFileTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"edit_file",
		"编辑文件（字符串替换）",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "文件路径",
				},
				"old_string": map[string]any{
					"type":        "string",
					"description": "要替换的字符串",
				},
				"new_string": map[string]any{
					"type":        "string",
					"description": "新字符串",
				},
				"replace_all": map[string]any{
					"type":        "boolean",
					"description": "是否替换所有匹配（默认 false）",
				},
			},
			"required": []string{"path", "old_string", "new_string"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			path, ok := args["path"].(string)
			if !ok {
				return "", fmt.Errorf("path must be a string")
			}

			oldStr, ok := args["old_string"].(string)
			if !ok {
				return "", fmt.Errorf("old_string must be a string")
			}

			newStr, ok := args["new_string"].(string)
			if !ok {
				return "", fmt.Errorf("new_string must be a string")
			}

			replaceAll := false
			if v, ok := args["replace_all"].(bool); ok {
				replaceAll = v
			}

			result, err := backend.EditFile(ctx, path, oldStr, newStr, replaceAll)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", result.Replacements, result.Path), nil
		},
	)
}

// NewGrepTool 创建 grep 工具
func NewGrepTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"grep",
		"搜索文件内容",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "搜索模式",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "搜索路径（可选）",
				},
				"glob": map[string]any{
					"type":        "string",
					"description": "文件匹配模式（可选）",
				},
			},
			"required": []string{"pattern"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			pattern, ok := args["pattern"].(string)
			if !ok {
				return "", fmt.Errorf("pattern must be a string")
			}

			path := ""
			if v, ok := args["path"].(string); ok {
				path = v
			}

			glob := ""
			if v, ok := args["glob"].(string); ok {
				glob = v
			}

			matches, err := backend.Grep(ctx, pattern, path, glob)
			if err != nil {
				return "", err
			}

			if len(matches) == 0 {
				return "No matches found", nil
			}

			result := fmt.Sprintf("Found %d matches:\n", len(matches))
			for _, match := range matches {
				result += fmt.Sprintf("%s:%d: %s\n", match.Path, match.LineNumber, match.Line)
			}
			return result, nil
		},
	)
}

// NewGlobTool 创建 glob 工具
func NewGlobTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"glob",
		"查找匹配的文件",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "文件匹配模式（如 *.go）",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "搜索路径（可选）",
				},
			},
			"required": []string{"pattern"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			pattern, ok := args["pattern"].(string)
			if !ok {
				return "", fmt.Errorf("pattern must be a string")
			}

			path := ""
			if v, ok := args["path"].(string); ok {
				path = v
			}

			files, err := backend.Glob(ctx, pattern, path)
			if err != nil {
				return "", err
			}

			if len(files) == 0 {
				return "No files found", nil
			}

			result := fmt.Sprintf("Found %d files:\n", len(files))
			for _, file := range files {
				result += fmt.Sprintf("  %s\n", file.Path)
			}
			return result, nil
		},
	)
}

// NewBashTool 创建 bash 工具
func NewBashTool() Tool {
	return NewBaseTool(
		"bash",
		"执行 bash 命令",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "要执行的 bash 命令",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "超时时间（秒），默认 30 秒",
				},
			},
			"required": []string{"command"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			command, ok := args["command"].(string)
			if !ok {
				return "", fmt.Errorf("command must be a string")
			}

			// 获取超时时间，默认 30 秒
			timeout := 30
			if v, ok := args["timeout"].(float64); ok {
				timeout = int(v)
			}

			// 创建带超时的 context
			execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
			defer cancel()

			// 执行命令
			cmd := exec.CommandContext(execCtx, "bash", "-c", command)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// 构建结果
			result := ""
			if stdout.Len() > 0 {
				result += "STDOUT:\n" + stdout.String()
			}
			if stderr.Len() > 0 {
				if result != "" {
					result += "\n\n"
				}
				result += "STDERR:\n" + stderr.String()
			}

			if err != nil {
				if result != "" {
					result += "\n\n"
				}
				result += fmt.Sprintf("ERROR: %v", err)

				// 检查是否超时
				if execCtx.Err() == context.DeadlineExceeded {
					result += fmt.Sprintf("\nCommand timed out after %d seconds", timeout)
				}
			}

			if result == "" {
				result = "Command executed successfully (no output)"
			}

			return result, nil
		},
	)
}
