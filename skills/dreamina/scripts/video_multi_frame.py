#!/usr/bin/env python3
"""
多帧生视频工具 (multiFrame2video)

用法:
    python video_multi_frame.py <frame1_uri> <frame2_uri> [frame3_uri...]
    python video_multi_frame.py tos-cn-i-xxx/f1.png tos-cn-i-xxx/f2.png tos-cn-i-xxx/f3.png --wait
    python video_multi_frame.py frame1.png frame2.png --duration 2,2 -o ./output/

参数:
    frame_uris: 关键帧图片的 TOS URI 列表 (至少 2 帧)
    --duration: 每帧时长列表 (逗号分隔，如 "2,2,2"，默认每帧 2 秒)
    --prompts: 每帧描述列表 (逗号分隔，可选)
    --wait: 等待生成完成
    -o/--output: 输出目录 (自动下载)
"""

import sys
import argparse
from pathlib import Path
from datetime import datetime
from typing import List

from config import ENDPOINTS, AGENT_SCENE
from utils import mcp_request, generate_and_wait, extract_video_url, download_file


def video_multi_frame(
    frame_uris: List[str],
    duration_list: List[int] = None,
    prompt_list: List[str] = None,
) -> dict:
    """
    多帧生视频

    Args:
        frame_uris: 关键帧图片的 TOS URI 列表 (至少 2 帧)
        duration_list: 每帧时长列表 (秒)
        prompt_list: 每帧描述列表

    Returns:
        API 响应
    """
    if len(frame_uris) < 2:
        raise ValueError("至少需要 2 个关键帧")

    # 默认每帧 2 秒
    if duration_list is None:
        duration_list = [2] * len(frame_uris)
    elif len(duration_list) != len(frame_uris):
        raise ValueError("时长列表长度必须与帧数一致")

    # 默认描述
    if prompt_list is None:
        prompt_list = [f"frame {i+1}" for i in range(len(frame_uris))]
    elif len(prompt_list) != len(frame_uris):
        raise ValueError("描述列表长度必须与帧数一致")

    # 构建媒体类型列表 (全部为 image)
    media_type_list = ["image"] * len(frame_uris)

    data = {
        "generate_type": "multiFrame2video",
        "agent_scene": AGENT_SCENE,
        "media_resource_uri_list": frame_uris,
        "duration_list": duration_list,
        "media_type_list": media_type_list,
        "prompt_list": prompt_list,
    }

    return mcp_request(ENDPOINTS["video_generate"], data)


def video_multi_frame_and_wait(
    frame_uris: List[str],
    duration_list: List[int] = None,
    prompt_list: List[str] = None,
    max_attempts: int = 180,
    interval: int = 10,
    verbose: bool = True,
) -> dict:
    """
    多帧生视频并等待结果

    Args:
        frame_uris: 关键帧图片的 TOS URI 列表
        duration_list: 每帧时长列表 (秒)
        prompt_list: 每帧描述列表
        max_attempts: 最大轮询次数
        interval: 轮询间隔
        verbose: 是否打印进度

    Returns:
        生成结果
    """
    if len(frame_uris) < 2:
        raise ValueError("至少需要 2 个关键帧")

    # 默认每帧 2 秒
    if duration_list is None:
        duration_list = [2] * len(frame_uris)

    # 默认描述
    if prompt_list is None:
        prompt_list = [f"frame {i+1}" for i in range(len(frame_uris))]

    media_type_list = ["image"] * len(frame_uris)

    data = {
        "generate_type": "multiFrame2video",
        "agent_scene": AGENT_SCENE,
        "media_resource_uri_list": frame_uris,
        "duration_list": duration_list,
        "media_type_list": media_type_list,
        "prompt_list": prompt_list,
    }

    return generate_and_wait(
        ENDPOINTS["video_generate"],
        data,
        max_attempts=max_attempts,
        interval=interval,
        verbose=verbose,
    )


def main():
    parser = argparse.ArgumentParser(description="多帧生视频 (multiFrame2video)")
    parser.add_argument(
        "frame_uris",
        nargs="+",
        help="关键帧图片的 TOS URI 列表 (至少 2 帧)",
    )
    parser.add_argument(
        "--duration",
        help="每帧时长 (逗号分隔，如 '2,2,2'，默认每帧 2 秒)",
    )
    parser.add_argument(
        "--prompts",
        help="每帧描述 (逗号分隔，可选)",
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

    # 解析时长
    duration_list = None
    if args.duration:
        duration_list = [int(d.strip()) for d in args.duration.split(",")]

    # 解析描述
    prompt_list = None
    if args.prompts:
        prompt_list = [p.strip() for p in args.prompts.split(",")]

    print("=" * 50)
    print("多帧生视频 (multiFrame2video)")
    print("=" * 50)
    print(f"关键帧数量: {len(args.frame_uris)}")
    for i, uri in enumerate(args.frame_uris):
        print(f"  帧 {i+1}: {uri}")
    if duration_list:
        print(f"每帧时长: {duration_list}")
    if prompt_list:
        print(f"每帧描述: {prompt_list}")
    print()

    try:
        if args.wait or args.output:
            print("注意: 多帧视频生成可能需要 2-5 分钟，请耐心等待...")
            print()

            result = video_multi_frame_and_wait(
                args.frame_uris,
                duration_list=duration_list,
                prompt_list=prompt_list,
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
                        filename = f"video_multi_frame_{timestamp}.mp4"

                        print(f"\n下载到: {output_dir / filename}")
                        download_file(video_url, str(output_dir / filename))
                else:
                    print("未获取到视频 URL")
            else:
                print("生成失败")
                sys.exit(1)
        else:
            result = video_multi_frame(
                args.frame_uris,
                duration_list=duration_list,
                prompt_list=prompt_list,
            )
            submit_id = result.get("data", {}).get("submit_id")
            print(f"submit_id: {submit_id}")
            print(f"\n多帧视频生成可能需要 2-5 分钟")
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
