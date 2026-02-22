package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 定义配置结构
type Config struct {
	// API 配置
	APIKey  string `yaml:"api_key" json:"api_key"`
	BaseUrl string `yaml:"base_url" json:"base_url"`
	Model   string `yaml:"model" json:"model"`

	// 工作目录配置
	WorkDir string `yaml:"work_dir" json:"work_dir"`

	// 系统提示词配置
	SystemPromptFile string `yaml:"system_prompt_file" json:"system_prompt_file"` // 系统提示词文件路径

	// Agent 配置
	MaxIterations int     `yaml:"max_iterations" json:"max_iterations"`
	MaxTokens     int     `yaml:"max_tokens" json:"max_tokens"`
	Temperature   float64 `yaml:"temperature" json:"temperature"`

	// 日志配置
	LogLevel  string `yaml:"log_level" json:"log_level"`   // debug, info, warn, error
	LogFile   string `yaml:"log_file" json:"log_file"`     // 日志文件路径，空表示输出到标准输出
	LogFormat string `yaml:"log_format" json:"log_format"` // text, json

	// 流式响应配置
	EnableStreaming bool `yaml:"enable_streaming" json:"enable_streaming"` // 启用流式响应
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Model:            "claude-sonnet-4-5-20250929",
		WorkDir:          "./",
		SystemPromptFile: "system_prompt.txt",
		MaxIterations:    25,
		MaxTokens:        4096,
		Temperature:      0.8,
		LogLevel:         "info",
		LogFormat:        "text",
	}
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	// 如果路径为空，尝试默认路径
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, ".deepagents", "config.yaml")
		}
	}

	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// 读取文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	cfg := DefaultConfig()
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析 YAML 配置失败: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析 JSON 配置失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的配置文件格式: %s", ext)
	}

	// 环境变量覆盖配置文件（优先级最高）
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" && cfg.APIKey == "" {
		cfg.APIKey = key
	}
	if url := os.Getenv("ANTHROPIC_BASE_URL"); url != "" && cfg.BaseUrl == "" {
		cfg.BaseUrl = url
	}

	return cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 序列化配置
	var data []byte
	var err error
	ext := filepath.Ext(path)

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
		if err != nil {
			return fmt.Errorf("序列化 YAML 配置失败: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化 JSON 配置失败: %w", err)
		}
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", ext)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Merge 合并配置（优先使用 other 中的非零值）
func (c *Config) Merge(other *Config) {
	if other.APIKey != "" {
		c.APIKey = other.APIKey
	}
	if other.Model != "" {
		c.Model = other.Model
	}
	if other.WorkDir != "" {
		c.WorkDir = other.WorkDir
	}
	if other.SystemPromptFile != "" {
		c.SystemPromptFile = other.SystemPromptFile
	}
	if other.MaxIterations > 0 {
		c.MaxIterations = other.MaxIterations
	}
	if other.MaxTokens > 0 {
		c.MaxTokens = other.MaxTokens
	}
	if other.Temperature > 0 {
		c.Temperature = other.Temperature
	}
	if other.LogLevel != "" {
		c.LogLevel = other.LogLevel
	}
	if other.LogFile != "" {
		c.LogFile = other.LogFile
	}
	if other.LogFormat != "" {
		c.LogFormat = other.LogFormat
	}
}

// LoadSystemPrompt 加载系统提示词
func (c *Config) LoadSystemPrompt() (string, error) {
	if c.SystemPromptFile == "" {
		return "", fmt.Errorf("system_prompt_file not configured")
	}

	// 检查文件是否存在
	if _, err := os.Stat(c.SystemPromptFile); os.IsNotExist(err) {
		return "", fmt.Errorf("system prompt file not found: %s", c.SystemPromptFile)
	}

	// 读取文件内容
	content, err := os.ReadFile(c.SystemPromptFile)
	if err != nil {
		return "", fmt.Errorf("failed to read system prompt file: %w", err)
	}

	return string(content), nil
}
