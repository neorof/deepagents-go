# Dreamina MCP API 参考

## 基础配置

### Base URL
```
https://jimeng.jianying.com
```

### PPE 泳道环境
```
ppe_upload_image_api
```

### 通用请求头
```python
headers = {
    "Content-Type": "application/json",
    "x-tt-env": "ppe_upload_image_api",
    "x-use-ppe": "1",
    "pf": "7",
    "appid": "513695",
    "Cookie": "<cookie>"
}
```

## API 端点

### 1. 图片上传

**端点**: `POST /dreamina/mcp/v1/upload`

**请求体**:
```json
{
    "image_data": "<base64_encoded_image>",
    "agent_scene": "infinite_canvas"
}
```

**响应**:
```json
{
    "ret": "0",
    "errmsg": "success",
    "data": {
        "resource_uri": "tos-cn-i-tb4s082cfz/xxxxx.png"
    }
}
```

---

### 2. 图片生成

**端点**: `POST /dreamina/mcp/v1/image_generate`

#### 2.1 文生图 (text2imageV2)

**请求体**:
```json
{
    "generate_type": "text2imageV2",
    "agent_scene": "infinite_canvas",
    "prompt": "一只可爱的橘猫",
    "ratio": "1:1",
    "model_key": "high_aes_general_v30l:general_v3.0_18b",
    "subject_id": "text2image_api"
}
```

#### 2.2 图生图 (seedEdit40)

**请求体**:
```json
{
    "generate_type": "seedEdit40",
    "agent_scene": "infinite_canvas",
    "prompt": "添加一顶红色帽子",
    "ratio": "1:1",
    "resource_uri_list": ["tos-cn-i-tb4s082cfz/xxxxx.png"],
    "subject_id": "seed_edit_api"
}
```

**响应**:
```json
{
    "ret": "0",
    "errmsg": "success",
    "data": {
        "submit_id": "b02d84e4-7cd0-4007-8952-e2d835f3855f"
    }
}
```

---

### 3. 视频生成

**端点**: `POST /dreamina/mcp/v1/video_generate`

#### 3.1 首帧生视频 (image2videoV2)

**请求体**:
```json
{
    "generate_type": "image2videoV2",
    "agent_scene": "infinite_canvas",
    "first_frame_resource_uri": "tos-cn-i-tb4s082cfz/xxxxx.png",
    "duration": 5,
    "prompt": "猫咪缓缓走动"
}
```

#### 3.2 首尾帧生视频 (startEnd2Video)

**请求体**:
```json
{
    "generate_type": "startEnd2Video",
    "agent_scene": "infinite_canvas",
    "first_frame_resource_uri": "tos-cn-i-tb4s082cfz/first.png",
    "last_frame_resource_uri": "tos-cn-i-tb4s082cfz/last.png",
    "duration": 5,
    "prompt": "smooth transition between frames"
}
```

#### 3.3 多帧生视频 (multiFrame2video)

**请求体**:
```json
{
    "generate_type": "multiFrame2video",
    "agent_scene": "infinite_canvas",
    "media_resource_uri_list": [
        "tos-cn-i-tb4s082cfz/frame1.png",
        "tos-cn-i-tb4s082cfz/frame2.png",
        "tos-cn-i-tb4s082cfz/frame3.png"
    ],
    "duration_list": [2, 2, 2],
    "media_type_list": ["image", "image", "image"],
    "prompt_list": ["frame 1", "frame 2", "frame 3"]
}
```

**响应**:
```json
{
    "ret": "0",
    "errmsg": "success",
    "data": {
        "submit_id": "b02d84e4-7cd0-4007-8952-e2d835f3855f"
    }
}
```

---

### 4. 结果查询

**端点**: `POST /mweb/v1/get_history_by_ids?aid=513695`

**请求体**:
```json
{
    "submit_ids": ["b02d84e4-7cd0-4007-8952-e2d835f3855f"],
    "need_batch": true,
    "history_ids": []
}
```

**响应 (图片)**:
```json
{
    "ret": "0",
    "data": {
        "history_list": [{
            "status": "succeed",
            "item_list": [{
                "image": {
                    "large_images": [{
                        "image_uri": "tos-cn-i-tb4s082cfz/xxx.png",
                        "image_url": "https://p26-dreamina-sign.byteimg.com/...",
                        "width": 2048,
                        "height": 2048
                    }]
                }
            }]
        }]
    }
}
```

**响应 (视频)**:
```json
{
    "ret": "0",
    "data": {
        "history_list": [{
            "status": "succeed",
            "item_list": [{
                "video": {
                    "video_resource": {
                        "video_url": "https://...",
                        "video_uri": "tos-cn-v-xxx/...",
                        "duration": 5000,
                        "width": 960,
                        "height": 960
                    }
                }
            }]
        }]
    }
}
```

---

## 参数说明

### 图片比例 (ratio)

| 值 | 说明 |
|----|------|
| 1:1 | 正方形 |
| 16:9 | 横版宽屏 |
| 9:16 | 竖版 |
| 3:2 | 横版 |
| 2:3 | 竖版 |
| 4:3 | 横版 |
| 3:4 | 竖版 |
| 21:9 | 超宽屏 |

### 图片模型 (model_key)

| 版本 | model_key |
|------|-----------|
| v3.0 | high_aes_general_v30l:general_v3.0_18b |
| v4.0 | high_aes_general_v40 |
| v4.1 | high_aes_general_v41 |
| v4.5 | high_aes_general_v40l |

### 视频时长 (duration)

| 值 | 说明 |
|----|------|
| 5 | 5 秒 |
| 10 | 10 秒 |

### 任务状态 (status)

| 状态 | 说明 |
|------|------|
| pending | 等待中 |
| processing | 处理中 |
| succeed | 成功 |
| failed | 失败 |

---

## 错误码

| ret | 说明 |
|-----|------|
| 0 | 成功 |
| 非0 | 失败，查看 errmsg |

常见错误:
- Cookie 过期
- 参数格式错误
- 资源不存在
- 配额不足

---

## 注意事项

1. **Cookie 认证**: 所有请求需要携带有效的 Cookie
2. **PPE 泳道**: 需要内网环境访问 PPE 泳道
3. **轮询间隔**: 图片建议 3-5 秒，视频建议 10 秒
4. **URL 有效期**: 生成的 URL 带签名，约 1-2 小时过期
5. **文生图返回**: 每次生成返回 4 张变体图片
