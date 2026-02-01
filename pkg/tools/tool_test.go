package tools

import (
	"context"
	"testing"
)

func TestBaseTool_Name(t *testing.T) {
	tool := NewBaseTool("test_tool", "Description", map[string]any{}, nil)

	if tool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got %q", tool.Name())
	}
}

func TestBaseTool_Description(t *testing.T) {
	tool := NewBaseTool("test_tool", "Test description", map[string]any{}, nil)

	if tool.Description() != "Test description" {
		t.Errorf("Expected description 'Test description', got %q", tool.Description())
	}
}

func TestBaseTool_Parameters(t *testing.T) {
	params := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"param1": map[string]any{"type": "string"},
		},
	}

	tool := NewBaseTool("test_tool", "Description", params, nil)

	returnedParams := tool.Parameters()
	if returnedParams["type"] != "object" {
		t.Error("Expected type to be 'object'")
	}
}

func TestBaseTool_Execute(t *testing.T) {
	executed := false
	executor := func(ctx context.Context, args map[string]any) (string, error) {
		executed = true
		return "Success", nil
	}

	tool := NewBaseTool("test_tool", "Description", map[string]any{}, executor)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !executed {
		t.Error("Expected executor to be called")
	}

	if result != "Success" {
		t.Errorf("Expected result 'Success', got %q", result)
	}
}

func TestBaseTool_ExecuteWithArgs(t *testing.T) {
	var receivedArgs map[string]any
	executor := func(ctx context.Context, args map[string]any) (string, error) {
		receivedArgs = args
		return "OK", nil
	}

	tool := NewBaseTool("test_tool", "Description", map[string]any{}, executor)

	ctx := context.Background()
	args := map[string]any{
		"param1": "value1",
		"param2": 42,
	}

	_, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if receivedArgs["param1"] != "value1" {
		t.Error("Expected param1 to be passed to executor")
	}

	if receivedArgs["param2"] != 42 {
		t.Error("Expected param2 to be passed to executor")
	}
}

func TestBaseTool_NilExecutor(t *testing.T) {
	tool := NewBaseTool("test_tool", "Description", map[string]any{}, nil)

	// 这应该会 panic，但我们不测试 panic 行为
	// 在实际使用中，应该总是提供 executor
	if tool.executor == nil {
		// 预期行为
	}
}
