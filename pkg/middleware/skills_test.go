package middleware

import (
	"context"
	"testing"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

func TestNewSkillsMiddleware(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	if m == nil {
		t.Fatal("Expected middleware to be created")
	}

	if m.disclosureMode != "progressive" {
		t.Errorf("Expected default disclosure mode to be 'progressive', got %s", m.disclosureMode)
	}
}

func TestSkillsMiddleware_LoadFromFile(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	// 测试加载不存在的文件
	err := m.LoadFromFile("nonexistent.md")
	if err == nil {
		t.Error("Expected error when loading nonexistent file")
	}

	// 测试加载 SKILLS.md
	err = m.LoadFromFile("../../SKILLS.md")
	if err != nil {
		t.Fatalf("Failed to load SKILLS.md: %v", err)
	}

	skills := m.GetSkills()
	if len(skills) == 0 {
		t.Error("Expected skills to be loaded")
	}

	// 验证技能内容
	fileOpSkill := m.GetSkillByName("文件操作")
	if fileOpSkill == nil {
		t.Error("Expected '文件操作' skill to be loaded")
	} else {
		if fileOpSkill.Description == "" {
			t.Error("Expected skill description to be set")
		}
		if len(fileOpSkill.Tools) == 0 {
			t.Error("Expected skill tools to be extracted")
		}
	}
}

func TestSkillsMiddleware_GetSkillByName(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	m.skills = []Skill{
		{Name: "测试技能", Description: "测试描述"},
	}

	skill := m.GetSkillByName("测试技能")
	if skill == nil {
		t.Error("Expected to find skill")
	}

	skill = m.GetSkillByName("不存在的技能")
	if skill != nil {
		t.Error("Expected nil for nonexistent skill")
	}
}

func TestSkillsMiddleware_EnableDisableSkill(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	// 测试启用技能
	m.EnableSkill("测试技能")
	if !m.IsSkillEnabled("测试技能") {
		t.Error("Expected skill to be enabled")
	}

	// 测试禁用技能
	m.DisableSkill("测试技能")
	if m.IsSkillEnabled("测试技能") {
		t.Error("Expected skill to be disabled")
	}
}

func TestSkillsMiddleware_SetDisclosureMode(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	m.SetDisclosureMode("all")
	if m.GetDisclosureMode() != "all" {
		t.Errorf("Expected disclosure mode to be 'all', got %s", m.GetDisclosureMode())
	}

	m.SetDisclosureMode("manual")
	if m.GetDisclosureMode() != "manual" {
		t.Errorf("Expected disclosure mode to be 'manual', got %s", m.GetDisclosureMode())
	}
}

func TestSkillsMiddleware_BeforeModel_NoSkills(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	req := &llm.ModelRequest{
		Messages:     []llm.Message{{Role: llm.RoleUser, Content: "测试"}},
		SystemPrompt: "原始提示",
	}

	err := m.BeforeModel(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 没有技能时，系统提示不应该改变
	if req.SystemPrompt != "原始提示" {
		t.Error("Expected system prompt to remain unchanged")
	}
}

func TestSkillsMiddleware_BeforeModel_AllMode(t *testing.T) {
	m := NewSkillsMiddleware(&SkillsConfig{DisclosureMode: "all"})
	m.skills = []Skill{
		{Name: "技能1", Description: "描述1", Tools: []string{"tool1"}},
		{Name: "技能2", Description: "描述2", Tools: []string{"tool2"}},
	}

	req := &llm.ModelRequest{
		Messages:     []llm.Message{{Role: llm.RoleUser, Content: "测试"}},
		SystemPrompt: "",
	}

	err := m.BeforeModel(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 应该包含所有技能
	if req.SystemPrompt == "" {
		t.Error("Expected system prompt to be set")
	}

	if !contains(req.SystemPrompt, "技能1") || !contains(req.SystemPrompt, "技能2") {
		t.Error("Expected all skills to be included in system prompt")
	}
}

func TestSkillsMiddleware_BeforeModel_ManualMode(t *testing.T) {
	m := NewSkillsMiddleware(&SkillsConfig{DisclosureMode: "manual"})
	m.skills = []Skill{
		{Name: "技能1", Description: "描述1"},
		{Name: "技能2", Description: "描述2"},
	}

	// 只启用技能1
	m.EnableSkill("技能1")

	req := &llm.ModelRequest{
		Messages:     []llm.Message{{Role: llm.RoleUser, Content: "测试"}},
		SystemPrompt: "",
	}

	err := m.BeforeModel(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 应该只包含启用的技能
	if !contains(req.SystemPrompt, "技能1") {
		t.Error("Expected enabled skill to be included")
	}

	if contains(req.SystemPrompt, "技能2") {
		t.Error("Expected disabled skill to be excluded")
	}
}

func TestSkillsMiddleware_BeforeModel_ProgressiveMode(t *testing.T) {
	m := NewSkillsMiddleware(&SkillsConfig{DisclosureMode: "progressive"})
	m.skills = []Skill{
		{Name: "文件操作", Description: "文件相关", Keywords: []string{"文件", "file"}},
		{Name: "任务规划", Description: "任务相关", Keywords: []string{"任务", "task"}},
	}
	m.buildKeywordMap()

	// 消息中包含"文件"关键词
	req := &llm.ModelRequest{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "请帮我读取文件"},
		},
		SystemPrompt: "",
	}

	err := m.BeforeModel(context.Background(), req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 应该包含相关技能
	if !contains(req.SystemPrompt, "文件操作") {
		t.Error("Expected relevant skill to be included")
	}
}

func TestSkillsMiddleware_BeforeAgent(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	state := agent.NewState()

	err := m.BeforeAgent(context.Background(), state)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 如果 SKILLS.md 存在，应该被加载
	if len(m.skills) > 0 {
		t.Log("SKILLS.md loaded successfully")
	}
}

func TestSkillsMiddleware_ParseSkills(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	content := `# Skills

## Skill: 测试技能

**描述**: 这是一个测试技能

**使用场景**: 用于测试

**可用工具**:
- ` + "`test_tool`" + `: 测试工具
- ` + "`another_tool`" + `: 另一个工具

**最佳实践**: 遵循最佳实践

**相关技能**: 技能A、技能B
`

	skills, err := m.parseSkills(content)
	if err != nil {
		t.Fatalf("Failed to parse skills: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("Expected 1 skill, got %d", len(skills))
	}

	skill := skills[0]
	if skill.Name != "测试技能" {
		t.Errorf("Expected name '测试技能', got '%s'", skill.Name)
	}

	if skill.Description != "这是一个测试技能" {
		t.Errorf("Expected description to be set, got '%s'", skill.Description)
	}

	if len(skill.Tools) < 2 {
		t.Errorf("Expected at least 2 tools, got %d", len(skill.Tools))
	}

	if len(skill.RelatedSkills) != 2 {
		t.Errorf("Expected 2 related skills, got %d", len(skill.RelatedSkills))
	}
}

func TestSkillsMiddleware_ExtractTools(t *testing.T) {
	m := NewSkillsMiddleware(nil)

	content := "使用 `read_file` 和 `write_file` 工具，还有 `SomeClass` 类"

	tools := m.extractTools(content)

	// 应该提取出工具名（包含下划线的）
	if !containsString(tools, "read_file") {
		t.Error("Expected to extract 'read_file'")
	}

	if !containsString(tools, "write_file") {
		t.Error("Expected to extract 'write_file'")
	}
}

func TestSkillsMiddleware_GetRelevantSkills(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	m.skills = []Skill{
		{Name: "文件操作", Keywords: []string{"文件", "file"}},
		{Name: "任务规划", Keywords: []string{"任务", "task"}},
		{Name: "代码开发", Keywords: []string{"代码", "code"}},
	}
	m.buildKeywordMap()

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "请帮我读取文件"},
	}

	relevant := m.getRelevantSkills(messages)

	// 应该返回相关技能
	if len(relevant) == 0 {
		t.Error("Expected to find relevant skills")
	}

	// 应该包含文件操作技能
	found := false
	for _, skill := range relevant {
		if skill.Name == "文件操作" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find '文件操作' skill")
	}
}

func TestSkillsMiddleware_GetRelevantSkills_NoMatch(t *testing.T) {
	m := NewSkillsMiddleware(nil)
	m.skills = []Skill{
		{Name: "技能1", Keywords: []string{"keyword1"}},
		{Name: "技能2", Keywords: []string{"keyword2"}},
		{Name: "技能3", Keywords: []string{"keyword3"}},
		{Name: "技能4", Keywords: []string{"keyword4"}},
	}
	m.buildKeywordMap()

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "没有匹配的关键词"},
	}

	relevant := m.getRelevantSkills(messages)

	// 没有匹配时，应该返回前3个技能
	if len(relevant) != 3 {
		t.Errorf("Expected 3 default skills, got %d", len(relevant))
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
