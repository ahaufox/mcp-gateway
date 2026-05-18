import asyncio
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).parent.parent))

from server import extract_audio
from scripts.logger import logger

async def test():
    share_link = "https://v.douyin.com/1T1sZtZmqsM/"
    logger.info(f"开始测试从链接提取音频: {share_link}")

    class MockContext:
        def info(self, msg): logger.info(f"[INFO] {msg}")
        def error(self, msg): logger.error(f"[ERROR] {msg}")
        async def report_progress(self, current, total): pass

    ctx = MockContext()

    try:
        result = await extract_audio(share_link=share_link, ctx=ctx)
        logger.info("链接提取测试结果:")
        logger.info(result)
    except Exception as e:
        logger.error(f"链接提取测试失败: {e}")

    downloads_dir = Path("downloads")
    videos = list(downloads_dir.glob("*.mp4"))
    if videos:
        local_video = videos[0]
        logger.info(f"开始测试从本地文件提取音频: {local_video}")
        try:
            result = await extract_audio(video_path=str(local_video), ctx=ctx)
            logger.info("本地文件提取测试结果:")
            logger.info(result)
        except Exception as e:
            logger.error(f"本地文件提取测试失败: {e}")
    else:
        logger.info("未找到本地视频文件，跳过本地提取测试。")

if __name__ == "__main__":
    asyncio.run(test())