# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Deep Agents Go 是一个用 Go 实现的 AI Agent 框架，支持工具调用、中间件系统和多种 LLM 后端（Anthropic Claude、OpenAI）。

## 常用命令

### 构建和测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 构建所有示例程序
make build

# 格式化代码（必须在修改代码后执行）
make fmt
# 或
gofmt -w .

# 运行 linter
make lint

# 更新依赖
make deps
```

### 运行单个测试
```bash
# 运行特定包的测试
go test -v ./pkg/agent

# 运行特定测试函数
go test -v ./pkg/agent -run TestExecutor

# 运行测试并显示覆盖率
go test -v -cover ./pkg/agent
```

### 运行示例程序
```bash
# 运行主 CLI 程序
go run ./cmd/deepagents

# 运行示例程序
make run-basic          # 基础示例
make run-filesystem     # 文件系统示例
make run-todo          # Todo 示例
make run-composite     # 组合后端示例
make run-skills        # 技能系统示例
make run-bash          # Bash 工具示例

# 或直接运行
go run ./cmd/examples/basic/main.go
```

## 核心架构

### 包结构和依赖关系

**关键约束**：避免循环依赖
- `pkg/agent` 定义核心接口（Agent, Middleware, Executor, Config）
- `pkg/middleware` 导入 `pkg/agent`，因此 `pkg/agent` **不能**反向导入 `pkg/middleware`
- `pkg/agentkit` 是高层封装包，可同时导入 `pkg/agent` 和 `pkg/middleware`

```
pkg/agent/          # 核心接口和执行器
  ├── agent.go      # Agent, Middleware 接口定义
  ├── executor.go   # Executor 实现（Invoke, InvokeStream）
  └── state.go      # Agent 状态管理

pkg/middleware/     # 中间件实现（依赖 pkg/agent）
  ├── filesystem.go # 文件系统中间件
  ├── agent_config.go # Agent 配置中间件
  ├── context_injection.go # 上下文注入中间件
  ├── memory.go     # 对话历史记录中间件
  ├── todo.go       # Todo 管理中间件
  ├── web.go        # Web 工具中间件
  ├── skills.go     # 技能系统中间件
  ├── summarization.go  # 摘要中间件
  ├── subagent.go   # 子 Agent 中间件
  ├── session_record.go # 会话记录数据结构
  └── counter.go    # 轮次计数器组件

pkg/agentkit/       # 高层封装（Option 模式）
  ├── agent_builder.go  # AgentBuilder 实现
  ├── options.go    # Option 函数
  └── callbacks.go  # 默认回调处理器

pkg/llm/            # LLM 客户端
  ├── client.go     # Client 接口
  ├── anthropic.go  # Anthropic 实现
  ├── openai.go     # OpenAI 实现
  └── message.go    # 消息和流式事件定义

pkg/tools/          # 工具系统
  ├── tool.go       # Tool 接口
  ├── registry.go   # 工具注册表
  ├── filesystem.go # 文件系统工具（Read, Write, Edit, Glob, Grep）
  ├── web.go        # Web 工具（WebFetch）
  └── web_search.go # Web 搜索工具

pkg/backend/        # 存储后端
  ├── backend.go    # Backend 接口
  ├── filesystem.go # 文件系统后端
  ├── state.go      # 内存状态后端
  ├── sandbox.go    # 沙箱后端（安全隔离）
  └── composite.go  # 组合后端（多后端聚合）

internal/           # 内部工具包
  ├── config/       # 配置管理
  ├── logger/       # 日志系统
  ├── progress/     # 进度显示
  ├── repl/         # 交互式 REPL
  └── color/        # 终端颜色输出
```

### 核心设计模式

#### 1. Option 模式构建 Agent
使用 `agentkit.New()` 通过链式调用构建 Agent：

```go
agent := agentkit.New(
    agentkit.WithLLM(llmClient),
    agentkit.WithConfig(agentkit.AgentConfig{
        SystemPrompt:  "...",
        MaxIterations: 25,
        MaxTokens:     4096,
        Temperature:   0.8,
    }),
    agentkit.WithFilesystem("./workspace"),
    agentkit.WithSkillsDirs("skills"),
    agentkit.WithMemoryPaths("memories"),
)
```

**重要**：
- `WithFilesystem()` 是必须的，`Build()` 会校验
- `Build()` 默认启用所有内置中间件：Filesystem、AgentConfig、ContextInjection、Memory、Todo、Web、Skills、SubAgent
- 只有 Summarization 是可选的，需要显式调用 `EnableSummarization()`
- `WithSkillsDirs()`、`WithSessionID()`、`WithWebConfig()` 用于传递配置参数
- `WithMemoryPaths()` 已废弃（保留向后兼容），新的 MemoryMiddleware 自动记录对话历史

#### 2. 中间件系统
中间件通过钩子函数拦截 Agent 执行流程：

```go
type Middleware interface {
    Name() string
    BeforeAgent(ctx context.Context, state *State) error
    BeforeModel(ctx context.Context, req *ModelRequest) error
    AfterModel(ctx context.Context, resp *ModelResponse, state *State) error
    BeforeTool(ctx context.Context, toolCall *ToolCall, state *State) error
    AfterTool(ctx context.Context, result *ToolResult, state *State) error
}
```

执行顺序：
1. `BeforeAgent` - Agent 开始执行前
2. 循环：
   - `BeforeModel` - 调用 LLM 前
   - LLM 生成响应
   - `AfterModel` - LLM 响应后
   - 对每个工具调用：
     - `BeforeTool` - 工具执行前
     - 执行工具
     - `AfterTool` - 工具执行后

#### 3. 流式响应
支持两种执行模式：
- `Invoke()` - 非流式，等待完整响应
- `InvokeStream()` - 流式，实时接收事件

流式事件类型：
- `AgentEventTypeStart` - Agent 开始
- `AgentEventTypeLLMStart` - LLM 开始生成
- `AgentEventTypeLLMText` - LLM 文本内容（增量）
- `AgentEventTypeLLMToolCall` - LLM 工具调用
- `AgentEventTypeLLMEnd` - LLM 生成结束
- `AgentEventTypeToolStart` - 工具开始执行
- `AgentEventTypeToolResult` - 工具执行结果
- `AgentEventTypeIterationEnd` - 迭代结束
- `AgentEventTypeEnd` - Agent 执行结束
- `AgentEventTypeError` - 错误

#### 4. 技能系统（Skills）
技能是可重用的提示词模板，存储在 `skills/` 目录：

```
skills/
  ├── file-operations/
  │   └── SKILL.md
  └── git-workflow/
      └── SKILL.md
