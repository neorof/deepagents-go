package middleware

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// HistoryMiddleware 自动记录对话历史到 JSON 文件
type HistoryMiddleware struct {
	*BaseMiddleware
	store            *SessionStore
	sessionID        string         // 会话 ID（UUID）
	sessionRecord    *SessionRecord // 当前会话记录
	dateDir          string         // 会话创建时固定的日期目录（修复跨午夜漂移）
	lastMessageCount int            // 上次记录的消息数量（用于跟踪新增消息）
	pendingToolCalls int            // 待完成的工具调用数量
	resumeMode       bool           // 是否为恢复模式
	mu               sync.Mutex     // 并发保护
}

// NewHistoryMiddleware 创建会话记录中间件
func NewHistoryMiddleware(store *SessionStore) *HistoryMiddleware {
	return &HistoryMiddleware{
		BaseMiddleware: NewBaseMiddleware("memory"),
		store:          store,
	}
}

// GetSessionStore 返回底层的 SessionStore（供 REPL 使用）
func (m *HistoryMiddleware) GetSessionStore() *SessionStore {
	return m.store
}

// BeforeAgent 在 Agent 开始执行前初始化会话记录
func (m *HistoryMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
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

	// 4. 初始化新会话，固定 dateDir
	m.dateDir = time.Now().Format("2006-01-02")
	m.sessionRecord = &SessionRecord{
		SessionID: m.sessionID,
		StartTime: time.Now().UTC().Format(time.RFC3339),
		Messages:  make([]MessageLog, 0),
	}

	return nil
}

// BeforeModel 在调用 LLM 前记录用户消息
func (m *HistoryMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 记录所有新增的 user 消息
	for i := m.lastMessageCount; i < len(req.Messages); i++ {
		msg := req.Messages[i]
		if msg.Role == "user" {
			msgLog := MessageLog{
				Role:      "user",
				Content:   msg.Content,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}

			// 记录工具结果
			for _, tr := range msg.ToolResults {
				msgLog.ToolResults = append(msgLog.ToolResults, ToolResultLog{
					ToolCallID: tr.ToolCallID,
					Content:    tr.Content,
					IsError:    tr.IsError,
				})
			}

			m.sessionRecord.Messages = append(m.sessionRecord.Messages, msgLog)
		}
	}

	// 更新消息计数
	m.lastMessageCount = len(req.Messages)

	return nil
}

// AfterModel 在 LLM 响应后记录助手响应和工具调用
func (m *HistoryMiddleware) AfterModel(ctx context.Context, resp *llm.ModelResponse, state *agent.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 创建 assistant 消息
	assistantMsg := MessageLog{
		Role:       "assistant",
		Content:    resp.Content,
		StopReason: resp.StopReason,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ToolCalls:  make([]ToolCallLog, 0),
	}

	// 2. 记录工具调用（此时工具还未执行，Output 为空）
	for _, tc := range resp.ToolCalls {
		toolCallLog := ToolCallLog{
			ID:        tc.ID,
			Name:      tc.Name,
			Input:     tc.Input,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, toolCallLog)
	}

	// 3. 添加到消息列表
	m.sessionRecord.Messages = append(m.sessionRecord.Messages, assistantMsg)

	// 4. 设置待完成的工具调用数量
	m.pendingToolCalls = len(resp.ToolCalls)

	// 5. 如果没有工具调用，立即持久化（不重置 lastMessageCount）
	if m.pendingToolCalls == 0 {
		if err := m.persistSession(ctx); err != nil {
			log.Printf("警告: 持久化会话记录失败: %v", err)
		}
	}

	return nil
}

// AfterTool 在工具执行后记录工具输出
func (m *HistoryMiddleware) AfterTool(ctx context.Context, result *llm.ToolResult, state *agent.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.sessionRecord.Messages) == 0 {
		return nil
	}

	// 1. 找到最后一条 assistant 消息（包含工具调用）
	var lastAssistantMsg *MessageLog
	for i := len(m.sessionRecord.Messages) - 1; i >= 0; i-- {
		if m.sessionRecord.Messages[i].Role == "assistant" {
			lastAssistantMsg = &m.sessionRecord.Messages[i]
			break
		}
	}

	if lastAssistantMsg == nil {
		return nil
	}

	// 2. 查找对应的 ToolCallLog 并更新
	for i := range lastAssistantMsg.ToolCalls {
		if lastAssistantMsg.ToolCalls[i].ID == result.ToolCallID {
			lastAssistantMsg.ToolCalls[i].Output = result.Content
			lastAssistantMsg.ToolCalls[i].IsError = result.IsError
			lastAssistantMsg.ToolCalls[i].Timestamp = time.Now().UTC().Format(time.RFC3339)
			break
		}
	}

	// 3. 减少待完成的工具调用数量
	m.pendingToolCalls--

	// 4. 如果所有工具都执行完毕，持久化（不重置 lastMessageCount）
	if m.pendingToolCalls <= 0 {
		if err := m.persistSession(ctx); err != nil {
			log.Printf("警告: 持久化会话记录失败: %v", err)
		}
	}

	return nil
}

// persistSession 持久化会话记录到文件（使用固定的 dateDir）
func (m *HistoryMiddleware) persistSession(ctx context.Context) error {
	m.sessionRecord.EndTime = time.Now().UTC().Format(time.RFC3339)
	return m.store.Save(ctx, m.sessionRecord, m.dateDir)
}

// loadExistingSession 尝试加载已有的会话记录
func (m *HistoryMiddleware) loadExistingSession(ctx context.Context) bool {
	record, dateDir, err := m.store.Load(ctx, m.sessionID, 30)
	if err != nil {
		return false
	}

	m.sessionRecord = record
	m.dateDir = dateDir
	m.lastMessageCount = len(record.Messages)
	return true
}

// SetResumeMode 设置为恢复模式（由 REPL 调用）
func (m *HistoryMiddleware) SetResumeMode(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionID = sessionID
	m.resumeMode = true
}

// ClearSession 清除当前会话状态（由 REPL /clear 调用）
func (m *HistoryMiddleware) ClearSession() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionRecord = nil
	m.sessionID = ""
	m.dateDir = ""
	m.lastMessageCount = 0
	m.pendingToolCalls = 0
	m.resumeMode = false
}

// LoadSessionMessages 加载会话的历史消息（用于恢复对话上下文）
func (m *HistoryMiddleware) LoadSessionMessages(ctx context.Context) ([]llm.Message, error) {
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

	messages := make([]llm.Message, 0, len(m.sessionRecord.Messages))

	// 从 SessionRecord 重建消息列表
	for _, msg := range m.sessionRecord.Messages {
		var role llm.Role
		switch msg.Role {
		case "user":
			role = llm.RoleUser
		case "assistant":
			role = llm.RoleAssistant
		default:
			role = llm.Role(msg.Role)
		}

		message := llm.Message{
			Role:    role,
			Content: msg.Content,
		}

		if msg.Role == "assistant" {
			// 恢复工具调用
			for _, tc := range msg.ToolCalls {
				message.ToolCalls = append(message.ToolCalls, llm.ToolCall{
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Input,
				})
			}
		} else if msg.Role == "user" {
			// 恢复工具结果
			for _, tr := range msg.ToolResults {
				message.ToolResults = append(message.ToolResults, llm.ToolResult{
					ToolCallID: tr.ToolCallID,
					Content:    tr.Content,
					IsError:    tr.IsError,
				})
			}
		}

		messages = append(messages, message)
	}

	return messages, nil
}
