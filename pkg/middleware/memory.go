package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// MemoryMiddleware 自动记录对话历史到 JSON 文件
type MemoryMiddleware struct {
	*BaseMiddleware
	backend          backend.Backend
	sessionID        string         // 会话 ID（UUID）
	sessionRecord    *SessionRecord // 当前会话记录
	currentIteration int            // 当前迭代编号
	iterationLog     *IterationLog  // 当前迭代日志
	resumeMode       bool           // 是否为恢复模式
	mu               sync.Mutex     // 并发保护
}

// NewMemoryMiddleware 创建记忆中间件
func NewMemoryMiddleware(b backend.Backend) *MemoryMiddleware {
	return &MemoryMiddleware{
		BaseMiddleware: NewBaseMiddleware("memory"),
		backend:        b,
	}
}

// BeforeAgent 在 Agent 开始执行前初始化会话记录
func (m *MemoryMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 从 state.Metadata 获取 session_id（由 CLI 传入）
	if sessionID, ok := state.GetMetadata("session_id"); ok {
		if sid, ok := sessionID.(string); ok {
			m.sessionID = sid
		}
	}

	// 如果没有 session_id，生成新的 UUID
	if m.sessionID == "" {
		// 简单的 UUID 生成（实际应使用 github.com/google/uuid）
		m.sessionID = fmt.Sprintf("%d", time.Now().UnixNano())
		state.SetMetadata("session_id", m.sessionID)
	}

	// 2. 如果是恢复模式，尝试加载已有会话
	if m.resumeMode {
		m.resumeMode = false // 无论成功与否，只执行一次
		if m.loadExistingSession(ctx) {
			return nil
		}
	}

	// 3. 如果已有会话记录（非首次调用），跳过初始化
	if m.sessionRecord != nil {
		return nil
	}

	// 4. 初始化新会话
	m.sessionRecord = &SessionRecord{
		SessionID:  m.sessionID,
		StartTime:  time.Now().UTC().Format(time.RFC3339),
		Iterations: make([]IterationLog, 0),
	}

	return nil
}

// BeforeModel 在调用 LLM 前记录用户消息
func (m *MemoryMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 从 req.Messages 提取最后一条用户消息
	var userMsg *MessageLog
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMsg = &MessageLog{
				Role:      "user",
				Content:   req.Messages[i].Content,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			break
		}
	}

	// 2. 创建新的 IterationLog
	m.currentIteration++
	m.iterationLog = &IterationLog{
		Iteration:   m.currentIteration,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		UserMessage: userMsg,
	}

	return nil
}

// AfterModel 在 LLM 响应后记录助手响应和工具调用
func (m *MemoryMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.iterationLog == nil {
		return nil
	}

	// 1. 记录助手响应
	assistantLog := &AssistantLog{
		Content:    resp.Content,
		StopReason: resp.StopReason,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ToolCalls:  make([]ToolCallLog, 0),
	}

	// 2. 记录工具调用请求（此时工具还未执行）
	for _, tc := range resp.ToolCalls {
		toolCallLog := ToolCallLog{
			ID:        tc.ID,
			Name:      tc.Name,
			Input:     tc.Input,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			// Output 和 IsError 将在 AfterTool 中填充
		}
		assistantLog.ToolCalls = append(assistantLog.ToolCalls, toolCallLog)
	}

	m.iterationLog.Assistant = assistantLog

	// 3. 如果没有工具调用，立即持久化
	if len(resp.ToolCalls) == 0 {
		m.sessionRecord.Iterations = append(m.sessionRecord.Iterations, *m.iterationLog)
		m.persistSession(ctx)
		m.iterationLog = nil
	}

	return nil
}

