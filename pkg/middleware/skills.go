package middleware

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
)

// SkillsMiddleware 管理 Agent 的技能
type SkillsMiddleware struct {
	*BaseMiddleware
	skills          []Skill
	enabledSkills   map[string]bool     // 已启用的技能
	disclosureMode  string              // 披露模式：all, progressive, manual
	contextKeywords map[string][]string // 技能关键词映射
}

// Skill 表示一个技能
type Skill struct {
	Name          string   // 技能名称
	Description   string   // 技能描述
	UseCases      string   // 使用场景
	Tools         []string // 可用工具
	BestPractices string   // 最佳实践
	Examples      string   // 示例
	RelatedSkills []string // 相关技能
	Keywords      []string // 关键词（用于渐进式披露）
}

// SkillsConfig 技能配置
type SkillsConfig struct {
	DisclosureMode string // 披露模式：all（全部显示）, progressive（渐进式）, manual（手动）
}

// NewSkillsMiddleware 创建技能中间件
func NewSkillsMiddleware(config *SkillsConfig) *SkillsMiddleware {
	if config == nil {
		config = &SkillsConfig{
			DisclosureMode: "progressive", // 默认渐进式披露
		}
	}

	return &SkillsMiddleware{
		BaseMiddleware:  NewBaseMiddleware("skills"),
		skills:          make([]Skill, 0),
		enabledSkills:   make(map[string]bool),
		disclosureMode:  config.DisclosureMode,
		contextKeywords: make(map[string][]string),
	}
}

// LoadFromFile 从文件加载技能
func (m *SkillsMiddleware) LoadFromFile(path string) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("技能文件不存在: %s", path)
	}

	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取技能文件失败: %w", err)
	}

	// 解析技能
	skills, err := m.parseSkills(string(content))
	if err != nil {
		return fmt.Errorf("解析技能失败: %w", err)
	}

	m.skills = skills

	// 构建关键词映射
	m.buildKeywordMap()

	return nil
}

// parseSkills 解析技能内容
func (m *SkillsMiddleware) parseSkills(content string) ([]Skill, error) {
	skills := make([]Skill, 0)

	// 使用正则表达式匹配技能块
	// 匹配 "## Skill: [技能名称]" 开始的块
	skillPattern := regexp.MustCompile(`(?m)^## Skill: (.+?)$`)
	matches := skillPattern.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		return skills, nil
	}

	for i, match := range matches {
		// 提取技能名称
		nameStart := match[2]
		nameEnd := match[3]
		name := strings.TrimSpace(content[nameStart:nameEnd])

		// 提取技能内容（从当前技能到下一个技能或文件末尾）
		contentStart := match[1]
		contentEnd := len(content)
		if i < len(matches)-1 {
			contentEnd = matches[i+1][0]
		}

		skillContent := content[contentStart:contentEnd]

		// 解析技能字段
		skill := Skill{
			Name:          name,
			Description:   m.extractField(skillContent, "描述"),
			UseCases:      m.extractField(skillContent, "使用场景"),
			BestPractices: m.extractField(skillContent, "最佳实践"),
			Examples:      m.extractField(skillContent, "示例"),
			Tools:         m.extractTools(skillContent),
			RelatedSkills: m.extractRelatedSkills(skillContent),
			Keywords:      m.extractKeywords(name, skillContent),
		}

		skills = append(skills, skill)
	}

	return skills, nil
}

