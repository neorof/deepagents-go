---
name: dreamina
description: 即梦 AI 创作技能 - 支持文生图、图生图、视频生成等 AI 创作功能。基于 Dreamina MCP API 实现。
version: 1.0.0
author: DeepAgent
license: MIT
activation:
  when:
    - "用户需要生成图片"
    - "用户需要编辑图片"
    - "用户需要生成视频"
    - "用户需要上传图片到即梦"
    - "用户需要使用 Dreamina/即梦 API"
  keywords:
    - "即梦"
    - "dreamina"
    - "文生图"
    - "图生图"
    - "图生视频"
    - "AI 生成"
    - "图片生成"
    - "视频生成"
---

# 即梦 AI 创作技能

本技能提供完整的即梦 (Dreamina) AI 创作能力，包括图片生成、图片编辑和视频生成。

## 快速开始

### 1. 配置认证

设置环境变量或配置文件：

```bash
# 方式 1: 环境变量
export DREAMINA_COOKIE="sessionid=xxx;uid_tt=xxx"

# 方式 2: 配置文件 ~/.dreamina.json
{
  "cookie": "sessionid=xxx;uid_tt=xxx"
}
```

Cookie 获取方式：
1. 登录 https://jimeng.jianying.com
2. 打开浏览器开发者工具 (F12)
3. 在 Application > Cookies 中找到 `sessionid`

### 2. 脚本调用方式

**重要**：为确保 `-o` 输出路径相对于当前工作目录正确解析，请使用以下调用方式：

```bash
# 设置脚本目录环境变量（技能激活时会自动提示此路径）
export DREAMINA_SCRIPTS=/home/zhoucx/go/deepagents-go/skills/dreamina/scripts

# 调用脚本（推荐方式）
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "提示词" -o ./output/
```

这样 `-o ./output/` 会正确解析为当前工作目录下的 `output` 文件夹。

### 3. 基本使用

```bash
# 文生图
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "一只可爱的橘猫" --ratio 1:1 --wait

# 图生图
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/edit_image.py tos-cn-i-xxx/image.png "添加一顶红色帽子" --wait

# 首帧生视频
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_first_frame.py tos-cn-i-xxx/image.png --prompt "猫咪缓缓走动" --wait
```

## 可用工具

### 图片功能

| 工具 | 脚本 | 功能 |
|------|------|------|
| 上传图片 | `upload_image.py` | 上传本地图片，获取 resource_uri |
| 文生图 | `gen_image.py` | 从文字描述生成图片 (text2imageV2) |
| 图生图 | `edit_image.py` | 基于参考图编辑生成 (seedEdit40) |

### 视频功能

| 工具 | 脚本 | 功能 |
|------|------|------|
| 首帧生视频 | `video_first_frame.py` | 图片作为首帧生成视频 (image2videoV2) |
| 首尾帧生视频 | `video_start_end.py` | 两张图片生成过渡视频 (startEnd2Video) |
| 多帧生视频 | `video_multi_frame.py` | 多张关键帧生成视频 (multiFrame2video) |

### 辅助工具

| 工具 | 脚本 | 功能 |
|------|------|------|
| 结果查询 | `query_result.py` | 查询生成任务状态和结果 |

## 工具详细说明

### 文生图 (gen_image.py)

```bash
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "提示词" [选项]

选项:
  --ratio    图片比例: 1:1, 16:9, 9:16, 3:2, 2:3, 4:3, 3:4, 21:9
  --model    模型版本: v3.0, v4.0, v4.1, v4.5
  --wait     等待生成完成
  -o         输出目录 (自动下载到当前工作目录的相对路径)

示例:
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "一只可爱的橘猫坐在沙发上" --ratio 16:9 --model v4.0 --wait
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "赛博朋克风格的城市夜景" -o ./output/
```

### 图生图 (edit_image.py)

```bash
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/edit_image.py <resource_uri> "编辑提示词" [选项]

选项:
  --ratio    输出图片比例
  --wait     等待生成完成
  -o         输出目录

示例:
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/edit_image.py tos-cn-i-tb4s082cfz/xxx.png "将背景改为夜景" --wait
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/edit_image.py tos-cn-i-tb4s082cfz/xxx.png "添加一顶红色帽子" --ratio 1:1 -o ./output/
```

