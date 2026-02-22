package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient 实现 OpenAI 客户端
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient 创建 OpenAI 客户端
func NewOpenAIClient(apiKey, model, baseURL string) *OpenAIClient {
	if model == "" {
		model = openai.GPT4o // 默认使用 GPT-4o
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	return &OpenAIClient{
		client: client,
		model:  model,
	}
}

// Generate 生成响应
func (c *OpenAIClient) Generate(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	// 记录请求信息
	logRequest("openai", c.model, req, false)

	// 转换消息格式
	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages)+1)

	// 如果有系统提示词，添加为第一条消息
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	// 添加用户和助手消息
	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleAssistant:
			// Assistant 消息：可能包含文本内容和工具调用
			assistantMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Content,
			}

			// 添加工具调用
			if len(msg.ToolCalls) > 0 {
				assistantMsg.ToolCalls = make([]openai.ToolCall, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					// 将参数转换为 JSON 字符串
					argsJSON, _ := json.Marshal(tc.Input)
					assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, openai.ToolCall{
						ID:   tc.ID,
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      tc.Name,
							Arguments: string(argsJSON),
						},
					})
				}
			}

			messages = append(messages, assistantMsg)

		case RoleUser:
			// User 消息：可能包含文本内容
			if msg.Content != "" {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: msg.Content,
				})
			}

			// 添加工具结果（作为独立的 tool 消息）
			for _, tr := range msg.ToolResults {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    tr.Content,
					ToolCallID: tr.ToolCallID,
				})
			}

		case RoleSystem:
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: msg.Content,
			})

		default:
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Content,
			})
		}
	}

	// 构建请求参数
	chatReq := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	if req.Temperature > 0 {
		chatReq.Temperature = float32(req.Temperature)
	}

	// 添加工具定义
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
		chatReq.Tools = tools
	}

	// 调用 API
	chatResp, err := c.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("openai api error: %w", err)
	}

	// 解析响应
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai")
	}

	choice := chatResp.Choices[0]
	resp := &ModelResponse{
		Content:    choice.Message.Content,
		StopReason: string(choice.FinishReason),
	}

	// 解析工具调用
	if len(choice.Message.ToolCalls) > 0 {
		resp.ToolCalls = make([]ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			if tc.Type != openai.ToolTypeFunction {
				continue
			}

			// 解析函数参数
			input := make(map[string]any)
			if tc.Function.Arguments != "" {
				// 使用 json.Unmarshal 解析参数
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
					// 解析失败，将原始字符串作为参数
					input["_raw"] = tc.Function.Arguments
				}
			}

			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
	}

	return resp, nil
}

// CountTokens 估算 token 数量（简化实现：4 字符 ≈ 1 token）
func (c *OpenAIClient) CountTokens(messages []Message) int {
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
func (c *OpenAIClient) StreamGenerate(ctx context.Context, req *ModelRequest) (<-chan StreamEvent, error) {
	// 记录请求信息
	logRequest("openai", c.model, req, true)

	// 创建事件 channel
	eventChan := make(chan StreamEvent, 10)

	// 转换消息格式
	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages)+1)

	// 如果有系统提示词，添加为第一条消息
	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	// 添加用户和助手消息
	for _, msg := range req.Messages {
		switch msg.Role {
		case RoleAssistant:
			// Assistant 消息：可能包含文本内容和工具调用
			assistantMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Content,
			}

			// 添加工具调用
			if len(msg.ToolCalls) > 0 {
				assistantMsg.ToolCalls = make([]openai.ToolCall, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					// 将参数转换为 JSON 字符串
					argsJSON, _ := json.Marshal(tc.Input)
					assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, openai.ToolCall{
						ID:   tc.ID,
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      tc.Name,
							Arguments: string(argsJSON),
						},
					})
				}
			}

			messages = append(messages, assistantMsg)

		case RoleUser:
			// User 消息：可能包含文本内容
			if msg.Content != "" {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: msg.Content,
				})
			}

			// 添加工具结果（作为独立的 tool 消息）
			for _, tr := range msg.ToolResults {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    tr.Content,
					ToolCallID: tr.ToolCallID,
				})
			}

		case RoleSystem:
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: msg.Content,
			})

		default:
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Content,
			})
		}
	}

	// 构建请求参数
	chatReq := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true, // 启用流式
	}

	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	if req.Temperature > 0 {
		chatReq.Temperature = float32(req.Temperature)
	}

	// 添加工具定义
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			})
		}
		chatReq.Tools = tools
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
		stream, err := c.client.CreateChatCompletionStream(ctx, chatReq)
		if err != nil {
			eventChan <- StreamEvent{
				Type:  StreamEventTypeError,
				Error: fmt.Errorf("openai stream error: %w", err),
				Done:  true,
			}
			return
		}
		defer stream.Close()

		// 用于累积工具调用信息
		toolCallsMap := make(map[int]*ToolCall)

		// 处理流式响应
		for {
			response, err := stream.Recv()
			if err != nil {
				// 流结束
				if err.Error() == "EOF" {
					eventChan <- StreamEvent{
						Type: StreamEventTypeEnd,
						Done: true,
					}
				} else {
					eventChan <- StreamEvent{
						Type:  StreamEventTypeError,
						Error: fmt.Errorf("openai stream error: %w", err),
						Done:  true,
					}
				}
				break
			}

			if len(response.Choices) == 0 {
				continue
			}

			choice := response.Choices[0]

			// 处理文本内容
			if choice.Delta.Content != "" {
				eventChan <- StreamEvent{
					Type:    StreamEventTypeText,
					Content: choice.Delta.Content,
					Done:    false,
				}
			}

			// 处理工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					index := 0
					if tc.Index != nil {
						index = *tc.Index
					}

					// 累积工具调用信息
					if _, exists := toolCallsMap[index]; !exists {
						toolCallsMap[index] = &ToolCall{
							ID:    tc.ID,
							Name:  tc.Function.Name,
							Input: make(map[string]any),
						}
						// 初始化 JSON 缓冲区
						toolCallsMap[index].Input["_json_buffer"] = ""
					}

					// 累积函数参数
					if tc.Function.Arguments != "" {
						jsonBuffer := toolCallsMap[index].Input["_json_buffer"].(string)
						jsonBuffer += tc.Function.Arguments
						toolCallsMap[index].Input["_json_buffer"] = jsonBuffer
					}
				}
			}

			// 检查是否结束
			if choice.FinishReason != "" {
				// 发送所有工具调用
				for _, toolCall := range toolCallsMap {
					// 解析完整的 JSON
					if jsonStr, ok := toolCall.Input["_json_buffer"].(string); ok && jsonStr != "" {
						var input map[string]any
						if err := json.Unmarshal([]byte(jsonStr), &input); err == nil {
							toolCall.Input = input
						} else {
							// 解析失败，保留原始字符串
							toolCall.Input = map[string]any{"_raw": jsonStr}
						}
					}

					eventChan <- StreamEvent{
						Type:     StreamEventTypeToolUse,
						ToolCall: toolCall,
						Done:     false,
					}
				}

				eventChan <- StreamEvent{
					Type:       StreamEventTypeEnd,
					StopReason: string(choice.FinishReason),
					Done:       true,
				}
				break
			}
		}
	}()

	return eventChan, nil
}
