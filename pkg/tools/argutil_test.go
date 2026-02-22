package tools

import (
	"reflect"
	"testing"
)

func TestGetStringArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		key        string
		defaultVal string
		expected   string
	}{
		{
			name:       "string value exists",
			args:       map[string]any{"key": "value"},
			key:        "key",
			defaultVal: "default",
			expected:   "value",
		},
		{
			name:       "key not found",
			args:       map[string]any{},
			key:        "key",
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "wrong type",
			args:       map[string]any{"key": 123},
			key:        "key",
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "empty string",
			args:       map[string]any{"key": ""},
			key:        "key",
			defaultVal: "default",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringArg(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("GetStringArg() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetIntArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		key        string
		defaultVal int
		expected   int
	}{
		{
			name:       "float64 value exists",
			args:       map[string]any{"key": float64(42)},
			key:        "key",
			defaultVal: 10,
			expected:   42,
		},
		{
			name:       "key not found",
			args:       map[string]any{},
			key:        "key",
			defaultVal: 10,
			expected:   10,
		},
		{
			name:       "wrong type",
			args:       map[string]any{"key": "not a number"},
			key:        "key",
			defaultVal: 10,
			expected:   10,
		},
		{
			name:       "zero value",
			args:       map[string]any{"key": float64(0)},
			key:        "key",
			defaultVal: 10,
			expected:   0,
		},
		{
			name:       "negative value",
			args:       map[string]any{"key": float64(-5)},
			key:        "key",
			defaultVal: 10,
			expected:   -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIntArg(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("GetIntArg() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetBoolArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		key        string
		defaultVal bool
		expected   bool
	}{
		{
			name:       "bool value true",
			args:       map[string]any{"key": true},
			key:        "key",
			defaultVal: false,
			expected:   true,
		},
		{
			name:       "bool value false",
			args:       map[string]any{"key": false},
			key:        "key",
			defaultVal: true,
			expected:   false,
		},
		{
			name:       "key not found",
			args:       map[string]any{},
			key:        "key",
			defaultVal: true,
			expected:   true,
		},
		{
			name:       "wrong type",
			args:       map[string]any{"key": "not a bool"},
			key:        "key",
			defaultVal: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBoolArg(tt.args, tt.key, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("GetBoolArg() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateStringArg(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		key       string
		expected  string
		expectErr bool
	}{
		{
			name:      "valid string",
			args:      map[string]any{"key": "value"},
			key:       "key",
			expected:  "value",
			expectErr: false,
		},
		{
			name:      "empty string",
			args:      map[string]any{"key": ""},
			key:       "key",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "key not found",
			args:      map[string]any{},
			key:       "key",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "wrong type",
			args:      map[string]any{"key": 123},
			key:       "key",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateStringArg(tt.args, tt.key)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ValidateStringArg() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStringArg() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("ValidateStringArg() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		expected int
	}{
		{
			name:     "value within range",
			value:    5,
			min:      1,
			max:      10,
			expected: 5,
		},
		{
			name:     "value below min",
			value:    0,
			min:      1,
			max:      10,
			expected: 1,
		},
		{
			name:     "value above max",
			value:    15,
			min:      1,
			max:      10,
			expected: 10,
		},
		{
			name:     "value equals min",
			value:    1,
			min:      1,
			max:      10,
			expected: 1,
		},
		{
			name:     "value equals max",
			value:    10,
			min:      1,
			max:      10,
			expected: 10,
		},
		{
			name:     "negative range",
			value:    -5,
			min:      -10,
			max:      -1,
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d", tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestArgUtilIntegration(t *testing.T) {
	// 模拟真实场景：从 JSON 解析的参数
	args := map[string]any{
		"query":       "test query",
		"max_results": float64(15),
		"enabled":     true,
	}

	// 测试组合使用
	query := GetStringArg(args, "query", "")
	if query != "test query" {
		t.Errorf("query = %q, want %q", query, "test query")
	}

	maxResults := GetIntArg(args, "max_results", 5)
	maxResults = Clamp(maxResults, 1, 10)
	if maxResults != 10 {
		t.Errorf("maxResults = %d, want %d", maxResults, 10)
	}

	enabled := GetBoolArg(args, "enabled", false)
	if !enabled {
		t.Errorf("enabled = %v, want %v", enabled, true)
	}

	// 测试缺失的参数
	missing := GetStringArg(args, "missing", "default")
	if missing != "default" {
		t.Errorf("missing = %q, want %q", missing, "default")
	}
}

func TestValidateStringArgErrorMessage(t *testing.T) {
	args := map[string]any{"key": ""}
	_, err := ValidateStringArg(args, "test_key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expected := "test_key 必须是非空字符串"
	if err.Error() != expected {
		t.Errorf("error message = %q, want %q", err.Error(), expected)
	}
}

func TestGetIntArgWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected int
	}{
		{"float64", float64(42), 42},
		{"int", 42, 10}, // int 不会被识别，返回默认值
		{"string", "42", 10},
		{"nil", nil, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]any{"key": tt.value}
			result := GetIntArg(args, "key", 10)
			if result != tt.expected {
				t.Errorf("GetIntArg() with %T = %d, want %d", tt.value, result, tt.expected)
			}
		})
	}
}

func TestAllFunctionsWithNilArgs(t *testing.T) {
	var args map[string]any

	// 测试所有函数都能处理 nil map
	if s := GetStringArg(args, "key", "default"); s != "default" {
		t.Errorf("GetStringArg with nil map = %q, want %q", s, "default")
	}

	if i := GetIntArg(args, "key", 42); i != 42 {
		t.Errorf("GetIntArg with nil map = %d, want %d", i, 42)
	}

	if b := GetBoolArg(args, "key", true); !b {
		t.Errorf("GetBoolArg with nil map = %v, want %v", b, true)
	}

	if _, err := ValidateStringArg(args, "key"); err == nil {
		t.Error("ValidateStringArg with nil map should return error")
	}

	if c := Clamp(5, 1, 10); c != 5 {
		t.Errorf("Clamp(5, 1, 10) = %d, want %d", c, 5)
	}
}

func TestGetStringArgTypes(t *testing.T) {
	// 测试各种类型的值
	testCases := []struct {
		value    any
		expected string
	}{
		{value: "string", expected: "string"},
		{value: 123, expected: "default"},
		{value: true, expected: "default"},
		{value: []string{"a"}, expected: "default"},
		{value: map[string]string{"a": "b"}, expected: "default"},
		{value: nil, expected: "default"},
	}

	for _, tc := range testCases {
		typeName := "nil"
		if tc.value != nil {
			typeName = reflect.TypeOf(tc.value).String()
		}
		t.Run(typeName, func(t *testing.T) {
			args := map[string]any{"key": tc.value}
			result := GetStringArg(args, "key", "default")
			if result != tc.expected {
				t.Errorf("GetStringArg with %T = %q, want %q", tc.value, result, tc.expected)
			}
		})
	}
}
