#!/usr/bin/env python3
"""
首尾帧生视频工具 (startEnd2Video)

用法:
    python video_start_end.py <first_frame_uri> <last_frame_uri>
    python video_start_end.py tos-cn-i-xxx/first.png tos-cn-i-xxx/last.png --wait
    python video_start_end.py tos-cn-i-xxx/first.png tos-cn-i-xxx/last.png --duration 10 -o ./output/

参数:
    first_frame_uri: 首帧图片的 TOS URI
    last_frame_uri: 尾帧图片的 TOS URI
    --prompt: 过渡描述 (可选)
    --duration: 视频时长 (5 或 10 秒，默认 5)
    --wait: 等待生成完成
    -o/--output: 输出目录 (自动下载)
"""

import sys
import argparse
from pathlib import Path
from datetime import datetime

from config import ENDPOINTS, AGENT_SCENE
from utils import mcp_request, generate_and_wait, extract_video_url, download_file


def video_start_end(
    first_frame_uri: str,
    last_frame_uri: str,
    prompt: str = "smooth transition between frames",
    duration: int = 5,
) -> dict:
    """
    首尾帧生视频

    Args:
        first_frame_uri: 首帧图片的 TOS URI
        last_frame_uri: 尾帧图片的 TOS URI
        prompt: 过渡描述
        duration: 视频时长 (5 或 10 秒)

    Returns:
        API 响应
    """
    data = {
        "generate_type": "startEnd2Video",
        "agent_scene": AGENT_SCENE,
        "prompt": prompt,
        "duration": duration,
        "first_frame_resource_uri": first_frame_uri,
        "last_frame_resource_uri": last_frame_uri,
    }

    return mcp_request(ENDPOINTS["video_generate"], data)


def video_start_end_and_wait(
    first_frame_uri: str,
    last_frame_uri: str,
    prompt: str = "smooth transition between frames",
    duration: int = 5,
    max_attempts: int = 120,
    interval: int = 10,
    verbose: bool = True,
) -> dict:
    """
    首尾帧生视频并等待结果

    Args:
        first_frame_uri: 首帧图片的 TOS URI
        last_frame_uri: 尾帧图片的 TOS URI
        prompt: 过渡描述
        duration: 视频时长 (5 或 10 秒)
        max_attempts: 最大轮询次数
        interval: 轮询间隔
        verbose: 是否打印进度

    Returns:
        生成结果
    """
    data = {
        "generate_type": "startEnd2Video",
        "agent_scene": AGENT_SCENE,
        "prompt": prompt,
        "duration": duration,
        "first_frame_resource_uri": first_frame_uri,
        "last_frame_resource_uri": last_frame_uri,
    }

    return generate_and_wait(
        ENDPOINTS["video_generate"],
        data,
        max_attempts=max_attempts,
        interval=interval,
        verbose=verbose,
    )


def main():
    parser = argparse.ArgumentParser(description="首尾帧生视频 (startEnd2Video)")
    parser.add_argument("first_frame_uri", help="首帧图片的 TOS URI")
    parser.add_argument("last_frame_uri", help="尾帧图片的 TOS URI")
    parser.add_argument(
        "--prompt",
        default="smooth transition between frames",
        help="过渡描述 (默认: smooth transition between frames)",
    )
    parser.add_argument(
        "--duration",
        type=int,
        default=5,
        choices=[5, 10],
        help="视频时长 (5 或 10 秒，默认 5)",
    )
    parser.add_argument(
        "--wait",
        action="store_true",
        help="等待生成完成",
    )
    parser.add_argument(
        "-o", "--output",
        help="输出目录 (自动下载视频)",
    )

    args = parser.parse_args()

    print("=" * 50)
    print("首尾帧生视频 (startEnd2Video)")
    print("=" * 50)
    print(f"首帧图片: {args.first_frame_uri}")
    print(f"尾帧图片: {args.last_frame_uri}")
    print(f"过渡描述: {args.prompt}")
    print(f"时长: {args.duration} 秒")
    print()

    try:
        if args.wait or args.output:
            print("注意: 视频生成通常需要 1-3 分钟，请耐心等待...")
            print()

            result = video_start_end_and_wait(
                args.first_frame_uri,
                args.last_frame_uri,
                prompt=args.prompt,
                duration=args.duration,
            )

            if result:
                video_url = extract_video_url(result)

                if video_url:
                    print(f"\n视频 URL: {video_url[:80]}...")

                    # 下载视频
                    if args.output:
                        output_dir = Path(args.output)
                        output_dir.mkdir(parents=True, exist_ok=True)
                        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                        filename = f"video_start_end_{timestamp}.mp4"

                        print(f"\n下载到: {output_dir / filename}")
                        download_file(video_url, str(output_dir / filename))
                else:
                    print("未获取到视频 URL")
            else:
                print("生成失败")
                sys.exit(1)
        else:
            result = video_start_end(
                args.first_frame_uri,
                args.last_frame_uri,
                prompt=args.prompt,
                duration=args.duration,
            )
            submit_id = result.get("data", {}).get("submit_id")
            print(f"submit_id: {submit_id}")
            print(f"\n视频生成通常需要 1-3 分钟")
            print(f"使用以下命令查询结果:")
            print(f"  python query_result.py {submit_id}")

    except ValueError as e:
        print(f"错误: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"生成失败: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
