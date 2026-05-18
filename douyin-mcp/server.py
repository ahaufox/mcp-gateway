#!/usr/bin/env python3
"""
抖音无水印视频下载并提取文本的 MCP 服务器

该服务器提供以下功能：
1. 解析抖音分享链接获取无水印视频链接
2. 下载视频并提取音频
3. 从音频中提取文本内容
4. 自动清理中间文件
"""

import os
import re
import json
import requests
import tempfile
from pathlib import Path
from typing import Optional, List
import ffmpeg
from urllib.parse import urlparse, parse_qs
from urllib import request
from http import HTTPStatus
import dashscope
import logging

from mcp.server.fastmcp import FastMCP
from mcp.server.fastmcp import Context
from mcp.server.transport_security import TransportSecuritySettings

# 导入自定义模块
from schemas import (
    VideoDownloadInfo, DownloadResult, AudioResult, 
    VideoBasicInfo, TextExtractionResult,
    AlbumDownloadResult, UserVideoList, AwemeItem
)
from utils import DouyinUtils

# 配置日志
logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)


# 创建 MCP 服务器实例
# Docker 内部网络通信需要禁用 DNS rebinding protection，
# 否则 MCP SDK 会因 Host 头为容器名（如 douyin-mcp:8000）而返回 421
mcp = FastMCP("Douyin MCP Server", 
              dependencies=["requests", "ffmpeg-python", "tqdm", "dashscope"],
              transport_security=TransportSecuritySettings(
                  enable_dns_rebinding_protection=False
              ))

# 请求头，模拟移动端访问
HEADERS = {
    'User-Agent': 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) EdgiOS/121.0.2277.107 Version/17.0 Mobile/15E148 Safari/604.1'
}

# 默认 API 配置
DEFAULT_MODEL = "paraformer-v2"


