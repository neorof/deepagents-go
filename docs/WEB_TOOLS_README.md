# WebSearch 和 WebFetch 工具使用指南

## 概述

WebSearch 和 WebFetch 是 deepagents-go 项目的网络工具，提供搜索网络内容和获取网页的功能。

## 功能特性

### web_search
- 搜索网络内容（使用 DuckDuckGo）
- 返回结果摘要（标题、链接、摘要）
- 支持限制结果数量
- Markdown 格式输出

### web_fetch
- 获取指定 URL 的内容
- 转换为 Markdown 格式
- 智能内容提取（过滤广告、导航栏）
- 支持超时控制和大小限制

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

    // 使用工具
    ctx := context.Background()
    result, err := webSearchTool.Execute(ctx, map[string]any{
        "query":       "Go 语言",
        "max_results": 5,
    })

    if err != nil {
        log.Fatalf("搜索失败: %v", err)
    }

    fmt.Println(result)
}
```

### 2. 与 Agent 集成

```go
// 创建 Agent
agentConfig := &agent.Config{
    LLMClient:    llmClient,
    ToolRegistry: toolRegistry,
    Middlewares:  []agent.Middleware{fsMiddleware, webMiddleware},
    SystemPrompt: "你是一个有用的 AI 助手...",
}

executor := agent.NewExecutor(agentConfig)

// 执行任务
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "搜索 'Go 语言最佳实践'，并保存结果",
        },
    },
})
```

## 配置

### 默认配置

```go
webConfig := config.DefaultWebConfig()
// {
//     SearchEngine:      "duckduckgo",
//     DefaultTimeout:    30,
//     MaxContentLength:  100000,
//     EnableReadability: true,
// }
```

### 自定义配置

```go
webConfig := config.WebConfig{
    SearchEngine:      "duckduckgo",
    DefaultTimeout:    60,              // 增加超时时间
    MaxContentLength:  50000,           // 减小最大内容长度
    EnableReadability: true,            // 启用智能提取
}
```

### 配置文件

创建 `config.yaml`:

```yaml
web:
  search_engine: "duckduckgo"
  default_timeout: 30
  max_content_length: 100000
  enable_readability: true
```

## 工具参数

### web_search

| 参数 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| query | string | 是 | - | 搜索关键词 |
| max_results | integer | 否 | 5 | 最多返回结果数（1-10） |

**示例**:
```go
result, _ := webSearchTool.Execute(ctx, map[string]any{
    "query":       "Go 语言最佳实践",
    "max_results": 3,
})
```

### web_fetch

| 参数 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| url | string | 是 | - | 要获取的 URL（仅支持 http/https） |
| timeout | integer | 否 | 30 | 超时时间（秒，1-300） |

**示例**:
```go
result, _ := webFetchTool.Execute(ctx, map[string]any{
    "url":     "https://go.dev",
    "timeout": 30,
})
```

## 常见任务

### 任务 1: 搜索并保存

```go
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "搜索 'Go 语言性能优化'，并将结果保存到 /performance.md",
        },
    },
})
```

### 任务 2: 获取网页内容

```go
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: "获取 https://go.dev/doc/ 的内容，并提取主要章节",
        },
    },
})
```

### 任务 3: 研究主题

```go
output, _ := executor.Invoke(ctx, &agent.InvokeInput{
    Messages: []llm.Message{
        {
            Role:    llm.RoleUser,
            Content: `研究 Go 语言测试最佳实践：
1. 搜索相关内容
2. 获取前 3 个结果的详细内容
3. 总结关键要点
4. 保存到 /testing_best_practices.md`,
        },
    },
})
```

## 错误处理

### 搜索失败

```go
result, err := webSearchTool.Execute(ctx, map[string]any{
    "query": "test",
})

if err != nil {
    if strings.Contains(err.Error(), "超时") {
        log.Println("请求超时，请稍后重试")
    } else if strings.Contains(err.Error(), "未找到") {
        log.Println("没有找到相关结果")
    } else {
        log.Printf("搜索失败: %v", err)
    }
    return
}
```

### 获取失败

```go
result, err := webFetchTool.Execute(ctx, map[string]any{
    "url": "https://example.com",
})

if err != nil {
    if strings.Contains(err.Error(), "无效的 URL") {
        log.Println("URL 格式错误")
    } else if strings.Contains(err.Error(), "HTTP 错误") {
        log.Println("网站返回错误")
    } else {
        log.Printf("获取失败: %v", err)
    }
    return
}
```

## 性能优化

### 1. 并发搜索

```go
var wg sync.WaitGroup
results := make(chan string, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        result, _ := webSearchTool.Execute(ctx, map[string]any{
            "query": q,
        })
        results <- result
    }(query)
}

wg.Wait()
close(results)
```

### 2. 超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := webFetchTool.Execute(ctx, map[string]any{
    "url": "https://example.com",
})
```

### 3. 调整配置

```go
// 增加超时时间
webConfig.DefaultTimeout = 60

// 减小内容长度
webConfig.MaxContentLength = 50000

// 禁用智能提取（提升性能）
webConfig.EnableReadability = false
```

## 安全注意事项

1. **URL 验证**: 工具会验证 URL 格式，仅支持 http/https 协议
2. **大小限制**: HTTP 响应限制为 10MB，工具输出限制为配置的 max_content_length
3. **超时控制**: 所有请求都有超时限制，避免长时间阻塞
4. **错误处理**: 完整的错误处理，避免程序崩溃

## 故障排查

### 问题 1: 搜索失败

**症状**: 搜索返回错误或无结果

**解决方案**:
1. 检查网络连接
2. 增加超时时间
3. 稍后重试

### 问题 2: 内容获取失败

**症状**: web_fetch 返回错误

**解决方案**:
1. 验证 URL 格式
2. 检查网站是否可访问
3. 增加超时时间

### 问题 3: 内容被截断

**症状**: 返回的内容不完整

**解决方案**:
1. 增加 max_content_length 值
2. 或使用智能提取减少无关内容

## 示例程序

运行完整示例：

```bash
export ANTHROPIC_API_KEY=your_api_key
go run ./cmd/examples/web/main.go
```

## 相关文档

- [快速开始](WEB_QUICKSTART.md)
- [配置指南](WEB_CONFIG.md)
- [使用示例](WEB_EXAMPLES.md)
- [实现总结](../WEB_IMPLEMENTATION_SUMMARY.md)

## 技术支持

如有问题或建议，请访问:
https://github.com/zhoucx/deepagents-go

---

**版本**: v1.0
**更新日期**: 2026-02-02
