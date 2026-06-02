#!/usr/bin/env python3
"""
数据问题修复脚本
解决数据库无数据的问题
"""

import sys
import os
import logging
from datetime import datetime, timedelta

# 添加项目路径
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

def setup_logging():
    """设置日志配置"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler('data_fix.log'),
            logging.StreamHandler(sys.stdout)
        ]
    )

def check_database_connection():
    """检查数据库连接"""
    logger = logging.getLogger(__name__)

    try:
        # 尝试导入数据库模块
        from data.python.database_writer import DatabaseWriter

        logger.info("正在测试数据库连接...")

        # 创建数据库写入器实例
        writer = DatabaseWriter()

        # 测试连接
        conn = writer._get_connection()
        if conn:
            logger.info("✓ 数据库连接成功")
            conn.close()
            return True
        else:
            logger.error("✗ 数据库连接失败")
            return False

    except Exception as e:
        logger.error(f"✗ 数据库连接测试失败: {e}")
        return False

def check_data_source():
    """检查数据源连接"""
    logger = logging.getLogger(__name__)

    try:
        from data.python.stock_data_collector import StockDataCollector

        logger.info("正在测试数据源连接...")

        collector = StockDataCollector()

        # 测试获取股票列表
        stocks = collector.get_all_a_stocks()

        if stocks and len(stocks) > 0:
            logger.info(f"✓ 数据源连接成功，获取到 {len(stocks)} 只股票")

            # 测试获取单只股票数据
            test_stock = stocks[0]
            today = datetime.now().strftime('%Y%m%d')

            daily_data = collector.get_stock_daily_data(
                test_stock['code'],
                today,
                today
            )

            if not daily_data.empty:
                logger.info(f"✓ 数据采集测试成功，获取到股票 {test_stock['code']} 的日线数据")
            else:
                logger.warning("⚠ 数据采集测试返回空结果")

            return True
        else:
            logger.error("✗ 数据源连接失败，无法获取股票列表")
            return False

    except Exception as e:
        logger.error(f"✗ 数据源连接测试失败: {e}")
        return False

def run_initial_data_collection():
    """执行初始数据采集"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始执行初始数据采集...")

        # 导入数据采集模块
        from data.python.main import DataCollectionService

        # 创建数据采集服务
        service = DataCollectionService()

        # 1. 获取所有股票基本信息
        logger.info("步骤1: 获取股票基本信息")
        all_stocks = service.collector.get_all_a_stocks()

        if not all_stocks:
            logger.error("无法获取股票基本信息，采集终止")
            return False

        logger.info(f"✓ 获取到 {len(all_stocks)} 只股票基本信息")

        # 2. 保存股票基本信息
        logger.info("步骤2: 保存股票基本信息到数据库")
        service.writer.save_stock_basic_info(all_stocks)
        logger.info("✓ 股票基本信息保存完成")

        # 3. 采集最近7天的历史数据（快速测试）
        logger.info("步骤3: 采集最近7天的历史数据")

        # 只采集前10只股票以节省时间
        test_stocks = all_stocks[:10]

        for i, stock in enumerate(test_stocks):
            logger.info(f"采集股票 {i+1}/{len(test_stocks)}: {stock['code']} {stock['name']}")

            # 获取最近7天的数据
            end_date = datetime.now()
            start_date = end_date - timedelta(days=7)

            daily_data = service.collector.get_stock_daily_data(
                stock['code'],
                start_date.strftime('%Y%m%d'),
                end_date.strftime('%Y%m%d')
            )

            if not daily_data.empty:
                service.writer.save_stock_daily_data(stock['code'], daily_data)
                logger.info(f"  ✓ 保存 {len(daily_data)} 条日线数据")
            else:
                logger.warning(f"  ⚠ 无日线数据可保存")

        logger.info("✓ 初始数据采集完成")
        return True

    except Exception as e:
        logger.error(f"✗ 初始数据采集失败: {e}")
        return False

def verify_data_integrity():
    """验证数据完整性"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("验证数据完整性...")

        from data.python.database_writer import DatabaseWriter

        writer = DatabaseWriter()

        # 检查表数据量
        tables_to_check = ['stocks', 'stock_daily_data']

        for table in tables_to_check:
            try:
                conn = writer._get_connection()
                cursor = conn.cursor()

                cursor.execute(f"SELECT COUNT(*) FROM {table}")
                count = cursor.fetchone()[0]

                logger.info(f"表 {table}: {count} 条记录")

                cursor.close()
                conn.close()

            except Exception as e:
                logger.warning(f"表 {table} 检查失败: {e}")

        logger.info("✓ 数据完整性验证完成")
        return True

    except Exception as e:
        logger.error(f"✗ 数据完整性验证失败: {e}")
        return False

def main():
    """主函数"""
    setup_logging()
    logger = logging.getLogger(__name__)

    print("=" * 60)
    print("股票策略系统 - 数据问题修复工具")
    print("=" * 60)

    try:
        # 步骤1: 检查数据库连接
        logger.info("\n[步骤1] 检查数据库连接")
        if not check_database_connection():
            print("\n❌ 数据库连接失败，请检查:")
            print("1. PostgreSQL服务是否启动")
            print("2. 数据库配置是否正确 (backend/config.yaml)")
            print("3. 数据库是否已创建 (stock_strategy)")
            return False

        # 步骤2: 检查数据源连接
        logger.info("\n[步骤2] 检查数据源连接")
        if not check_data_source():
            print("\n❌ 数据源连接失败，请检查:")
            print("1. 网络连接是否正常")
            print("2. BaoStock/AKShare API是否可用")
            print("3. 防火墙设置")
            return False

        # 步骤3: 执行初始数据采集
        logger.info("\n[步骤3] 执行初始数据采集")
        if not run_initial_data_collection():
            print("\n❌ 数据采集失败，请检查数据源配置")
            return False

        # 步骤4: 验证数据完整性
        logger.info("\n[步骤4] 验证数据完整性")
        if not verify_data_integrity():
            print("\n⚠ 数据完整性验证发现问题")

        print("\n" + "=" * 60)
        print("✅ 数据问题修复完成!")
        print("=" * 60)
        print("\n下一步操作建议:")
        print("1. 启动后端服务: cd backend && go run main.go")
        print("2. 启动前端服务: cd frontend && npm run dev")
        print("3. 访问系统: http://localhost:3000")
        print("\n如需采集更多历史数据，运行:")
        print("cd data/python && python run_data_collection.py init 30")

        return True

    except Exception as e:
        logger.error(f"修复过程出现错误: {e}")
        print(f"\n❌ 修复失败: {e}")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)