class DouyinProcessor:
    """抖音视频处理器"""
    
    def __init__(self, api_key: str, model: Optional[str] = None):
        self.api_key = api_key
        self.model = model or DEFAULT_MODEL
        self.temp_dir = Path(tempfile.mkdtemp())
        # 设置阿里云百炼API密钥
        dashscope.api_key = api_key
    
    def __del__(self):
        """清理临时目录"""
        import shutil
        if hasattr(self, 'temp_dir') and self.temp_dir.exists():
            shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    def parse_share_url(self, share_text: str) -> dict:
        """从分享文本中提取无水印视频链接"""
        # 提取分享链接
        urls = re.findall(r'http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+', share_text)
        if not urls:
            raise ValueError("未找到有效的分享链接")
        
        share_url = urls[0]
        # 获取重定向后的链接
        resp = requests.get(share_url, headers=HEADERS, allow_redirects=True, timeout=10)
        final_url = resp.url
        
        # 提取 ID
        item_id = ""
        if "/video/" in final_url:
            item_id = final_url.split("/video/")[1].split("/")[0].split("?")[0]
        elif "/note/" in final_url:
            item_id = final_url.split("/note/")[1].split("/")[0].split("?")[0]
        else:
            # 尝试通过 iesdouyin 获取
            video_id = final_url.split("?")[0].strip("/").split("/")[-1]
            item_id = video_id

        # 构造 API URL (使用 web 接口获取更多信息)
        api_url = f"https://www.douyin.com/aweme/v1/web/aweme/detail/?aweme_id={item_id}&aid=1128&version_name=23.5.0&device_platform=webapp&os_version=17.0"
        
        # 签名
        signed_url = DouyinUtils.sign_url(api_url)
        
        # 获取详细信息
        headers = HEADERS.copy()
        ttwid = DouyinUtils.get_ttwid()
        if ttwid:
            headers['Cookie'] = f"ttwid={ttwid}"
            
        api_response = requests.get(signed_url, headers=headers, timeout=10)
        api_response.raise_for_status()
        try:
            data = api_response.json()
        except (json.JSONDecodeError, ValueError):
            log.warning(f"API Response is not JSON. Status: {api_response.status_code}")
            # 降级方案：使用之前的 HTML 解析逻辑
            return self._parse_html_fallback(final_url)
        
        if "aweme_detail" not in data:
            log.warning(f"aweme_detail not found in API response: {data}")
            # 降级方案：使用之前的 HTML 解析逻辑
            return self._parse_html_fallback(final_url)

        aweme = data["aweme_detail"]
        desc = aweme.get("desc", "").strip() or f"douyin_{item_id}"
        desc = re.sub(r'[\\/:*?"<>|]', '_', desc)
        
        # 类型判断
        is_album = aweme.get("images") is not None
        
        video_url = ""
        images = []
        
        if is_album:
            images = [img["url_list"][0] for img in aweme["images"]]
            video_url = images[0] if images else ""
        else:
            # 获取 1080p 链接
            uri = aweme["video"]["play_addr"]["uri"]
            video_url = f"https://aweme.snssdk.com/aweme/v1/play/?video_id={uri}&ratio=1080p&line=0"

        return {
            "url": video_url,
            "title": desc,
            "video_id": item_id,
            "is_album": is_album,
            "images": images
        }

    def _parse_html_fallback(self, share_url: str) -> dict:
        """HTML 解析备选方案"""
        # ... (之前的逻辑，略微优化以支持图集声明) ...
        response = requests.get(share_url, headers=HEADERS)
        response.raise_for_status()
        
        # 尝试匹配 _ROUTER_DATA
        pattern = re.compile(pattern=r"window\._ROUTER_DATA\s*=\s*(.*?)</script>", flags=re.DOTALL)
        find_res = pattern.search(response.text)

        # 如果 _ROUTER_DATA 失败，尝试 _RENDER_DATA (新版抖音可能使用)
        if not find_res:
            pattern_render = re.compile(pattern=r"window\._RENDER_DATA\s*=\s*(.*?)</script>", flags=re.DOTALL)
            find_res = pattern_render.search(response.text)

        # 如果还是失败，可能是 _SSR_DATA
        if not find_res:
             pattern_ssr = re.compile(pattern=r"window\._SSR_DATA\s*=\s*(.*?)</script>", flags=re.DOTALL)
             find_res = pattern_ssr.search(response.text)

        if not find_res or not find_res.group(1):
             # 记录一下获取到的HTML便于调试
             log.error(f"HTML parsing failed. URL: {share_url}, Content Length: {len(response.text)}")
             raise ValueError(f"解析页面数据失败，未找到数据节点. 页面长度: {len(response.text)}")

        try:
            json_data = json.loads(find_res.group(1).strip())
        except (json.JSONDecodeError, ValueError) as e:
            log.error(f"Failed to parse HTML data: {str(e)}")
            raise ValueError(f"无法解析网页中的 JSON 数据: {str(e)}")
        VIDEO_ID_PAGE_KEY = "video_(id)/page"
        NOTE_ID_PAGE_KEY = "note_(id)/page"
        
        if VIDEO_ID_PAGE_KEY in json_data["loaderData"]:
            original_video_info = json_data["loaderData"][VIDEO_ID_PAGE_KEY]["videoInfoRes"]
        elif NOTE_ID_PAGE_KEY in json_data["loaderData"]:
            original_video_info = json_data["loaderData"][NOTE_ID_PAGE_KEY]["videoInfoRes"]
        else:
            raise Exception("无法从 HTML JS 数据中提取信息")

        data = original_video_info["item_list"][0]
        video_id = data["aweme_id"]
        
        is_album = data.get("images") is not None
        images = []
        if is_album:
            images = [img["url_list"][0] for img in data["images"]]
            video_url = images[0]
        else:
            raw_video_url = data["video"]["play_addr"]["url_list"][0]
            if "video_id=http" in raw_video_url:
                parsed_raw = urlparse(raw_video_url)
                video_url = parse_qs(parsed_raw.query).get('video_id', [raw_video_url])[0]
            else:
                video_url = raw_video_url.replace("playwm", "play") if "playwm" in raw_video_url else raw_video_url
            
        desc = data.get("desc", "").strip() or f"douyin_{video_id}"
        desc = re.sub(r'[\\/:*?"<>|]', '_', desc)
        
        return {
            "url": video_url,
            "title": desc,
            "video_id": video_id,
            "is_album": is_album,
            "images": images
        }
    
    async def download_video(self, video_info: dict, ctx: Context) -> Path:
        """异步下载视频到临时目录"""
        filename = f"{video_info['video_id']}.mp4"
        filepath = self.temp_dir / filename
        
        await ctx.info(f"正在下载视频: {video_info['title']}")
        
        response = requests.get(video_info['url'], headers=HEADERS, stream=True)
        # 调试信息
        log.info(f"下载请求状态码: {response.status_code}, Headers: {dict(response.headers)}")
        response.raise_for_status()
        
        # 获取文件大小
        total_size = int(response.headers.get('content-length', 0))
        
        # 异步下载文件，显示进度
        with open(filepath, 'wb') as f:
            downloaded = 0
            for chunk in response.iter_content(chunk_size=8192):
                if chunk:
                    f.write(chunk)
                    downloaded += len(chunk)
                    if total_size > 0:
                        await ctx.report_progress(downloaded, total_size)
        
        await ctx.info(f"视频下载完成: {filepath}")
        return filepath
    
    def extract_audio(self, video_path: Path) -> Path:
        """从视频文件中提取音频"""
        audio_path = video_path.with_suffix('.mp3')
        
        try:
            (
                ffmpeg
                .input(str(video_path))
                .output(str(audio_path), acodec='libmp3lame', q=0)
                .run(capture_stdout=True, capture_stderr=True, overwrite_output=True)
            )
            return audio_path
        except Exception as e:
            stderr = ""
            if hasattr(e, 'stderr'):
                stderr = f"\nFFmpeg stderr: {e.stderr.decode()}"
            raise Exception(f"提取音频时出错: {str(e)}{stderr}")
    
    def extract_text_from_video_url(self, video_url: str) -> str:
        """从视频URL中提取文字（使用阿里云百炼API）"""
        try:
            # 发起异步转录任务
            task_response = dashscope.audio.asr.Transcription.async_call(
                model=self.model,
                file_urls=[video_url],
                language_hints=['zh', 'en']
            )
            
            # 等待转录完成
            transcription_response = dashscope.audio.asr.Transcription.wait(
                task=task_response.output.task_id
            )
            
            if transcription_response.status_code == HTTPStatus.OK:
                # 获取转录结果
                for transcription in transcription_response.output['results']:
                    url = transcription['transcription_url']
                    result = json.loads(request.urlopen(url).read().decode('utf8'))
                    
                    # 保存结果到临时文件
                    temp_json_path = self.temp_dir / 'transcription.json'
                    with open(temp_json_path, 'w') as f:
                        json.dump(result, f, indent=4, ensure_ascii=False)
                    
                    # 提取文本内容
                    if 'transcripts' in result and len(result['transcripts']) > 0:
                        return result['transcripts'][0]['text']
                    else:
                        return "未识别到文本内容"
                        
            else:
                raise Exception(f"转录失败: {transcription_response.output.message}")
                
        except Exception as e:
            raise Exception(f"提取文字时出错: {str(e)}")
    
    async def download_album(self, video_info: dict, ctx: Context) -> List[Path]:
        """下载图文作品的所有图片"""
        image_urls = video_info.get("images", [])
        if not image_urls:
             if video_info.get("url"):
                 image_urls = [video_info["url"]]
             else:
                 raise ValueError("该作品不包含图片内容")
                 
        # 使用当前目录下的 downloads 文件夹
        downloads_base = Path.cwd() / "downloads"
        album_dir = downloads_base / f"album_{video_info['video_id']}"
        album_dir.mkdir(parents=True, exist_ok=True)
        
        saved_paths = []
        if ctx:
            await ctx.info(f"正在下载图集: {video_info['title']}，共 {len(image_urls)} 张图片...")
        
        for i, url in enumerate(image_urls):
            filename = f"{i+1}.jpg"
            save_path = album_dir / filename
            
            response = requests.get(url, headers=HEADERS, stream=True, timeout=20)
            response.raise_for_status()
            
            with open(save_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            
            saved_paths.append(save_path)
            
        return saved_paths

    def get_user_videos(self, sec_uid: str, count: int = 35, cursor: int = 0) -> dict:
        """获取用户的作品列表"""
        api_url = f"https://www.douyin.com/aweme/v1/web/aweme/post/?sec_user_id={sec_uid}&count={count}&max_cursor={cursor}&aid=1128&version_name=23.5.0&device_platform=webapp&os_version=17.0"
        signed_url = DouyinUtils.sign_url(api_url)
        
        headers = HEADERS.copy()
        ttwid = DouyinUtils.get_ttwid()
        if ttwid:
            headers['Cookie'] = f"ttwid={ttwid}"
            
        response = requests.get(signed_url, headers=headers, timeout=10)
        response.raise_for_status()
        data = response.json()
        
        import time
        aweme_list = []
        for item in data.get("aweme_list", []):
            is_album = item.get("images") is not None
            aweme_list.append({
                "aweme_id": item["aweme_id"],
                "desc": item.get("desc", ""),
                "type": "image" if is_album else "video",
                "author": item["author"]["nickname"],
                "create_time": time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(item["create_time"])),
                "url": item["images"][0]["url_list"][0] if is_album else item["video"]["play_addr"]["url_list"][0]
            })
            
        return {
            "user_id": sec_uid,
            "nickname": data.get("aweme_list", [{}])[0].get("author", {}).get("nickname", "Unknown") if data.get("aweme_list") else "",
            "aweme_list": aweme_list,
            "has_more": data.get("has_more", False),
            "cursor": data.get("max_cursor", 0)
        }

    def cleanup_files(self, *file_paths: Path):
        """清理指定的文件"""
        for file_path in file_paths:
            if file_path.exists():
                file_path.unlink()


