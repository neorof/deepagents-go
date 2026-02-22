# Dreamina 工作流示例

## 示例 1: 基础文生图

### 场景
生成一张可爱猫咪的图片并下载。

### 步骤

```bash
cd skills/dreamina/scripts

# 1. 生成图片
python gen_image.py "一只可爱的橘猫坐在窗台上，阳光透过窗户照进来，温馨氛围，高清" \
    --ratio 1:1 \
    --model v4.0 \
    --wait \
    -o ./output/cats/

# 输出:
# 文生图 (text2imageV2)
# 提示词: 一只可爱的橘猫...
# 发送生成请求...
# submit_id: xxx
# 查询中... (1/60)
# 状态: succeed
# 生成成功!
# 生成的图片:
#   [1] https://...
#   [2] https://...
#   [3] https://...
#   [4] https://...
# 下载到: ./output/cats/
```

---

## 示例 2: 图片编辑链路

### 场景
上传一张照片，添加创意元素，然后生成视频。

### 步骤

```bash
# 1. 上传本地图片
python upload_image.py ./my_photo.jpg
# 输出: resource_uri: tos-cn-i-tb4s082cfz/abc123.jpg

# 2. 编辑图片 - 添加帽子
python edit_image.py tos-cn-i-tb4s082cfz/abc123.jpg "添加一顶圣诞帽" --wait
# 输出: 新的 resource_uri: tos-cn-i-tb4s082cfz/def456.png

# 3. 基于编辑后的图片生成视频
python video_first_frame.py tos-cn-i-tb4s082cfz/def456.png \
    --prompt "人物微微转头，帽子上的绒球轻轻晃动" \
    --duration 5 \
    --wait \
    -o ./output/videos/
```

---

## 示例 3: 多图过渡视频

### 场景
生成季节变换的过渡视频。

### 步骤

```bash
# 1. 生成春天场景
python gen_image.py "春天的樱花树，粉色花瓣飘落，阳光明媚，清新氛围，高清" \
    --ratio 16:9 --wait
# 记录 resource_uri: spring_uri

# 2. 生成秋天场景
python gen_image.py "秋天的枫树，金黄色落叶，夕阳西下，温暖色调，高清" \
    --ratio 16:9 --wait
# 记录 resource_uri: autumn_uri

# 3. 生成过渡视频
python video_start_end.py <spring_uri> <autumn_uri> \
    --prompt "季节缓缓变换，春去秋来，树叶颜色渐变" \
    --duration 5 \
    --wait \
    -o ./output/seasons/
```

---

## 示例 4: 多帧故事视频

### 场景
使用多张图片制作一个简短的故事视频。

### 步骤

```bash
# 1. 生成场景 1 - 开始
python gen_image.py "一个男孩站在山脚下，仰望山顶，充满期待，清晨光线，电影感" \
    --ratio 16:9 --wait
# frame1_uri

# 2. 生成场景 2 - 攀登中
python gen_image.py "同一个男孩在山间小路上行走，背着背包，阳光正好，电影感" \
    --ratio 16:9 --wait
# frame2_uri

# 3. 生成场景 3 - 到达山顶
python gen_image.py "男孩站在山顶，眺望远方的云海，黄昏光线，壮观场景，电影感" \
    --ratio 16:9 --wait
# frame3_uri

# 4. 合成多帧视频
python video_multi_frame.py <frame1_uri> <frame2_uri> <frame3_uri> \
    --duration 2,3,3 \
    --prompts "仰望山顶充满期待,在山路上稳步前行,站在山顶眺望远方" \
    --wait \
    -o ./output/story/
```

---

## 示例 5: 批量生成

### 场景
批量生成多个主题的图片。

### Python 脚本

```python
#!/usr/bin/env python3
"""批量生成图片示例"""

import sys
sys.path.insert(0, 'skills/dreamina/scripts')

from gen_image import gen_image_and_wait
from utils import extract_image_urls, download_file
from pathlib import Path
from datetime import datetime

# 主题列表
themes = [
    ("可爱猫咪", "一只可爱的橘猫，毛茸茸，大眼睛，高清"),
    ("萌犬", "一只金毛犬，开心的笑容，阳光草地，高清"),
    ("小兔子", "一只白色兔子，红宝石眼睛，花园背景，高清"),
]

# 创建输出目录
timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
output_dir = Path(f"./output/batch_{timestamp}")
output_dir.mkdir(parents=True, exist_ok=True)

# 批量生成
for name, prompt in themes:
    print(f"\n生成: {name}")
    print(f"提示词: {prompt}")

    result = gen_image_and_wait(prompt, ratio="1:1", model="v4.0")

    if result:
        urls = extract_image_urls(result)
        for i, url in enumerate(urls):
            filename = f"{name}_{i+1}.png"
            download_file(url, str(output_dir / filename))
    else:
        print(f"生成失败: {name}")

print(f"\n批量生成完成，输出目录: {output_dir}")
```

---

## 示例 6: 错误处理

### 场景
处理生成过程中可能出现的错误。

### Python 脚本

```python
#!/usr/bin/env python3
"""带错误处理的生成示例"""

import sys
sys.path.insert(0, 'skills/dreamina/scripts')

from gen_image import gen_image_and_wait
from utils import extract_image_urls

def safe_generate(prompt, max_retries=3):
    """带重试的安全生成函数"""
    for attempt in range(max_retries):
        try:
            print(f"尝试 {attempt + 1}/{max_retries}...")
            result = gen_image_and_wait(prompt, ratio="1:1")

            if result:
                urls = extract_image_urls(result)
                if urls:
                    print(f"生成成功!")
                    return urls
                else:
                    print("未获取到图片 URL")
            else:
                print("生成失败，重试中...")

        except ValueError as e:
            print(f"API 错误: {e}")
            if "Cookie" in str(e):
                print("Cookie 可能已过期，请更新")
                break
        except Exception as e:
            print(f"未知错误: {e}")

    return None

# 使用
urls = safe_generate("一只可爱的猫咪")
if urls:
    print(f"获得 {len(urls)} 张图片")
else:
    print("所有尝试都失败了")
```

---

## 常用命令速查

```bash
# 文生图
python gen_image.py "提示词" --ratio 1:1 --model v4.0 --wait

# 图生图
python edit_image.py <uri> "编辑提示词" --wait

# 上传图片
python upload_image.py ./image.png

# 首帧生视频
python video_first_frame.py <uri> --prompt "运动描述" --wait

# 首尾帧生视频
python video_start_end.py <first_uri> <last_uri> --wait

# 多帧生视频
python video_multi_frame.py <uri1> <uri2> <uri3> --wait

# 查询结果
python query_result.py <submit_id> --download ./output/
```
