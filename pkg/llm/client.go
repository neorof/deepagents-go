package llm

import "context"

// Client 定义 LLM 客户端接口
type Client interface {
	// Generate 生成响应（非流式）
	Generate(ctx context.Context, req *ModelRequest) (*ModelResponse, error)

	// StreamGenerate 生成响应（流式）
	// 返回一个只读 channel，通过它接收流式事件
	// 当生成完成或发生错误时，channel 会被关闭
	StreamGenerate(ctx context.Context, req *ModelRequest) (<-chan StreamEvent, error)

	// CountTokens 估算 token 数量
	CountTokens(messages []Message) int
}
