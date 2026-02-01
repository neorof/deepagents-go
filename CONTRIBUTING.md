# 贡献指南

感谢你对 Deep Agents Go 的关注！

## 开发环境设置

1. 克隆仓库
```bash
git clone https://github.com/zhoucx/deepagents-go.git
cd deepagents-go
```

2. 安装依赖
```bash
make deps
```

3. 运行测试
```bash
make test
```

## 代码规范

### Go 代码风格
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 使用有意义的变量名
- 添加必要的注释

### 提交规范
- 提交信息使用中文
- 格式：`<类型>: <描述>`
- 类型：
  - `feat`: 新功能
  - `fix`: 修复 bug
  - `docs`: 文档更新
  - `test`: 测试相关
  - `refactor`: 重构
  - `chore`: 构建/工具相关

示例：
```
feat: 添加 OpenAI 客户端支持
fix: 修复文件路径遍历漏洞
docs: 更新 README 示例代码
```

## 开发流程

1. **创建分支**
```bash
git checkout -b feature/your-feature-name
```

2. **编写代码**
- 遵循现有代码风格
- 添加必要的测试
- 确保测试通过

3. **提交代码**
```bash
make fmt          # 格式化代码
make test         # 运行测试
git add .
git commit -m "feat: 你的功能描述"
```

4. **推送并创建 PR**
```bash
git push origin feature/your-feature-name
```

## 测试要求

- 新功能必须包含单元测试
- 测试覆盖率应保持在 80% 以上
- 所有测试必须通过

```bash
# 运行测试
make test

# 查看覆盖率
make test-coverage
```

## 文档要求

- 公开的函数和类型必须有注释
- 复杂的逻辑需要添加说明
- 更新相关的 README 和文档

## 问题反馈

如果你发现 bug 或有功能建议，请：

1. 在 GitHub Issues 中搜索是否已有相关问题
2. 如果没有，创建新的 Issue
3. 提供详细的描述和复现步骤

## 代码审查

所有 PR 都需要经过代码审查：

- 代码风格是否符合规范
- 测试是否充分
- 文档是否完整
- 是否有潜在的 bug

## 许可证

通过提交代码，你同意你的贡献将使用 MIT 许可证。