@mcp.tool()
def get_douyin_download_link(share_link: str) -> VideoDownloadInfo:
    """
    获取抖音视频的无水印下载链接
    
    参数:
    - share_link: 抖音分享链接或包含链接的文本
    
    返回:
    - 结构化的无水印下载链接信息
    """
    try:
        processor = DouyinProcessor("")  # 获取下载链接不需要API密钥
        video_info = processor.parse_share_url(share_link)
        
        return VideoDownloadInfo(
            status="success",
            video_id=video_info["video_id"],
            title=video_info["title"],
            download_url=video_info["url"],
            description=f"视频标题: {video_info['title']}",
            usage_tip="可以直接使用此链接下载无水印视频"
        )
        
    except Exception as e:
        return VideoDownloadInfo(
            status="error",
            error=f"获取下载链接失败: {str(e)}"
        )


@mcp.tool()
async def download_video(
    share_link: str,
    save_path: Optional[str] = None,
    ctx: Context = None
) -> DownloadResult:
    """
    下载抖音无水印视频到本地
    
    参数:
    - share_link: 抖音分享链接或包含链接的文本
    - save_path: 保存路径（可选，默认为当前目录下的 downloads 文件夹）
    
    返回:
    - 保存成功的详细信息
    """
    try:
        # 获取下载链接不需要API密钥
        processor = DouyinProcessor("")
        
        await ctx.info("正在解析抖音分享链接...")
        video_info = processor.parse_share_url(share_link)
        
        if video_info.get("is_album"):
             return DownloadResult(
                status="error",
                message="该链接指向的是图文作品，请使用 download_album 工具下载",
                error="Type mismatch: Expected video, got album"
            )
        
        # 确定保存完整路径
        import shutil
        current_dir = Path.cwd()
        
        if save_path:
            target_path = Path(save_path)
            if target_path.is_dir():
                # 如果是目录，则使用视频标题并清理非法字符作为文件名
                safe_title = re.sub(r'[\\/:*?"<>|]', '_', video_info['title'])
                final_path = target_path / f"{safe_title}.mp4"
            else:
                final_path = target_path
        else:
            # 默认保存在当前目录的 downloads 文件夹
            downloads_dir = current_dir / "downloads"
            downloads_dir.mkdir(exist_ok=True)
            safe_title = re.sub(r'[\\/:*?"<>|]', '_', video_info['title'])
            final_path = downloads_dir / f"{safe_title}.mp4"
            
        # 确保父目录存在
        final_path.parent.mkdir(parents=True, exist_ok=True)
            
        # 下载到临时文件
        temp_file = await processor.download_video(video_info, ctx)
        
        # 移动到最终位置
        shutil.move(str(temp_file), str(final_path))
        
        await ctx.info(f"视频已成功下载至: {final_path.absolute()}")
        return DownloadResult(
            status="success",
            message="视频下载成功",
            video_id=video_info["video_id"],
            title=video_info["title"],
            path=str(final_path.absolute())
        )
        
    except Exception as e:
        if ctx:
            await ctx.error(f"下载过程中出现错误: {str(e)}")
        return DownloadResult(
            status="error",
            message="视频下载失败",
            error=str(e)
        )

