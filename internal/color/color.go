package color

import (
	"os"
	"strings"
)

// ANSI 颜色码
const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[90m"
	bold    = "\033[1m"
)

// Enabled 指示是否启用颜色输出
var Enabled = isColorSupported()

// isColorSupported 检测终端是否支持颜色
func isColorSupported() bool {
	// 检查 NO_COLOR 环境变量
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// 检查 TERM 环境变量
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// 检查是否为终端
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	return true
}

// Colorize 使用指定颜色包装文本
func Colorize(colorCode, text string) string {
	if !Enabled || text == "" {
		return text
	}
	return colorCode + text + reset
}

// Red 返回红色文本
func Red(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return red + text + reset
}

// Green 返回绿色文本
func Green(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return green + text + reset
}

// Yellow 返回黄色文本
func Yellow(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return yellow + text + reset
}

// Blue 返回蓝色文本
func Blue(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return blue + text + reset
}

// Cyan 返回青色文本
func Cyan(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return cyan + text + reset
}

// Gray 返回灰色文本
func Gray(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return gray + text + reset
}

// Magenta 返回品红色文本
func Magenta(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return magenta + text + reset
}

// Bold 返回粗体文本
func Bold(text string) string {
	if !Enabled || text == "" {
		return text
	}
	return bold + text + reset
}

// DisableColor 禁用颜色输出
func DisableColor() {
	Enabled = false
}

// EnableColor 启用颜色输出
func EnableColor() {
	Enabled = true
}

// StripColors 移除文本中的 ANSI 颜色码
func StripColors(text string) string {
	// 简单的实现：移除常见的 ANSI 转义序列
	result := text
	colorCodes := []string{reset, red, green, yellow, blue, magenta, cyan, gray, bold}
	for _, code := range colorCodes {
		result = strings.ReplaceAll(result, code, "")
	}
	return result
}
