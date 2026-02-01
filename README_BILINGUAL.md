# Deep Agents Go

[English](#english) | [中文](#中文)

---

## 中文

基于 Go 语言实现的 AI Agent 框架，提供任务规划、文件操作、多后端存储等能力。

### 特性

- ✅ **中间件架构**：模块化设计，功能可组合
- ✅ **多 LLM 支持**：支持 Anthropic Claude
- ✅ **文件系统操作**：完整的文件读写、搜索、编辑功能
- ✅ **工具系统**：可扩展的工具注册机制
- ✅ **多后端存储**：内存、磁盘、组合后端
- ✅ **任务规划**：Todo 列表管理
- ✅ **大结果驱逐**：自动保存大结果到文件
- ✅ **虚拟模式**：路径安全验证

### 快速开始

#### 安装

```bash
go get github.com/zhoucx/deepagents-go
```

#### 前置要求

- Go 1.21 或更高版本
- Anthropic API Key

#### 基础使用

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
    output, _ := executor.Invoke(context.Background(), &agent.InvokeInput{
        Messages: []llm.Message{
            {Role: llm.RoleUser, Content: "创建文件 /test.txt"},
        },
    })
}
```

### 示例程序

```bash
export ANTHROPIC_API_KEY=your_api_key

# 基础示例
go run ./cmd/examples/basic/main.go

# 文件系统示例
go run ./cmd/examples/filesystem/main.go

# Todo 管理示例
go run ./cmd/examples/todo/main.go

# 多后端路由示例
go run ./cmd/examples/composite/main.go
```

### 测试

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make test-coverage
```

### 文档

- [实现计划](IMPLEMENTATION_PLAN.md)
- [阶段 1 总结](STAGE1_SUMMARY.md)
- [项目总结](PROJECT_SUMMARY.md)
- [贡献指南](CONTRIBUTING.md)

### 许可证

MIT License

---

## English

An AI Agent framework implemented in Go, providing task planning, file operations, multi-backend storage, and more.

### Features

- ✅ **Middleware Architecture**: Modular design with composable functionality
- ✅ **Multi-LLM Support**: Supports Anthropic Claude
- ✅ **File System Operations**: Complete file read/write, search, and edit capabilities
- ✅ **Tool System**: Extensible tool registration mechanism
- ✅ **Multi-Backend Storage**: Memory, disk, and composite backends
- ✅ **Task Planning**: Todo list management
- ✅ **Large Result Eviction**: Automatically save large results to files
- ✅ **Virtual Mode**: Path security validation

### Quick Start

#### Installation

```bash
go get github.com/zhoucx/deepagents-go
```

#### Prerequisites

- Go 1.21 or higher
- Anthropic API Key

#### Basic Usage

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
    // Create LLM client
    llmClient := llm.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY"), "")

    // Create tool registry and backend
    toolRegistry := tools.NewRegistry()
    backend := backend.NewStateBackend()

    // Create middleware
    fsMiddleware := middleware.NewFilesystemMiddleware(backend, toolRegistry)

    // Create Agent
    config := &agent.Config{
        LLMClient:    llmClient,
        ToolRegistry: toolRegistry,
        Middlewares:  []agent.Middleware{fsMiddleware},
    }
    executor := agent.NewExecutor(config)

    // Execute task
    output, _ := executor.Invoke(context.Background(), &agent.InvokeInput{
        Messages: []llm.Message{
            {Role: llm.RoleUser, Content: "Create file /test.txt"},
        },
    })
}
```

### Examples

```bash
export ANTHROPIC_API_KEY=your_api_key

# Basic example
go run ./cmd/examples/basic/main.go

# Filesystem example
go run ./cmd/examples/filesystem/main.go

# Todo management example
go run ./cmd/examples/todo/main.go

# Multi-backend routing example
go run ./cmd/examples/composite/main.go
```

### Testing

```bash
# Run all tests
make test

# Generate coverage report
make test-coverage
```

### Documentation

- [Implementation Plan](IMPLEMENTATION_PLAN.md)
- [Stage 1 Summary](STAGE1_SUMMARY.md)
- [Project Summary](PROJECT_SUMMARY.md)
- [Contributing Guide](CONTRIBUTING.md)

### License

MIT License
