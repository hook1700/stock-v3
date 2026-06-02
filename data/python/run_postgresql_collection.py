#!/usr/bin/env python3
"""
PostgreSQL数据采集启动脚本 - 专为stock-v3项目设计
将采集的数据直接写入PostgreSQL数据库，供Golang后端使用
"""

import sys
import os
import logging
from datetime import datetime, timedelta

# 添加项目路径
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from postgresql_data_collector import PostgreSQLDataCollector, collect_daily_data_to_postgresql, test_postgresql_collector

def setup_logging():
    """设置日志配置"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler('postgresql_data_collection.log'),
            logging.StreamHandler(sys.stdout)
        ]
    )

def initialize_postgresql_database():
    """初始化PostgreSQL数据库，创建必要的数据"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始初始化PostgreSQL数据库...")

        collector = PostgreSQLDataCollector()

        # 1. 获取所有A股股票基本信息
        logger.info("步骤1: 获取A股股票基本信息")
        all_stocks = collector.get_all_a_stocks()
        logger.info(f"获取到 {len(all_stocks)} 只A股股票")

        if not all_stocks:
            logger.error("无法获取股票列表，任务终止")
            return False

        # 2. 保存股票基本信息到PostgreSQL
        logger.info("步骤2: 保存股票基本信息到PostgreSQL")
        success = collector.save_stocks_to_postgresql(all_stocks)
        if not success:
            logger.error("保存股票基本信息失败")
            return False

        # 3. 采集最近30天的历史数据
        logger.info("步骤3: 采集最近30天的历史数据")
        collector.collect_and_save_historical_data(days=30)

        collector.close()
        logger.info("PostgreSQL数据库初始化完成")
        return True

    except Exception as e:
        logger.error(f"PostgreSQL数据库初始化失败: {e}")
        return False

def run_daily_postgresql_collection():
    """执行每日PostgreSQL数据采集"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始执行每日PostgreSQL数据采集...")

        result = collect_daily_data_to_postgresql()

        logger.info("每日PostgreSQL数据采集完成")

    except Exception as e:
        logger.error(f"每日PostgreSQL数据采集失败: {e}")
        raise

def run_historical_postgresql_collection(days: int = 365):
    """执行历史数据采集到PostgreSQL"""
    logger = logging.getLogger(__name__)

    try:
        logger.info(f"开始采集最近{days}天的历史数据到PostgreSQL...")

        collector = PostgreSQLDataCollector()
        collector.collect_and_save_historical_data(days)
        collector.close()

        logger.info("历史数据采集到PostgreSQL完成")

    except Exception as e:
        logger.error(f"历史数据采集到PostgreSQL失败: {e}")
        raise

def check_postgresql_connection():
    """检查PostgreSQL连接状态"""
    logger = logging.getLogger(__name__)

    try:
        collector = PostgreSQLDataCollector()

        # 检查PostgreSQL连接
        if collector.pg_conn:
            cursor = collector.pg_conn.cursor()
            cursor.execute("SELECT version();")
            version = cursor.fetchone()[0]
            logger.info(f"PostgreSQL连接正常，版本: {version}")
        else:
            logger.error("PostgreSQL连接失败")

        # 检查BaoStock连接
        if collector.bs_connected:
            logger.info("BaoStock连接正常")
        else:
            logger.warning("BaoStock连接异常")

        collector.close()
        return True

    except Exception as e:
        logger.error(f"PostgreSQL连接检查失败: {e}")
        return False

def main():
    """主函数"""
    setup_logging()
    logger = logging.getLogger(__name__)

    if len(sys.argv) < 2:
        print("""
PostgreSQL数据采集系统使用说明 - stock-v3项目专用

用法: python run_postgresql_collection.py [命令] [参数]

命令列表:
  init [days]     初始化PostgreSQL数据库并采集历史数据（默认30天）
  daily           执行每日数据采集到PostgreSQL
  history [days]  采集历史数据到PostgreSQL（默认365天）
  status          检查PostgreSQL连接状态
  test            测试PostgreSQL数据采集功能

示例:
  python run_postgresql_collection.py init 30    # 初始化并采集30天历史数据
  python run_postgresql_collection.py daily     # 执行每日数据采集
  python run_postgresql_collection.py status     # 检查连接状态
""")
        return

    command = sys.argv[1]

    try:
        if command == "init":
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 30
            initialize_postgresql_database()

        elif command == "daily":
            run_daily_postgresql_collection()

        elif command == "history":
            days = int(sys.argv[2]) if len(sys.argv) > 2 else 365
            run_historical_postgresql_collection(days)

        elif command == "status":
            check_postgresql_connection()

        elif command == "test":
            test_postgresql_collector()

        else:
            logger.error(f"未知命令: {command}")
            print("请使用 'init', 'daily', 'history', 'status' 或 'test' 命令")

    except Exception as e:
        logger.error(f"执行命令失败: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()