// AfterTool 在工具执行后记录工具输出
func (m *MemoryMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.iterationLog == nil || m.iterationLog.Assistant == nil {
		return nil
	}

	// 1. 查找对应的 ToolCallLog 并更新
	for i := range m.iterationLog.Assistant.ToolCalls {
		if m.iterationLog.Assistant.ToolCalls[i].ID == result.ToolCallID {
			m.iterationLog.Assistant.ToolCalls[i].Output = result.Content
			m.iterationLog.Assistant.ToolCalls[i].IsError = result.IsError
			m.iterationLog.Assistant.ToolCalls[i].Timestamp = time.Now().UTC().Format(time.RFC3339)
			break
		}
	}

	// 2. 检查是否所有工具都已执行完毕
	allToolsCompleted := true
	for _, tc := range m.iterationLog.Assistant.ToolCalls {
		if tc.Output == "" {
			allToolsCompleted = false
			break
		}
	}

	// 3. 如果所有工具都执行完毕，持久化迭代记录
	if allToolsCompleted {
		m.sessionRecord.Iterations = append(m.sessionRecord.Iterations, *m.iterationLog)
		m.persistSession(ctx)
		m.iterationLog = nil
	}

	return nil
}

// persistSession 持久化会话记录到文件
func (m *MemoryMiddleware) persistSession(ctx context.Context) error {
	// 1. 更新 EndTime
	m.sessionRecord.EndTime = time.Now().UTC().Format(time.RFC3339)

	// 2. 序列化为 JSON
	data, err := json.MarshalIndent(m.sessionRecord, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化会话记录失败: %w", err)
	}

	// 3. 生成文件路径：memory/sessions/YYYY-MM-DD/{session_id}.json
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	filePath := fmt.Sprintf("/memory/sessions/%s/%s.json", dateDir, m.sessionID)

	// 4. 写入文件
	_, err = m.backend.WriteFile(ctx, filePath, string(data))
	if err != nil {
		// 持久化失败不应导致 Agent 崩溃，记录错误但继续执行
		return fmt.Errorf("写入会话记录失败: %w", err)
	}

	return nil
}

// loadExistingSession 尝试加载已有的会话记录
// 返回 true 表示成功加载，false 表示未找到或加载失败
func (m *MemoryMiddleware) loadExistingSession(ctx context.Context) bool {
	// 尝试从最近 30 天的目录中查找会话文件
	now := time.Now()
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i)
		dateDir := date.Format("2006-01-02")
		filePath := fmt.Sprintf("/memory/sessions/%s/%s.json", dateDir, m.sessionID)

		content, err := m.backend.ReadFile(ctx, filePath, 0, 0)
		if err != nil {
			continue // 文件不存在，尝试下一个日期
		}

		var record SessionRecord
		if err := json.Unmarshal([]byte(content), &record); err != nil {
			continue // 解析失败，尝试下一个日期
		}

		// 成功加载
		m.sessionRecord = &record
		m.currentIteration = len(record.Iterations)
		return true
	}

	return false // 未找到会话
}

// SetResumeMode 设置为恢复模式（由 REPL 调用）
func (m *MemoryMiddleware) SetResumeMode(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionID = sessionID
	m.resumeMode = true
}

// GetSessionInfo 获取会话元信息
func (m *MemoryMiddleware) GetSessionInfo(ctx context.Context, sessionID, dateDir string) (*SessionInfo, error) {
	filePath := fmt.Sprintf("/memory/sessions/%s/%s.json", dateDir, sessionID)

	content, err := m.backend.ReadFile(ctx, filePath, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("读取会话文件失败: %w", err)
	}

	record, err := ParseSessionRecord(content)
	if err != nil {
		return nil, err
	}

	return record.ToSessionInfo(dateDir), nil
}

// LoadSessionMessages 加载会话的历史消息（用于恢复对话上下文）
func (m *MemoryMiddleware) LoadSessionMessages(ctx context.Context) ([]llm.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果 sessionRecord 尚未加载，主动从文件加载
	if m.sessionRecord == nil {
		if m.sessionID == "" {
			return nil, fmt.Errorf("会话 ID 未设置")
		}
		if !m.loadExistingSession(ctx) {
			return nil, fmt.Errorf("未找到会话: %s", m.sessionID)
		}
	}

	messages := make([]llm.Message, 0)

	// 从 SessionRecord 重建消息列表
	for _, iter := range m.sessionRecord.Iterations {
		// 添加用户消息
		if iter.UserMessage != nil {
			messages = append(messages, llm.Message{
				Role:    llm.RoleUser,
				Content: iter.UserMessage.Content,
			})
		}

		// 添加助手响应
		if iter.Assistant != nil {
			messages = append(messages, llm.Message{
				Role:    llm.RoleAssistant,
				Content: iter.Assistant.Content,
			})
		}
	}

	return messages, nil
}
