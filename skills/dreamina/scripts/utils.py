#!/usr/bin/env python3
"""
Dreamina API 工具函数模块
"""

import json
import time
import requests
from typing import Optional, Dict, Any, List
from pathlib import Path

from config import get_base_url, get_headers, ENDPOINTS, APP_ID


def mcp_request(endpoint: str, data: Dict[str, Any], timeout: int = 30) -> Dict[str, Any]:
    """
    发送 MCP API 请求

    Args:
        endpoint: API 端点路径
        data: 请求数据
        timeout: 超时时间(秒)

    Returns:
        API 响应 JSON

    Raises:
        requests.RequestException: 网络请求错误
        ValueError: API 返回错误
    """
    url = f"{get_base_url()}{endpoint}"
    headers = get_headers()

    # 禁用代理（内网访问）
    proxies = {"http": None, "https": None}

    response = requests.post(url, headers=headers, json=data, timeout=timeout, proxies=proxies)
    response.raise_for_status()

    result = response.json()

    # 检查 API 错误
    if result.get("ret") != "0" and result.get("ret") != 0:
        error_msg = result.get("errmsg", result.get("msg", "未知错误"))
        raise ValueError(f"API 错误: {error_msg}")

    return result


def query_result(
    submit_id: str,
    max_attempts: int = 30,
    interval: int = 5,
    verbose: bool = True,
) -> Optional[Dict[str, Any]]:
    """
    轮询查询生成结果

    Args:
        submit_id: 任务提交 ID
        max_attempts: 最大尝试次数
        interval: 轮询间隔(秒)
        verbose: 是否打印进度

    Returns:
        成功返回结果数据，失败返回 None
    """
    url = f"{get_base_url()}{ENDPOINTS['query_result']}"
    headers = get_headers()
    params = {"aid": APP_ID}

    for attempt in range(1, max_attempts + 1):
        if verbose:
            print(f"查询中... ({attempt}/{max_attempts})")

        try:
            # 禁用代理（内网访问）
            proxies = {"http": None, "https": None}

            response = requests.post(
                url,
                params=params,
                headers=headers,
                json={
                    "submit_ids": [submit_id],
                    "need_batch": True,
                    "history_ids": [],
                },
                timeout=30,
                proxies=proxies,
            )
            response.raise_for_status()
            result = response.json()

            # 响应结构: data.<submit_id> 或 data.history_list
            data = result.get("data", {})

            # 尝试直接通过 submit_id 获取
            history = data.get(submit_id)

            # 如果没有，尝试 history_list 结构
            if not history:
                history_list = data.get("history_list", [])
                if history_list:
                    history = history_list[0]

            if not history:
                if verbose:
                    print("  无历史记录，等待...")
                time.sleep(interval)
                continue

            status = history.get("status")

            if verbose:
                print(f"  状态: {status}")

            # status 50 表示成功
            if status == 50 or status == "succeed":
                if verbose:
                    print("生成成功!")
                return history

            elif status == 30 or status == "failed":
                fail_msg = history.get("fail_msg", "未知错误")
                if verbose:
                    print(f"生成失败: {fail_msg}")
                return None

            elif status in (20, 42, 45, "processing", "pending"):
                if verbose:
                    print(f"  处理中，等待 {interval} 秒...")
                time.sleep(interval)

            else:
                if verbose:
                    print(f"  状态: {status}，等待...")
                time.sleep(interval)

        except requests.RequestException as e:
            if verbose:
                print(f"  请求错误: {e}")
            time.sleep(interval)

    if verbose:
        print(f"超时: 已达到最大尝试次数 ({max_attempts})")
    return None


def extract_image_urls(result: Dict[str, Any]) -> List[str]:
    """
    从查询结果中提取图片 URL

    Args:
        result: query_result 返回的结果

    Returns:
        图片 URL 列表
    """
    urls = []
    item_list = result.get("item_list", [])
    for item in item_list:
        image = item.get("image", {})
        large_images = image.get("large_images", [])
        for img in large_images:
            url = img.get("image_url")
            if url:
                urls.append(url)
    return urls


def extract_image_uris(result: Dict[str, Any]) -> List[str]:
    """
    从查询结果中提取图片 URI (用于后续生成)

    Args:
        result: query_result 返回的结果

    Returns:
        图片 URI 列表 (tos-cn-i-xxx/xxx 格式)
    """
    uris = []
    item_list = result.get("item_list", [])
    for item in item_list:
        image = item.get("image", {})
        large_images = image.get("large_images", [])
        for img in large_images:
            uri = img.get("image_uri")
            if uri:
                uris.append(uri)
    return uris


def extract_video_url(result: Dict[str, Any]) -> Optional[str]:
    """
    从查询结果中提取视频 URL

    Args:
        result: query_result 返回的结果

    Returns:
        视频 URL 或 None
    """
    item_list = result.get("item_list", [])
    if item_list:
        video = item_list[0].get("video", {})

        # 尝试多种可能的路径
        # 路径 1: video.transcoded_video.origin.video_url (公网 API)
        transcoded = video.get("transcoded_video", {})
        if transcoded:
            origin = transcoded.get("origin", {})
            url = origin.get("video_url")
            if url:
                return url

        # 路径 2: video.video_resource.video_url (MCP API)
        video_resource = video.get("video_resource", {})
        url = video_resource.get("video_url")
        if url:
            return url

        # 路径 3: video.origin_video.video_url
        origin_video = video.get("origin_video", {})
        if origin_video:
            url = origin_video.get("video_url")
            if url:
                return url

    return None


def download_file(url: str, output_path: str, timeout: int = 120) -> int:
    """
    下载文件

    Args:
        url: 文件 URL
        output_path: 输出路径
        timeout: 超时时间(秒)

    Returns:
        下载的字节数
    """
    # 禁用代理（内网访问）
    proxies = {"http": None, "https": None}

    response = requests.get(url, timeout=timeout, stream=True, proxies=proxies)
    response.raise_for_status()

    output_file = Path(output_path)
    output_file.parent.mkdir(parents=True, exist_ok=True)

    total_bytes = 0
    with open(output_file, "wb") as f:
        for chunk in response.iter_content(chunk_size=8192):
            if chunk:
                f.write(chunk)
                total_bytes += len(chunk)

    print(f"已下载: {output_path} ({total_bytes} bytes)")
    return total_bytes


def generate_and_wait(
    endpoint: str,
    data: Dict[str, Any],
    max_attempts: int = 60,
    interval: int = 5,
    verbose: bool = True,
) -> Optional[Dict[str, Any]]:
    """
    发送生成请求并等待结果

    Args:
        endpoint: API 端点
        data: 请求数据
        max_attempts: 最大轮询次数
        interval: 轮询间隔(秒)
        verbose: 是否打印进度

    Returns:
        生成结果或 None
    """
    if verbose:
        print("发送生成请求...")

    result = mcp_request(endpoint, data)
    submit_id = result.get("data", {}).get("submit_id")

    if not submit_id:
        if verbose:
            print("错误: 未获取到 submit_id")
        return None

    if verbose:
        print(f"submit_id: {submit_id}")
        print(f"开始轮询结果 (最多 {max_attempts} 次，间隔 {interval} 秒)...")

    return query_result(submit_id, max_attempts, interval, verbose)


if __name__ == "__main__":
    print("=== Dreamina Utils 测试 ===")
    print(f"Base URL: {get_base_url()}")
    print(f"Query Endpoint: {ENDPOINTS['query_result']}")
