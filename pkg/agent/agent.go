package agent

import (
	"context"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// Agent 定义 Agent 接口
type Agent interface {
	// Invoke 执行 Agent（非流式）
	Invoke(ctx context.Context, input *InvokeInput) (*InvokeOutput, error)

	// InvokeStream 执行 Agent（流式）
	InvokeStream(ctx context.Context, input *InvokeInput) (<-chan AgentEvent, error)
}

// InvokeInput Agent 输入
type InvokeInput struct {
	Messages []llm.Message     `json:"messages"`
	Files    map[string]string `json:"files,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

// InvokeOutput Agent 输出
type InvokeOutput struct {
	Messages []llm.Message     `json:"messages"`
	Files    map[string]string `json:"files"`
	Metadata map[string]any    `json:"metadata"`
}

// AgentEventType 定义 Agent 事件类型
type AgentEventType string

const (
	AgentEventTypeStart        AgentEventType = "start"         // Agent 开始执行
	AgentEventTypeLLMStart     AgentEventType = "llm_start"     // LLM 开始生成
	AgentEventTypeLLMText      AgentEventType = "llm_text"      // LLM 文本内容
	AgentEventTypeLLMToolCall  AgentEventType = "llm_tool_call" // LLM 工具调用
	AgentEventTypeLLMEnd       AgentEventType = "llm_end"       // LLM 生成结束
	AgentEventTypeToolStart    AgentEventType = "tool_start"    // 工具开始执行
	AgentEventTypeToolResult   AgentEventType = "tool_result"   // 工具执行结果
	AgentEventTypeIterationEnd AgentEventType = "iteration_end" // 迭代结束
	AgentEventTypeEnd          AgentEventType = "end"           // Agent 执行结束
	AgentEventTypeError        AgentEventType = "error"         // 错误
)

// AgentEvent 表示 Agent 执行事件
type AgentEvent struct {
	Type       AgentEventType  `json:"type"`
	Content    string          `json:"content,omitempty"`     // 文本内容
	ToolCall   *llm.ToolCall   `json:"tool_call,omitempty"`   // 工具调用
	ToolResult *llm.ToolResult `json:"tool_result,omitempty"` // 工具结果
	Iteration  int             `json:"iteration,omitempty"`   // 当前迭代次数
	Error      error           `json:"error,omitempty"`       // 错误信息
	Metadata   map[string]any  `json:"metadata,omitempty"`    // 元数据
	Done       bool            `json:"done"`                  // 是否完成
}

// Middleware 定义中间件接口
type Middleware interface {
	// Name 返回中间件名称
	Name() string

	// BeforeAgent 在 Agent 执行前调用
	BeforeAgent(ctx context.Context, state *State) error

	// BeforeModel 在调用 LLM 前调用
	BeforeModel(ctx context.Context, req *llm.ModelRequest) error

	// AfterModel 在 LLM 响应后调用
	AfterModel(ctx context.Context, resp *llm.ModelResponse, state *State) error

	// BeforeTool 在执行工具前调用
	BeforeTool(ctx context.Context, toolCall *llm.ToolCall, state *State) error

	// AfterTool 在工具执行后调用
	AfterTool(ctx context.Context, result *llm.ToolResult, state *State) error
}
