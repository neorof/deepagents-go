#!/usr/bin/env python3
"""
图片上传工具

用法:
    python upload_image.py <image_path>
    python upload_image.py /path/to/image.png

返回:
    resource_uri (例如: tos-cn-i-tb4s082cfz/xxxxx.png)
"""

import sys
import base64
import argparse
from pathlib import Path

from config import ENDPOINTS, AGENT_SCENE
from utils import mcp_request


def upload_image(image_path: str) -> str:
    """
    上传图片到 Dreamina

    Args:
        image_path: 图片文件路径

    Returns:
        resource_uri (TOS URI)

    Raises:
        FileNotFoundError: 文件不存在
        ValueError: API 错误
    """
    path = Path(image_path)
    if not path.exists():
        raise FileNotFoundError(f"文件不存在: {image_path}")

    # 读取并编码图片
    with open(path, "rb") as f:
        image_data = base64.b64encode(f.read()).decode("utf-8")

    # 发送上传请求
    data = {
        "image_data": image_data,
        "agent_scene": AGENT_SCENE,
    }

    result = mcp_request(ENDPOINTS["upload"], data)

    resource_uri = result.get("data", {}).get("resource_uri")
    if not resource_uri:
        raise ValueError("未获取到 resource_uri")

    return resource_uri


def upload_image_base64(image_base64: str) -> str:
    """
    上传 Base64 编码的图片

    Args:
        image_base64: Base64 编码的图片数据

    Returns:
        resource_uri (TOS URI)
    """
    # 移除可能的 data URI 前缀
    if image_base64.startswith("data:"):
        image_base64 = image_base64.split(",", 1)[1]

    data = {
        "image_data": image_base64,
        "agent_scene": AGENT_SCENE,
    }

    result = mcp_request(ENDPOINTS["upload"], data)

    resource_uri = result.get("data", {}).get("resource_uri")
    if not resource_uri:
        raise ValueError("未获取到 resource_uri")

    return resource_uri


def main():
    parser = argparse.ArgumentParser(description="上传图片到 Dreamina")
    parser.add_argument("image_path", help="图片文件路径")

    args = parser.parse_args()

    print(f"上传图片: {args.image_path}")

    try:
        resource_uri = upload_image(args.image_path)
        print(f"\n上传成功!")
        print(f"resource_uri: {resource_uri}")
        print(f"\n可用于后续操作:")
        print(f"  python edit_image.py {resource_uri} '添加一顶红色帽子'")
        print(f"  python video_first_frame.py {resource_uri}")
        return resource_uri
    except FileNotFoundError as e:
        print(f"错误: {e}")
        sys.exit(1)
    except ValueError as e:
        print(f"错误: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"上传失败: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