// extractField 提取字段内容
func (m *SkillsMiddleware) extractField(content, fieldName string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?m)\*\*%s\*\*:\s*\n?(.*?)(?:\n\n|\*\*|$)`, fieldName))
	matches := pattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractTools 提取工具列表
func (m *SkillsMiddleware) extractTools(content string) []string {
	tools := make([]string, 0)
	pattern := regexp.MustCompile("`([a-z_]+)`")
	matches := pattern.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			tool := match[1]
			// 只保留看起来像工具名的（包含下划线或全小写）
			if strings.Contains(tool, "_") || tool == strings.ToLower(tool) {
				if !seen[tool] {
					tools = append(tools, tool)
					seen[tool] = true
				}
			}
		}
	}
	return tools
}

// extractRelatedSkills 提取相关技能
func (m *SkillsMiddleware) extractRelatedSkills(content string) []string {
	pattern := regexp.MustCompile(`(?m)\*\*相关技能\*\*:\s*(.+)`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		skillsStr := matches[1]
		skills := strings.Split(skillsStr, "、")
		result := make([]string, 0, len(skills))
		for _, skill := range skills {
			result = append(result, strings.TrimSpace(skill))
		}
		return result
	}
	return []string{}
}

// extractKeywords 提取关键词
func (m *SkillsMiddleware) extractKeywords(name, content string) []string {
	keywords := make([]string, 0)

	// 从技能名称提取关键词
	keywords = append(keywords, strings.ToLower(name))

	// 从描述和使用场景提取关键词
	description := m.extractField(content, "描述")
	useCases := m.extractField(content, "使用场景")

	text := description + " " + useCases
	text = strings.ToLower(text)

	// 提取常见关键词
	commonKeywords := []string{
		"文件", "搜索", "任务", "代码", "测试", "文档",
		"读取", "写入", "编辑", "查找", "创建", "修改",
		"file", "search", "task", "code", "test", "doc",
		"read", "write", "edit", "find", "create", "modify",
	}

	for _, keyword := range commonKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, keyword)
		}
	}

	return keywords
}

// buildKeywordMap 构建关键词映射
func (m *SkillsMiddleware) buildKeywordMap() {
	m.contextKeywords = make(map[string][]string)

	for _, skill := range m.skills {
		for _, keyword := range skill.Keywords {
			m.contextKeywords[keyword] = append(m.contextKeywords[keyword], skill.Name)
		}
	}
}

// GetSkills 获取所有技能
func (m *SkillsMiddleware) GetSkills() []Skill {
	return m.skills
}

// GetSkillByName 根据名称获取技能
func (m *SkillsMiddleware) GetSkillByName(name string) *Skill {
	for _, skill := range m.skills {
		if skill.Name == name {
			return &skill
		}
	}
	return nil
}

// EnableSkill 启用技能
func (m *SkillsMiddleware) EnableSkill(name string) {
	m.enabledSkills[name] = true
}

// DisableSkill 禁用技能
func (m *SkillsMiddleware) DisableSkill(name string) {
	delete(m.enabledSkills, name)
}

// IsSkillEnabled 检查技能是否启用
func (m *SkillsMiddleware) IsSkillEnabled(name string) bool {
	return m.enabledSkills[name]
}

// SetDisclosureMode 设置披露模式
func (m *SkillsMiddleware) SetDisclosureMode(mode string) {
	m.disclosureMode = mode
}

// GetDisclosureMode 获取披露模式
func (m *SkillsMiddleware) GetDisclosureMode() string {
	return m.disclosureMode
}

// BeforeAgent 在 Agent 开始前加载默认技能
func (m *SkillsMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	// 尝试加载默认的 SKILLS.md 文件
	defaultPaths := []string{
		"SKILLS.md",
		".skills/SKILLS.md",
		"docs/SKILLS.md",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			if len(m.skills) == 0 {
				m.LoadFromFile(path)
			}
			break
		}
	}

	return nil
}

// BeforeModel 在调用模型前注入技能到系统提示
func (m *SkillsMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	if len(m.skills) == 0 {
		return nil
	}

	// 根据披露模式决定显示哪些技能
	var skillsToShow []Skill

	switch m.disclosureMode {
	case "all":
		// 显示所有技能
		skillsToShow = m.skills
	case "progressive":
		// 渐进式披露：根据上下文关键词显示相关技能
		skillsToShow = m.getRelevantSkills(req.Messages)
	case "manual":
		// 手动模式：只显示已启用的技能
		for _, skill := range m.skills {
			if m.IsSkillEnabled(skill.Name) {
				skillsToShow = append(skillsToShow, skill)
			}
		}
	default:
		// 默认显示所有技能
		skillsToShow = m.skills
	}

	if len(skillsToShow) == 0 {
		return nil
	}

	// 构建技能内容
	var skillsContent strings.Builder
	skillsContent.WriteString("\n\n=== 可用技能 ===\n")

	for _, skill := range skillsToShow {
		skillsContent.WriteString(fmt.Sprintf("\n## %s\n", skill.Name))

		if skill.Description != "" {
			skillsContent.WriteString(fmt.Sprintf("**描述**: %s\n", skill.Description))
		}

		if len(skill.Tools) > 0 {
			skillsContent.WriteString(fmt.Sprintf("**工具**: %s\n", strings.Join(skill.Tools, ", ")))
		}

		if skill.BestPractices != "" {
			skillsContent.WriteString(fmt.Sprintf("**最佳实践**: %s\n", skill.BestPractices))
		}

		skillsContent.WriteString("\n")
	}

	// 注入到系统提示
	if req.SystemPrompt == "" {
		req.SystemPrompt = skillsContent.String()
	} else {
		req.SystemPrompt += skillsContent.String()
	}

	return nil
}

// getRelevantSkills 根据上下文获取相关技能
func (m *SkillsMiddleware) getRelevantSkills(messages []llm.Message) []Skill {
	// 提取最近几条消息的内容
	recentMessages := messages
	if len(messages) > 5 {
		recentMessages = messages[len(messages)-5:]
	}

	// 构建上下文文本
	var contextText strings.Builder
	for _, msg := range recentMessages {
		contextText.WriteString(strings.ToLower(msg.Content))
		contextText.WriteString(" ")
	}
	context := contextText.String()

	// 查找匹配的技能
	relevantSkills := make(map[string]bool)
	for keyword, skillNames := range m.contextKeywords {
		if strings.Contains(context, keyword) {
			for _, skillName := range skillNames {
				relevantSkills[skillName] = true
			}
		}
	}

	// 如果没有匹配的技能，返回前3个技能作为默认
	if len(relevantSkills) == 0 {
		maxSkills := 3
		if len(m.skills) < maxSkills {
			maxSkills = len(m.skills)
		}
		return m.skills[:maxSkills]
	}

	// 返回相关技能
	result := make([]Skill, 0)
	for _, skill := range m.skills {
		if relevantSkills[skill.Name] {
			result = append(result, skill)
		}
	}

	return result
}
