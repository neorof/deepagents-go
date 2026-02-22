package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
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
		`读取文件内容。

**使用场景**：
- 查看文件内容
- 在编辑文件前必须先读取
- 了解代码结构和实现

**参数说明**：
- path: 文件的绝对或相对路径
- offset: 起始行号（可选，用于大文件分页读取）
- limit: 读取行数（可选，用于大文件分页读取）

**输出格式**：
- 每行格式：[空格][行号][Tab][实际内容]
- 示例：    42→    func example() {

**注意**：
- 使用 edit_file 前必须先调用此工具
- 对于大文件，可使用 offset 和 limit 分页读取`,
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
		`写入文件内容（创建或覆盖）。

**使用场景**：
- 创建新文件
- 完全覆盖现有文件（谨慎使用）

**使用原则**：
- 优先使用 edit_file 编辑现有文件
- 覆盖现有文件前必须先用 read_file 读取
- 不主动创建文档文件（README.md 等）

**参数说明**：
- path: 文件路径
- content: 要写入的完整内容`,
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
		`编辑文件（字符串替换）。

**使用前提**：
- 必须先用 read_file 读取文件内容
- 无例外：即使文件刚创建、内容已知，也必须先读取

**唯一性要求**：
- old_string 必须在文件中唯一，否则编辑失败
- 通过包含更多上下文（如函数名、注释）使 old_string 唯一
- 使用 replace_all: true 可替换所有匹配项（如重命名变量）

**参数说明**：
- path: 文件路径
- old_string: 要替换的字符串（必须与文件内容完全匹配）
- new_string: 替换后的新字符串
- replace_all: 是否替换所有匹配（默认 false）

**常见错误**：
- 未先读取文件
- old_string 不唯一
- 缩进不匹配（Tab vs 空格）
- 包含了行号前缀

**示例**：
要替换文件中 3 处相同代码中的 2 处，通过包含函数名使每处唯一：
old_string: "func a() {\n    x := 1"
new_string: "func a() {\n    x := 2"`,
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "文件路径",
				},
				"old_string": map[string]any{
					"type":        "string",
					"description": "要替换的字符串（必须与文件内容完全匹配）",
				},
				"new_string": map[string]any{
					"type":        "string",
					"description": "新字符串",
				},
				"replace_all": map[string]any{
					"type":        "boolean",
					"description": "是否替换所有匹配（默认 false，用于批量重命名时设为 true）",
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
		`搜索文件内容（支持正则表达式和 .gitignore）。

**使用场景**：
- 在代码库中搜索特定模式
- 查找函数定义、变量使用
- 定位错误信息来源

**参数说明**：
- pattern: 搜索模式（支持正则表达式，如 "func.*Handler"）
- path: 搜索路径（可选，默认当前目录）
- glob: 文件匹配模式（可选，如 "*.go"）

**正则表达式示例**：
- 搜索函数定义：pattern="func\\s+\\w+Handler"
- 搜索注释：pattern="//\\s*TODO"
- 多选一：pattern="error|Error|ERROR"
- 字面字符串：pattern="TODO"（简单字符串自动工作）

**过滤规则**：
- 自动遵循 .gitignore 规则（跳过被忽略的文件和目录）
- 始终跳过 .git 目录
- 跳过二进制文件和大文件（>1MB）

**注意**：
- 优先使用此工具而非 bash grep
- 返回匹配的文件路径、行号和内容
- 无效的正则表达式会自动降级为字面字符串匹配`,
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "搜索模式（支持正则表达式）",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "搜索路径（可选，默认当前目录）",
				},
				"glob": map[string]any{
					"type":        "string",
					"description": "文件匹配模式（可选，如 *.go）",
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

			var result strings.Builder
			fmt.Fprintf(&result, "Found %d matches:\n", len(matches))
			for _, match := range matches {
				fmt.Fprintf(&result, "%s:%d: %s\n", match.Path, match.LineNumber, match.Line)
			}
			return result.String(), nil
		},
	)
}

// NewGlobTool 创建 glob 工具
func NewGlobTool(backend backend.Backend) Tool {
	return NewBaseTool(
		"glob",
		`查找匹配的文件（支持通配符）。

**使用场景**：
- 查找特定类型的文件
- 探索项目结构
- 定位配置文件

**参数说明**：
- pattern: 文件匹配模式（支持 *, **, ? 通配符）
- path: 搜索路径（可选，默认当前目录）

**常用模式**：
- "*.go": 当前目录下的 Go 文件
- "**/*.go": 递归查找所有 Go 文件
- "src/**/*.ts": src 目录下所有 TypeScript 文件
- "*.{js,ts}": 所有 JS 和 TS 文件

**注意**：
- 优先使用此工具而非 bash find 或 ls
- 返回匹配的文件路径列表`,
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{
					"type":        "string",
					"description": "文件匹配模式（如 **/*.go）",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "搜索路径（可选，默认当前目录）",
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

			var result strings.Builder
			fmt.Fprintf(&result, "Found %d files:\n", len(files))
			for _, file := range files {
				fmt.Fprintf(&result, "  %s\n", file.Path)
			}
			return result.String(), nil
		},
	)
}

// NewBashTool 创建 bash 工具
func NewBashTool() Tool {
	return NewBaseTool(
		"bash",
		`执行 bash 命令。

**正确用途**：
- git 操作（git status, git commit, git push 等）
- 构建命令（go build, go test, npm install 等）
- 系统命令（docker, curl 等）

**禁止用途**：
- 文件读写（用 read_file, write_file, edit_file）
- 文件搜索（用 grep, glob）
- 不要用 cat, head, tail, sed, awk, find, ls

**命令组合**：
- 独立命令：可并行调用多次 bash
- 依赖命令：用 && 串联（如 git add && git commit）

**Git 安全规则**：
- 禁止 --force, --no-verify, reset --hard（除非用户明确要求）
- 优先 git add 具体文件而非 git add -A

**参数说明**：
- command: 要执行的命令
- timeout: 超时时间（秒），默认 30 秒`,
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