### 首帧生视频 (video_first_frame.py)

```bash
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_first_frame.py <resource_uri> [选项]

选项:
  --prompt     视频运动描述
  --duration   视频时长: 5 或 10 秒
  --wait       等待生成完成
  -o           输出目录

示例:
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_first_frame.py tos-cn-i-xxx/image.png --prompt "镜头缓缓推进" --duration 5 --wait
```

### 首尾帧生视频 (video_start_end.py)

```bash
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_start_end.py <first_frame_uri> <last_frame_uri> [选项]

选项:
  --prompt     过渡描述
  --duration   视频时长: 5 或 10 秒
  --wait       等待生成完成
  -o           输出目录

示例:
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_start_end.py tos-cn-i-xxx/first.png tos-cn-i-xxx/last.png --wait
```

### 多帧生视频 (video_multi_frame.py)

```bash
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_multi_frame.py <frame1> <frame2> [frame3...] [选项]

选项:
  --duration   每帧时长 (逗号分隔，如 "2,2,2")
  --prompts    每帧描述 (逗号分隔)
  --wait       等待生成完成
  -o           输出目录

示例:
  PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_multi_frame.py f1.png f2.png f3.png --duration 2,3,2 --wait
```

## 典型工作流

### 工作流 1: 文生图 + 图生视频

```bash
# 1. 生成图片
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "一只可爱的柴犬" --wait
# 获得 resource_uri: tos-cn-i-tb4s082cfz/xxx.png

# 2. 图片生成视频
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_first_frame.py tos-cn-i-tb4s082cfz/xxx.png --prompt "狗狗摇着尾巴" --wait
```

### 工作流 2: 上传图片 + 编辑 + 生成视频

```bash
# 1. 上传本地图片
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/upload_image.py ./my_photo.png
# 获得 resource_uri: tos-cn-i-tb4s082cfz/uploaded.png

# 2. 编辑图片
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/edit_image.py tos-cn-i-tb4s082cfz/uploaded.png "添加晚霞背景" --wait
# 获得新的 resource_uri

# 3. 生成视频
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_first_frame.py <new_uri> --prompt "夕阳缓缓下沉" --wait
```

### 工作流 3: 多图片制作过渡视频

```bash
# 1. 生成起始图片
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "春天的樱花树" --wait
# 获得 uri_spring

# 2. 生成结束图片
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/gen_image.py "秋天的枫叶树" --wait
# 获得 uri_autumn

# 3. 生成过渡视频
PYTHONPATH=$DREAMINA_SCRIPTS python $DREAMINA_SCRIPTS/video_start_end.py <uri_spring> <uri_autumn> --prompt "季节变换过渡" --wait
```

## 技能资源

### 参考文档
- `references/api-reference.md` - API 完整参考
- `references/prompt-guide.md` - 提示词编写指南
- `references/models.md` - 模型说明

### 示例
- `assets/examples/workflow_examples.md` - 工作流示例

## 注意事项

1. **Cookie 有效期**: Cookie 会过期，需要定期更新
2. **生成时间**: 图片约 10-30 秒，视频约 1-5 分钟
3. **URL 过期**: 生成的 URL 带签名，约 1-2 小时过期，请及时下载
4. **网络环境**: 需要字节内网环境访问 PPE 泳道
5. **每次生成**: 文生图每次返回 4 张变体图片
6. **输出路径**: 使用 `-o` 参数时，必须通过 `PYTHONPATH` 方式调用脚本（见"脚本调用方式"），否则相对路径会相对于脚本目录而非当前工作目录

## 常见问题

### Q: 如何获取 Cookie?
登录 jimeng.jianying.com，F12 打开开发者工具，在 Application > Cookies 中找到 sessionid。

### Q: 生成失败怎么办?
1. 检查 Cookie 是否过期
2. 检查网络是否在内网环境
3. 查看错误信息，可能是参数问题

### Q: 视频生成很慢?
视频生成需要 1-5 分钟，这是正常的。使用 `--wait` 参数会自动轮询等待。

### Q: 如何批量生成?
可以编写脚本循环调用，或使用 Python 直接导入模块：

```python
from gen_image import gen_image_and_wait

prompts = ["猫", "狗", "鸟"]
for prompt in prompts:
    result = gen_image_and_wait(prompt)
    print(result)
```
