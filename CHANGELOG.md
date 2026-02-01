# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Planned
- SubAgentMiddleware - 子 Agent 委派
- SummarizationMiddleware - 上下文摘要
- MemoryMiddleware - 记忆系统
- SkillsMiddleware - 技能系统
- SandboxBackend - 沙箱执行
- OpenAI 客户端支持

## [0.2.0] - 2026-01-29

### Added
- CompositeBackend - 多后端路由支持
- TodoMiddleware - 任务规划和跟踪
- CLI 工具 (deepagents)
- 大结果驱逐机制 (>80,000 字符)
- 虚拟模式安全验证
- 4 个完整的示例程序
- 完整的文档 (15个文档)

### Changed
- 提高测试覆盖率到 60%+
- 优化中间件架构
- 改进错误处理

### Fixed
- 修复循环导入问题
- 修复并发安全问题

## [0.1.0] - 2026-01-29

### Added
- Agent 核心执行器
- LLM 客户端 (Anthropic Claude)
- 中间件系统
- 工具注册表
- StateBackend (内存存储)
- FilesystemBackend (磁盘存储)
- 6 个文件系统工具:
  - ls - 列出目录
  - read_file - 读取文件
  - write_file - 写入文件
  - edit_file - 编辑文件
  - grep - 搜索文件内容
  - glob - 查找匹配的文件
- FilesystemMiddleware
- 基础测试覆盖
- README 和基础文档

### Technical Details
- Go 1.21+ 支持
- 并发安全设计
- 清晰的接口定义
- 模块化架构

---

## Version History

- **v0.2.0** (2026-01-29) - 核心功能完成
- **v0.1.0** (2026-01-29) - MVP 完成

---

## Upgrade Guide

### From 0.1.0 to 0.2.0

#### New Features
1. **CompositeBackend** - 多后端路由
```go
// 旧方式
backend := backend.NewStateBackend()

// 新方式
composite := backend.NewCompositeBackend(backend.NewStateBackend())
composite.AddRoute("/data", dataBackend)
```

2. **TodoMiddleware** - 任务规划
```go
// 添加 Todo 中间件
todoMiddleware := middleware.NewTodoMiddleware(backend, toolRegistry)
config.Middlewares = append(config.Middlewares, todoMiddleware)
```

3. **CLI 工具**
```bash
# 新增 CLI 工具
./bin/deepagents -prompt "你的任务"
```

#### Breaking Changes
无破坏性变更，完全向后兼容。

---

## Future Roadmap

### v0.3.0 (计划中)
- SubAgentMiddleware
- SummarizationMiddleware
- MemoryMiddleware

### v0.4.0 (计划中)
- SkillsMiddleware
- SandboxBackend
- 性能优化

### v1.0.0 (计划中)
- 完整功能
- 生产级性能
- 完整文档

---

**最后更新**: 2026-01-29
