package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// SummarizationMiddleware 管理上下文摘要
type SummarizationMiddleware struct {
	*BaseMiddleware
	maxTokens       int // 触发摘要的最大 token 数
	targetTokens    int // 摘要后的目标 token 数
	llmClient       llm.Client
	enabled         bool
	summaryCount    int // 已生成的摘要数量
	lastSummarySize int // 上次摘要的大小
}

// SummarizationConfig 摘要配置
type SummarizationConfig struct {
	MaxTokens    int        // 触发摘要的最大 token 数（默认 8000）
	TargetTokens int        // 摘要后的目标 token 数（默认 2000）
	LLMClient    llm.Client // LLM 客户端（用于生成摘要）
	Enabled      bool       // 是否启用摘要（默认 true）
}

// NewSummarizationMiddleware 创建摘要中间件
func NewSummarizationMiddleware(config *SummarizationConfig) *SummarizationMiddleware {
	if config == nil {
		config = &SummarizationConfig{
			MaxTokens:    8000,
			TargetTokens: 2000,
			Enabled:      true,
		}
	}

	// 设置默认值
	if config.MaxTokens == 0 {
		config.MaxTokens = 8000
	}
	if config.TargetTokens == 0 {
		config.TargetTokens = 2000
	}

	return &SummarizationMiddleware{
		BaseMiddleware: NewBaseMiddleware("summarization"),
		maxTokens:      config.MaxTokens,
		targetTokens:   config.TargetTokens,
		llmClient:      config.LLMClient,
		enabled:        config.Enabled,
	}
}

// SetEnabled 设置是否启用摘要
func (m *SummarizationMiddleware) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// IsEnabled 检查是否启用摘要
func (m *SummarizationMiddleware) IsEnabled() bool {
	return m.enabled
}

// GetSummaryCount 获取已生成的摘要数量
func (m *SummarizationMiddleware) GetSummaryCount() int {
	return m.summaryCount
}

// GetLastSummarySize 获取上次摘要的大小
func (m *SummarizationMiddleware) GetLastSummarySize() int {
	return m.lastSummarySize
}

// BeforeModel 在调用模型前检查是否需要摘要
func (m *SummarizationMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	if !m.enabled || m.llmClient == nil {
		return nil
	}

	// 计算当前消息的 token 数
	totalTokens := m.llmClient.CountTokens(req.Messages)

	// 如果未超过阈值，不需要摘要
	if totalTokens <= m.maxTokens {
		return nil
	}

	// 生成摘要
	summary, err := m.generateSummary(ctx, req.Messages)
	if err != nil {
		// 摘要失败不应该阻止执行，只记录错误
		return nil
	}

	// 替换消息为摘要
	m.replaceWithSummary(req, summary)
	m.summaryCount++
	m.lastSummarySize = len(summary)

	return nil
}

// generateSummary 生成消息摘要
func (m *SummarizationMiddleware) generateSummary(ctx context.Context, messages []llm.Message) (string, error) {
	// 构建摘要提示
	var content strings.Builder
	content.WriteString("请总结以下对话的关键信息，保留重要的上下文和决策：\n\n")

	// 只摘要前面的消息，保留最近的几条
	messagesToSummarize := messages
	if len(messages) > 5 {
		messagesToSummarize = messages[:len(messages)-3] // 保留最后 3 条消息
	}

	for _, msg := range messagesToSummarize {
		content.WriteString(fmt.Sprintf("%s: %s\n\n", msg.Role, msg.Content))
	}

	// 调用 LLM 生成摘要
	summaryReq := &llm.ModelRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: content.String(),
			},
		},
		SystemPrompt: "你是一个专业的对话摘要助手。请简洁地总结对话的关键信息，保留重要的上下文、决策和结论。",
		MaxTokens:    m.targetTokens,
		Temperature:  0.3, // 使用较低的温度以获得更一致的摘要
	}

	resp, err := m.llmClient.Generate(ctx, summaryReq)
	if err != nil {
		return "", fmt.Errorf("生成摘要失败: %w", err)
	}

	return resp.Content, nil
}

// replaceWithSummary 用摘要替换旧消息
func (m *SummarizationMiddleware) replaceWithSummary(req *llm.ModelRequest, summary string) {
	// 保留最后几条消息
	recentMessages := []llm.Message{}
	if len(req.Messages) > 3 {
		recentMessages = req.Messages[len(req.Messages)-3:]
	} else {
		recentMessages = req.Messages
	}

	// 创建摘要消息
	summaryMessage := llm.Message{
		Role:    llm.RoleAssistant,
		Content: fmt.Sprintf("=== 对话摘要 ===\n\n%s\n\n=== 继续对话 ===", summary),
	}

	// 替换消息列表
	req.Messages = append([]llm.Message{summaryMessage}, recentMessages...)
}

// ShouldSummarize 检查是否应该触发摘要
func (m *SummarizationMiddleware) ShouldSummarize(messages []llm.Message) bool {
	if !m.enabled || m.llmClient == nil {
		return false
	}

	totalTokens := m.llmClient.CountTokens(messages)
	return totalTokens > m.maxTokens
}

// GetTokenCount 获取消息的 token 数
func (m *SummarizationMiddleware) GetTokenCount(messages []llm.Message) int {
	if m.llmClient == nil {
		// 如果没有 LLM 客户端，使用简单估算
		total := 0
		for _, msg := range messages {
			total += len(msg.Content) / 3
		}
		return total
	}
	return m.llmClient.CountTokens(messages)
}

// SetMaxTokens 设置最大 token 数
func (m *SummarizationMiddleware) SetMaxTokens(maxTokens int) {
	m.maxTokens = maxTokens
}

// SetTargetTokens 设置目标 token 数
func (m *SummarizationMiddleware) SetTargetTokens(targetTokens int) {
	m.targetTokens = targetTokens
}

// GetMaxTokens 获取最大 token 数
func (m *SummarizationMiddleware) GetMaxTokens() int {
	return m.maxTokens
}

// GetTargetTokens 获取目标 token 数
func (m *SummarizationMiddleware) GetTargetTokens() int {
	return m.targetTokens
}
