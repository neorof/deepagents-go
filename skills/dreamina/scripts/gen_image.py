#!/usr/bin/env python3
"""
文生图工具 (text2imageV2)

用法:
    python gen_image.py "提示词"
    python gen_image.py "一只可爱的橘猫" --ratio 1:1 --model v4.0 --wait
    python gen_image.py "一只可爱的橘猫" -o ./output/

参数:
    prompt: 图片描述文字
    --ratio: 图片比例 (1:1, 16:9, 9:16, 等)
    --model: 模型版本 (v3.0, v4.0, v4.1, v4.5)
    --wait: 等待生成完成
    -o/--output: 输出目录 (自动下载)
"""

import sys
import argparse
from pathlib import Path
from datetime import datetime

from config import ENDPOINTS, AGENT_SCENE, IMAGE_MODELS, IMAGE_RATIOS
from utils import mcp_request, generate_and_wait, extract_image_urls, extract_image_uris, download_file


def gen_image(
    prompt: str,
    ratio: str = "1:1",
    model: str = "v3.0",
    subject_id: str = "text2image_api",
) -> dict:
    """
    文生图

    Args:
        prompt: 图片描述文字
        ratio: 图片比例
        model: 模型版本
        subject_id: 任务标识

    Returns:
        API 响应
    """
    model_key = IMAGE_MODELS.get(model, IMAGE_MODELS["v3.0"])

    data = {
        "generate_type": "text2imageV2",
        "agent_scene": AGENT_SCENE,
        "prompt": prompt,
        "ratio": ratio,
        "model_key": model_key,
        "subject_id": subject_id,
    }

    return mcp_request(ENDPOINTS["image_generate"], data)


def gen_image_and_wait(
    prompt: str,
    ratio: str = "1:1",
    model: str = "v3.0",
    max_attempts: int = 60,
    interval: int = 3,
    verbose: bool = True,
) -> dict:
    """
    文生图并等待结果

    Args:
        prompt: 图片描述文字
        ratio: 图片比例
        model: 模型版本
        max_attempts: 最大轮询次数
        interval: 轮询间隔
        verbose: 是否打印进度

    Returns:
        生成结果
    """
    model_key = IMAGE_MODELS.get(model, IMAGE_MODELS["v3.0"])

    data = {
        "generate_type": "text2imageV2",
        "agent_scene": AGENT_SCENE,
        "prompt": prompt,
        "ratio": ratio,
        "model_key": model_key,
        "subject_id": "text2image_api",
    }

    return generate_and_wait(
        ENDPOINTS["image_generate"],
        data,
        max_attempts=max_attempts,
        interval=interval,
        verbose=verbose,
    )


def main():
    parser = argparse.ArgumentParser(description="文生图 (text2imageV2)")
    parser.add_argument("prompt", help="图片描述文字")
    parser.add_argument(
        "--ratio",
        default="1:1",
        choices=IMAGE_RATIOS,
        help="图片比例 (默认: 1:1)",
    )
    parser.add_argument(
        "--model",
        default="v3.0",
        choices=list(IMAGE_MODELS.keys()),
        help="模型版本 (默认: v3.0)",
    )
    parser.add_argument(
        "--wait",
        action="store_true",
        help="等待生成完成",
    )
    parser.add_argument(
        "-o", "--output",
        help="输出目录 (自动下载图片)",
    )

    args = parser.parse_args()

    print("=" * 50)
    print("文生图 (text2imageV2)")
    print("=" * 50)
    print(f"提示词: {args.prompt}")
    print(f"比例: {args.ratio}")
    print(f"模型: {args.model}")
    print()

    try:
        if args.wait or args.output:
            result = gen_image_and_wait(
                args.prompt,
                ratio=args.ratio,
                model=args.model,
            )

            if result:
                urls = extract_image_urls(result)
                uris = extract_image_uris(result)

                print("\n生成的图片:")
                for i, url in enumerate(urls):
                    print(f"  [{i+1}] {url[:80]}...")

                print("\n图片 URI (可用于后续操作):")
                for i, uri in enumerate(uris):
                    print(f"  [{i+1}] {uri}")

                # 下载图片
                if args.output:
                    output_dir = Path(args.output)
                    output_dir.mkdir(parents=True, exist_ok=True)
                    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

                    print(f"\n下载到: {output_dir}")
                    for i, url in enumerate(urls):
                        filename = f"image_{timestamp}_{i+1}.png"
                        download_file(url, str(output_dir / filename))

            else:
                print("生成失败")
                sys.exit(1)
        else:
            result = gen_image(args.prompt, ratio=args.ratio, model=args.model)
            submit_id = result.get("data", {}).get("submit_id")
            print(f"submit_id: {submit_id}")
            print(f"\n使用以下命令查询结果:")
            print(f"  python query_result.py {submit_id}")

    except ValueError as e:
        print(f"错误: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"生成失败: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
