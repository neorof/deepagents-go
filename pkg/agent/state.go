package agent

import (
	"sync"

	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// State 表示 Agent 的状态
type State struct {
	mu       sync.RWMutex
	Messages []llm.Message     `json:"messages"`
	Files    map[string]string `json:"files"`
	Metadata map[string]any    `json:"metadata"`
}

// NewState 创建新的状态
func NewState() *State {
	return &State{
		Messages: []llm.Message{},
		Files:    make(map[string]string),
		Metadata: make(map[string]any),
	}
}

// AddMessage 添加消息
func (s *State) AddMessage(msg llm.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
}

// GetMessages 获取所有消息
func (s *State) GetMessages() []llm.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Messages
}

// SetFile 设置文件内容
func (s *State) SetFile(path, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Files[path] = content
}

// GetFile 获取文件内容
func (s *State) GetFile(path string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	content, ok := s.Files[path]
	return content, ok
}

// SetMetadata 设置元数据
func (s *State) SetMetadata(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Metadata[key] = value
}

// GetMetadata 获取元数据
func (s *State) GetMetadata(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.Metadata[key]
	return value, ok
}
