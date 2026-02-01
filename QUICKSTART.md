# Deep Agents Go - 快速开始指南

## 目录
- [安装](#安装)
- [基础概念](#基础概念)
- [第一个 Agent](#第一个-agent)
- [文件操作](#文件操作)
- [任务规划](#任务规划)
- [多后端存储](#多后端存储)
- [自定义工具](#自定义工具)
- [自定义中间件](#自定义中间件)

## 安装

### 前置要求
- Go 1.21 或更高版本
- Anthropic API Key（从 https://console.anthropic.com/ 获取）

### 安装依赖
```bash
go get github.com/zhoucx/deepagents-go
```

### 设置环境变量
```bash
export ANTHROPIC_API_KEY=your_api_key_here
```

## 基础概念

### Agent
Agent 是执行任务的核心组件，它：
- 接收用户消息
- 调用 LLM 生成响应
- 执行工具调用
- 返回结果

### 中间件
中间件在 Agent 执行过程中提供额外功能：
- **FilesystemMiddleware**: 文件系统操作
- **TodoMiddleware**: 任务规划
- 可以自定义中间件扩展功能

### 后端
后端负责数据存储：
- **StateBackend**: 内存存储（临时数据）
- **FilesystemBackend**: 磁盘存储（持久化）
- **CompositeBackend**: 多后端路由

### 工具
工具是 Agent 可以调用的函数：
- 文件操作：ls, read_file, write_file, edit_file
- 搜索：grep, glob
- 任务管理：write_todos

## 第一个 Agent

### 最简单的例子

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/zhoucx/deepagents-go/pkg/agent"
    "github.com/zhoucx/deepagents-go/pkg/backend"
    "github.com/zhoucx/deepagents-go/pkg/llm"
    "github.com/zhoucx/deepagents-go/pkg/middleware"
    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 1. 创建 LLM 客户端
    llmClient := llm.NewAnthropicClient(
        os.Getenv("ANTHROPIC_API_KEY"),
        "claude-3-5-sonnet-20241022",
    )

    // 2. 创建工具注册表
    toolRegistry := tools.NewRegistry()

    // 3. 创建后端
    backend := backend.NewStateBackend()

    // 4. 创建中间件
    fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

    // 5. 创建 Agent 配置
    config := &agent.Config{
        LLMClient:    llmClient,
        ToolRegistry: toolRegistry,
        Middlewares:  []agent.Middleware{fsMiddleware},
        SystemPrompt: "你是一个有用的 AI 助手。",
    }

    // 6. 创建 Agent 执行器
    executor := agent.NewExecutor(config)

    // 7. 执行任务
    ctx := context.Background()
    output, err := executor.Invoke(ctx, &agent.InvokeInput{
        Messages: []llm.Message{
            {
                Role:    llm.RoleUser,
                Content: "你好！请介绍一下你自己。",
            },
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    // 8. 打印结果
    for _, msg := range output.Messages {
        if msg.Role == llm.RoleAssistant {
            fmt.Println(msg.Content)
        }
    }
}
```

### 运行
```bash
go run main.go
```

## 文件操作

### 创建和读取文件

```go
// 创建文件
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "创建文件 /hello.txt，内容为 'Hello, World!'",
        },
    },
})

// 读取文件
output, _ = executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "读取文件 /hello.txt 的内容",
        },
    },
})
```

### 编辑文件

```go
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "将 /hello.txt 中的 'World' 替换为 'Go'",
        },
    },
})
```

### 搜索文件

```go
// 使用 grep 搜索内容
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "搜索所有包含 'Hello' 的文件",
        },
    },
})

// 使用 glob 查找文件
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "查找所有 .txt 文件",
        },
    },
})
```

## 任务规划

### 使用 Todo 中间件

```go
// 添加 Todo 中间件
todoMiddleware := middleware.NewTodoMiddleware(backend, toolRegistry)

config := &agent.Config{
    LLMClient:    llmClient,
    ToolRegistry: toolRegistry,
    Middlewares:  []agent.Middleware{
        fsMiddleware,
        todoMiddleware, // 添加 Todo 中间件
    },
    SystemPrompt: "你是一个擅长任务规划的 AI 助手。",
}

executor := agent.NewExecutor(config)

