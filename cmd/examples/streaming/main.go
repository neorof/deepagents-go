package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func main() {
	fmt.Println("=== LLM 流式响应示例 ===")
	fmt.Println()

	// 获取 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	// 演示1: Anthropic 流式响应
	fmt.Println("演示1: Anthropic Claude 流式响应")
	fmt.Println("-----------------------------------")
	demoAnthropicStreaming(apiKey)

	// 演示2: OpenAI 流式响应（如果有 API Key）
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		fmt.Println("\n演示2: OpenAI GPT 流式响应")
		fmt.Println("-----------------------------------")
		demoOpenAIStreaming(openaiKey)
	}

	// 演示3: 对比非流式和流式响应
	fmt.Println("\n演示3: 对比非流式和流式响应")
	fmt.Println("-----------------------------------")
	demoComparison(apiKey)

	fmt.Println("\n=== 演示完成 ===")
}

// demoAnthropicStreaming 演示 Anthropic 流式响应
func demoAnthropicStreaming(apiKey string) {
	client := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "用一句话介绍什么是 Go 语言，然后用3个要点说明它的优势。",
			},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	}

	// 调用流式 API
	stream, err := client.StreamGenerate(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	fmt.Print("Claude: ")

	// 接收流式事件
	for event := range stream {
		switch event.Type {
		case llm.StreamEventTypeStart:
			// 开始生成
			continue

		case llm.StreamEventTypeText:
			// 接收到文本内容（增量）
			fmt.Print(event.Content)

		case llm.StreamEventTypeToolUse:
			// 工具调用
			fmt.Printf("\n[工具调用: %s]\n", event.ToolCall.Name)

		case llm.StreamEventTypeEnd:
			// 生成结束
			fmt.Printf("\n[结束，原因: %s]\n", event.StopReason)

		case llm.StreamEventTypeError:
			// 错误
			fmt.Printf("\n[错误: %v]\n", event.Error)
		}
	}
}

// demoOpenAIStreaming 演示 OpenAI 流式响应
func demoOpenAIStreaming(apiKey string) {
	client := llm.NewOpenAIClient(apiKey, "gpt-4o-mini", "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "用一句话介绍什么是 Go 语言，然后用3个要点说明它的优势。",
			},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	}

	// 调用流式 API
	stream, err := client.StreamGenerate(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	fmt.Print("GPT: ")

	// 接收流式事件
	for event := range stream {
		switch event.Type {
		case llm.StreamEventTypeStart:
			continue

		case llm.StreamEventTypeText:
			fmt.Print(event.Content)

		case llm.StreamEventTypeToolUse:
			fmt.Printf("\n[工具调用: %s]\n", event.ToolCall.Name)

		case llm.StreamEventTypeEnd:
			fmt.Printf("\n[结束，原因: %s]\n", event.StopReason)

		case llm.StreamEventTypeError:
			fmt.Printf("\n[错误: %v]\n", event.Error)
		}
	}
}

// demoComparison 对比非流式和流式响应
func demoComparison(apiKey string) {
	client := llm.NewAnthropicClient(apiKey, "claude-3-5-sonnet-20241022", "")

	ctx := context.Background()

	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "写一首关于编程的打油诗（4行）。",
			},
		},
		MaxTokens:   300,
		Temperature: 0.9,
	}

	// 非流式响应
	fmt.Println("非流式响应:")
	start := time.Now()
	resp, err := client.Generate(ctx, req)
	if err != nil {
		log.Printf("Generate error: %v", err)
	} else {
		duration := time.Since(start)
		fmt.Printf("\n%s\n", resp.Content)
		fmt.Printf("[耗时: %v, 一次性返回]\n", duration)
	}

	fmt.Println()

	// 流式响应
	fmt.Println("流式响应:")
	start = time.Now()
	stream, err := client.StreamGenerate(ctx, req)
	if err != nil {
		log.Printf("StreamGenerate error: %v", err)
		return
	}

	firstByteTime := time.Duration(0)
	totalContent := ""

	for event := range stream {
		if event.Type == llm.StreamEventTypeText {
			if firstByteTime == 0 {
				firstByteTime = time.Since(start)
			}
			fmt.Print(event.Content)
			totalContent += event.Content
		}
		if event.Type == llm.StreamEventTypeEnd {
			totalDuration := time.Since(start)
			fmt.Printf("\n[首字节耗时: %v, 总耗时: %v, 逐字返回]\n", firstByteTime, totalDuration)
		}
	}

	fmt.Println()
	fmt.Println("✓ 流式响应的优势:")
	fmt.Println("  - 首字节延迟更低，用户体验更好")
	fmt.Println("  - 逐字显示，更有交互感")
	fmt.Println("  - 可以提前终止（节省成本）")
}
