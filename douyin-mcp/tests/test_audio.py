import asyncio
import sys
from pathlib import Path

# 添加父目录到路径（server.py 所在位置）
sys.path.append(str(Path(__file__).parent.parent))

from server import extract_audio

async def test():
    # 测试方式 1: 通过分享链接提取
    share_link = "https://v.douyin.com/1T1sZtZmqsM/"
    print(f"开始测试从链接提取音频: {share_link}")
    
    class MockContext:
        def info(self, msg): print(f"[INFO] {msg}")
        def error(self, msg): print(f"[ERROR] {msg}")
        async def report_progress(self, current, total): pass

    ctx = MockContext()
    
    try:
        result = await extract_audio(share_link=share_link, ctx=ctx)
        print("\n链接提取测试结果:")
        print(result)
    except Exception as e:
        print(f"\n链接提取测试失败: {e}")

    # 测试方式 2: 通过本地路径提取
    # 查找刚才下载的视频
    downloads_dir = Path("downloads")
    videos = list(downloads_dir.glob("*.mp4"))
    if videos:
        local_video = videos[0]
        print(f"\n开始测试从本地文件提取音频: {local_video}")
        try:
            result = await extract_audio(video_path=str(local_video), ctx=ctx)
            print("\n本地文件提取测试结果:")
            print(result)
        except Exception as e:
            print(f"\n本地文件提取测试失败: {e}")
    else:
        print("\n未找到本地视频文件，跳过本地提取测试。")

if __name__ == "__main__":
    asyncio.run(test())
