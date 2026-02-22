package middleware

import (
	"encoding/json"
	"fmt"
	"time"
)

// SessionRecord 会话记录（顶层结构）
type SessionRecord struct {
	SessionID string       `json:"session_id"` // UUID
	StartTime string       `json:"start_time"` // RFC3339 格式
	EndTime   string       `json:"end_time"`   // RFC3339 格式
	Messages  []MessageLog `json:"messages"`   // 完整的对话消息列表
}

// SessionInfo 会话元信息（用于列表展示）
type SessionInfo struct {
	SessionID    string    `json:"session_id"`    // 完整的会话 ID
	StartTime    time.Time `json:"start_time"`    // 开始时间
	EndTime      time.Time `json:"end_time"`      // 结束时间
	MessageCount int       `json:"message_count"` // 消息数量
	DateDir      string    `json:"date_dir"`      // 日期目录（YYYY-MM-DD）
}

// Duration 计算会话持续时间
func (s *SessionInfo) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return 0
	}
	return s.EndTime.Sub(s.StartTime)
}

// ShortID 返回会话 ID 的前 8 位
func (s *SessionInfo) ShortID() string {
	if len(s.SessionID) >= 8 {
		return s.SessionID[:8]
	}
	return s.SessionID
}

// MessageLog 消息记录
type MessageLog struct {
	Role        string          `json:"role"`                   // "user" 或 "assistant"
	Content     string          `json:"content"`                // 消息内容
	ToolCalls   []ToolCallLog   `json:"tool_calls,omitempty"`   // 工具调用列表（仅 assistant 消息）
	ToolResults []ToolResultLog `json:"tool_results,omitempty"` // 工具结果列表（仅 user 消息）
	StopReason  string          `json:"stop_reason,omitempty"`  // 停止原因（仅 assistant 消息）
	Timestamp   string          `json:"timestamp"`              // RFC3339 格式
}

// ToolCallLog 工具调用记录
type ToolCallLog struct {
	ID        string         `json:"id"`        // 工具调用 ID
	Name      string         `json:"name"`      // 工具名称
	Input     map[string]any `json:"input"`     // 工具参数
	Output    string         `json:"output"`    // 工具输出
	IsError   bool           `json:"is_error"`  // 是否错误
	Timestamp string         `json:"timestamp"` // RFC3339 格式
}

// ToolResultLog 工具结果记录
type ToolResultLog struct {
	ToolCallID string `json:"tool_call_id"` // 对应的工具调用 ID
	Content    string `json:"content"`      // 工具输出内容
	IsError    bool   `json:"is_error"`     // 是否错误
}

// ParseSessionRecord 从 JSON 字符串解析会话记录
func ParseSessionRecord(data string) (*SessionRecord, error) {
	var record SessionRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, fmt.Errorf("解析会话记录失败: %w", err)
	}
	return &record, nil
}

// ToSessionInfo 将 SessionRecord 转换为 SessionInfo
func (r *SessionRecord) ToSessionInfo(dateDir string) *SessionInfo {
	startTime, _ := time.Parse(time.RFC3339, r.StartTime)
	endTime, _ := time.Parse(time.RFC3339, r.EndTime)

	return &SessionInfo{
		SessionID:    r.SessionID,
		StartTime:    startTime,
		EndTime:      endTime,
		MessageCount: len(r.Messages),
		DateDir:      dateDir,
	}
}
