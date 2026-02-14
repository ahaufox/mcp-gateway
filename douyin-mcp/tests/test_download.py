import asyncio
import os
import sys
from pathlib import Path

# 添加当前目录到路径
# 添加父目录到路径（server.py 所在位置）
sys.path.append(str(Path(__file__).parent.parent))

from server import download_video

async def test():
    # https://v.douyin.com/MuqYMtekyQk/ 非视频
    # share_link = "https://v.douyin.com/1T1sZtZmqsM/"
    share_link = "https://v.douyin.com/MuqYMtekyQk/"
    print(f"开始测试下载: {share_link}")
    
    # 模拟 Context 对象
    class MockContext:
        async def info(self, msg): print(f"[INFO] {msg}")
        async def error(self, msg): print(f"[ERROR] {msg}")
        async def report_progress(self, current, total):
            percent = (current / total) * 100
            print(f"[PROGRESS] {percent:.2f}% ({current}/{total})", end='\r')

    ctx = MockContext()
    
    # 1. 测试视频下载（如果链接是图文，应该报错）
    print("\n--- 测试 download_video ---")
    try:
        result = await download_video(share_link=share_link, ctx=ctx)
        print("测试结果:")
        print(result)
    except Exception as e:
        print(f"测试执行失败: {e}")

    # 2. 如果是图文，测试 download_album
    from server import download_album
    print("\n--- 测试 download_album ---")
    try:
        result = await download_album(share_link=share_link, ctx=ctx)
        print("测试结果:")
        print(result)
    except Exception as e:
        print(f"测试执行失败: {e}")

if __name__ == "__main__":
    asyncio.run(test())
