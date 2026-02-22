#!/usr/bin/env python3
"""
Dreamina API 配置模块

配置优先级:
1. 环境变量 DREAMINA_COOKIE
2. 配置文件 ~/.dreamina.json
3. 默认值
"""

import os
import json
from pathlib import Path
from typing import Optional

# API 配置
BASE_URL = "https://jimeng.jianying.com"
PPE_ENV = "ppe_upload_image_api"
AGENT_SCENE = "infinite_canvas"
APP_ID = "513695"

# API 端点
ENDPOINTS = {
    "upload": "/dreamina/mcp/v1/upload",
    "image_generate": "/dreamina/mcp/v1/image_generate",
    "video_generate": "/dreamina/mcp/v1/video_generate",
    "query_result": "/mweb/v1/get_history_by_ids",
}

# 模型配置
IMAGE_MODELS = {
    "v3.0": "high_aes_general_v30l:general_v3.0_18b",
    "v4.0": "high_aes_general_v40",
    "v4.1": "high_aes_general_v41",
    "v4.5": "high_aes_general_v40l",
}

VIDEO_MODELS = {
    "v3.0_fast": "dreamina_ic_generate_video_model_vgfm_3.0_fast",
    "v3.0": "dreamina_ic_generate_video_model_vgfm_3.0",
}

# 图片比例
IMAGE_RATIOS = ["1:1", "16:9", "9:16", "3:2", "2:3", "4:3", "3:4", "21:9"]


def get_config_file_path() -> Path:
    """获取配置文件路径"""
    return Path.home() / ".dreamina.json"


def load_config_file() -> dict:
    """从配置文件加载配置"""
    config_path = get_config_file_path()
    if config_path.exists():
        try:
            with open(config_path, "r", encoding="utf-8") as f:
                return json.load(f)
        except (json.JSONDecodeError, IOError):
            pass
    return {}


def get_cookie() -> Optional[str]:
    """
    获取 Cookie

    优先级:
    1. 环境变量 DREAMINA_COOKIE
    2. 配置文件 ~/.dreamina.json 中的 cookie 字段
    """
    # 1. 环境变量
    cookie = os.environ.get("DREAMINA_COOKIE")
    if cookie:
        return cookie

    # 2. 配置文件
    config = load_config_file()
    cookie = config.get("cookie")
    if cookie:
        return cookie

    return None


def get_base_url() -> str:
    """获取 API 基础 URL"""
    config = load_config_file()
    return config.get("base_url", BASE_URL)


def get_ppe_env() -> str:
    """获取 PPE 泳道环境"""
    config = load_config_file()
    return config.get("ppe_env", PPE_ENV)


def get_headers() -> dict:
    """
    获取 API 请求头

    Returns:
        dict: 请求头字典

    Raises:
        ValueError: 如果未配置 Cookie
    """
    cookie = get_cookie()
    if not cookie:
        raise ValueError(
            "未配置 Cookie。请设置环境变量 DREAMINA_COOKIE 或创建配置文件 ~/.dreamina.json"
        )

    return {
        "Content-Type": "application/json",
        "x-tt-env": get_ppe_env(),
        "x-use-ppe": "1",
        "pf": "7",
        "appid": APP_ID,
        "Cookie": cookie,
    }


def save_config(cookie: str, base_url: Optional[str] = None, ppe_env: Optional[str] = None):
    """
    保存配置到配置文件

    Args:
        cookie: Cookie 字符串
        base_url: API 基础 URL (可选)
        ppe_env: PPE 泳道环境 (可选)
    """
    config = load_config_file()
    config["cookie"] = cookie
    if base_url:
        config["base_url"] = base_url
    if ppe_env:
        config["ppe_env"] = ppe_env

    config_path = get_config_file_path()
    with open(config_path, "w", encoding="utf-8") as f:
        json.dump(config, f, indent=2, ensure_ascii=False)

    print(f"配置已保存到: {config_path}")


if __name__ == "__main__":
    # 测试配置加载
    print("=== Dreamina 配置测试 ===")
    print(f"Base URL: {get_base_url()}")
    print(f"PPE 环境: {get_ppe_env()}")

    cookie = get_cookie()
    if cookie:
        print(f"Cookie: {cookie[:50]}..." if len(cookie) > 50 else f"Cookie: {cookie}")
    else:
        print("Cookie: 未配置")
        print("\n请设置环境变量或创建配置文件:")
        print("  export DREAMINA_COOKIE='sessionid=xxx;uid_tt=xxx'")
        print("  或")
        print(f"  创建 {get_config_file_path()} 文件")