```

SKILL.md 格式：
```markdown
---
name: skill-name
description: 技能描述
allowed-tools:
  - Read
  - Write
---

# 技能内容
...
```

技能通过 `SkillsMiddleware` 加载并注入到系统提示词中。

## 开发规范

### 错误处理
- 返回 `error` 而非 `panic`（panic 仅用于不可恢复的错误）
- 使用 `fmt.Errorf("context: %w", err)` 包装错误
- 每个返回 `error` 的函数调用都必须检查
- 快速失败：尽早返回错误，避免深层嵌套

### 测试规范
- 测试文件命名：`*_test.go`
- 测试函数命名：`TestXxx` 或 `TestXxx_Scenario`
- 使用表驱动测试（table-driven tests）
- Mock LLM 客户端用于测试（避免真实 API 调用）
- 测试覆盖率目标：70%+

### 代码格式化
**关键**：修改代码后必须运行 `gofmt -w .` 或 `make fmt`
- Write 工具生成的代码可能不符合 gofmt 标准
- CI 会检查格式，未格式化的代码会导致构建失败

### 命名规范
- 包名：小写单词，不使用下划线或驼峰
- 导出标识符：大写字母开头
- 接口名：通常以 `-er` 结尾（Reader, Writer, Executor）
- 简短清晰优于冗长（`i` 而非 `index`，`buf` 而非 `buffer`）

## 常见任务

### 添加新的中间件
1. 在 `pkg/middleware/` 创建新文件
2. 实现 `agent.Middleware` 接口（可继承 `BaseMiddleware`）
3. 在 `pkg/agentkit/options.go` 添加 Option 函数
4. 在 `pkg/agentkit/agent_builder.go` 的 `build()` 方法中处理
5. 添加测试文件 `*_test.go`

### 添加新的工具
1. 在 `pkg/tools/` 实现 `Tool` 接口
2. 在相应的中间件中注册工具到 `Registry`
3. 添加测试覆盖

### 添加新的 LLM 客户端
1. 在 `pkg/llm/` 实现 `Client` 接口
2. 实现 `Generate()` 和 `StreamGenerate()` 方法
3. 添加测试（使用 mock 避免真实 API 调用）

### 调试技巧
```bash
# 启用详细日志
export LOG_LEVEL=debug
go run ./cmd/deepagents

# 运行单个测试并显示详细输出
go test -v ./pkg/agent -run TestExecutor

# 使用 delve 调试器
dlv test ./pkg/agent -- -test.run TestExecutor
```

## 配置文件

主 CLI 程序支持配置文件（`~/.deepagents/config.yaml`）：

```yaml
api_key: "your-api-key"
base_url: "https://api.anthropic.com"  # 可选
model: "claude-sonnet-4-5-20250929"
work_dir: "./"
system_prompt_file: "system_prompt.txt"
max_iterations: 25
max_tokens: 4096
temperature: 0.8
log_level: "info"  # debug, info, warn, error
log_file: ""       # 空表示输出到标准输出
log_format: "text" # text, json
enable_streaming: true
```

## 重要注意事项

1. **循环依赖**：`pkg/agent` 不能导入 `pkg/middleware`
2. **测试中的 log.Fatalf**：会直接退出进程，无法用 defer/recover 捕获，需直接测试 `build()` 方法
3. **gofmt 必须运行**：Write 工具生成的缩进可能不符合标准
4. **中间件依赖**：在 `build()` 阶段统一处理，Option 声明顺序无关
5. **REPL 依赖**：`internal/repl` 依赖 `*agent.Executor`，不需要改动

## 项目状态

- **完成度**：95%
- **测试覆盖率**：76.8%
- **代码规模**：13,000+ 行，68+ 个 Go 文件
- **示例程序**：13 个
- **中间件**：6 个（Filesystem, Todo, Web, Skills, Memory, Summarization, SubAgent）
- **工具**：8+ 个（Read, Write, Edit, Glob, Grep, Bash, WebFetch, WebSearch）
- **后端**：4 个（Filesystem, State, Sandbox, Composite）
