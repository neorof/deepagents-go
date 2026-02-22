# WebSearch 和 WebFetch 工具快速开始

## 5分钟快速上手

### 1. 安装依赖

```bash
cd /home/zhoucx/go/deepagents-go
go mod tidy
```

### 2. 运行示例程序

```bash
# 设置 API Key
export ANTHROPIC_API_KEY=your_api_key

# 运行 Web 工具示例
go run ./cmd/examples/web/main.go
```

### 3. 查看结果

示例程序会执行以下任务：
1. 搜索 "Go 语言 2026 新特性"
2. 获取 https://go.dev 首页内容
3. 搜索 AI 新闻并保存到文件
4. 获取技术文章并保存

所有结果保存在 `./workspace` 目录。

## 基本使用

### 在代码中使用

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
    // 1. 创建工具注册表
    toolRegistry := tools.NewRegistry()

    // 2. 创建 Web 配置（使用默认配置）
    webConfig := config.DefaultWebConfig()

    // 3. 创建 Web 中间件
    webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

    // 4. 获取工具
    webSearchTool, _ := toolRegistry.Get("web_search")

    // 5. 使用工具
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

### 自定义配置

```go
// 创建自定义配置
webConfig := config.WebConfig{
    SearchEngine:      "duckduckgo",
    DefaultTimeout:    60,              // 增加超时时间
    MaxContentLength:  50000,           // 减小最大内容长度
    EnableReadability: true,            // 启用智能提取
}

// 使用自定义配置创建中间件
webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)
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

## 工具说明

### web_search

搜索网络内容并返回结果摘要。

**参数**:
- `query` (string, required): 搜索关键词
- `max_results` (integer, optional): 最多返回结果数（默认 5）

**示例**:
```go
result, _ := webSearchTool.Execute(ctx, map[string]any{
    "query":       "Go 语言最佳实践",
    "max_results": 3,
})
```

### web_fetch

获取指定 URL 的内容并转换为 Markdown。

**参数**:
- `url` (string, required): 要获取的 URL
- `timeout` (integer, optional): 超时时间（秒，默认 30）

**示例**:
```go
result, _ := webFetchTool.Execute(ctx, map[string]any{
    "url":     "https://go.dev",
    "timeout": 30,
})
```

## 与 Agent 集成

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

    // 创建文件系统中间件
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
                Content: "搜索 'Go 语言并发编程'，并将结果保存到 /go_concurrency.md",
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

## 验证安装

运行验证脚本：

```bash
./scripts/verify_web.sh
```

预期输出：
```
==========================================
Web 工具实现验证
==========================================

1. 编译检查...
✅ 编译通过

2. 运行单元测试...
✅ 单元测试通过

3. 运行集成测试...
✅ 集成测试通过

4. 生成测试覆盖率...
✅ 覆盖率检查完成

5. 检查代码格式...
✅ 代码格式正确

6. 检查文件完整性...
✅ 所有文件完整

7. 检查依赖...
✅ 依赖检查通过

8. 代码统计...
✅ 统计完成

==========================================
✅ 所有验证通过！
==========================================
```

## 故障排查

### 问题 1: 搜索失败

**症状**: 搜索返回错误

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

### 问题 3: 编译错误

**症状**: 缺少依赖

**解决方案**:
```bash
go mod tidy
go mod download
```

## 下一步

1. **查看完整文档**:
   - 配置指南: `docs/WEB_CONFIG.md`
   - 使用示例: `docs/WEB_EXAMPLES.md`
   - 实现总结: `WEB_IMPLEMENTATION_SUMMARY.md`

2. **运行示例程序**:
   ```bash
   go run ./cmd/examples/web/main.go
   ```

3. **集成到你的项目**:
   - 参考上面的代码示例
   - 根据需求自定义配置

4. **反馈和改进**:
   - 遇到问题请提 Issue
   - 欢迎贡献代码

## 相关资源

- 项目主页: https://github.com/zhoucx/deepagents-go
- 配置指南: docs/WEB_CONFIG.md
- 使用示例: docs/WEB_EXAMPLES.md
- 实现总结: WEB_IMPLEMENTATION_SUMMARY.md

---

**快速开始完成！** 🎉

现在你可以开始使用 WebSearch 和 WebFetch 工具了。
