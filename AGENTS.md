# Agent 内存示例

本文件演示如何使用 MemoryMiddleware 为你的 AI agent 提供上下文和内存功能。

## 系统信息

该 agent 旨在帮助进行文件管理和任务执行。

## 可用工具

- `write_file`: 创建或覆盖文件
- `read_file`: 读取文件内容
- `edit_file`: 使用字符串替换编辑文件
- `ls`: 列出目录内容
- `grep`: 在文件中搜索模式
- `glob`: 查找匹配模式的文件
- `write_todos`: 创建任务计划

## 最佳实践

1. 操作前始终验证文件路径
2. 使用描述性的提交信息
3. 提交前测试更改
4. 将文件组织在适当的目录中

## 项目上下文

这是一个基于 Go 的 AI agent 框架，具有以下结构：
- `pkg/`: 核心包（agent、backend、llm、middleware、tools）
- `cmd/`: 命令行工具和示例
- `internal/`: 内部包（config、logger、repl、progress）
- `tests/`: 集成测试

## 指南

- 遵循 Go 编码规范
- 为新功能编写测试
- 保持函数小而专注
- 使用有意义的变量名
- 为导出的函数编写文档
