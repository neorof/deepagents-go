package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RequestLog 记录请求的完整信息
type RequestLog struct {
	Timestamp    string       `json:"timestamp"`
	Model        string       `json:"model"`
	Provider     string       `json:"provider"` // "anthropic" 或 "openai"
	Messages     []Message    `json:"messages"`
	SystemPrompt string       `json:"system_prompt,omitempty"`
	Tools        []ToolSchema `json:"tools,omitempty"`
	MaxTokens    int          `json:"max_tokens,omitempty"`
	Temperature  float64      `json:"temperature,omitempty"`
	IsStreaming  bool         `json:"is_streaming"`
}

// logRequest 将请求信息记录到文件
func logRequest(provider, model string, req *ModelRequest, isStreaming bool) {
	// 创建日志目录
	logDir := filepath.Join(".", "logs", "llm_requests")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// 静默失败，不影响主流程
		return
	}

	// 生成日志文件名（按日期）
	date := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("requests_%s.jsonl", date))

	// 构建日志记录
	log := RequestLog{
		Timestamp:    time.Now().Format(time.RFC3339),
		Model:        model,
		Provider:     provider,
		Messages:     req.Messages,
		SystemPrompt: req.SystemPrompt,
		Tools:        req.Tools,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		IsStreaming:  isStreaming,
	}

	// 序列化为 JSON
	data, err := json.Marshal(log)
	if err != nil {
		return
	}

	// 追加到文件（JSONL 格式，每行一个 JSON 对象）
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	// 写入日志行
	f.Write(data)
	f.WriteString("\n")
}