// 执行复杂任务
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role: llm.RoleUser,
            Content: `创建一个 Go Web 服务器项目，包括：
1. main.go - 主程序
2. handlers.go - 请求处理
3. README.md - 文档

请先制定任务计划，然后逐步执行。`,
        },
    },
})
```

Agent 会自动：
1. 使用 `write_todos` 创建任务列表
2. 逐步执行每个任务
3. 更新任务状态
4. 完成后给出总结

## 多后端存储

### 创建组合后端

```go
// 1. 创建内存后端（用于临时数据）
memoryBackend := backend.NewStateBackend()

// 2. 创建组合后端
composite := backend.NewCompositeBackend(memoryBackend)

// 3. 添加文件系统后端
dataBackend, _ := backend.NewFilesystemBackend("./data", true)
configBackend, _ := backend.NewFilesystemBackend("./config", true)

// 4. 添加路由规则
composite.AddRoute("/data", dataBackend)
composite.AddRoute("/config", configBackend)

// 5. 使用组合后端
fsMiddleware := middleware.NewFilesystemMiddleware(composite, toolRegistry)
```

### 路由规则

```
/data/file.txt    -> dataBackend (磁盘: ./data/file.txt)
/config/app.yaml  -> configBackend (磁盘: ./config/app.yaml)
/session.txt      -> memoryBackend (内存)
```

### 使用示例

```go
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role: llm.RoleUser,
            Content: `
创建以下文件：
- /data/users.json - 用户数据（持久化到磁盘）
- /config/app.yaml - 配置文件（持久化到磁盘）
- /session.txt - 会话信息（临时存储在内存）
`,
        },
    },
})
```

## 自定义工具

### 创建简单工具

```go
// 创建一个计算器工具
calculatorTool := tools.NewBaseTool(
    "calculator",
    "执行数学计算",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "expression": map[string]any{
                "type":        "string",
                "description": "数学表达式，如 '2 + 2'",
            },
        },
        "required": []string{"expression"},
    },
    func(ctx context.Context, args map[string]any) (string, error) {
        expr := args["expression"].(string)
        // 这里应该使用安全的表达式求值库
        // 简化示例：
        result := "4" // 假设计算结果
        return fmt.Sprintf("计算结果: %s", result), nil
    },
)

// 注册工具
toolRegistry.Register(calculatorTool)
```

### 创建复杂工具

```go
// 创建一个 HTTP 请求工具
type HTTPTool struct {
    client *http.Client
}

func NewHTTPTool() *HTTPTool {
    return &HTTPTool{
        client: &http.Client{Timeout: 10 * time.Second},
    }
}

func (t *HTTPTool) Name() string {
    return "http_request"
}

func (t *HTTPTool) Description() string {
    return "发送 HTTP 请求"
}

func (t *HTTPTool) Parameters() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "url": map[string]any{
                "type":        "string",
                "description": "请求 URL",
            },
            "method": map[string]any{
                "type":        "string",
                "description": "HTTP 方法（GET, POST 等）",
                "enum":        []string{"GET", "POST", "PUT", "DELETE"},
            },
        },
        "required": []string{"url", "method"},
    }
}

func (t *HTTPTool) Execute(ctx context.Context, args map[string]any) (string, error) {
    url := args["url"].(string)
    method := args["method"].(string)

    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return "", err
    }

    resp, err := t.client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    return string(body), nil
}

// 注册工具
toolRegistry.Register(NewHTTPTool())
```

## 自定义中间件

### 创建简单中间件

```go
type LoggingMiddleware struct {
    *middleware.BaseMiddleware
}

func NewLoggingMiddleware() *LoggingMiddleware {
    return &LoggingMiddleware{
        BaseMiddleware: middleware.NewBaseMiddleware("logging"),
    }
}

func (m *LoggingMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
    log.Printf("调用 LLM，消息数量: %d", len(req.Messages))
    return nil
}

func (m *LoggingMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
    log.Printf("LLM 响应，工具调用数量: %d", len(resp.ToolCalls))
    return nil
}

func (m *LoggingMiddleware) BeforeTool(ctx context.Context, toolCall *llm.ToolCall, state *agent.State) error {
    log.Printf("执行工具: %s", toolCall.Name)
    return nil
}

