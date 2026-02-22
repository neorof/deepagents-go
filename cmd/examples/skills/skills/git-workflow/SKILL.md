---
name: git-workflow
description: Git 工作流最佳实践，包括分支管理、提交规范、PR 流程
allowed-tools:
  - Bash
---

# Git 工作流技能

## 分支命名规范

- `feature/xxx` - 新功能开发
- `fix/xxx` - Bug 修复
- `refactor/xxx` - 代码重构
- `docs/xxx` - 文档更新

## 提交信息规范

格式：`<type>: <description>`

类型：
- `feat` - 新功能
- `fix` - Bug 修复
- `refactor` - 重构
- `docs` - 文档
- `test` - 测试
- `chore` - 构建/工具

示例：
```
feat: 添加用户登录功能
fix: 修复空指针异常
refactor: 优化数据库查询性能
```

## 工作流程

1. 创建功能分支
2. 小步提交，保持原子性
3. 提交前运行测试
4. 创建 PR 并请求 review
5. 合并后删除功能分支

## 常用命令

```bash
# 查看状态
git status

# 查看差异
git diff
git diff --staged

# 提交
git add <files>
git commit -m "type: description"

# 分支操作
git checkout -b feature/xxx
git push -u origin feature/xxx
```

## 注意事项

- 不要直接推送到 main/master
- 不要使用 --force（除非明确需要）
- 不要跳过 pre-commit hooks
