package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// SearchEngine 搜索引擎接口
type SearchEngine interface {
	Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error)
}

// SearchResult 搜索结果
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// DuckDuckGoEngine DuckDuckGo 搜索引擎实现
type DuckDuckGoEngine struct {
	client  *http.Client
	timeout time.Duration
}

// NewDuckDuckGoEngine 创建 DuckDuckGo 搜索引擎
func NewDuckDuckGoEngine(timeout time.Duration) *DuckDuckGoEngine {
	return &DuckDuckGoEngine{
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

// Search 执行搜索
func (e *DuckDuckGoEngine) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := e.client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("请求超时（%v）", e.timeout)
		}
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %d %s", resp.StatusCode, resp.Status)
	}

	results, err := e.parseHTML(resp.Body, maxResults)
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到相关结果")
	}

	return results, nil
}

// parseHTML 解析 DuckDuckGo HTML 结果
func (e *DuckDuckGoEngine) parseHTML(body io.Reader, maxResults int) ([]SearchResult, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	e.walkNodes(doc, maxResults, &results)
	return results, nil
}

// walkNodes 遍历 HTML 节点查找搜索结果
func (e *DuckDuckGoEngine) walkNodes(n *html.Node, maxResults int, results *[]SearchResult) {
	if len(*results) >= maxResults {
		return
	}

	if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "result") {
		if result := e.parseResult(n); result != nil {
			*results = append(*results, *result)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.walkNodes(c, maxResults, results)
	}
}

// parseResult 解析单个搜索结果
func (e *DuckDuckGoEngine) parseResult(n *html.Node) *SearchResult {
	result := &SearchResult{}
	e.extractResultFields(n, result)

	if result.Title == "" || result.URL == "" {
		return nil
	}
	return result
}

// extractResultFields 从节点中提取搜索结果字段
func (e *DuckDuckGoEngine) extractResultFields(node *html.Node, result *SearchResult) {
	if node.Type == html.ElementNode && node.Data == "a" {
		if hasClass(node, "result__a") {
			result.Title = getTextContent(node)
			result.URL = getAttr(node, "href")
		} else if hasClass(node, "result__snippet") {
			result.Snippet = getTextContent(node)
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		e.extractResultFields(c, result)
	}
}

// hasClass 检查节点是否包含指定的 class
func hasClass(n *html.Node, className string) bool {
	classAttr := getAttr(n, "class")
	if classAttr == "" {
		return false
	}
	for _, c := range strings.Fields(classAttr) {
		if c == className {
			return true
		}
	}
	return false
}

// getAttr 获取节点属性值
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// getTextContent 获取节点的文本内容
func getTextContent(n *html.Node) string {
	var buf strings.Builder
	collectText(n, &buf)
	return strings.TrimSpace(buf.String())
}

// collectText 递归收集节点文本
func collectText(n *html.Node, buf *strings.Builder) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
}

// SerpAPIEngine SerpAPI 搜索引擎实现（预留）
type SerpAPIEngine struct {
	apiKey string
}

// NewSerpAPIEngine 创建 SerpAPI 搜索引擎
func NewSerpAPIEngine(apiKey string) *SerpAPIEngine {
	return &SerpAPIEngine{apiKey: apiKey}
}

// Search 执行搜索（暂未实现）
func (e *SerpAPIEngine) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	return nil, fmt.Errorf("SerpAPI 搜索引擎暂未实现")
}
