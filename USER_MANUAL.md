# Deep Agents Go - 使用手册

## 目录
- [安装](#安装)
- [CLI 工具](#cli-工具)
- [API 使用](#api-使用)
- [配置](#配置)
- [工具](#工具)
- [中间件](#中间件)
- [后端](#后端)
- [常见问题](#常见问题)

## 安装

### 从源码安装

```bash
git clone https://github.com/zhoucx/deepagents-go.git
cd deepagents-go
make build
```

### 使用 go get

```bash
go get github.com/zhoucx/deepagents-go
```

## CLI 工具

### 基础用法

```bash
# 设置 API Key
export ANTHROPIC_API_KEY=your_api_key

# 执行任务
./bin/deepagents -prompt "创建文件 /test.txt，内容为 'Hello World'"

# 指定工作目录
./bin/deepagents -work-dir ./my-workspace -prompt "列出当前目录的文件"

# 指定模型
./bin/deepagents -model claude-3-5-sonnet-20241022 -prompt "你好"

# 指定最大迭代次数
./bin/deepagents -max-iter 10 -prompt "创建一个 Go 项目"
```

### 命令行参数

```
-api-key string
    Anthropic API Key（默认从 ANTHROPIC_API_KEY 环境变量读取）

-model string
    LLM 模型（默认: claude-3-5-sonnet-20241022）

-work-dir string
    工作目录（默认: ./workspace）

-max-iter int
    最大迭代次数（默认: 25）

-prompt string
    用户提示词（必需）
```

### 示例

#### 1. 创建文件

```bash
./bin/deepagents -prompt "创建文件 /hello.txt，内容为 'Hello, Deep Agents!'"
```

#### 2. 搜索文件

```bash
./bin/deepagents -prompt "搜索所有包含 'Hello' 的文件"
```

#### 3. 创建项目

```bash
./bin/deepagents -prompt "创建一个简单的 Go Web 服务器项目，包括 main.go 和 README.md"
```

#### 4. 任务规划

```bash
./bin/deepagents -prompt "创建一个 TODO 应用，先制定任务计划，然后逐步实现"
```

## API 使用

### 基础示例

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/zhoucx/deepagents-go/pkg/agent"
    "github.com/zhoucx/deepagents-go/pkg/backend"
    "github.com/zhoucx/deepagents-go/pkg/llm"
    "github.com/zhoucx/deepagents-go/pkg/middleware"
    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 创建 LLM 客户端
    llmClient := llm.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY"), "")

    // 创建工具注册表和后端
    toolRegistry := tools.NewRegistry()
    backend := backend.NewStateBackend()

    // 创建中间件
    fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

    // 创建 Agent
    config := &agent.Config{
        LLMClient:    llmClient,
        ToolRegistry: toolRegistry,
        Middlewares:  []agent.Middleware{fsMiddleware},
    }
    executor := agent.NewExecutor(config)

    // 执行任务
    output, err := executor.Invoke(context.Background(), &agent.InvokeInput{
        Messages: []llm.Message{
            {Role: llm.RoleUser, Content: "你的任务"},
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    // 处理结果
    for _, msg := range output.Messages {
        log.Println(msg.Content)
    }
}
```

## 配置

### Agent 配置

```go
type Config struct {
    LLMClient     llm.Client           // LLM 客户端（必需）
    ToolRegistry  *tools.Registry      // 工具注册表（必需）
    Middlewares   []Middleware         // 中间件列表
    SystemPrompt  string               // 系统提示词
    MaxIterations int                  // 最大迭代次数（默认: 25）
    MaxTokens     int                  // 最大 token 数（默认: 4096）
    Temperature   float64              // 温度参数（默认: 0.7）
}
```

### 示例配置

```go
config := &agent.Config{
    LLMClient:     llmClient,
    ToolRegistry:  toolRegistry,
    Middlewares:   []agent.Middleware{fsMiddleware, todoMiddleware},
    SystemPrompt:  "你是一个专业的代码助手...",
    MaxIterations: 30,
    MaxTokens:     8192,
    Temperature:   0.5,
}
```

## 工具

### 内置工具

#### 文件系统工具

1. **ls** - 列出目录
```go
// 参数：
// - path: 目录路径
```

2. **read_file** - 读取文件
```go
// 参数：
// - path: 文件路径
// - offset: 起始行号（可选）
// - limit: 读取行数（可选）
```

3. **write_file** - 写入文件
```go
// 参数：
// - path: 文件路径
// - content: 文件内容
```

4. **edit_file** - 编辑文件
```go
// 参数：
// - path: 文件路径
// - old_string: 要替换的字符串
// - new_string: 新字符串
// - replace_all: 是否替换所有（可选，默认 false）
```

5. **grep** - 搜索文件内容
```go
// 参数：
// - pattern: 搜索模式
// - path: 搜索路径（可选）
// - glob: 文件匹配模式（可选）
```

6. **glob** - 查找匹配的文件
```go
// 参数：
// - pattern: 文件匹配模式（如 *.go）
// - path: 搜索路径（可选）
```

#### 任务管理工具

7. **write_todos** - 创建/更新 Todo 列表
```go
// 参数：
// - todos: Todo 项列表
//   - id: Todo 项 ID
//   - title: 标题
//   - status: 状态（pending, in_progress, completed）
//   - description: 描述（可选）
```

### 自定义工具

```go
// 创建自定义工具
customTool := tools.NewBaseTool(
    "my_tool",
    "工具描述",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param1": map[string]any{
                "type":        "string",
                "description": "参数描述",
            },
        },
        "required": []string{"param1"},
    },
    func(ctx context.Context, args map[string]any) (string, error) {
        // 工具实现
        param1 := args["param1"].(string)
        return fmt.Sprintf("处理结果: %s", param1), nil
    },
)

// 注册工具
toolRegistry.Register(customTool)
```

## 中间件

### 内置中间件

#### FilesystemMiddleware

提供文件系统操作能力。

```go
fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)
```

特性：
- 注册 6 个文件系统工具
- 大结果驱逐（>80,000 字符）
- 自动保存到 `/large_tool_results/`

#### TodoMiddleware

提供任务规划能力。

```go
todoMiddleware := middleware.NewTodoMiddleware(backend, toolRegistry)
```

特性：
- 注册 write_todos 工具
- 自动加载 Todo 列表
- 注入到系统提示词

### 自定义中间件

```go
type MyMiddleware struct {
    *middleware.BaseMiddleware
}

func NewMyMiddleware() *MyMiddleware {
    return &MyMiddleware{
        BaseMiddleware: middleware.NewBaseMiddleware("my_middleware"),
    }
}

func (m *MyMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
    // 在调用 LLM 前执行
    log.Println("调用 LLM...")
    return nil
}

func (m *MyMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
    // 在工具执行后执行
    log.Printf("工具执行完成: %s", result.ToolCallID)
    return nil
}
```

## 后端

### StateBackend（内存存储）

```go
backend := backend.NewStateBackend()
```

特点：
- 数据存储在内存中
- 进程结束后数据丢失
- 适合临时数据

### FilesystemBackend（磁盘存储）

```go
backend, err := backend.NewFilesystemBackend("./workspace", true)
```

参数：
- `rootDir`: 根目录
- `virtualMode`: 虚拟模式（true: 限制在根目录内，false: 支持绝对路径）

特点：
- 数据持久化到磁盘
- 虚拟模式防止路径遍历
- 适合持久化数据

### CompositeBackend（多后端路由）

```go
// 创建组合后端
memoryBackend := backend.NewStateBackend()
composite := backend.NewCompositeBackend(memoryBackend)

// 添加路由
dataBackend, _ := backend.NewFilesystemBackend("./data", true)
configBackend, _ := backend.NewFilesystemBackend("./config", true)

composite.AddRoute("/data", dataBackend)
composite.AddRoute("/config", configBackend)
```

特点：
- 根据路径前缀自动路由
- 最长前缀匹配
- 跨后端搜索和聚合

路由规则：
```
/data/file.txt    -> dataBackend
/config/app.yaml  -> configBackend
/session.txt      -> memoryBackend（默认）
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

A: FilesystemMiddleware 会自动驱逐超过 80,000 字符的结果到文件系统。你也可以使用 `read_file` 的 `offset` 和 `limit` 参数分段读取：

```go
// 读取前 100 行
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {Role: llm.RoleUser, Content: "读取 /large_file.txt 的前 100 行"},
    },
})
```

### Q: 如何使用真实文件系统？

A: 使用 FilesystemBackend：

```go
backend, _ := backend.NewFilesystemBackend("./workspace", true)
middleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)
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
    SystemPrompt: "你是一个专业的代码审查助手，擅长发现代码中的问题...",
}
```

### Q: 如何处理错误？

A: Agent 会将工具执行错误返回给 LLM，让 LLM 决定如何处理。你也可以在中间件中捕获和处理错误：

```go
func (m *MyMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
    if result.IsError {
        log.Printf("工具执行出错: %s", result.Content)
        // 自定义错误处理
    }
    return nil
}
```

### Q: 如何设置超时？

A: 使用 context.WithTimeout：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := executor.Invoke(ctx, input)
```

### Q: 如何查看 Agent 的执行过程？

A: 可以创建一个日志中间件：

```go
type LoggingMiddleware struct {
    *middleware.BaseMiddleware
}

func (m *LoggingMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
    log.Printf("调用 LLM，消息数量: %d", len(req.Messages))
    return nil
}

func (m *LoggingMiddleware) BeforeTool(ctx context.Context, toolCall *llm.ToolCall, state *agent.State) error {
    log.Printf("执行工具: %s", toolCall.Name)
    return nil
}
```

### Q: 如何测试 Agent？

A: 使用 mock LLM 客户端：

```go
type MockLLMClient struct {
    responses []*llm.ModelResponse
    callCount int
}

func (m *MockLLMClient) Generate(ctx context.Context, req *llm.ModelRequest) (*llm.ModelResponse, error) {
    if m.callCount >= len(m.responses) {
        return &llm.ModelResponse{Content: "Done"}, nil
    }
    resp := m.responses[m.callCount]
    m.callCount++
    return resp, nil
}

// 在测试中使用
mockClient := &MockLLMClient{
    responses: []*llm.ModelResponse{
        {Content: "测试响应"},
    },
}

config := &agent.Config{
    LLMClient: mockClient,
    // ...
}
```

## 性能优化

### 1. 使用流式读取

对于大文件，使用 `offset` 和 `limit` 参数：

```go
// 分段读取
for offset := 0; offset < totalLines; offset += 1000 {
    // 读取 1000 行
}
```

### 2. 缓存 LLM 响应

创建缓存中间件：

```go
type CachingMiddleware struct {
    cache map[string]string
}

func (m *CachingMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
    key := generateKey(req.Messages)
    if cached, ok := m.cache[key]; ok {
        // 使用缓存
    }
    return nil
}
```

### 3. 并行执行工具

如果有多个独立的工具调用，可以并行执行（需要自定义实现）。

### 4. 限制上下文长度

使用 `MaxTokens` 限制上下文长度：

```go
config := &agent.Config{
    MaxTokens: 4096, // 限制为 4096 tokens
}
```

## 安全性

### 1. 虚拟模式

使用虚拟模式防止路径遍历：

```go
backend, _ := backend.NewFilesystemBackend("./workspace", true) // virtualMode = true
```

### 2. 输入验证

在自定义工具中验证输入：

```go
func (t *MyTool) Execute(ctx context.Context, args map[string]any) (string, error) {
    input := args["input"].(string)

    // 验证输入
    if len(input) > 10000 {
        return "", errors.New("input too long")
    }

    // 处理输入
    return process(input), nil
}
```

### 3. 超时控制

使用 context.WithTimeout 防止长时间运行：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## 更多资源

- [快速开始指南](QUICKSTART.md)
- [API 文档](API.md)
- [示例程序](../cmd/examples/)
- [贡献指南](CONTRIBUTING.md)
- [GitHub Issues](https://github.com/zhoucx/deepagents-go/issues)
