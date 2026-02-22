#!/usr/bin/env python3
"""
结果查询工具

用法:
    python query_result.py <submit_id>
    python query_result.py <submit_id> --max-attempts 60 --interval 5
    python query_result.py <submit_id> --download ./output/

参数:
    submit_id: 任务提交 ID
    --max-attempts: 最大轮询次数 (默认 60)
    --interval: 轮询间隔秒数 (默认 5)
    --download: 下载目录
"""

import sys
import argparse
from pathlib import Path
from datetime import datetime

from utils import (
    query_result,
    extract_image_urls,
    extract_image_uris,
    extract_video_url,
    download_file,
)


def main():
    parser = argparse.ArgumentParser(description="查询 Dreamina 生成结果")
    parser.add_argument("submit_id", help="任务提交 ID")
    parser.add_argument(
        "--max-attempts",
        type=int,
        default=60,
        help="最大轮询次数 (默认 60)",
    )
    parser.add_argument(
        "--interval",
        type=int,
        default=5,
        help="轮询间隔秒数 (默认 5)",
    )
    parser.add_argument(
        "--download",
        help="下载目录",
    )

    args = parser.parse_args()

    print("=" * 50)
    print("查询生成结果")
    print("=" * 50)
    print(f"submit_id: {args.submit_id}")
    print(f"最大尝试: {args.max_attempts} 次")
    print(f"轮询间隔: {args.interval} 秒")
    print()

    try:
        result = query_result(
            args.submit_id,
            max_attempts=args.max_attempts,
            interval=args.interval,
        )

        if result:
            # 检查是图片还是视频
            image_urls = extract_image_urls(result)
            video_url = extract_video_url(result)

            if image_urls:
                print("\n生成的图片:")
                for i, url in enumerate(image_urls):
                    print(f"  [{i+1}] {url[:80]}...")

                uris = extract_image_uris(result)
                print("\n图片 URI:")
                for i, uri in enumerate(uris):
                    print(f"  [{i+1}] {uri}")

                # 下载
                if args.download:
                    output_dir = Path(args.download)
                    output_dir.mkdir(parents=True, exist_ok=True)
                    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

                    print(f"\n下载到: {output_dir}")
                    for i, url in enumerate(image_urls):
                        filename = f"image_{timestamp}_{i+1}.png"
                        download_file(url, str(output_dir / filename))

            elif video_url:
                print(f"\n视频 URL: {video_url[:80]}...")

                # 下载
                if args.download:
                    output_dir = Path(args.download)
                    output_dir.mkdir(parents=True, exist_ok=True)
                    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                    filename = f"video_{timestamp}.mp4"

                    print(f"\n下载到: {output_dir / filename}")
                    download_file(video_url, str(output_dir / filename))

            else:
                print("\n未找到图片或视频资源")
                print("原始结果:")
                import json
                print(json.dumps(result, indent=2, ensure_ascii=False))

        else:
            print("查询失败或超时")
            sys.exit(1)

    except ValueError as e:
        print(f"错误: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"查询失败: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
