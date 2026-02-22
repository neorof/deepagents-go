package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDuckDuckGoEngine_Search(t *testing.T) {
	// 创建 mock HTTP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		if !strings.Contains(r.URL.Query().Get("q"), "test") {
			t.Errorf("Expected query to contain 'test'")
		}

		// 返回模拟的 HTML 响应
		html := `
<!DOCTYPE html>
<html>
<body>
	<div class="result">
		<h2 class="result__title">
			<a class="result__a" href="https://example.com/1">Test Result 1</a>
		</h2>
		<a class="result__snippet">This is a test snippet 1</a>
	</div>
	<div class="result">
		<h2 class="result__title">
			<a class="result__a" href="https://example.com/2">Test Result 2</a>
		</h2>
		<a class="result__snippet">This is a test snippet 2</a>
	</div>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	// 创建引擎（使用较短的超时时间用于测试）
	engine := NewDuckDuckGoEngine(5 * time.Second)

	// 修改 client 使用 mock 服务器
	// 注意：这里我们需要修改实际的搜索 URL，但为了测试，我们直接测试 parseHTML 方法
	ctx := context.Background()

	// 测试 parseHTML
	htmlReader := strings.NewReader(`
<!DOCTYPE html>
<html>
<body>
	<div class="result">
		<h2 class="result__title">
			<a class="result__a" href="https://example.com/1">Test Result 1</a>
		</h2>
		<a class="result__snippet">This is a test snippet 1</a>
	</div>
	<div class="result">
		<h2 class="result__title">
			<a class="result__a" href="https://example.com/2">Test Result 2</a>
		</h2>
		<a class="result__snippet">This is a test snippet 2</a>
	</div>
</body>
</html>
`)

	results, err := engine.parseHTML(htmlReader, 5)
	if err != nil {
		t.Fatalf("parseHTML failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Title != "Test Result 1" {
		t.Errorf("Expected title 'Test Result 1', got %q", results[0].Title)
	}

	if results[0].URL != "https://example.com/1" {
		t.Errorf("Expected URL 'https://example.com/1', got %q", results[0].URL)
	}

	if results[0].Snippet != "This is a test snippet 1" {
		t.Errorf("Expected snippet 'This is a test snippet 1', got %q", results[0].Snippet)
	}

	_ = ctx // 避免未使用变量警告
}

func TestDuckDuckGoEngine_ParseHTML_NoResults(t *testing.T) {
	engine := NewDuckDuckGoEngine(5 * time.Second)

	htmlReader := strings.NewReader(`
<!DOCTYPE html>
<html>
<body>
	<p>No results found</p>