@mcp.tool()
async def extract_audio(
    share_link: Optional[str] = None,
    video_path: Optional[str] = None,
    save_path: Optional[str] = None,
    ctx: Context = None
) -> AudioResult:
    """
    从抖音视频中提取音频 (MP3)
    
    参数:
    - share_link: 抖音分享链接（如果提供，将自动下载视频并行提取音频）
    - video_path: 本地视频文件路径（如果提供，直接提取该文件的音频）
    - save_path: 音频保存路径（可选，默认为 downloads 文件夹）
    
    返回:
    - 提取成功的详细信息
    """
    try:
        if not share_link and not video_path:
            raise ValueError("必须提供 share_link 或 video_path 其中之一")

        processor = DouyinProcessor("")
        import shutil
        current_dir = Path.cwd()
        downloads_dir = current_dir / "downloads"
        downloads_dir.mkdir(exist_ok=True)

        final_video_path = None
        title = "audio_extraction"

        if share_link:
            await ctx.info("正在解析并下载视频以提取音频...")
            video_info = processor.parse_share_url(share_link)
            title = video_info['title']
            final_video_path = await processor.download_video(video_info, ctx)
        else:
            final_video_path = Path(video_path)
            if not final_video_path.exists():
                raise FileNotFoundError(f"未找到视频文件: {video_path}")
            title = final_video_path.stem

        await ctx.info("正在提取音频...")
        # 调用处理器内部方法提取音频（生成在临时目录或同目录）
        temp_audio_path = processor.extract_audio(final_video_path)

        # 确定最终保存路径
        safe_title = re.sub(r'[\\/:*?"<>|]', '_', title)
        if save_path:
            target_save_path = Path(save_path)
            if target_save_path.is_dir():
                final_audio_path = target_save_path / f"{safe_title}.mp3"
            else:
                final_audio_path = target_save_path
        else:
            final_audio_path = downloads_dir / f"{safe_title}.mp3"

        # 确保父目录存在
        final_audio_path.parent.mkdir(parents=True, exist_ok=True)

        # 移动音频文件
        shutil.move(str(temp_audio_path), str(final_audio_path))
        
        await ctx.info(f"音频已成功提取至: {final_audio_path.absolute()}")
        return AudioResult(
            status="success",
            message="音频提取成功",
            title=title,
            path=str(final_audio_path.absolute())
        )

    except Exception as e:
        if ctx:
            await ctx.error(f"提取音频过程中出现错误: {str(e)}")
        return AudioResult(
            status="error",
            message="音频提取失败",
            error=str(e)
        )


