# Web 工具配置指南

## 概述

Web 工具提供网络搜索和网页内容获取功能，支持通过配置文件自定义行为。

## 配置文件格式

支持 YAML 和 JSON 格式。

### YAML 格式示例

```yaml
# config.yaml
web:
  search_engine: "duckduckgo"  # 搜索引擎类型
  serpapi_key: ""               # SerpAPI Key（可选）
  default_timeout: 30           # 默认超时时间（秒）
  max_content_length: 100000    # 最大内容长度（字符）
  enable_readability: true      # 是否启用智能内容提取
```

### JSON 格式示例

```json
{
  "web": {
    "search_engine": "duckduckgo",
    "serpapi_key": "",
    "default_timeout": 30,
    "max_content_length": 100000,
    "enable_readability": true
  }
}
```

## 配置项说明

### search_engine

**类型**: `string`
**默认值**: `"duckduckgo"`
**可选值**: `"duckduckgo"`, `"serpapi"`

搜索引擎类型：
- `duckduckgo`: 使用 DuckDuckGo HTML 搜索（免费，无需 API Key）
- `serpapi`: 使用 SerpAPI（需要 API Key，暂未实现）

### serpapi_key

**类型**: `string`
**默认值**: `""`
**说明**: SerpAPI 的 API Key（仅在 `search_engine` 为 `serpapi` 时需要）

### default_timeout

**类型**: `int`
**默认值**: `30`
**单位**: 秒
**范围**: 1-300

HTTP 请求的默认超时时间。

### max_content_length

**类型**: `int`
**默认值**: `100000`
**单位**: 字符

工具返回内容的最大长度。超过此长度的内容将被截断。

### enable_readability

**类型**: `bool`
**默认值**: `true`

是否启用智能内容提取：
- `true`: 使用 go-readability 提取网页主要内容，过滤广告、导航栏等
- `false`: 直接转换整个 HTML 为 Markdown

## 使用示例

### 代码中使用

```go
package main

import (
    "github.com/zhoucx/deepagents-go/pkg/config"
    "github.com/zhoucx/deepagents-go/pkg/middleware"
    "github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
    // 使用默认配置
    webConfig := config.DefaultWebConfig()

    // 或自定义配置
    webConfig := config.WebConfig{
        SearchEngine:      "duckduckgo",
        DefaultTimeout:    60,
        MaxContentLength:  50000,
        EnableReadability: true,
    }

    // 创建工具注册表和中间件
    toolRegistry := tools.NewRegistry()
    webMiddleware := middleware.NewWebMiddleware(toolRegistry, webConfig)

    // 使用中间件...
}
```

### 从文件加载配置

```go
package main

import (
    "log"

    "github.com/zhoucx/deepagents-go/internal/config"
)

func main() {
    // 加载配置文件
    cfg, err := config.Load("config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 使用 Web 配置
    webConfig := cfg.Web

    // 创建中间件...
}
```

## 工具说明

### web_search

搜索网络内容并返回结果摘要。

**参数**:
- `query` (string, required): 搜索关键词
- `max_results` (integer, optional): 最多返回结果数（默认 5，范围 1-10）

**返回格式**:
```markdown
搜索关键词: {query}
找到 {count} 条结果:

1. **[标题1](链接1)**
   摘要内容...

2. **[标题2](链接2)**
   摘要内容...
```

### web_fetch

获取指定 URL 的内容并转换为 Markdown。

**参数**:
- `url` (string, required): 要获取的 URL（仅支持 http/https）
- `timeout` (integer, optional): 超时时间（秒，默认 30，范围 1-300）

**返回格式**:
```markdown
# 页面标题

**作者**: 作者名称（如有）

**来源**: URL

---

页面内容（Markdown 格式）...
```

## 性能优化建议

### 1. 调整超时时间

对于网络较慢的环境，可以增加超时时间：

```yaml
web:
  default_timeout: 60  # 增加到 60 秒
```

### 2. 限制内容长度

为避免 Token 超限，可以减小最大内容长度：

```yaml
web:
  max_content_length: 50000  # 减小到 50KB
```

### 3. 禁用智能提取

如果不需要智能内容提取，可以禁用以提升性能：

```yaml
web:
  enable_readability: false
```

## 故障排查

### 问题 1: 搜索失败

**症状**: 搜索返回错误或无结果

**可能原因**:
1. 网络连接问题
2. DuckDuckGo 服务不可用
3. 超时时间过短

**解决方案**:
1. 检查网络连接
2. 增加超时时间
3. 稍后重试

### 问题 2: 内容获取失败

**症状**: web_fetch 返回错误

**可能原因**:
1. URL 无效或不可访问
2. 网站返回非 HTML 内容
3. 超时

**解决方案**:
1. 验证 URL 格式
2. 检查网站是否可访问
3. 增加超时时间

### 问题 3: 内容被截断

**症状**: 返回的内容不完整

**可能原因**:
- 内容超过 `max_content_length` 限制

**解决方案**:
- 增加 `max_content_length` 值
- 或使用智能提取减少无关内容

## 安全注意事项

1. **URL 验证**: 工具会验证 URL 格式，仅支持 http/https 协议
2. **大小限制**: HTTP 响应限制为 10MB，工具输出限制为配置的 `max_content_length`
3. **超时控制**: 所有请求都有超时限制，避免长时间阻塞
4. **User-Agent**: 请求使用标识 `Mozilla/5.0 (compatible; DeepAgents/1.0)`

## 未来计划

- [ ] 支持 SerpAPI 搜索引擎
- [ ] 支持更多搜索引擎（Google, Bing 等）
- [ ] 添加缓存机制
- [ ] 添加限流保护
- [ ] 支持代理配置
- [ ] 支持自定义 User-Agent

## 相关文档

- [快速开始指南](QUICKSTART.md)
- [使用手册](USER_MANUAL.md)
- [示例程序](cmd/examples/web/main.go)
