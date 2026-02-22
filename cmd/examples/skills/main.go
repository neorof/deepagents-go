package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zhoucx/deepagents-go/pkg/agent"
	"github.com/zhoucx/deepagents-go/pkg/backend"
	"github.com/zhoucx/deepagents-go/pkg/llm"
	"github.com/zhoucx/deepagents-go/pkg/middleware"
	"github.com/zhoucx/deepagents-go/pkg/tools"
)

func main() {
	fmt.Println("=== Deep Agents Go - Skills 懒加载示例 ===")

	// 检查 API Key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 ANTHROPIC_API_KEY 环境变量")
	}

	// 创建 LLM 客户端
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	llmClient := llm.NewAnthropicClient(apiKey, "", baseURL)

	// 创建工具注册表和后端
	toolRegistry := tools.NewRegistry()
	stateBackend := backend.NewStateBackend()

	// 创建中间件
	fsMiddleware := middleware.NewFilesystemMiddleware(stateBackend, toolRegistry)

	// 创建技能中间件（懒加载模式）
	// - 在 system prompt 中只显示轻量技能列表（名称+描述）
	// - 模型通过 use_skill 工具按需加载完整技能内容
	skillsMiddleware := middleware.NewSkillsMiddleware(toolRegistry)

	// 从目录加载技能（Claude Code 格式）
	// 目录结构：skills/<skill-name>/SKILL.md
	if err := skillsMiddleware.LoadFromDirectory("skills"); err != nil {
		log.Printf("警告: 无法加载技能目录: %v\n", err)
	}

	skills := skillsMiddleware.GetSkills()
	fmt.Printf("已加载 %d 个技能\n\n", len(skills))

	// 显示已加载的技能（轻量列表）
	fmt.Println("可用技能:")
	for i, skill := range skills {
		fmt.Printf("%d. %s - %s\n", i+1, skill.Name, skill.Description)
	}
	fmt.Println()

	// 创建 Agent
	config := &agent.Config{
		LLMClient:    llmClient,
		ToolRegistry: toolRegistry,
		Middlewares: []agent.Middleware{
			fsMiddleware,
			skillsMiddleware,
		},
		SystemPrompt: "你是一个有用的 AI 助手，可以帮助用户完成各种任务。",
	}
	executor := agent.NewRunnable(config)

	// 示例 1: 文件操作任务
	fmt.Println("=== 示例 1: 文件操作任务 ===")
	fmt.Println("任务: 创建一个测试文件")
	fmt.Println("说明: 模型会根据需要决定是否调用 use_skill 获取详细知识")

	output1, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "请创建一个文件 /test.txt，内容为 'Hello from Skills Middleware!'",
			},
		},
	})

	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		if len(output1.Messages) > 0 {
			lastMsg := output1.Messages[len(output1.Messages)-1]
			fmt.Printf("\n响应: %s\n", lastMsg.Content)
		}
		fmt.Printf("消息数: %d\n\n", len(output1.Messages))
	}

	// 示例 2: 询问技能详情（会触发 Skill 工具调用）
	fmt.Println("=== 示例 2: 询问技能详情 ===")
	fmt.Println("任务: 询问文件操作技能的详细信息")
	fmt.Println("说明: 模型应该调用 Skill 工具获取完整技能内容")

	output2, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "请告诉我 'file-operations' 技能的详细信息和最佳实践",
			},
		},
	})

	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		if len(output2.Messages) > 0 {
			lastMsg := output2.Messages[len(output2.Messages)-1]
			fmt.Printf("\n响应: %s\n", lastMsg.Content)
		}
		fmt.Printf("消息数: %d\n\n", len(output2.Messages))
	}

	// 检查技能加载状态
	fmt.Println("=== 技能加载状态 ===")
	for _, skill := range skillsMiddleware.GetSkills() {
		loaded := "未加载"
		if skillsMiddleware.IsSkillLoaded(skill.Name) {
			loaded = "已加载"
		}
		fmt.Printf("- %s: %s\n", skill.Name, loaded)
	}

	fmt.Println("\n=== 示例完成 ===")
}