@mcp.tool()
async def extract_douyin_text(
    share_link: str,
    model: Optional[str] = None,
    ctx: Context = None
) -> TextExtractionResult:
    """
    从抖音分享链接提取视频中的文本内容
    
    参数:
    - share_link: 抖音分享链接或包含链接的文本
    - model: 语音识别模型（可选，默认使用paraformer-v2）
    
    返回:
    - 结构化的文本提取结果
    
    注意: 需要设置环境变量 API_KEY
    """
    try:
        # 从环境变量获取API密钥
        api_key = os.getenv('API_KEY')
        if not api_key:
            raise ValueError("未设置环境变量 API_KEY，请在配置中添加阿里云百炼API密钥")
        
        processor = DouyinProcessor(api_key, model)
        
        # 解析视频链接
        await ctx.info("正在解析抖音分享链接...")
        video_info = processor.parse_share_url(share_link)
        
        # 直接使用视频URL进行文本提取
        await ctx.info("正在从视频中提取文本...")
        text_content = processor.extract_text_from_video_url(video_info['url'])
        
        await ctx.info("文本提取完成!")
        return TextExtractionResult(
            status="success",
            text=text_content,
            title=video_info['title']
        )
        
    except Exception as e:
        if ctx:
            await ctx.error(f"处理过程中出现错误: {str(e)}")
        return TextExtractionResult(
            status="error",
            error=str(e)
        )


@mcp.tool()
async def download_album(
    share_link: str,
    ctx: Context = None
) -> AlbumDownloadResult:
    """
    下载抖音图文作品（所有原图）
    
    参数:
    - share_link: 抖音分享链接或包含链接的文本
    
    返回:
    - 下载成功的详细信息和本地路径列表
    """
    try:
        processor = DouyinProcessor("")
        if ctx:
            await ctx.info("正在解析图集信息...")
        video_info = processor.parse_share_url(share_link)
        
        if not video_info.get("is_album"):
             raise ValueError("该链接指向的不是图文作品")
             
        saved_paths = await processor.download_album(video_info, ctx)
        
        return AlbumDownloadResult(
            status="success",
            message=f"成功下载图集，共 {len(saved_paths)} 张图片",
            video_id=video_info["video_id"],
            title=video_info["title"],
            image_paths=[str(p.absolute()) for p in saved_paths]
        )
    except Exception as e:
        if ctx:
            await ctx.error(f"下载图集失败: {str(e)}")
        return AlbumDownloadResult(
            status="error",
            message="图集下载失败",
            error=str(e)
        )

