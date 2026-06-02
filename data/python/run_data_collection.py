#!/usr/bin/env python3
"""
数据采集启动脚本
用于首次启动和日常数据采集
"""

import sys
import os
import logging
from datetime import datetime, timedelta

# 添加项目路径
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from main import DataCollectionService
from stock_data_collector import StockDataCollector

def setup_logging():
    """设置日志配置"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler('data_collection.log'),
            logging.StreamHandler(sys.stdout)
        ]
    )

def initialize_database():
    """初始化数据库，创建必要的数据"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始初始化数据库...")

        # 创建数据采集服务
        service = DataCollectionService()

        # 1. 获取所有A股股票基本信息
        logger.info("步骤1: 获取A股股票基本信息")
        all_stocks = service.collector.get_all_a_stocks()
        logger.info(f"获取到 {len(all_stocks)} 只A股股票")

        # 2. 保存股票基本信息到数据库
        logger.info("步骤2: 保存股票基本信息到数据库")
        service.writer.save_stock_basic_info(all_stocks)

        # 3. 采集最近30天的历史数据
        logger.info("步骤3: 采集最近30天的历史数据")
        service.historical_data_collection(30)

        # 4. 执行今日数据采集
        logger.info("步骤4: 执行今日数据采集")
        service.daily_collection_task()

        logger.info("数据库初始化完成")

    except Exception as e:
        logger.error(f"数据库初始化失败: {e}")
        raise

def run_daily_collection():
    """执行每日数据采集"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始执行每日数据采集...")

        service = DataCollectionService()

        # 执行数据采集
        result = service.daily_collection_task()

        if result:
            logger.info(f"每日数据采集完成，处理了 {len(result.get('liquid_stocks', []))} 只流动性股票")
        else:
            logger.warning("数据采集返回空结果")

    except Exception as e:
        logger.error(f"每日数据采集失败: {e}")
        raise

def run_historical_collection(days=365):
    """执行历史数据采集"""
    logger = logging.getLogger(__name__)

    try:
        logger.info(f"开始采集最近{days}天的历史数据...")

        service = DataCollectionService()
        service.historical_data_collection(days)

        logger.info("历史数据采集完成")

    except Exception as e:
        logger.error(f"历史数据采集失败: {e}")
        raise

def check_data_status():
    """检查数据状态"""
    logger = logging.getLogger(__name__)

    try:
        collector = StockDataCollector()

        # 检查股票基本信息
        all_stocks = collector.get_all_a_stocks()
        logger.info(f"数据源可获取股票数量: {len(all_stocks)}")

        # 检查数据源连接
        if collector.bs_connected:
            logger.info("BaoStock连接正常")
        else:
            logger.warning("BaoStock连接异常")

        # 检查数据采集能力
        if all_stocks:
            # 测试获取一只股票的日线数据
            test_stock = all_stocks[0]
            today = datetime.now().strftime('%Y%m%d')

            daily_data = collector.get_stock_daily_data(
                test_stock['code'],
                today,
                today
            )

            if not daily_data.empty:
                logger.info(f"数据采集测试成功，获取到股票 {test_stock['code']} 的日线数据")
            else:
                logger.warning(f"数据采集测试失败，无法获取股票 {test_stock['code']} 的日线数据")

        return True

    except Exception as e:
        logger.error(f"数据状态检查失败: {e}")
        return False

def main():
    """主函数"""
    setup_logging()
    logger = logging.getLogger(__name__)

    if len(sys.argv) < 2:
        print("""
数据采集系统使用说明:

用法: python run_data_collection.py [命令] [参数]

命令列表:
  init [days]     初始化数据库并采集历史数据（默认30天）
  daily           执行每日数据采集
  history [days]  采集历史数据（默认365天）
  improved_history [days]  使用改进采集器采集历史数据（默认250天）
  improved_daily   使用改进采集器执行每日数据采集
  status          检查数据状态
  test            测试数据采集功能

示例:
  python run_data_collection.py init 30    # 初始化并采集30天历史数据
  python run_data_collection.py daily     # 执行每日数据采集
  python run_data_collection.py improved_history 250  # 使用改进采集器采集250天历史数据
  python run_data_collection.py status     # 检查数据状态
""")
        return

    command = sys.argv[1]

    try:
        if command == "init":
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 30
            initialize_database()

        elif command == "daily":
            run_daily_collection()

        elif command == "history":
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 365
            run_historical_collection(days)

        elif command == "improved_history":
            # 使用改进采集器采集历史数据
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 250
            from improved_data_collector import collect_historical_data_improved
            logger.info(f"使用改进采集器采集最近{days}天历史数据...")
            collect_historical_data_improved(days=days, save_to_db=True)
            logger.info("改进采集器历史数据采集完成")

        elif command == "improved_daily":
            # 使用改进采集器执行每日采集
            from improved_data_collector import collect_daily_data_improved
            logger.info("使用改进采集器执行每日数据采集...")
            result = collect_daily_data_improved(save_to_db=True)
            logger.info(f"改进采集器每日数据采集完成，保存了 {result.get('saved_count', 0)} 只股票数据")

        elif command == "status":
            check_data_status()

        elif command == "test":
            # 测试数据采集功能
            logger.info("开始测试数据采集功能...")

            from improved_data_collector import ImprovedDataCollector
            collector = ImprovedDataCollector()

            # 测试获取股票列表
            stocks = collector.get_all_a_stocks()
            logger.info(f"测试结果: 获取到 {len(stocks)} 只股票")

            if stocks:
                # 测试获取日线数据
                test_stock = stocks[0]
                today = datetime.now().strftime('%Y%m%d')

                daily_data = collector.get_stock_daily_data(
                    test_stock['code'],
                    today,
                    today
                )

                if not daily_data.empty:
                    logger.info(f"日线数据测试成功: {len(daily_data)} 条记录")
                else:
                    logger.warning("日线数据测试失败")

            collector.close()
            logger.info("数据采集测试完成")

        else:
            logger.error(f"未知命令: {command}")
            print("请使用 'init', 'daily', 'history', 'improved_history', 'improved_daily', 'status' 或 'test' 命令")

    except Exception as e:
        logger.error(f"执行命令失败: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()