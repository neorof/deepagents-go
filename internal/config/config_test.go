package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Model != "claude-sonnet-4-5-20250929" {
		t.Errorf("Expected default model, got %s", cfg.Model)
	}

	if cfg.MaxIterations != 25 {
		t.Errorf("Expected MaxIterations 25, got %d", cfg.MaxIterations)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel 'info', got %s", cfg.LogLevel)
	}
}

func TestLoad_YAML(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
api_key: test-key
model: claude-3-opus-20240229
work_dir: /tmp/test
max_iterations: 10
temperature: 0.5
log_level: debug
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 加载配置
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got %s", cfg.APIKey)
	}

	if cfg.Model != "claude-3-opus-20240229" {
		t.Errorf("Expected Model 'claude-3-opus-20240229', got %s", cfg.Model)
	}

	if cfg.MaxIterations != 10 {
		t.Errorf("Expected MaxIterations 10, got %d", cfg.MaxIterations)
	}

	if cfg.Temperature != 0.5 {
		t.Errorf("Expected Temperature 0.5, got %f", cfg.Temperature)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got %s", cfg.LogLevel)
	}
}

func TestLoad_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
  "api_key": "json-key",
  "model": "claude-3-sonnet-20240229",
  "max_iterations": 15
}`

	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.APIKey != "json-key" {
		t.Errorf("Expected APIKey 'json-key', got %s", cfg.APIKey)
	}

	if cfg.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected Model 'claude-3-sonnet-20240229', got %s", cfg.Model)
	}
}

func TestLoad_FileNotExists(t *testing.T) {
	// 加载不存在的文件应该返回默认配置
	cfg, err := Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 应该返回默认配置
	if cfg.Model != "claude-sonnet-4-5-20250929" {
		t.Error("Expected default config when file not exists")
	}
}

func TestSave_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		APIKey:        "save-test-key",
		Model:         "test-model",
		MaxIterations: 20,
		Temperature:   0.8,
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// 重新加载并验证
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}

	if loaded.APIKey != cfg.APIKey {
		t.Errorf("Expected APIKey %s, got %s", cfg.APIKey, loaded.APIKey)
	}

	if loaded.Model != cfg.Model {
		t.Errorf("Expected Model %s, got %s", cfg.Model, loaded.Model)
	}
}

func TestSave_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		APIKey: "json-save-key",
		Model:  "json-model",
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}

	if loaded.APIKey != cfg.APIKey {
		t.Error("APIKey mismatch after save/load")
	}
}

func TestMerge(t *testing.T) {
	base := DefaultConfig()
	base.APIKey = "base-key"
	base.Model = "base-model"

	override := &Config{
		APIKey:        "override-key",
		MaxIterations: 50,
	}

	base.Merge(override)

	if base.APIKey != "override-key" {
		t.Errorf("Expected APIKey 'override-key', got %s", base.APIKey)
	}

	if base.MaxIterations != 50 {
		t.Errorf("Expected MaxIterations 50, got %d", base.MaxIterations)
	}

	// Model 应该保持不变（override 中为空）
	if base.Model != "base-model" {
		t.Errorf("Expected Model 'base-model', got %s", base.Model)
	}
}

func TestMerge_EmptyValues(t *testing.T) {
	base := DefaultConfig()
	base.APIKey = "original-key"

	override := &Config{
		APIKey: "", // 空值不应该覆盖
	}

	base.Merge(override)

	if base.APIKey != "original-key" {
		t.Error("Empty value should not override existing value")
	}
}

func TestLoad_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.txt")

	if err := os.WriteFile(configPath, []byte("invalid"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
api_key: test
  invalid: yaml
    syntax
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}
