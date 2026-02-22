#!/usr/bin/env python3
"""
首帧生视频工具 (image2videoV2)

用法:
    python video_first_frame.py <resource_uri>
    python video_first_frame.py tos-cn-i-tb4s082cfz/xxx.png --prompt "猫咪缓缓走动" --wait
    python video_first_frame.py tos-cn-i-tb4s082cfz/xxx.png --duration 10 -o ./output/

参数:
    resource_uri: 首帧图片的 TOS URI
    --prompt: 视频运动描述 (可选)
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


def video_first_frame(
    resource_uri: str,
    prompt: str = "",
    duration: int = 5,
) -> dict:
    """
    首帧生视频

    Args:
        resource_uri: 首帧图片的 TOS URI
        prompt: 视频运动描述 (可选)
        duration: 视频时长 (5 或 10 秒)

    Returns:
        API 响应
    """
    data = {
        "generate_type": "image2videoV2",
        "agent_scene": AGENT_SCENE,
        "first_frame_resource_uri": resource_uri,
        "duration": duration,
    }

    if prompt:
        data["prompt"] = prompt

    return mcp_request(ENDPOINTS["video_generate"], data)


def video_first_frame_and_wait(
    resource_uri: str,
    prompt: str = "",
    duration: int = 5,
    max_attempts: int = 120,
    interval: int = 10,
    verbose: bool = True,
) -> dict:
    """
    首帧生视频并等待结果

    Args:
        resource_uri: 首帧图片的 TOS URI
        prompt: 视频运动描述 (可选)
        duration: 视频时长 (5 或 10 秒)
        max_attempts: 最大轮询次数
        interval: 轮询间隔
        verbose: 是否打印进度

    Returns:
        生成结果
    """
    data = {
        "generate_type": "image2videoV2",
        "agent_scene": AGENT_SCENE,
        "first_frame_resource_uri": resource_uri,
        "duration": duration,
    }

    if prompt:
        data["prompt"] = prompt

    return generate_and_wait(
        ENDPOINTS["video_generate"],
        data,
        max_attempts=max_attempts,
        interval=interval,
        verbose=verbose,
    )


def main():
    parser = argparse.ArgumentParser(description="首帧生视频 (image2videoV2)")
    parser.add_argument("resource_uri", help="首帧图片的 TOS URI")
    parser.add_argument(
        "--prompt",
        default="",
        help="视频运动描述 (可选)",
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
    print("首帧生视频 (image2videoV2)")
    print("=" * 50)
    print(f"首帧图片: {args.resource_uri}")
    if args.prompt:
        print(f"提示词: {args.prompt}")
    print(f"时长: {args.duration} 秒")
    print()

    try:
        if args.wait or args.output:
            print("注意: 视频生成通常需要 1-3 分钟，请耐心等待...")
            print()

            result = video_first_frame_and_wait(
                args.resource_uri,
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
                        filename = f"video_{timestamp}.mp4"

                        print(f"\n下载到: {output_dir / filename}")
                        download_file(video_url, str(output_dir / filename))
                else:
                    print("未获取到视频 URL")
            else:
                print("生成失败")
                sys.exit(1)
        else:
            result = video_first_frame(
                args.resource_uri,
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
