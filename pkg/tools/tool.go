package tools

import "context"

// Tool 定义工具接口
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// Parameters 返回工具参数的 JSON Schema
	Parameters() map[string]any

	// Execute 执行工具
	Execute(ctx context.Context, args map[string]any) (string, error)
}

// BaseTool 提供工具的基础实现
type BaseTool struct {
	name        string
	description string
	parameters  map[string]any
	executor    func(ctx context.Context, args map[string]any) (string, error)
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, description string, parameters map[string]any, executor func(ctx context.Context, args map[string]any) (string, error)) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
		executor:    executor,
	}
}

func (t *BaseTool) Name() string {
	return t.name
}

func (t *BaseTool) Description() string {
	return t.description
}

func (t *BaseTool) Parameters() map[string]any {
	return t.parameters
}

func (t *BaseTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	return t.executor(ctx, args)
}
