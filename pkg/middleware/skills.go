package middleware

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/tools"
	"gopkg.in/yaml.v3"
)

// SkillsMiddleware 管理 Agent 的技能（懒加载模式）
// Skills 是可复用的指令、上下文和资源包，用于扩展 Agent 的领域知识
// 在 system prompt 中声明轻量技能列表，模型通过 Skill 工具按需加载完整内容
type SkillsMiddleware struct {
	*BaseMiddleware
	skills       []Skill
	skillPaths   map[string]string // 技能名称 -> 技能目录路径
	toolRegistry *tools.Registry
	loadedSkills map[string]bool // 已加载完整内容的技能
}

// Skill 表示一个技能（对齐 Claude Code 格式）
type Skill struct {
	// YAML Frontmatter 字段
	Name         string   `yaml:"name"`          // 技能标识符（小写，用连字符分隔）
	Description  string   `yaml:"description"`   // 技能描述（说明做什么、何时使用）
	AllowedTools []string `yaml:"allowed-tools"` // 允许使用的工具列表（可选）

	// 解析后的内容
	Content  string // SKILL.md 的 Markdown 正文（不含 frontmatter）
	BasePath string // 技能目录路径（用于加载资源文件）
}

// NewSkillsMiddleware 创建技能中间件
func NewSkillsMiddleware(toolRegistry *tools.Registry) *SkillsMiddleware {
	m := &SkillsMiddleware{
		BaseMiddleware: NewBaseMiddleware("skills"),
		skills:         make([]Skill, 0),
		skillPaths:     make(map[string]string),
		toolRegistry:   toolRegistry,
		loadedSkills:   make(map[string]bool),
	}

	// 注册 Skill 工具
	if toolRegistry != nil {
		m.registerSkillTool()
	}

	return m
}

// registerSkillTool 注册 Skill 工具（对齐 Claude Code 的 Skill tool）
func (m *SkillsMiddleware) registerSkillTool() {
	m.toolRegistry.Register(tools.NewBaseTool(
		"Skill",
		m.buildSkillToolDescription(),
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"skill": map[string]any{
					"type":        "string",
					"description": "要激活的技能名称",
				},
			},
			"required": []string{"skill"},
		},
		func(ctx context.Context, args map[string]any) (string, error) {
			skillName, ok := args["skill"].(string)
			if !ok {
				return "", fmt.Errorf("skill must be a string")
			}

			skill := m.GetSkillByName(skillName)
			if skill == nil {
				var available strings.Builder
				fmt.Fprintf(&available, "技能 '%s' 不存在。\n\n可用技能：\n", skillName)
				for _, s := range m.skills {
					fmt.Fprintf(&available, "- %s: %s\n", s.Name, s.Description)
				}
				return available.String(), nil
			}

			// 标记技能已加载
			m.loadedSkills[skillName] = true

			// 返回技能内容和基础路径
			var result strings.Builder
			fmt.Fprintf(&result, "Skill '%s' activated.\n", skill.Name)
			fmt.Fprintf(&result, "Base path: %s\n\n", skill.BasePath)
			result.WriteString(skill.Content)

			return result.String(), nil
		},
	))
}

// buildSkillToolDescription 构建 Skill 工具的描述（包含可用技能列表）
func (m *SkillsMiddleware) buildSkillToolDescription() string {
	var desc strings.Builder
	desc.WriteString(`激活一个技能，获取该技能的详细指令和领域知识。

技能是可复用的指令包，包含：
- 领域知识和最佳实践
- 项目规范和工作流程
- 模板和示例

`)

	if len(m.skills) > 0 {
		desc.WriteString("<available_skills>\n")
		for _, skill := range m.skills {
			fmt.Fprintf(&desc, "- %s: %s\n", skill.Name, skill.Description)
		}
		desc.WriteString("</available_skills>\n")
	}

	return desc.String()
}

