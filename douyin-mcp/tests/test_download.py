import asyncio
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).parent.parent))

from server import download_video
from scripts.logger import logger

async def test():
    share_link = "https://v.douyin.com/MuqYMtekyQk/"
    logger.info(f"开始测试下载: {share_link}")

    class MockContext:
        async def info(self, msg): logger.info(f"[INFO] {msg}")
        async def error(self, msg): logger.error(f"[ERROR] {msg}")
        async def report_progress(self, current, total):
            percent = (current / total) * 100
            logger.info(f"[PROGRESS] {percent:.2f}% ({current}/{total})")

    ctx = MockContext()

    logger.info("--- 测试 download_video ---")
    try:
        result = await download_video(share_link=share_link, ctx=ctx)
        logger.info("测试结果:")
        logger.info(result)
    except Exception as e:
        logger.error(f"测试执行失败: {e}")

    from server import download_album
    logger.info("--- 测试 download_album ---")
    try:
        result = await download_album(share_link=share_link, ctx=ctx)
        logger.info("测试结果:")
        logger.info(result)
    except Exception as e:
        logger.error(f"测试执行失败: {e}")

if __name__ == "__main__":
    asyncio.run(test())