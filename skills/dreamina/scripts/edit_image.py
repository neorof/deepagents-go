#!/usr/bin/env python3
"""
图生图工具 (seedEdit40)

用法:
    python edit_image.py <resource_uri> "编辑提示词"
    python edit_image.py tos-cn-i-tb4s082cfz/xxx.png "添加一顶红色帽子" --wait
    python edit_image.py tos-cn-i-tb4s082cfz/xxx.png "将背景改为夜景" --ratio 16:9 -o ./output/

参数:
    resource_uri: 原图片的 TOS URI (通过 upload_image.py 获取)
    prompt: 编辑提示词
    --ratio: 输出图片比例
    --wait: 等待生成完成
    -o/--output: 输出目录 (自动下载)
"""

import sys
import argparse
from pathlib import Path
from datetime import datetime

from config import ENDPOINTS, AGENT_SCENE, IMAGE_RATIOS
from utils import mcp_request, generate_and_wait, extract_image_urls, extract_image_uris, download_file


def edit_image(
    resource_uri: str,
    prompt: str,
    ratio: str = "1:1",
    subject_id: str = "seed_edit_api",
) -> dict:
    """
    图生图 (基于参考图编辑)

    Args:
        resource_uri: 原图片的 TOS URI
        prompt: 编辑提示词
        ratio: 输出图片比例
        subject_id: 任务标识

    Returns:
        API 响应
    """
    data = {
        "generate_type": "seedEdit40",
        "agent_scene": AGENT_SCENE,
        "prompt": prompt,
        "ratio": ratio,
        "resource_uri_list": [resource_uri],
        "subject_id": subject_id,
    }

    return mcp_request(ENDPOINTS["image_generate"], data)


def edit_image_and_wait(
    resource_uri: str,
    prompt: str,
    ratio: str = "1:1",
    max_attempts: int = 60,
    interval: int = 3,
    verbose: bool = True,
) -> dict:
    """
    图生图并等待结果

    Args:
        resource_uri: 原图片的 TOS URI
        prompt: 编辑提示词
        ratio: 输出图片比例
        max_attempts: 最大轮询次数
        interval: 轮询间隔
        verbose: 是否打印进度

    Returns:
        生成结果
    """
    data = {
        "generate_type": "seedEdit40",
        "agent_scene": AGENT_SCENE,
        "prompt": prompt,
        "ratio": ratio,
        "resource_uri_list": [resource_uri],
        "subject_id": "seed_edit_api",
    }

    return generate_and_wait(
        ENDPOINTS["image_generate"],
        data,
        max_attempts=max_attempts,
        interval=interval,
        verbose=verbose,
    )


def main():
    parser = argparse.ArgumentParser(description="图生图 (seedEdit40)")
    parser.add_argument("resource_uri", help="原图片的 TOS URI")
    parser.add_argument("prompt", help="编辑提示词")
    parser.add_argument(
        "--ratio",
        default="1:1",
        choices=IMAGE_RATIOS,
        help="输出图片比例 (默认: 1:1)",
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
    print("图生图 (seedEdit40)")
    print("=" * 50)
    print(f"原图: {args.resource_uri}")
    print(f"提示词: {args.prompt}")
    print(f"比例: {args.ratio}")
    print()

    try:
        if args.wait or args.output:
            result = edit_image_and_wait(
                args.resource_uri,
                args.prompt,
                ratio=args.ratio,
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
                        filename = f"edit_{timestamp}_{i+1}.png"
                        download_file(url, str(output_dir / filename))

            else:
                print("生成失败")
                sys.exit(1)
        else:
            result = edit_image(
                args.resource_uri,
                args.prompt,
                ratio=args.ratio,
            )
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