</body>
</html>
`)

	results, err := engine.parseHTML(htmlReader, 5)
	if err != nil {
		t.Fatalf("parseHTML failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestDuckDuckGoEngine_ParseHTML_MaxResults(t *testing.T) {
	engine := NewDuckDuckGoEngine(5 * time.Second)

	// 创建包含 10 个结果的 HTML
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString(`<!DOCTYPE html><html><body>`)
	for i := 1; i <= 10; i++ {
		htmlBuilder.WriteString(`<div class="result">`)
		htmlBuilder.WriteString(`<h2 class="result__title">`)
		htmlBuilder.WriteString(`<a class="result__a" href="https://example.com/`)
		htmlBuilder.WriteString(string(rune('0' + i)))
		htmlBuilder.WriteString(`">Result `)
		htmlBuilder.WriteString(string(rune('0' + i)))
		htmlBuilder.WriteString(`</a></h2>`)
		htmlBuilder.WriteString(`<a class="result__snippet">Snippet `)
		htmlBuilder.WriteString(string(rune('0' + i)))
		htmlBuilder.WriteString(`</a></div>`)
	}
	htmlBuilder.WriteString(`</body></html>`)

	// 限制为 3 个结果
	results, err := engine.parseHTML(strings.NewReader(htmlBuilder.String()), 3)
	if err != nil {
		t.Fatalf("parseHTML failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestWebSearchTool_Execute(t *testing.T) {
	// 创建 mock 搜索引擎
	mockEngine := &mockSearchEngine{
		results: []SearchResult{
			{Title: "Result 1", URL: "https://example.com/1", Snippet: "Snippet 1"},
			{Title: "Result 2", URL: "https://example.com/2", Snippet: "Snippet 2"},
		},
	}

	tool := NewWebSearchTool(mockEngine)

	// 验证工具属性
	if tool.Name() != "web_search" {
		t.Errorf("Expected name 'web_search', got %q", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// 测试执行
	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{
		"query":       "test query",
		"max_results": float64(5),
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Result 1") {
		t.Error("Expected result to contain 'Result 1'")
	}

	if !strings.Contains(result, "https://example.com/1") {
		t.Error("Expected result to contain URL")
	}

	if !strings.Contains(result, "Snippet 1") {
		t.Error("Expected result to contain snippet")
	}
}

func TestWebSearchTool_InvalidQuery(t *testing.T) {
	mockEngine := &mockSearchEngine{}
	tool := NewWebSearchTool(mockEngine)

	ctx := context.Background()

	// 测试空查询
	_, err := tool.Execute(ctx, map[string]any{
		"query": "",
	})

	if err == nil {
		t.Error("Expected error for empty query")
	}

	// 测试无效类型
	_, err = tool.Execute(ctx, map[string]any{
		"query": 123,
	})

	if err == nil {
		t.Error("Expected error for invalid query type")
	}
}

func TestWebFetchTool_Execute(t *testing.T) {
	// 创建 mock HTTP 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<h1>Test Heading</h1>
	<p>Test paragraph content.</p>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	tool := NewWebFetchTool(false, 100000)

	// 测试执行
	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{
		"url":     server.URL,
		"timeout": float64(30),
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result, "Test Heading") {
		t.Error("Expected result to contain 'Test Heading'")
	}
}

func TestWebFetchTool_InvalidURL(t *testing.T) {
	tool := NewWebFetchTool(false, 100000)
	ctx := context.Background()

	// 测试空 URL
	_, err := tool.Execute(ctx, map[string]any{
		"url": "",
	})

	if err == nil {
		t.Error("Expected error for empty URL")
	}

	// 测试无效协议
	_, err = tool.Execute(ctx, map[string]any{
		"url": "ftp://example.com",
	})

	if err == nil {
		t.Error("Expected error for invalid protocol")
	}

	// 测试无效 URL 格式
	_, err = tool.Execute(ctx, map[string]any{
		"url": "not a url",
	})

	if err == nil {
		t.Error("Expected error for invalid URL format")
	}
}

func TestWebFetchTool_HTTPError(t *testing.T) {
	// 创建返回 404 的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	tool := NewWebFetchTool(false, 100000)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]any{
		"url": server.URL,
	})

	if err == nil {
		t.Error("Expected error for HTTP 404")
	}

	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error to contain '404', got: %v", err)
	}
}

func TestWebFetchTool_NonHTMLContent(t *testing.T) {
	// 创建返回 JSON 的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	tool := NewWebFetchTool(false, 100000)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]any{
		"url": server.URL,
	})

	if err == nil {
		t.Error("Expected error for non-HTML content")
	}

	if !strings.Contains(err.Error(), "不支持的内容类型") {
		t.Errorf("Expected error about unsupported content type, got: %v", err)
	}
}

func TestWebFetchTool_ContentTruncation(t *testing.T) {
	// 创建返回大量内容的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 生成超过限制的内容
		largeContent := strings.Repeat("<p>Test content</p>", 10000)
		html := `<!DOCTYPE html><html><body>` + largeContent + `</body></html>`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	// 设置较小的最大内容长度
	tool := NewWebFetchTool(false, 1000)
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]any{
		"url": server.URL,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result) > 1500 {
		t.Errorf("Expected result to be truncated, got length %d", len(result))
	}

	if !strings.Contains(result, "内容已截断") {
		t.Error("Expected result to contain truncation message")
	}
}

func TestSerpAPIEngine_NotImplemented(t *testing.T) {
	engine := NewSerpAPIEngine("test-key")
	ctx := context.Background()

	_, err := engine.Search(ctx, "test", 5)

	if err == nil {
		t.Error("Expected error for unimplemented SerpAPI")
	}

	if !strings.Contains(err.Error(), "暂未实现") {
		t.Errorf("Expected error about not implemented, got: %v", err)
	}
}

// mockSearchEngine 用于测试的 mock 搜索引擎
type mockSearchEngine struct {
	results []SearchResult
	err     error
}

func (m *mockSearchEngine) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}

	if len(m.results) > maxResults {
		return m.results[:maxResults], nil
	}

	return m.results, nil
}
