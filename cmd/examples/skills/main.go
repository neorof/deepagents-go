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
	fmt.Println("=== Deep Agents Go - Skills Middleware 示例 ===")

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

	// 创建技能中间件（渐进式披露模式）
	skillsMiddleware := middleware.NewSkillsMiddleware(&middleware.SkillsConfig{
		DisclosureMode: "progressive", // 渐进式披露
	})

	// 加载技能文件
	if err := skillsMiddleware.LoadFromFile("../../SKILLS.md"); err != nil {
		log.Printf("警告: 无法加载 SKILLS.md: %v\n", err)
	} else {
		skills := skillsMiddleware.GetSkills()
		fmt.Printf("✅ 已加载 %d 个技能\n\n", len(skills))

		// 显示已加载的技能
		fmt.Println("已加载的技能:")
		for i, skill := range skills {
			fmt.Printf("%d. %s - %s\n", i+1, skill.Name, skill.Description)
		}
		fmt.Println()
	}

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
	executor := agent.NewExecutor(config)

	// 示例 1: 文件操作任务（应该触发"文件操作"技能）
	fmt.Println("=== 示例 1: 文件操作任务 ===")
	fmt.Println("任务: 创建一个测试文件")

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

	// 示例 2: 搜索任务（应该触发"文件搜索"技能）
	fmt.Println("=== 示例 2: 文件搜索任务 ===")
	fmt.Println("任务: 搜索文件")

	output2, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "请列出根目录下的所有文件",
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

	// 示例 3: 切换到"全部显示"模式
	fmt.Println("=== 示例 3: 全部显示模式 ===")
	skillsMiddleware.SetDisclosureMode("all")
	fmt.Println("已切换到全部显示模式，所有技能都会被注入到系统提示中")

	output3, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "你有哪些技能？",
			},
		},
	})

	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		if len(output3.Messages) > 0 {
			lastMsg := output3.Messages[len(output3.Messages)-1]
			fmt.Printf("\n响应: %s\n\n", lastMsg.Content)
		}
	}

	// 示例 4: 手动模式
	fmt.Println("=== 示例 4: 手动模式 ===")
	skillsMiddleware.SetDisclosureMode("manual")
	skillsMiddleware.EnableSkill("文件操作")
	skillsMiddleware.EnableSkill("任务规划")
	fmt.Println("已切换到手动模式，只启用了'文件操作'和'任务规划'技能")

	output4, err := executor.Invoke(context.Background(), &agent.InvokeInput{
		Messages: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: "你现在可以使用哪些技能？",
			},
		},
	})

	if err != nil {
		log.Printf("错误: %v\n", err)
	} else {
		if len(output4.Messages) > 0 {
			lastMsg := output4.Messages[len(output4.Messages)-1]
			fmt.Printf("\n响应: %s\n\n", lastMsg.Content)
		}
	}

	fmt.Println("=== 示例完成 ===")
}
