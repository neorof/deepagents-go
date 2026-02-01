package llm

// Role 定义消息角色
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message 表示一条对话消息
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ToolCall 表示工具调用请求
type ToolCall struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// ToolResult 表示工具执行结果
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}

// ModelRequest 表示发送给 LLM 的请求
type ModelRequest struct {
	Messages     []Message    `json:"messages"`
	SystemPrompt string       `json:"system_prompt,omitempty"`
	Tools        []ToolSchema `json:"tools,omitempty"`
	MaxTokens    int          `json:"max_tokens,omitempty"`
	Temperature  float64      `json:"temperature,omitempty"`
}

// ModelResponse 表示 LLM 的响应
type ModelResponse struct {
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	StopReason string     `json:"stop_reason"`
}

// ToolSchema 定义工具的 JSON Schema
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// StreamEventType 定义流式事件类型
type StreamEventType string

const (
	StreamEventTypeStart      StreamEventType = "start"       // 开始生成
	StreamEventTypeText       StreamEventType = "text"        // 文本内容
	StreamEventTypeToolUse    StreamEventType = "tool_use"    // 工具调用
	StreamEventTypeToolResult StreamEventType = "tool_result" // 工具结果
	StreamEventTypeEnd        StreamEventType = "end"         // 生成结束
	StreamEventTypeError      StreamEventType = "error"       // 错误
	StreamEventTypePing       StreamEventType = "ping"        // 心跳
	StreamEventTypeMetadata   StreamEventType = "metadata"    // 元数据
)

// StreamEvent 表示流式响应事件
type StreamEvent struct {
	Type       StreamEventType `json:"type"`
	Content    string          `json:"content,omitempty"`     // 文本内容（增量）
	ToolCall   *ToolCall       `json:"tool_call,omitempty"`   // 工具调用
	StopReason string          `json:"stop_reason,omitempty"` // 停止原因
	Error      error           `json:"error,omitempty"`       // 错误信息
	Metadata   map[string]any  `json:"metadata,omitempty"`    // 元数据
	Done       bool            `json:"done"`                  // 是否完成
}
