# 流式响应使用指南

## 📋 概述

**实施日期**: 2026-01-31
**功能**: LLM 流式响应支持
**状态**: ✅ 已完成

## 🎯 功能说明

Deep Agents Go 现在支持流式响应，可以实时显示 LLM 的生成内容，提升用户体验。

### 主要优势

1. **更低的首字节延迟** - 无需等待完整响应，立即开始显示
2. **更好的用户体验** - 逐字显示，更有交互感
3. **可提前终止** - 节省 token 成本
4. **统一的接口** - Anthropic 和 OpenAI 使用相同的流式 API

## 📦 核心组件

### 1. StreamEvent 类型

流式事件定义在 `pkg/llm/message.go`:

```go
type StreamEvent struct {
    Type       StreamEventType // 事件类型
    Content    string          // 文本内容（增量）
    ToolCall   *ToolCall       // 工具调用
    StopReason string          // 停止原因
    Error      error           // 错误信息
    Metadata   map[string]any  // 元数据
    Done       bool            // 是否完成
}
```

### 2. 事件类型

```go
const (
    StreamEventTypeStart      // 开始生成
    StreamEventTypeText       // 文本内容
    StreamEventTypeToolUse    // 工具调用
    StreamEventTypeEnd        // 生成结束
    StreamEventTypeError      // 错误
    StreamEventTypePing       // 心跳
    StreamEventTypeMetadata   // 元数据
)
```

### 3. Client 接口

扩展的 `llm.Client` 接口：

```go
type Client interface {
    // Generate 生成响应（非流式）
    Generate(ctx context.Context, req *ModelRequest) (*ModelResponse, error)

    // StreamGenerate 生成响应（流式）
    StreamGenerate(ctx context.Context, req *ModelRequest) (<-chan StreamEvent, error)

    // CountTokens 估算 token 数量
    CountTokens(messages []Message) int
}
```

## 🚀 使用方法

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "github.com/zhoucx/deepagents-go/pkg/llm"
)

func main() {
    // 创建客户端
    client := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", "")

    // 构建请求
    req := &llm.ModelRequest{
        Messages: []llm.Message{
            {Role: llm.RoleUser, Content: "介绍一下 Go 语言"},
        },
        MaxTokens:   500,
        Temperature: 0.7,
    }

    // 调用流式 API
    stream, err := client.StreamGenerate(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    // 处理流式事件
    for event := range stream {
        switch event.Type {
        case llm.StreamEventTypeText:
            fmt.Print(event.Content) // 实时显示文本
        case llm.StreamEventTypeEnd:
            fmt.Println("\n[完成]")
        case llm.StreamEventTypeError:
            fmt.Printf("[错误: %v]\n", event.Error)
        }
    }
}
```

### 完整示例

参见 `cmd/examples/streaming/main.go`:

- Anthropic 流式响应
- OpenAI 流式响应
- 对比非流式和流式响应

### 运行示例

```bash
# 设置 API Key
export ANTHROPIC_API_KEY=your_key

# 可选：设置 OpenAI Key
export OPENAI_API_KEY=your_openai_key

# 运行示例
go run ./cmd/examples/streaming/main.go
```

## 🎨 事件处理模式

### 模式 1：简单显示

```go
for event := range stream {
    if event.Type == llm.StreamEventTypeText {
        fmt.Print(event.Content)
    }
}
```

### 模式 2：累积内容

```go
var fullContent string

for event := range stream {
    switch event.Type {
    case llm.StreamEventTypeText:
        fullContent += event.Content
        fmt.Print(event.Content)
    case llm.StreamEventTypeEnd:
        // fullContent 现在包含完整内容
        saveToDatabase(fullContent)
    }
}
```

### 模式 3：处理工具调用

```go
for event := range stream {
    switch event.Type {
    case llm.StreamEventTypeText:
        fmt.Print(event.Content)
    case llm.StreamEventTypeToolUse:
        fmt.Printf("\n[调用工具: %s]\n", event.ToolCall.Name)
        // 执行工具...
    case llm.StreamEventTypeEnd:
        fmt.Println("\n[完成]")
    }
}
```

### 模式 4：提前终止

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

stream, _ := client.StreamGenerate(ctx, req)

for event := range stream {
    if event.Type == llm.StreamEventTypeText {
        fmt.Print(event.Content)

        // 如果生成了足够的内容，提前终止
        if len(fullContent) > 1000 {
            cancel() // 取消请求
            break
        }
    }
}
```

## 🔧 实现细节

### Anthropic 客户端

- 使用 `Messages.NewStreaming()` API
- 处理 `ContentBlockDelta` 事件
- 解析工具调用 JSON
- 支持心跳事件

### OpenAI 客户端

- 使用 `CreateChatCompletionStream()` API
- 处理 `Delta` 增量
- 累积工具调用参数
- 支持多个并发工具调用

## ⚠️ 注意事项

1. **资源清理**: channel 会自动关闭，但建议使用 `defer cancel()` 确保资源释放

2. **错误处理**: 流式过程中的错误会通过 `StreamEventTypeError` 事件发送

3. **超时控制**: 建议使用 `context.WithTimeout()` 设置超时

4. **并发安全**: 每个流式请求应在独立的 goroutine 中处理

## 📊 性能对比

### 非流式响应
- ✓ 简单易用
- ✓ 一次性返回
- ✗ 等待时间长
- ✗ 无法提前终止

### 流式响应
- ✓ 首字节延迟低
- ✓ 用户体验好
- ✓ 可提前终止
- ✗ 代码稍复杂

## 🔮 未来计划

- [x] Agent 层面的流式执行 ✅ 已完成
- [ ] CLI 工具集成流式显示
- [ ] 流式响应的性能优化
- [ ] 更多流式事件类型

## 🎯 Agent 流式执行

### 使用方法

```go
executor := agent.NewExecutor(config)

stream, err := executor.InvokeStream(ctx, input)
if err != nil {
    log.Fatal(err)
}

for event := range stream {
    switch event.Type {
    case agent.AgentEventTypeLLMText:
        fmt.Print(event.Content) // 实时显示
    case agent.AgentEventTypeToolStart:
        fmt.Printf("[调用工具: %s]\n", event.ToolCall.Name)
    case agent.AgentEventTypeToolResult:
        fmt.Printf("[结果: %s]\n", event.ToolResult.Content)
    case agent.AgentEventTypeEnd:
        fmt.Println("[完成]")
    }
}
```

### Agent 事件类型

| 事件类型 | 说明 | 数据 |
|---------|------|------|
| `Start` | Agent 开始 | - |
| `LLMStart` | LLM 开始生成 | Iteration |
| `LLMText` | LLM 文本内容 | Content, Iteration |
| `LLMToolCall` | LLM 工具调用 | ToolCall, Iteration |
| `LLMEnd` | LLM 生成结束 | Iteration |
| `ToolStart` | 工具开始执行 | ToolCall, Iteration |
| `ToolResult` | 工具执行结果 | ToolResult, Iteration |
| `IterationEnd` | 迭代结束 | Iteration |
| `End` | Agent 完成 | Metadata |
| `Error` | 错误 | Error |

### 示例程序

参见 `cmd/examples/agent_streaming/main.go`


## 📚 相关文档

- [LLM Client 接口](../pkg/llm/client.go)
- [流式事件定义](../pkg/llm/message.go)
- [Anthropic 实现](../pkg/llm/anthropic.go)
- [OpenAI 实现](../pkg/llm/openai.go)
- [示例程序](../cmd/examples/streaming/main.go)

---

**最后更新**: 2026-01-31
