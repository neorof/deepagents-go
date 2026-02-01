package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicClient 实现 Anthropic Claude 客户端
type AnthropicClient struct {
	client *anthropic.Client
	model  string
}

// NewAnthropicClient 创建 Anthropic 客户端
func NewAnthropicClient(apiKey, model, baseUrl string) *AnthropicClient {
	if model == "" {
		model = "claude-sonnet-4-5-20250929"
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseUrl != "" {
		opts = append(opts, option.WithBaseURL(baseUrl))
	}

	client := anthropic.NewClient(opts...)

	return &AnthropicClient{
		client: client,
		model:  model,
	}
}

// Generate 生成响应
func (c *AnthropicClient) Generate(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	// 转换消息格式
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
	}

	// 构建请求参数
	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		Messages:  anthropic.F(messages),
		MaxTokens: anthropic.F(int64(req.MaxTokens)),
	}

	if req.SystemPrompt != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(req.SystemPrompt),
		})
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.F(req.Temperature)
	}

	// 添加工具定义
	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolParam, 0, len(req.Tools))
		for _, tool := range req.Tools {
			tools = append(tools, anthropic.ToolParam{
				Name:        anthropic.F(tool.Name),
				Description: anthropic.F(tool.Description),
				InputSchema: anthropic.F(any(tool.InputSchema)),
			})
		}
		params.Tools = anthropic.F(tools)
	}

	// 调用 API
	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic api error: %w", err)
	}

	// 解析响应
	resp := &ModelResponse{
		StopReason: string(message.StopReason),
	}

	for _, block := range message.Content {
		switch block.Type {
		case anthropic.ContentBlockTypeText:
			resp.Content += block.Text
		case anthropic.ContentBlockTypeToolUse:
			// 解析工具调用
			inputJSON, _ := json.Marshal(block.Input)
			var input map[string]any
			json.Unmarshal(inputJSON, &input)

			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: input,
			})
		}
	}

	return resp, nil
}

// CountTokens 估算 token 数量（简化实现：4 字符 ≈ 1 token）
func (c *AnthropicClient) CountTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		// 英文：4 字符 ≈ 1 token
		// 中文：1.5 字符 ≈ 1 token
		chars := len(msg.Content)
		total += chars / 3 // 折中估算
	}
	return total
}

// StreamGenerate 生成流式响应
func (c *AnthropicClient) StreamGenerate(ctx context.Context, req *ModelRequest) (<-chan StreamEvent, error) {
	// 创建事件 channel
	eventChan := make(chan StreamEvent, 10)

	// 转换消息格式
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
	}

	// 构建请求参数
	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		Messages:  anthropic.F(messages),
		MaxTokens: anthropic.F(int64(req.MaxTokens)),
	}

	if req.SystemPrompt != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(req.SystemPrompt),
		})
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.F(req.Temperature)
	}

	// 添加工具定义
	if len(req.Tools) > 0 {
		tools := make([]anthropic.ToolParam, 0, len(req.Tools))
		for _, tool := range req.Tools {
			tools = append(tools, anthropic.ToolParam{
				Name:        anthropic.F(tool.Name),
				Description: anthropic.F(tool.Description),
				InputSchema: anthropic.F(any(tool.InputSchema)),
			})
		}
		params.Tools = anthropic.F(tools)
	}

	// 启动 goroutine 处理流式响应
	go func() {
		defer close(eventChan)

		// 发送开始事件
		eventChan <- StreamEvent{
			Type: StreamEventTypeStart,
			Done: false,
		}

		// 创建流式请求
		stream := c.client.Messages.NewStreaming(ctx, params)

		// 累积文本内容和工具调用
		var fullContent string
		var toolCalls []ToolCall
		var currentToolID string
		var currentToolName string
		var currentToolInput string

		// 处理流式事件
		for stream.Next() {
			event := stream.Current()

			// 简化处理：仅处理最基本的事件类型
			switch event.Type {
			case anthropic.MessageStreamEventTypeContentBlockDelta:
				// 尝试从 delta 中提取文本
				// 由于类型是 interface{}，我们需要使用类型断言或反射
				// 为了简化，我们先检查是否有文本字段
				deltaJSON, _ := json.Marshal(event.Delta)
				var deltaMap map[string]interface{}
				if err := json.Unmarshal(deltaJSON, &deltaMap); err == nil {
					// 检查是否是文本 delta
					if text, ok := deltaMap["text"].(string); ok {
						fullContent += text
						eventChan <- StreamEvent{
							Type:    StreamEventTypeText,
							Content: text,
							Done:    false,
						}
					}
					// 检查是否是工具调用参数 delta
					if partialJSON, ok := deltaMap["partial_json"].(string); ok {
						currentToolInput += partialJSON
					}
				}

			case anthropic.MessageStreamEventTypeContentBlockStart:
				// 内容块开始 - 可能是文本或工具调用
				if event.ContentBlock != nil {
					blockJSON, _ := json.Marshal(event.ContentBlock)
					var blockMap map[string]interface{}
					if err := json.Unmarshal(blockJSON, &blockMap); err == nil {
						if blockType, ok := blockMap["type"].(string); ok && blockType == "tool_use" {
							// 工具调用开始
							if id, ok := blockMap["id"].(string); ok {
								currentToolID = id
							}
							if name, ok := blockMap["name"].(string); ok {
								currentToolName = name
							}
							currentToolInput = ""
						}
					}
				}

			case anthropic.MessageStreamEventTypeContentBlockStop:
				// 内容块结束
				if currentToolID != "" {
					// 解析工具调用参数
					var input map[string]any
					if currentToolInput != "" {
						if err := json.Unmarshal([]byte(currentToolInput), &input); err != nil {
							input = map[string]any{"_raw": currentToolInput}
						}
					} else {
						input = make(map[string]any)
					}

					toolCall := ToolCall{
						ID:    currentToolID,
						Name:  currentToolName,
						Input: input,
					}
					toolCalls = append(toolCalls, toolCall)

					// 发送工具调用事件
					eventChan <- StreamEvent{
						Type:     StreamEventTypeToolUse,
						ToolCall: &toolCall,
						Done:     false,
					}

					// 重置当前工具调用
					currentToolID = ""
					currentToolName = ""
					currentToolInput = ""
				}

			case anthropic.MessageStreamEventTypeMessageStop:
				// 消息结束
				stopReason := ""
				if event.Message.StopReason != "" {
					stopReason = string(event.Message.StopReason)
				}
				eventChan <- StreamEvent{
					Type:       StreamEventTypeEnd,
					StopReason: stopReason,
					Done:       true,
				}
			}
		}

		// 检查错误
		if err := stream.Err(); err != nil {
			eventChan <- StreamEvent{
				Type:  StreamEventTypeError,
				Error: fmt.Errorf("anthropic stream error: %w", err),
				Done:  true,
			}
		}
	}()

	return eventChan, nil
}
