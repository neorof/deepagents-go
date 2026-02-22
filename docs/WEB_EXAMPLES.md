# Web 工具使用示例

## 快速开始

### 1. 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/zhoucx/deepagents-go/pkg/config"
    "github.com/zhoucx/deepagents-go/pkg/middleware"
    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 创建工具注册表
    toolRegistry := tools.NewRegistry()

    // 创建 Web 配置
    webConfig := config.DefaultWebConfig()

    // 创建 Web 中间件
    webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

    // 获取工具
    webSearchTool, _ := toolRegistry.Get("web_search")
    webFetchTool, _ := toolRegistry.Get("web_fetch")

    ctx := context.Background()

    // 使用 web_search 工具
    searchResult, err := webSearchTool.Execute(ctx, map[string]any{
        "query":       "Go 语言 2026 新特性",
        "max_results": 5,
    })
    if err != nil {
        log.Fatalf("搜索失败: %v", err)
    }
    fmt.Println(searchResult)

    // 使用 web_fetch 工具
    fetchResult, err := webFetchTool.Execute(ctx, map[string]any{
        "url":     "https://go.dev",
        "timeout": 30,
    })
    if err != nil {
        log.Fatalf("获取失败: %v", err)
    }
    fmt.Println(fetchResult)
}
```

### 2. 与 Agent 集成

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/zhoucx/deepagents-go/pkg/agent"
    "github.com/zhoucx/deepagents-go/pkg/backend"
    "github.com/zhoucx/deepagents-go/pkg/config"
    "github.com/zhoucx/deepagents-go/pkg/llm"
    "github.com/zhoucx/deepagents-go/pkg/middleware"
    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 创建 LLM 客户端
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    llmClient := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", "")

    // 创建工具注册表
    toolRegistry := tools.NewRegistry()

    // 创建文件系统后端和中间件
    fsBackend := backend.NewStateBackend()
    fsMiddleware := middleware.NewFilesystemMiddleware(fsBackend, toolRegistry)

    // 创建 Web 中间件
    webConfig := config.DefaultWebConfig()
    webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

    // 创建 Agent
    agentConfig := &agent.Config{
        LLMClient:    llmClient,
        ToolRegistry: toolRegistry,
        Middlewares:  []agent.Middleware{fsMiddleware, webMiddleware},
        SystemPrompt: `你是一个有用的 AI 助手，可以搜索网络内容和管理文件。

可用工具：
- web_search: 搜索网络内容
- web_fetch: 获取网页内容
- write_file: 写入文件
- read_file: 读取文件

请根据用户需求使用这些工具。`,
        MaxIterations: 25,
    }

    executor := agent.NewExecutor(agentConfig)

    // 执行任务
    ctx := context.Background()
    output, err := executor.Invoke(ctx, &agent.InvokeInput{
        Messages: []llm.Message{
            {
                Role:    llm.RoleUser,
                Content: "搜索 'Go 语言最佳实践'，并将结果保存到 /go_practices.md 文件中",
            },
        },
    })

    if err != nil {
        log.Fatalf("执行失败: %v", err)
    }

    // 打印结果
    for _, msg := range output.Messages {
        if msg.Role == llm.RoleAssistant && msg.Content != "" {
            log.Printf("[助手] %s\n", msg.Content)
        }
    }
}
```

### 3. 自定义配置

```go
package main

import (
    "time"

    "github.com/zhoucx/deepagents-go/pkg/config"
    "github.com/zhoucx/deepagents-go/pkg/middleware"
    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 自定义配置
    webConfig := config.WebConfig{
        SearchEngine:      "duckduckgo",
        DefaultTimeout:    60,              // 增加超时时间
        MaxContentLength:  50000,           // 减小最大内容长度
        EnableReadability: true,            // 启用智能提取
    }

    toolRegistry := tools.NewRegistry()
    webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

    // 使用中间件...
}
```

### 4. 直接使用搜索引擎

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 创建 DuckDuckGo 搜索引擎
    engine := tools.NewDuckDuckGoEngine(30 * time.Second)

    ctx := context.Background()

    // 执行搜索
    results, err := engine.Search(ctx, "Go 语言", 5)
    if err != nil {
        log.Fatalf("搜索失败: %v", err)
    }

    // 打印结果
    for i, result := range results {
        fmt.Printf("%d. %s\n", i+1, result.Title)
        fmt.Printf("   URL: %s\n", result.URL)
        fmt.Printf("   摘要: %s\n\n", result.Snippet)
    }
}
```

## 常见使用场景

### 场景 1: 研究技术主题

```go
// 任务：研究 Go 语言并发编程
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role: llm.RoleUser,
            Content: `研究 Go 语言并发编程：
1. 搜索 'Go 并发编程最佳实践'
2. 获取前 3 个结果的详细内容
3. 总结关键要点
4. 保存到 /go_concurrency_research.md`,
        },
    },
})
```

### 场景 2: 收集新闻

```go
// 任务：收集 AI 新闻
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role: llm.RoleUser,
            Content: `收集今天的 AI 新闻：
1. 搜索 '2026 AI 新闻'
2. 获取前 5 条结果
3. 整理成 Markdown 格式
4. 保存到 /ai_news_today.md`,
        },
    },
})
```

### 场景 3: 文档收集

```go
// 任务：收集官方文档
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role: llm.RoleUser,
            Content: `收集 Go 官方文档：
1. 获取 https://go.dev/doc/ 的内容
2. 提取主要章节
3. 保存到 /go_docs.md`,
        },
    },
})
```

### 场景 4: 竞品分析

```go
// 任务：竞品分析
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role: llm.RoleUser,
            Content: `分析 AI Agent 框架：
1. 搜索 'AI Agent 框架对比'
2. 获取前 5 个结果
3. 提取关键特性
4. 制作对比表格
5. 保存到 /agent_frameworks_comparison.md`,
        },
    },
})
```

## 高级用法

### 1. 批量搜索

```go
queries := []string{
    "Go 语言性能优化",
    "Go 语言测试最佳实践",
    "Go 语言错误处理",
}