@mcp.tool()
def get_user_videos(
    sec_uid: str,
    count: int = 35,
    cursor: int = 0
) -> UserVideoList:
    """
    获取指定用户的作品列表（支持分页）
    
    参数:
    - sec_uid: 用户的 sec_uid（可通过解析用户分享链接获得）
    - count: 每页获取数量（默认 35）
    - cursor: 分页游标（默认 0）
    """
    try:
        processor = DouyinProcessor("")
        result = processor.get_user_videos(sec_uid, count, cursor)
        
        aweme_items = [AwemeItem(**item) for item in result["aweme_list"]]
        
        return UserVideoList(
            status="success",
            user_id=result["user_id"],
            nickname=result["nickname"],
            aweme_list=aweme_items,
            has_more=result["has_more"],
            cursor=result["cursor"]
        )
    except Exception as e:
        return UserVideoList(
            status="error",
            error=str(e)
        )


@mcp.tool()
def parse_douyin_video_info(share_link: str) -> VideoBasicInfo:
    """
    解析抖音分享链接，获取视频基本信息
    
    参数:
    - share_link: 抖音分享链接或包含链接的文本
    
    返回:
    - 结构化的视频基本信息
    """
    try:
        processor = DouyinProcessor("")  # 不需要API密钥来解析链接
        video_info = processor.parse_share_url(share_link)
        
        return VideoBasicInfo(
            status="success",
            video_id=video_info["video_id"],
            title=video_info["title"],
            download_url=video_info["url"]
        )
        
    except Exception as e:
        return VideoBasicInfo(
            status="error",
            error=str(e)
        )


@mcp.resource("douyin://video/{video_id}")
def get_video_info(video_id: str) -> str:
    """
    获取指定视频ID的详细信息
    
    参数:
    - video_id: 抖音视频ID
    
    返回:
    - 视频详细信息
    """
    share_url = f"https://www.iesdouyin.com/share/video/{video_id}"
    try:
        processor = DouyinProcessor("")
        video_info = processor.parse_share_url(share_url)
        return json.dumps(video_info, ensure_ascii=False, indent=2)
    except Exception as e:
        return f"获取视频信息失败: {str(e)}"


@mcp.prompt()
def douyin_tool_usage_guide() -> str:
    """抖音工具集使用指南"""
    return """
# 抖音 MCP 工具集使用指南

该服务器提供了一系列工具用于解析、下载及处理抖音内容（视频与图文）。所有工具均返回结构化数据对象。

## 1. 核心工具说明
- `extract_douyin_text`: 从分享链接中提取视频语音并转录为文本（需 API_KEY）。
- `download_video`: 下载抖音无水印视频到指定本地目录（支持 1080p 高解析度）。
- `download_album`: 下载抖音图文作品的所有原图到独立文件夹。
- `get_user_videos`: 通过用户的 sec_uid 批量获取其发布的视频/图文列表。
- `extract_audio`: 从链接或本地视频中提取 MP3 音频。
- `get_douyin_download_link`: 仅获取无水印视频的 1080p 下载 URL。
- `parse_douyin_video_info`: 快速解析内容 ID、标题及类型（视频/图文）。

## 2. 环境配置
- `API_KEY`: 使用文本提取功能时，需在环境变量中配置阿里云百炼 API 密钥。

## 3. 使用场景示例
- **下载视频**：调用 `download_video(share_link="...")`。
- **批量获取用户作品**：先解析视频获取 `sec_uid`，再调用 `get_user_videos(sec_uid="...")`。
- **获取图文原图**：调用 `download_album(share_link="...")`。

## 4. 资源访问
- 使用 `douyin://video/{video_id}` 资源可以直接获取视频的元数据 JSON。
"""


def main():
    """启动 MCP 服务器"""
    # 默认使用 stdio 传输，支持通过环境变量切换到 sse 或 http
    transport = os.getenv("MCP_TRANSPORT", "stdio")
    
    if transport in ["sse", "streamable-http"]:
        import uvicorn
        port = int(os.getenv("MCP_PORT", 8000))
        app = mcp.sse_app if transport == "sse" else mcp.streamable_http_app()
        uvicorn.run(app, host="0.0.0.0", port=port)
    else:
        mcp.run(transport="stdio")


if __name__ == "__main__":
    main()