package stringutil

// SplitLines splits a string into lines by newline characters.
// It preserves empty lines and handles strings without trailing newlines.
func SplitLines(s string) []string {
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

// JoinLines joins a slice of strings with newline characters.
func JoinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// IndexOf returns the index of the first occurrence of substr in s, or -1 if not found.
func IndexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Contains reports whether substr is within s.
func Contains(s, substr string) bool {
	return IndexOf(s, substr) >= 0
}
