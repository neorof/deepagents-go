package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/go-shiori/go-readability"
)

const (
	defaultMaxResults   = 5
	maxResultsLimit     = 10
	defaultFetchTimeout = 30
	maxFetchTimeout     = 300
	maxBodySize         = 10 * 1024 * 1024 // 10MB
	defaultUserAgent    = "Mozilla/5.0 (compatible; DeepAgents/1.0)"
)

// NewWebSearchTool 创建 web_search 工具
func NewWebSearchTool(engine SearchEngine) Tool {
	return NewBaseTool(
		"web_search",
		"搜索网络内容并返回结果摘要",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "搜索关键词",
				},
				"max_results": map[string]any{
					"type":        "integer",
					"description": "最多返回结果数（默认 5）",
					"default":     defaultMaxResults,
				},
			},
			"required": []string{"query"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			query, ok := args["query"].(string)
			if !ok || query == "" {
				return "", fmt.Errorf("query 必须是非空字符串")
			}

			maxResults := GetIntArg(args, "max_results", defaultMaxResults)
			maxResults = Clamp(maxResults, 1, maxResultsLimit)

			results, err := engine.Search(ctx, query, maxResults)
			if err != nil {
				return "", fmt.Errorf("搜索失败: %w", err)
			}

			return formatSearchResults(query, results), nil
		},
	)
}

// formatSearchResults 格式化搜索结果为 Markdown
func formatSearchResults(query string, results []SearchResult) string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("搜索关键词: %s\n", query))
	output.WriteString(fmt.Sprintf("找到 %d 条结果:\n\n", len(results)))

	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. **[%s](%s)**\n", i+1, result.Title, result.URL))
		if result.Snippet != "" {
			output.WriteString(fmt.Sprintf("   %s\n\n", result.Snippet))
		} else {
			output.WriteString("\n")
		}
	}

	return output.String()
}

// NewWebFetchTool 创建 web_fetch 工具
func NewWebFetchTool(enableReadability bool, maxContentLength int) Tool {
	return NewBaseTool(
		"web_fetch",
		"获取指定 URL 的内容并转换为 Markdown",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"url": map[string]any{
					"type":        "string",
					"description": "要获取的 URL",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "超时时间（秒），默认 30 秒",
					"default":     defaultFetchTimeout,
				},
			},
			"required": []string{"url"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			urlStr, ok := args["url"].(string)
			if !ok || urlStr == "" {
				return "", fmt.Errorf("url 必须是非空字符串")
			}

			parsedURL, err := url.Parse(urlStr)
			if err != nil {
				return "", fmt.Errorf("无效的 URL: %w", err)
			}

			if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
				return "", fmt.Errorf("只支持 http 和 https 协议")
			}

			timeout := GetIntArg(args, "timeout", defaultFetchTimeout)
			timeout = Clamp(timeout, 1, maxFetchTimeout)

			content, err := fetchURL(ctx, urlStr, parsedURL, timeout, enableReadability)
			if err != nil {
				return "", err
			}

			return truncateContent(content, maxContentLength), nil
		},
	)
}

// fetchURL 获取 URL 内容并转换为 Markdown
func fetchURL(ctx context.Context, urlStr string, parsedURL *url.URL, timeout int, enableReadability bool) (string, error) {
	fetchCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(fetchCtx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if fetchCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("请求超时（%d 秒）", timeout)
		}
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 错误: %d %s", resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "html") {
		return "", fmt.Errorf("不支持的内容类型: %s（仅支持 HTML）", contentType)
	}

	limitedReader := io.LimitReader(resp.Body, maxBodySize)

	if !enableReadability {
		return convertHTMLToMarkdown(limitedReader)
	}

	return extractWithReadability(limitedReader, parsedURL, urlStr)
}

// extractWithReadability 使用 readability 提取主要内容
func extractWithReadability(reader io.Reader, parsedURL *url.URL, urlStr string) (string, error) {
	article, err := readability.FromReader(reader, parsedURL)
	if err != nil {
		// 降级到直接转换
		return convertHTMLToMarkdown(reader)
	}

	markdown, err := convertHTMLToMarkdown(strings.NewReader(article.Content))
	if err != nil {
		return "", fmt.Errorf("Markdown 转换失败: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# %s\n\n", article.Title))
	if article.Byline != "" {
		output.WriteString(fmt.Sprintf("**作者**: %s\n\n", article.Byline))
	}
	output.WriteString(fmt.Sprintf("**来源**: %s\n\n", urlStr))
	output.WriteString("---\n\n")
	output.WriteString(markdown)

	return output.String(), nil
}

// truncateContent 截断内容并添加提示
func truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	originalLen := len(content)
	return fmt.Sprintf("%s\n\n... (内容已截断，原始长度: %d 字符，已显示: %d 字符)",
		content[:maxLength], originalLen, maxLength)
}

// convertHTMLToMarkdown 将 HTML 转换为 Markdown
func convertHTMLToMarkdown(reader io.Reader) (string, error) {
	markdownBytes, err := htmltomarkdown.ConvertReader(reader)
	if err != nil {
		return "", fmt.Errorf("Markdown 转换失败: %w", err)
	}
	return string(markdownBytes), nil
}