func (m *LoggingMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
    log.Printf("工具执行完成，结果长度: %d", len(result.Content))
    return nil
}

// 使用中间件
config := &agent.Config{
    Middlewares: []agent.Middleware{
        NewLoggingMiddleware(),
        fsMiddleware,
        todoMiddleware,
    },
}
```

### 创建复杂中间件

```go
type CachingMiddleware struct {
    *middleware.BaseMiddleware
    cache map[string]string
    mu    sync.RWMutex
}

func NewCachingMiddleware() *CachingMiddleware {
    return &CachingMiddleware{
        BaseMiddleware: middleware.NewBaseMiddleware("caching"),
        cache:          make(map[string]string),
    }
}

func (m *CachingMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
    // 生成缓存键
    key := m.generateCacheKey(req.Messages)

    m.mu.RLock()
    cached, ok := m.cache[key]
    m.mu.RUnlock()

    if ok {
        log.Printf("缓存命中: %s", key)
        // 可以在这里修改请求，使用缓存的响应
    }

    return nil
}

func (m *CachingMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
    // 缓存响应
    key := m.generateCacheKey(state.GetMessages())

    m.mu.Lock()
    m.cache[key] = resp.Content
    m.mu.Unlock()

    return nil
}

func (m *CachingMiddleware) generateCacheKey(messages []llm.Message) string {
    // 简化实现：使用最后一条消息作为键
    if len(messages) > 0 {
        return messages[len(messages)-1].Content
    }
    return ""
}
```

## 最佳实践

### 1. 错误处理

```go
output, err := executor.Invoke(ctx, input)
if err != nil {
    log.Printf("执行失败: %v", err)
    // 处理错误
    return
}

// 检查工具执行错误
for _, msg := range output.Messages {
    if strings.Contains(msg.Content, "error") {
        log.Printf("工具执行出错: %s", msg.Content)
    }
}
```

### 2. 超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := executor.Invoke(ctx, input)
```

### 3. 配置管理

```go
type AgentConfig struct {
    APIKey        string
    Model         string
    MaxIterations int
    Temperature   float64
}

func NewAgentFromConfig(cfg *AgentConfig) *agent.Executor {
    llmClient := llm.NewAnthropicClient(cfg.APIKey, cfg.Model)

    config := &agent.Config{
        LLMClient:     llmClient,
        MaxIterations: cfg.MaxIterations,
        Temperature:   cfg.Temperature,
        // ...
    }

    return agent.NewExecutor(config)
}
```

### 4. 测试

```go
func TestAgent(t *testing.T) {
    // 使用 mock LLM 客户端
    mockClient := &MockLLMClient{
        responses: []*llm.ModelResponse{
            {Content: "测试响应"},
        },
    }

    config := &agent.Config{
        LLMClient: mockClient,
        // ...
    }

    executor := agent.NewExecutor(config)

    output, err := executor.Invoke(context.Background(), input)
    assert.NoError(t, err)
    assert.NotEmpty(t, output.Messages)
}
```

## 常见问题

### Q: 如何限制 Agent 的迭代次数？
A: 在配置中设置 `MaxIterations`：
```go
config := &agent.Config{
    MaxIterations: 10, // 最多 10 次迭代
}
```

### Q: 如何处理大文件？
A: FilesystemMiddleware 会自动驱逐超过 80,000 字符的结果到文件系统。

### Q: 如何使用真实文件系统？
A: 使用 FilesystemBackend：
```go
backend, _ := backend.NewFilesystemBackend("./workspace", true)
```

### Q: 如何禁用某个工具？
A: 从工具注册表中移除：
```go
toolRegistry.Remove("tool_name")
```

### Q: 如何自定义系统提示词？
A: 在配置中设置 `SystemPrompt`：
```go
config := &agent.Config{
    SystemPrompt: "你是一个专业的代码审查助手...",
}
```

## 下一步

- 查看 [示例程序](../cmd/examples/) 了解更多用法
- 阅读 [API 文档](API.md) 了解详细接口
- 参考 [架构设计](ARCHITECTURE.md) 了解内部实现
- 查看 [贡献指南](CONTRIBUTING.md) 参与开发

## 获取帮助

- GitHub Issues: https://github.com/zhoucx/deepagents-go/issues
- 文档: https://github.com/zhoucx/deepagents-go/docs
