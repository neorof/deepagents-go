package stringutil

import (
	"reflect"
	"testing"
)

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line without newline",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "single line with newline",
			input:    "hello\n",
			expected: []string{"hello"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "multiple lines with trailing newline",
			input:    "line1\nline2\nline3\n",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines",
			input:    "line1\n\nline3",
			expected: []string{"line1", "", "line3"},
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: []string{"", "", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitLines(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitLines(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJoinLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: "",
		},
		{
			name:     "single line",
			input:    []string{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple lines",
			input:    []string{"line1", "line2", "line3"},
			expected: "line1\nline2\nline3",
		},
		{
			name:     "with empty lines",
			input:    []string{"line1", "", "line3"},
			expected: "line1\n\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinLines(tt.input)
			if result != tt.expected {
				t.Errorf("JoinLines(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{
			name:     "found at beginning",
			s:        "hello world",
			substr:   "hello",
			expected: 0,
		},
		{
			name:     "found in middle",
			s:        "hello world",
			substr:   "world",
			expected: 6,
		},
		{
			name:     "not found",
			s:        "hello world",
			substr:   "foo",
			expected: -1,
		},
		{
			name:     "empty substr",
			s:        "hello",
			substr:   "",
			expected: 0,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "hello",
			expected: -1,
		},
		{
			name:     "both empty",
			s:        "",
			substr:   "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IndexOf(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("IndexOf(%q, %q) = %d, want %d", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "contains",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "does not contain",
			s:        "hello world",
			substr:   "foo",
			expected: false,
		},
		{
			name:     "empty substr",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("Contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestSplitJoinRoundtrip(t *testing.T) {
	tests := []string{
		"line1\nline2\nline3",
		"single line",
		"",
		"line1\n\nline3",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			lines := SplitLines(input)
			result := JoinLines(lines)
			// Note: trailing newlines are not preserved in roundtrip
			expected := input
			if len(input) > 0 && input[len(input)-1] == '\n' {
				expected = input[:len(input)-1]
			}
			if result != expected {
				t.Errorf("Roundtrip failed: input=%q, result=%q, expected=%q", input, result, expected)
			}
		})
	}
}
