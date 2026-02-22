package config

// Config 主配置结构
type Config struct {
	Web WebConfig `yaml:"web" json:"web"`
}

// WebConfig Web 工具配置
type WebConfig struct {
	SearchEngine      string `yaml:"search_engine" json:"search_engine"`           // 搜索引擎类型：duckduckgo, serpapi
	SerpAPIKey        string `yaml:"serpapi_key" json:"serpapi_key"`               // SerpAPI Key（可选）
	DefaultTimeout    int    `yaml:"default_timeout" json:"default_timeout"`       // 默认超时时间（秒）
	MaxContentLength  int    `yaml:"max_content_length" json:"max_content_length"` // 最大内容长度（字符）
	EnableReadability bool   `yaml:"enable_readability" json:"enable_readability"` // 是否启用智能内容提取
}

// DefaultWebConfig 返回默认 Web 配置
func DefaultWebConfig() WebConfig {
	return WebConfig{
		SearchEngine:      "duckduckgo",
		SerpAPIKey:        "",
		DefaultTimeout:    30,
		MaxContentLength:  100000,
		EnableReadability: true,
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Web: DefaultWebConfig(),
	}
}
