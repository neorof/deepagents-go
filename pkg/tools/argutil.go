package tools

import "fmt"

// GetStringArg 从参数中获取字符串值，支持默认值
func GetStringArg(args map[string]any, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

// GetIntArg 从参数中获取整数值，支持默认值
func GetIntArg(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}

// GetBoolArg 从参数中获取布尔值，支持默认值
func GetBoolArg(args map[string]any, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}

// ValidateStringArg 验证必需的字符串参数
func ValidateStringArg(args map[string]any, key string) (string, error) {
	v, ok := args[key].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("%s 必须是非空字符串", key)
	}
	return v, nil
}

// Clamp 将值限制在指定范围内
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
