package agentkit

import (
	"fmt"
	"strings"

	"github.com/zhoucx/deepagents-go/internal/color"
)

// DefaultToolCallHandler 默认的工具调用回调处理器
func DefaultToolCallHandler(toolName string, input map[string]any) {
	fmt.Printf("\n%s %s", color.Cyan("Tool:"), color.Bold(toolName))

	// 提取并显示关键参数
	paramStr := extractKeyParam(input)
	if paramStr != "" {
		fmt.Printf(" %s", color.Gray(fmt.Sprintf("(%s)", paramStr)))
	}
	fmt.Println()
}

// DefaultToolResultHandler 默认的工具结果回调处理器
func DefaultToolResultHandler(toolName string, result string, isError bool) {
	if isError {
		displayResult := truncateByLines(result, 10)
		fmt.Printf("   %s %s\n", color.Red("❌ Error:"), displayResult)
	} else {
		displayResult := truncateByLines(result, 10)
		fmt.Printf("   %s %s\n", color.Green("✓ Done:"), color.Gray(displayResult))
	}
}

// truncateByLines 按行数截断文本，超过指定行数则添加省略号
func truncateByLines(text string, maxLines int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}
	return strings.Join(lines[:maxLines], "\n") + "\n..."
}

// extractKeyParam 从输入参数中提取关键参数用于显示
func extractKeyParam(input map[string]any) string {
	// 优先级：command > path > pattern > query > url
	keyParams := []string{"command", "path", "pattern", "query", "url"}

	for _, key := range keyParams {
		if val, ok := input[key].(string); ok && len(val) > 0 {
			// 对于 command 参数，如果太长则截断
			if key == "command" && len(val) > 50 {
				return fmt.Sprintf("%s=%s...", key, val[:50])
			}
			return fmt.Sprintf("%s=%s", key, val)
		}
	}

	return ""
}
