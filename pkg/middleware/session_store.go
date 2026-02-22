package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhoucx/deepagents-go/pkg/backend"
)

// SessionStore 会话存储层，负责文件 I/O 和查询
type SessionStore struct {
	backend backend.Backend
}

// NewSessionStore 创建会话存储
func NewSessionStore(b backend.Backend) *SessionStore {
	return &SessionStore{backend: b}
}

// Save 保存会话记录到文件
func (s *SessionStore) Save(ctx context.Context, record *SessionRecord, dateDir string) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化会话记录失败: %w", err)
	}

	filePath := fmt.Sprintf("memory/sessions/%s/%s.json", dateDir, record.SessionID)

	_, err = s.backend.WriteFile(ctx, filePath, string(data))
	if err != nil {
		return fmt.Errorf("写入会话记录失败: %w", err)
	}

	return nil
}

// Load 按 ID 加载会话，搜索最近 searchDays 天的目录
// 返回会话记录和所在的 dateDir
func (s *SessionStore) Load(ctx context.Context, sessionID string, searchDays int) (*SessionRecord, string, error) {
	now := time.Now()
	for i := 0; i < searchDays; i++ {
		date := now.AddDate(0, 0, -i)
		dateDir := date.Format("2006-01-02")
		filePath := fmt.Sprintf("memory/sessions/%s/%s.json", dateDir, sessionID)

		content, err := s.backend.ReadFile(ctx, filePath, 0, 0)
		if err != nil {
			continue
		}

		var record SessionRecord
		if err := json.Unmarshal([]byte(content), &record); err != nil {
			continue
		}

		return &record, dateDir, nil
	}

	return nil, "", fmt.Errorf("未找到会话: %s", sessionID)
}

// GetSessionInfo 获取单个会话的元信息
func (s *SessionStore) GetSessionInfo(ctx context.Context, sessionID, dateDir string) (*SessionInfo, error) {
	filePath := fmt.Sprintf("memory/sessions/%s/%s.json", dateDir, sessionID)

	content, err := s.backend.ReadFile(ctx, filePath, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("读取会话文件失败: %w", err)
	}

	record, err := ParseSessionRecord(content)
	if err != nil {
		return nil, err
	}

	return record.ToSessionInfo(dateDir), nil
}

// ListSessions 列出最近 days 天的所有会话
func (s *SessionStore) ListSessions(ctx context.Context, days int) ([]*SessionInfo, error) {
	var sessions []*SessionInfo
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dateDir := date.Format("2006-01-02")
		dirPath := fmt.Sprintf("memory/sessions/%s", dateDir)

		files, err := s.backend.ListFiles(ctx, dirPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir || !strings.HasSuffix(file.Path, ".json") {
				continue
			}

			sessionID := strings.TrimSuffix(filepath.Base(file.Path), ".json")

			info, err := s.GetSessionInfo(ctx, sessionID, dateDir)
			if err != nil {
				continue
			}

			sessions = append(sessions, info)
		}
	}

	return sessions, nil
}

// FindSession 通过 ID 前缀查找会话（唯一匹配）
// 如果匹配多个会话，返回错误
func (s *SessionStore) FindSession(ctx context.Context, idPrefix string, days int) (*SessionInfo, error) {
	var matched *SessionInfo
	now := time.Now()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dateDir := date.Format("2006-01-02")
		dirPath := fmt.Sprintf("memory/sessions/%s", dateDir)

		files, err := s.backend.ListFiles(ctx, dirPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir || !strings.HasSuffix(file.Path, ".json") {
				continue
			}

			sessionID := strings.TrimSuffix(filepath.Base(file.Path), ".json")

			if !strings.HasPrefix(sessionID, idPrefix) {
				continue
			}

			if matched != nil {
				return nil, fmt.Errorf("会话 ID 前缀 '%s' 匹配多个会话，请提供更长的前缀", idPrefix)
			}

			info, err := s.GetSessionInfo(ctx, sessionID, dateDir)
			if err != nil {
				continue
			}
			matched = info
		}
	}

	if matched == nil {
		return nil, fmt.Errorf("未找到会话: %s", idPrefix)
	}

	return matched, nil
}
