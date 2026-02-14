from pydantic import BaseModel, Field
from typing import Optional, List

class VideoDownloadInfo(BaseModel):
    status: str = Field(..., description="状态 (success/error)")
    video_id: Optional[str] = Field(None, description="抖音视频 ID")
    title: Optional[str] = Field(None, description="视频标题")
    download_url: Optional[str] = Field(None, description="无水印下载链接")
    description: Optional[str] = Field(None, description="详细描述")
    usage_tip: Optional[str] = Field(None, description="使用说明")
    error: Optional[str] = Field(None, description="错误信息")

class DownloadResult(BaseModel):
    status: str = Field(..., description="状态")
    message: str = Field(..., description="提示消息")
    video_id: Optional[str] = Field(None, description="视频 ID")
    title: Optional[str] = Field(None, description="标题")
    path: Optional[str] = Field(None, description="本地保存物理路径")
    error: Optional[str] = Field(None, description="错误详情")

class AudioResult(BaseModel):
    status: str = Field(..., description="状态")
    message: str = Field(..., description="提示消息")
    title: Optional[str] = Field(None, description="标题")
    path: Optional[str] = Field(None, description="本地音频物理路径")
    error: Optional[str] = Field(None, description="错误详情")

class VideoBasicInfo(BaseModel):
    status: str = Field(..., description="状态")
    video_id: Optional[str] = Field(None, description="视频 ID")
    title: Optional[str] = Field(None, description="标题")
    download_url: Optional[str] = Field(None, description="下载链接")
    error: Optional[str] = Field(None, description="错误详情")

class TextExtractionResult(BaseModel):
    status: str = Field(..., description="状态")
    text: Optional[str] = Field(None, description="提取的文本内容")
    title: Optional[str] = Field(None, description="视频标题")
    error: Optional[str] = Field(None, description="错误详情")

class AlbumImage(BaseModel):
    url: str = Field(..., description="图片下载链接")
    uri: Optional[str] = Field(None, description="图片 ID/URI")

class AlbumDownloadResult(BaseModel):
    status: str = Field(..., description="状态")
    message: str = Field(..., description="提示消息")
    video_id: Optional[str] = Field(None, description="作品 ID")
    title: Optional[str] = Field(None, description="标题")
    image_paths: List[str] = Field(default_factory=list, description="本地图片物理路径列表")
    error: Optional[str] = Field(None, description="错误详情")

class AwemeItem(BaseModel):
    aweme_id: str = Field(..., description="作品 ID")
    desc: str = Field(..., description="描述/标题")
    type: str = Field(..., description="类型 (video/image)")
    author: str = Field(..., description="作者昵称")
    create_time: str = Field(..., description="创建时间")
    url: Optional[str] = Field(None, description="下载链接或预览图链接")

class UserVideoList(BaseModel):
    status: str = Field(..., description="状态")
    user_id: Optional[str] = Field(None, description="用户 ID")
    nickname: Optional[str] = Field(None, description="用户昵称")
    aweme_list: List[AwemeItem] = Field(default_factory=list, description="作品列表")
    has_more: bool = Field(False, description="是否还有更多")
    cursor: int = Field(0, description="下一页游标")
    error: Optional[str] = Field(None, description="错误详情")
