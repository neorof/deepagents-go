package agent

import (
	"context"
	"fmt"

	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

// Config Agent 配置
type Config struct {
	LLMClient     llm.Client
	ToolRegistry  *tools.Registry
	Middlewares   []Middleware
	SystemPrompt  string
	MaxIterations int
	MaxTokens     int
	Temperature   float64
	OnToolCall    func(toolName string, input map[string]any)        // 工具调用回调
	OnToolResult  func(toolName string, result string, isError bool) // 工具结果回调
}

// Executor 实现 Agent 执行器
type Executor struct {
	config      *Config
	middlewares []Middleware
}

// NewExecutor 创建 Agent 执行器
func NewExecutor(config *Config) *Executor {
	if config.MaxIterations == 0 {
		config.MaxIterations = 25
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	return &Executor{
		config:      config,
		middlewares: config.Middlewares,
	}
}

// Invoke 执行 Agent
func (e *Executor) Invoke(ctx context.Context, input *InvokeInput) (*InvokeOutput, error) {
	// 初始化状态
	state := NewState()
	state.Messages = input.Messages
	if input.Files != nil {
		state.Files = input.Files
	}
	if input.Metadata != nil {
		state.Metadata = input.Metadata
	}

	// 执行 BeforeAgent 钩子
	for _, m := range e.middlewares {
		if err := m.BeforeAgent(ctx, state); err != nil {
			return nil, fmt.Errorf("before agent hook failed: %w", err)
		}
	}

	// 主循环
	for i := 0; i < e.config.MaxIterations; i++ {
		// 构建 LLM 请求
		req := &llm.ModelRequest{
			Messages:     state.GetMessages(),
			SystemPrompt: e.config.SystemPrompt,
			MaxTokens:    e.config.MaxTokens,
			Temperature:  e.config.Temperature,
		}

		// 添加工具定义
		toolsList := e.config.ToolRegistry.List()
		req.Tools = make([]llm.ToolSchema, 0, len(toolsList))
		for _, tool := range toolsList {
			req.Tools = append(req.Tools, llm.ToolSchema{
				Name:        tool.Name(),
				Description: tool.Description(),
				InputSchema: tool.Parameters(),
			})
		}

		// 执行 BeforeModel 钩子
		for _, m := range e.middlewares {
			if err := m.BeforeModel(ctx, req); err != nil {
				return nil, fmt.Errorf("before model hook failed: %w", err)
			}
		}

		// 调用 LLM
		resp, err := e.config.LLMClient.Generate(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("llm generate failed: %w", err)
		}

		// 添加助手消息
		if resp.Content != "" {
			state.AddMessage(llm.Message{
				Role:    llm.RoleAssistant,
				Content: resp.Content,
			})
		}

		// 执行 AfterModel 钩子
		for _, m := range e.middlewares {
			if err := m.AfterModel(ctx, resp, state); err != nil {
				return nil, fmt.Errorf("after model hook failed: %w", err)
			}
		}

		// 检查是否需要执行工具
		if len(resp.ToolCalls) == 0 {
			// 没有工具调用，结束循环
			break
		}

		// 执行工具调用
		for _, toolCall := range resp.ToolCalls {
			// 通知工具调用开始
			if e.config.OnToolCall != nil {
				e.config.OnToolCall(toolCall.Name, toolCall.Input)
			}

			// 执行 BeforeTool 钩子
			for _, m := range e.middlewares {
				if err := m.BeforeTool(ctx, &toolCall, state); err != nil {
					return nil, fmt.Errorf("before tool hook failed: %w", err)
				}
			}

			// 获取工具
			tool, ok := e.config.ToolRegistry.Get(toolCall.Name)
			if !ok {
				// 工具不存在，返回错误
				result := &llm.ToolResult{
					ToolCallID: toolCall.ID,
					Content:    fmt.Sprintf("Tool not found: %s", toolCall.Name),
					IsError:    true,
				}

				// 通知工具结果
				if e.config.OnToolResult != nil {
					e.config.OnToolResult(toolCall.Name, result.Content, true)
				}

				for _, m := range e.middlewares {
					if err := m.AfterTool(ctx, result, state); err != nil {
						return nil, fmt.Errorf("after tool hook failed: %w", err)
					}
				}
				continue
			}

			// 执行工具
			output, err := tool.Execute(ctx, toolCall.Input)
			result := &llm.ToolResult{
				ToolCallID: toolCall.ID,
				Content:    output,
				IsError:    err != nil,
			}
			if err != nil {
				result.Content = fmt.Sprintf("Tool execution error: %v", err)
			}

			// 通知工具结果
			if e.config.OnToolResult != nil {
				e.config.OnToolResult(toolCall.Name, result.Content, result.IsError)
			}

			// 执行 AfterTool 钩子
			for _, m := range e.middlewares {
				if err := m.AfterTool(ctx, result, state); err != nil {
					return nil, fmt.Errorf("after tool hook failed: %w", err)
				}
			}

			// 添加工具结果到消息
			state.AddMessage(llm.Message{
				Role:    llm.RoleUser,
				Content: fmt.Sprintf("Tool %s result: %s", toolCall.Name, result.Content),
			})
		}
	}

	// 返回结果
	return &InvokeOutput{
		Messages: state.GetMessages(),
		Files:    state.Files,
		Metadata: state.Metadata,
	}, nil
}

// InvokeStream 执行 Agent（流式）
func (e *Executor) InvokeStream(ctx context.Context, input *InvokeInput) (<-chan AgentEvent, error) {
	eventChan := make(chan AgentEvent, 20)

	go func() {
		defer close(eventChan)

		// 初始化状态
		state := NewState()
		state.Messages = input.Messages
		if input.Files != nil {
			state.Files = input.Files
		}
		if input.Metadata != nil {
			state.Metadata = input.Metadata
		}

		// 发送开始事件
		eventChan <- AgentEvent{
			Type: AgentEventTypeStart,
			Done: false,
		}

		// 执行 BeforeAgent 钩子
		for _, m := range e.middlewares {
			if err := m.BeforeAgent(ctx, state); err != nil {
				eventChan <- AgentEvent{
					Type:  AgentEventTypeError,
					Error: fmt.Errorf("before agent hook failed: %w", err),
					Done:  true,
				}
				return
			}
		}

		// 主循环
		for i := 0; i < e.config.MaxIterations; i++ {
			// 构建 LLM 请求
			req := &llm.ModelRequest{
				Messages:     state.GetMessages(),
				SystemPrompt: e.config.SystemPrompt,
				MaxTokens:    e.config.MaxTokens,
				Temperature:  e.config.Temperature,
			}

			// 添加工具定义
			toolsList := e.config.ToolRegistry.List()
			req.Tools = make([]llm.ToolSchema, 0, len(toolsList))
			for _, tool := range toolsList {
				req.Tools = append(req.Tools, llm.ToolSchema{
					Name:        tool.Name(),
					Description: tool.Description(),
					InputSchema: tool.Parameters(),
				})
			}

			// 执行 BeforeModel 钩子
			for _, m := range e.middlewares {
				if err := m.BeforeModel(ctx, req); err != nil {
					eventChan <- AgentEvent{
						Type:  AgentEventTypeError,
						Error: fmt.Errorf("before model hook failed: %w", err),
						Done:  true,
					}
					return
				}
			}

			// 发送 LLM 开始事件
			eventChan <- AgentEvent{
				Type:      AgentEventTypeLLMStart,
				Iteration: i + 1,
				Done:      false,
			}

			// 调用流式 LLM
			stream, err := e.config.LLMClient.StreamGenerate(ctx, req)
			if err != nil {
				eventChan <- AgentEvent{
					Type:  AgentEventTypeError,
					Error: fmt.Errorf("llm stream generate failed: %w", err),
					Done:  true,
				}
				return
			}

			// 累积响应内容和工具调用
			var fullContent string
			var toolCalls []llm.ToolCall
			var stopReason string

			// 处理流式事件
			for streamEvent := range stream {
				switch streamEvent.Type {
				case llm.StreamEventTypeText:
					// 转发文本事件
					fullContent += streamEvent.Content
					eventChan <- AgentEvent{
						Type:      AgentEventTypeLLMText,
						Content:   streamEvent.Content,
						Iteration: i + 1,
						Done:      false,
					}

				case llm.StreamEventTypeToolUse:
					// 转发工具调用事件
					if streamEvent.ToolCall != nil {
						toolCalls = append(toolCalls, *streamEvent.ToolCall)
						eventChan <- AgentEvent{
							Type:      AgentEventTypeLLMToolCall,
							ToolCall:  streamEvent.ToolCall,
							Iteration: i + 1,
							Done:      false,
						}
					}

				case llm.StreamEventTypeEnd:
					// LLM 生成结束
					stopReason = streamEvent.StopReason
					eventChan <- AgentEvent{
						Type:      AgentEventTypeLLMEnd,
						Iteration: i + 1,
						Metadata:  map[string]any{"stop_reason": stopReason},
						Done:      false,
					}

				case llm.StreamEventTypeError:
					// 错误事件
					eventChan <- AgentEvent{
						Type:  AgentEventTypeError,
						Error: streamEvent.Error,
						Done:  true,
					}
					return
				}
			}

			// 构建响应对象（用于中间件）
			resp := &llm.ModelResponse{
				Content:    fullContent,
				ToolCalls:  toolCalls,
				StopReason: stopReason,
			}

			// 添加助手消息
			if resp.Content != "" {
				state.AddMessage(llm.Message{
					Role:    llm.RoleAssistant,
					Content: resp.Content,
				})
			}

			// 执行 AfterModel 钩子
			for _, m := range e.middlewares {
				if err := m.AfterModel(ctx, resp, state); err != nil {
					eventChan <- AgentEvent{
						Type:  AgentEventTypeError,
						Error: fmt.Errorf("after model hook failed: %w", err),
						Done:  true,
					}
					return
				}
			}

			// 检查是否需要执行工具
			if len(resp.ToolCalls) == 0 {
				// 没有工具调用，结束循环
				eventChan <- AgentEvent{
					Type:      AgentEventTypeIterationEnd,
					Iteration: i + 1,
					Done:      false,
				}
				break
			}

			// 执行工具调用
			for _, toolCall := range resp.ToolCalls {
				// 发送工具开始事件
				eventChan <- AgentEvent{
					Type:      AgentEventTypeToolStart,
					ToolCall:  &toolCall,
					Iteration: i + 1,
					Done:      false,
				}

				// 通知工具调用开始
				if e.config.OnToolCall != nil {
					e.config.OnToolCall(toolCall.Name, toolCall.Input)
				}

				// 执行 BeforeTool 钩子
				for _, m := range e.middlewares {
					if err := m.BeforeTool(ctx, &toolCall, state); err != nil {
						eventChan <- AgentEvent{
							Type:  AgentEventTypeError,
							Error: fmt.Errorf("before tool hook failed: %w", err),
							Done:  true,
						}
						return
					}
				}

				// 获取工具
				tool, ok := e.config.ToolRegistry.Get(toolCall.Name)
				var result *llm.ToolResult

				if !ok {
					// 工具不存在，返回错误
					result = &llm.ToolResult{
						ToolCallID: toolCall.ID,
						Content:    fmt.Sprintf("Tool not found: %s", toolCall.Name),
						IsError:    true,
					}
				} else {
					// 执行工具
					output, err := tool.Execute(ctx, toolCall.Input)
					result = &llm.ToolResult{
						ToolCallID: toolCall.ID,
						Content:    output,
						IsError:    err != nil,
					}
					if err != nil {
						result.Content = fmt.Sprintf("Tool execution error: %v", err)
					}
				}

				// 发送工具结果事件
				eventChan <- AgentEvent{
					Type:       AgentEventTypeToolResult,
					ToolResult: result,
					Iteration:  i + 1,
					Done:       false,
				}

				// 通知工具结果
				if e.config.OnToolResult != nil {
					e.config.OnToolResult(toolCall.Name, result.Content, result.IsError)
				}

				// 执行 AfterTool 钩子
				for _, m := range e.middlewares {
					if err := m.AfterTool(ctx, result, state); err != nil {
						eventChan <- AgentEvent{
							Type:  AgentEventTypeError,
							Error: fmt.Errorf("after tool hook failed: %w", err),
							Done:  true,
						}
						return
					}
				}

				// 添加工具结果到消息
				state.AddMessage(llm.Message{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("Tool %s result: %s", toolCall.Name, result.Content),
				})
			}

			// 迭代结束
			eventChan <- AgentEvent{
				Type:      AgentEventTypeIterationEnd,
				Iteration: i + 1,
				Done:      false,
			}
		}

		// 发送结束事件
		eventChan <- AgentEvent{
			Type: AgentEventTypeEnd,
			Metadata: map[string]any{
				"messages": state.GetMessages(),
				"files":    state.Files,
				"metadata": state.Metadata,
			},
			Done: true,
		}
	}()

	return eventChan, nil
}
