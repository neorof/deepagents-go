package middleware

import "github.com/zhoucx/deepagents-go/pkg/llm"

// appendToLastMessage 将内容追加到 req.Messages 最后一条消息的 Content 末尾。
// 如果没有消息，fallback 到 SystemPrompt。
func appendToLastMessage(req *llm.ModelRequest, content string) {
	if len(req.Messages) > 0 {
		req.Messages[len(req.Messages)-1].Content += content
	} else {
		req.SystemPrompt += content
	}
}