for _, query := range queries {
    result, _ := webSearchTool.Execute(ctx, map[string]any{
        "query":       query,
        "max_results": 3,
    })
    fmt.Println(result)
}
```

### 2. 错误处理

```go
result, err := webSearchTool.Execute(ctx, map[string]any{
    "query":       "Go 语言",
    "max_results": 5,
})

if err != nil {
    // 处理错误
    if strings.Contains(err.Error(), "超时") {
        log.Println("请求超时，请稍后重试")
    } else if strings.Contains(err.Error(), "未找到") {
        log.Println("没有找到相关结果")
    } else {
        log.Printf("搜索失败: %v", err)
    }
    return
}

fmt.Println(result)
```

### 3. 超时控制

```go
// 创建带超时的 context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := webFetchTool.Execute(ctx, map[string]any{
    "url":     "https://example.com",
    "timeout": 5, // 工具内部超时
})

if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("总体超时")
    } else {
        log.Printf("获取失败: %v", err)
    }
    return
}
```

### 4. 结果过滤

```go
result, _ := webSearchTool.Execute(ctx, map[string]any{
    "query":       "Go 语言",
    "max_results": 10,
})

// 解析结果并过滤
lines := strings.Split(result, "\n")
for _, line := range lines {
    if strings.Contains(line, "官方") {
        fmt.Println(line)
    }
}
```

## 性能优化建议

### 1. 并发搜索

```go
var wg sync.WaitGroup
results := make(chan string, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        result, _ := webSearchTool.Execute(ctx, map[string]any{
            "query":       q,
            "max_results": 3,
        })
        results <- result
    }(query)
}

wg.Wait()
close(results)

for result := range results {
    fmt.Println(result)
}
```

### 2. 缓存结果

```go
type Cache struct {
    mu    sync.RWMutex
    data  map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

// 使用缓存
cache := &Cache{data: make(map[string]string)}

query := "Go 语言"
if cached, ok := cache.Get(query); ok {
    fmt.Println(cached)
} else {
    result, _ := webSearchTool.Execute(ctx, map[string]any{
        "query": query,
    })
    cache.Set(query, result)
    fmt.Println(result)
}
```

### 3. 限流

```go
import "golang.org/x/time/rate"

// 创建限流器（每秒 2 个请求）
limiter := rate.NewLimiter(2, 1)

for _, query := range queries {
    // 等待限流器允许
    limiter.Wait(ctx)

    result, _ := webSearchTool.Execute(ctx, map[string]any{
        "query": query,
    })
    fmt.Println(result)
}
```

## 故障排查

### 问题 1: 搜索无结果

```go
result, err := webSearchTool.Execute(ctx, map[string]any{
    "query": "非常罕见的搜索词",
})

if err != nil && strings.Contains(err.Error(), "未找到") {
    // 尝试更通用的搜索词
    result, err = webSearchTool.Execute(ctx, map[string]any{
        "query": "更通用的搜索词",
    })
}
```

### 问题 2: 内容被截断

```go
// 方法 1: 增加最大内容长度
webConfig := config.WebConfig{
    MaxContentLength: 200000, // 增加到 200KB
}

// 方法 2: 启用智能提取
webConfig := config.WebConfig{
    EnableReadability: true, // 只提取主要内容
}
```

### 问题 3: 请求超时

```go
// 方法 1: 增加超时时间
result, _ := webFetchTool.Execute(ctx, map[string]any{
    "url":     "https://slow-website.com",
    "timeout": 60, // 增加到 60 秒
})

// 方法 2: 使用重试机制
maxRetries := 3
for i := 0; i < maxRetries; i++ {
    result, err := webFetchTool.Execute(ctx, map[string]any{
        "url": "https://example.com",
    })
    if err == nil {
        break
    }
    if i < maxRetries-1 {
        time.Sleep(time.Second * time.Duration(i+1))
    }
}
```

## 最佳实践

1. **使用合理的超时时间**: 根据网络环境调整
2. **限制并发请求**: 避免被封禁
3. **缓存搜索结果**: 减少重复请求
4. **错误处理**: 优雅处理各种错误情况
5. **日志记录**: 记录关键操作便于调试
6. **资源清理**: 及时释放资源

## 相关文档

- [配置指南](../docs/WEB_CONFIG.md)
- [实现总结](../WEB_IMPLEMENTATION_SUMMARY.md)
- [快速开始](../QUICKSTART.md)