// LoadFromDirectory 从目录加载技能
// 支持的目录结构：
//
//	skills/
//	├── git-workflow/
//	│   ├── SKILL.md
//	│   └── templates/
//	└── code-review/
//	    └── SKILL.md
func (m *SkillsMiddleware) LoadFromDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取技能目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(dir, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")

		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			continue
		}

		skill, err := m.loadSkillFromFile(skillFile, skillDir)
		if err != nil {
			continue // 跳过无效的技能
		}

		m.skills = append(m.skills, *skill)
		m.skillPaths[skill.Name] = skillDir
	}

	// 更新工具描述
	if m.toolRegistry != nil {
		m.toolRegistry.Remove("Skill")
		m.registerSkillTool()
	}

	return nil
}

// loadSkillFromFile 从文件加载单个技能
func (m *SkillsMiddleware) loadSkillFromFile(filePath, basePath string) (*Skill, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取技能文件失败: %w", err)
	}

	return m.parseSkill(string(content), basePath)
}

// parseSkill 解析技能内容（YAML frontmatter + Markdown body）
func (m *SkillsMiddleware) parseSkill(content, basePath string) (*Skill, error) {
	// 检查是否有 YAML frontmatter
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("技能文件缺少 YAML frontmatter")
	}

	// 分离 frontmatter 和 body
	parts := strings.SplitN(content[3:], "---", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("技能文件格式错误：缺少 frontmatter 结束标记")
	}

	frontmatter := strings.TrimSpace(parts[0])
	body := strings.TrimSpace(parts[1])

	// 解析 YAML frontmatter
	var skill Skill
	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		return nil, fmt.Errorf("解析 frontmatter 失败: %w", err)
	}

	// 验证必需字段
	if skill.Name == "" {
		return nil, fmt.Errorf("技能缺少 name 字段")
	}
	if skill.Description == "" {
		return nil, fmt.Errorf("技能缺少 description 字段")
	}

	skill.Content = body
	skill.BasePath = basePath

	return &skill, nil
}

// AddSkill 添加技能
func (m *SkillsMiddleware) AddSkill(skill Skill) {
	m.skills = append(m.skills, skill)
	if skill.BasePath != "" {
		m.skillPaths[skill.Name] = skill.BasePath
	}

	// 更新工具描述
	if m.toolRegistry != nil {
		m.toolRegistry.Remove("Skill")
		m.registerSkillTool()
	}
}

// GetSkills 获取所有技能
func (m *SkillsMiddleware) GetSkills() []Skill {
	return m.skills
}

// GetSkillByName 根据名称获取技能
func (m *SkillsMiddleware) GetSkillByName(name string) *Skill {
	for i := range m.skills {
		if m.skills[i].Name == name {
			return &m.skills[i]
		}
	}
	return nil
}

// IsSkillLoaded 检查技能是否已加载完整内容
func (m *SkillsMiddleware) IsSkillLoaded(name string) bool {
	return m.loadedSkills[name]
}

// BeforeAgent 在 Agent 开始前加载默认技能
func (m *SkillsMiddleware) BeforeAgent(ctx context.Context, state *agent.State) error {
	// 按优先级加载技能目录
	// 1. 项目级技能（最高优先级）
	// 2. 用户级技能
	skillDirs := []string{
		"skills",
	}

	for _, dir := range skillDirs {
		if _, err := os.Stat(dir); err == nil {
			m.LoadFromDirectory(dir)
		}
	}

	return nil
}

// BeforeModel 在调用模型前注入轻量技能列表到系统提示
func (m *SkillsMiddleware) BeforeModel(ctx context.Context, req *llm.ModelRequest) error {
	if len(m.skills) == 0 {
		return nil
	}

	// 构建轻量技能列表（只有名称和简短描述）
	var skillsList strings.Builder
	skillsList.WriteString("\n\n## 可用技能\n\n")
	skillsList.WriteString("以下技能可通过 `Skill` 工具激活，获取详细指令和领域知识：\n\n")

	for _, skill := range m.skills {
		fmt.Fprintf(&skillsList, "- **%s**: %s\n", skill.Name, skill.Description)
	}

	skillsList.WriteString("\n当任务需要特定领域知识时，调用 `Skill` 工具激活相应技能。\n")

	req.SystemPrompt += skillsList.String()

	return nil
}
