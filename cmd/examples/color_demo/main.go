package main

import (
	"fmt"

	"github.com/zhoucx/deepagents-go/internal/color"
)

func main() {
	fmt.Println(color.Bold("=== CLI 界面颜色优化演示 ==="))
	fmt.Println()

	// 用户输入提示符
	fmt.Println("用户输入提示符:")
	fmt.Print(color.Cyan("❯") + " ")
	fmt.Println(color.Gray("请输入您的问题..."))
	fmt.Println()

	// 模型输出
	fmt.Println("模型输出:")
	fmt.Printf("%s 我已经为您创建了 hello.txt 文件。\n", color.Green("🤖"))
	fmt.Println()

	// 工具调用
	fmt.Println("工具调用:")
	fmt.Printf("%s %s %s\n",
		color.Cyan("🔧 Tool:"),
		color.Bold("write_file"),
		color.Gray("(path=hello.txt)"))
	fmt.Printf("   %s %s\n",
		color.Green("✓ Done:"),
		color.Gray("文件写入成功"))
	fmt.Println()

	// 工具调用（错误）
	fmt.Println("工具调用（错误）:")
	fmt.Printf("%s %s %s\n",
		color.Cyan("🔧 Tool:"),
		color.Bold("read_file"),
		color.Gray("(path=missing.txt)"))
	fmt.Printf("   %s 文件不存在\n",
		color.Red("❌ Error:"))
	fmt.Println()

	// 系统消息
	fmt.Println("系统消息:")
	fmt.Println(color.Gray("可用命令:"))
	fmt.Println(color.Gray("  help, h, ?    - 显示此帮助信息"))
	fmt.Println(color.Gray("  exit, quit, q - 退出程序"))
	fmt.Println()

	// 状态切换
	fmt.Println("状态切换:")
	fmt.Println(color.Yellow("流式响应已启用"))
	fmt.Println()

	// 错误信息
	fmt.Println("错误信息:")
	fmt.Println(color.Red("❌ 错误: 连接 API 失败"))
	fmt.Println()

	// 完整对话示例
	fmt.Println(color.Bold("=== 完整对话示例 ==="))
	fmt.Println()
	fmt.Print(color.Cyan("❯") + " ")
	fmt.Println("帮我创建一个 hello.txt 文件")
	fmt.Println()
	fmt.Printf("%s %s %s\n",
		color.Cyan("🔧 Tool:"),
		color.Bold("write_file"),
		color.Gray("(path=hello.txt)"))
	fmt.Printf("   %s %s\n",
		color.Green("✓ Done:"),
		color.Gray("文件写入成功"))
	fmt.Println()
	fmt.Printf("%s 我已经为您创建了 hello.txt 文件。\n", color.Green("🤖"))
	fmt.Println()

	// 颜色对比
	fmt.Println(color.Bold("=== 颜色对比 ==="))
	fmt.Println()
	fmt.Println("无颜色版本:")
	fmt.Println("> 帮我创建一个 hello.txt 文件")
	fmt.Println()
	fmt.Println("🔧 调用工具: write_file (path=hello.txt)")
	fmt.Println("   ✓ 完成: 文件写入成功")
	fmt.Println()
	fmt.Println("我已经为您创建了 hello.txt 文件。")
	fmt.Println()
	fmt.Println("有颜色版本:")
	fmt.Print(color.Cyan("❯") + " ")
	fmt.Println("帮我创建一个 hello.txt 文件")
	fmt.Println()
	fmt.Printf("%s %s %s\n",
		color.Cyan("🔧 Tool:"),
		color.Bold("write_file"),
		color.Gray("(path=hello.txt)"))
	fmt.Printf("   %s %s\n",
		color.Green("✓ Done:"),
		color.Gray("文件写入成功"))
	fmt.Println()
	fmt.Printf("%s 我已经为您创建了 hello.txt 文件。\n", color.Green("🤖"))
	fmt.Println()

	// 禁用颜色测试
	fmt.Println(color.Bold("=== 禁用颜色测试 ==="))
	fmt.Println()
	color.DisableColor()
	fmt.Println("颜色已禁用:")
	fmt.Print(color.Cyan("❯") + " ")
	fmt.Println("这应该没有颜色")
	fmt.Printf("%s %s\n", color.Green("✓"), "测试文本")
	color.EnableColor()
	fmt.Println()
	fmt.Println("颜色已重新启用:")
	fmt.Print(color.Cyan("❯") + " ")
	fmt.Println("这应该有颜色")
	fmt.Printf("%s %s\n", color.Green("✓"), color.Green("测试文本"))
	fmt.Println()

	fmt.Println(color.Bold("=== 演示结束 ==="))
}
