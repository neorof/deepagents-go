package middleware

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func TestNewSkillsMiddleware(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	if m == nil {
		t.Fatal("Expected middleware to be created")
	}
	if len(m.skills) != 0 {
		t.Error("Expected empty skills list")
	}
}

func TestNewSkillsMiddleware_WithRegistry(t *testing.T) {
	registry := tools.NewRegistry()
	m := NewSkillsMiddleware(registry)
	if m == nil {
		t.Fatal("Expected middleware to be created")
	}

	tool, ok := registry.Get("Skill")
	if !ok {
		t.Error("Expected Skill tool to be registered")
	}
	if tool.Name() != "Skill" {
		t.Errorf("Expected tool name 'Skill', got '%s'", tool.Name())
	}
}

func TestSkillsMiddleware_ParseSkill(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	content := `---
name: git-workflow
description: Git 工作流最佳实践
allowed-tools:
  - Bash
  - Read
---

# Git 工作流

## 分支命名规范
- feature/xxx - 新功能
`

	skill, err := m.parseSkill(content, "/path/to/skill")
	if err != nil {
		t.Fatalf("Failed to parse skill: %v", err)
	}
	if skill.Name != "git-workflow" {
		t.Errorf("Expected name 'git-workflow', got '%s'", skill.Name)
	}
	if skill.Description == "" {
		t.Error("Expected description to be set")
	}
	if len(skill.AllowedTools) != 2 {
		t.Errorf("Expected 2 allowed tools, got %d", len(skill.AllowedTools))
	}
	if skill.BasePath != "/path/to/skill" {
		t.Errorf("Expected base path '/path/to/skill', got '%s'", skill.BasePath)
	}
}

func TestSkillsMiddleware_ParseSkill_Errors(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	// Missing frontmatter
	_, err := m.parseSkill("# No Frontmatter", "")
	if err == nil {
		t.Error("Expected error for missing frontmatter")
	}

	// Missing name
	_, err = m.parseSkill("---\ndescription: test\n---\n# Content", "")
	if err == nil {
		t.Error("Expected error for missing name")
	}

	// Missing description
	_, err = m.parseSkill("---\nname: test\n---\n# Content", "")
	if err == nil {
		t.Error("Expected error for missing description")
	}
}

func TestSkillsMiddleware_AddAndGetSkill(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	m.AddSkill(Skill{Name: "test-skill", Description: "A test skill"})

	if len(m.skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(m.skills))
	}

	skill := m.GetSkillByName("test-skill")
	if skill == nil {
		t.Error("Expected to find skill")
	}

	skill = m.GetSkillByName("nonexistent")
	if skill != nil {
		t.Error("Expected nil for nonexistent skill")
	}
}

func TestSkillsMiddleware_IsSkillLoaded(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	if m.IsSkillLoaded("test") {
		t.Error("Expected skill to not be loaded initially")
	}
	m.loadedSkills["test"] = true
	if !m.IsSkillLoaded("test") {
		t.Error("Expected skill to be loaded")
	}
}

func TestSkillsMiddleware_BeforeModel(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	// No skills - prompt unchanged
	req := &llm.ModelRequest{SystemPrompt: "original"}
	m.BeforeModel(context.Background(), req)
	if req.SystemPrompt != "original" {
		t.Error("Expected prompt unchanged with no skills")
	}

	// With skills - prompt updated
	m.AddSkill(Skill{Name: "git-workflow", Description: "Git best practices"})
	req = &llm.ModelRequest{SystemPrompt: ""}
	m.BeforeModel(context.Background(), req)
	if req.SystemPrompt == "" {
		t.Error("Expected prompt to be set")
	}
}

func TestSkillsMiddleware_SkillTool(t *testing.T) {
	registry := tools.NewRegistry()
	m := NewSkillsMiddleware(registry)
	m.AddSkill(Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Content:     "# Test Content",
		BasePath:    "/test/path",
	})

	tool, _ := registry.Get("Skill")

	// Existing skill
	result, err := tool.Execute(context.Background(), map[string]any{"skill": "test-skill"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !contains(result, "Test Content") {
		t.Error("Expected result to contain skill content")
	}
	if !m.IsSkillLoaded("test-skill") {
		t.Error("Expected skill to be marked as loaded")
	}

	// Nonexistent skill
	result, _ = tool.Execute(context.Background(), map[string]any{"skill": "nonexistent"})
	if !contains(result, "不存在") {
		t.Error("Expected error message for nonexistent skill")
	}
}

func TestSkillsMiddleware_LoadFromDirectory(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "skills-test")
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "git-workflow")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: git-workflow
description: Git best practices
---
# Content`), 0644)

	m := NewSkillsMiddleware(nil)
	m.LoadFromDirectory(tmpDir)

	if len(m.skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(m.skills))
	}
}

func TestSkillsMiddleware_BeforeAgent(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	err := m.BeforeAgent(context.Background(), agent.NewState())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
