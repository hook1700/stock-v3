#!/usr/bin/env python3
"""
股票数据采集主程序
每日定时获取股票数据并存储到数据库
"""

import logging
import sys
import os
import schedule
import time
from datetime import datetime, timedelta

# 添加项目路径
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from stock_data_collector import StockDataCollector, collect_daily_data
from database_writer import DatabaseWriter

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('data_collection.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class DataCollectionService:
    """数据采集服务"""

    def __init__(self):
        self.collector = StockDataCollector()
        self.writer = DatabaseWriter()

    def daily_collection_task(self):
        """每日数据采集任务"""
        logger.info("开始执行每日数据采集任务")

        try:
            # 1. 执行数据采集
            result = collect_daily_data()

            if not result:
                logger.error("数据采集失败")
                return

            # 2. 写入数据库
            self.writer.save_stock_basic_info(result.get('all_stocks', []))
            self.writer.save_liquid_stocks(result.get('liquid_stocks', []))
            self.writer.save_industry_data(result.get('industry_data', []))
            self.writer.save_fund_flow_data(result.get('fund_flow_data', []))

            logger.info("每日数据采集任务完成")

        except Exception as e:
            logger.error(f"每日数据采集任务失败: {e}")

    def historical_data_collection(self, days=365):
        """历史数据采集（用于初始化）"""
        logger.info(f"开始采集最近{days}天的历史数据")

        try:
            # 获取所有股票代码
            all_stocks = self.collector.get_all_a_stocks()
            logger.info(f"获取到 {len(all_stocks)} 只股票")

            # 计算日期范围
            end_date = datetime.now()
            start_date = end_date - timedelta(days=days)

            # 分批采集历史数据
            batch_size = 100
            for i in range(0, len(all_stocks), batch_size):
                batch = all_stocks[i:i + batch_size]
                self._collect_batch_historical_data(batch, start_date, end_date)
                logger.info(f"已完成批次 {i//batch_size + 1}/{(len(all_stocks)-1)//batch_size + 1}")

            logger.info("历史数据采集完成")

        except Exception as e:
            logger.error(f"历史数据采集失败: {e}")

    def _collect_batch_historical_data(self, stock_codes, start_date, end_date):
        """批量采集历史数据"""
        for stock_code in stock_codes:
            try:
                # 获取股票历史数据
                daily_data = self.collector.get_stock_daily_data(
                    stock_code,
                    start_date.strftime('%Y-%m-%d'),
                    end_date.strftime('%Y-%m-%d')
                )

                if not daily_data.empty:
                    # 保存到数据库
                    self.writer.save_stock_daily_data(stock_code, daily_data)

            except Exception as e:
                logger.error(f"采集股票 {stock_code} 历史数据失败: {e}")

    def start_scheduler(self):
        """启动定时任务调度器"""
        # 每日17:45执行数据采集
        schedule.every().day.at("17:45").do(self.daily_collection_task)

        # 周末跳过
        if datetime.now().weekday() >= 5:  # 5=周六, 6=周日
            logger.info("周末跳过数据采集")
            return

        logger.info("数据采集调度器已启动，每日17:45执行")

        # 立即执行一次
        self.daily_collection_task()

        # 保持程序运行
        while True:
            schedule.run_pending()
            time.sleep(60)  # 每分钟检查一次

def main():
    """主函数"""
    service = DataCollectionService()

    # 解析命令行参数
    if len(sys.argv) > 1:
        if sys.argv[1] == "--init":
            # 初始化模式：采集历史数据
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 365
            service.historical_data_collection(days)
        elif sys.argv[1] == "--once":
            # 单次执行模式
            service.daily_collection_task()
        elif sys.argv[1] == "--schedule":
            # 定时任务模式
            service.start_scheduler()
        else:
            print("用法: python main.py [--init [days] | --once | --schedule]")
            print("  --init [days]   初始化并采集历史数据（默认365天）")
            print("  --once          单次执行数据采集")
            print("  --schedule      启动定时任务调度器")
    else:
        # 默认单次执行
        service.daily_collection_task()

if __name__ == "__main__":
    main()