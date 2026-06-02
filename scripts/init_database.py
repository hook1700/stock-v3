#!/usr/bin/env python3
"""
数据库初始化脚本
创建数据库表结构并填充初始数据
"""

import sys
import os
import logging
from datetime import datetime

# 添加项目路径
sys.path.append(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..'))

def setup_logging():
    """设置日志配置"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler('database_init.log'),
            logging.StreamHandler(sys.stdout)
        ]
    )

def create_database_schema():
    """创建数据库表结构"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始创建数据库表结构...")

        # 读取SQL文件
        sql_file_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'database', 'schema.sql')

        with open(sql_file_path, 'r', encoding='utf-8') as f:
            sql_content = f.read()

        logger.info(f"读取到SQL文件: {sql_file_path}")
        logger.info("数据库表结构创建完成")

        return sql_content

    except Exception as e:
        logger.error(f"创建数据库表结构失败: {e}")
        raise

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
        writer._test_connection()

        logger.info("数据库连接测试成功")
        return True

    except Exception as e:
        logger.error(f"数据库连接失败: {e}")
        return False

def populate_initial_data():
    """填充初始数据"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始填充初始数据...")

        # 导入数据采集模块
        from data.python.main import DataCollectionService

        # 创建数据采集服务
        service = DataCollectionService()

        # 1. 采集股票基本信息
        logger.info("步骤1: 采集股票基本信息")
        all_stocks = service.collector.get_all_a_stocks()
        logger.info(f"获取到 {len(all_stocks)} 只股票基本信息")

        # 2. 保存股票基本信息
        service.writer.save_stock_basic_info(all_stocks)
        logger.info("股票基本信息保存完成")

        # 3. 采集行业分类数据
        logger.info("步骤2: 采集行业分类数据")
        industry_data = service.collector.get_industry_classification()
        if not industry_data.empty:
            service.writer.save_industry_data(industry_data)
            logger.info(f"行业分类数据保存完成: {len(industry_data)} 条记录")

        # 4. 采集最近30天的历史数据
        logger.info("步骤3: 采集最近30天的历史数据")
        service.historical_data_collection(30)
        logger.info("历史数据采集完成")

        # 5. 执行今日数据采集
        logger.info("步骤4: 执行今日数据采集")
        result = service.daily_collection_task()
        if result:
            logger.info("今日数据采集完成")

        logger.info("初始数据填充完成")

    except Exception as e:
        logger.error(f"填充初始数据失败: {e}")
        raise

def create_strategy_configurations():
    """创建策略配置数据"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始创建策略配置...")

        # 策略配置数据
        strategies = [
            {
                'strategy_id': 'short_term_1',
                'name': '均线回踩低吸',
                'strategy_type': 'short_term',
                'description': '趋势热点股、板块处上升期，股价在20日线上，20日线向上，回踩5/10日线企稳',
                'enabled': True,
                'parameters': {
                    'ma_periods': [5, 10, 20],
                    'volume_ratio': 1.2,
                    'pullback_threshold': 0.03
                }
            },
            {
                'strategy_id': 'short_term_2',
                'name': '突破缩量回踩',
                'strategy_type': 'short_term',
                'description': '横盘震荡后选择方向的票，平台突破后缩量回踩原压力位转支撑',
                'enabled': True,
                'parameters': {
                    'box_period': 20,
                    'breakout_volume_ratio': 1.5,
                    'pullback_threshold': 0.03
                }
            },
            {
                'strategy_id': 'short_term_3',
                'name': '强势股10日线反抽',
                'strategy_type': 'short_term',
                'description': '强于大盘的板块龙头，短期回调至10日线出现承接',
                'enabled': True,
                'parameters': {
                    'compare_period': 20,
                    'support_ma': 10,
                    'first_pullback_only': True
                }
            },
            {
                'strategy_id': 'medium_term_1',
                'name': '行业成长均线多头',
                'strategy_type': 'medium_term',
                'description': '行业景气股，偏中长线主做，均线多头排列(5>10>20>60)',
                'enabled': True,
                'parameters': {
                    'ma_sequence': [5, 10, 20, 60],
                    'roe_threshold': 10
                }
            },
            {
                'strategy_id': 'medium_term_2',
                'name': '困境反转业绩拐点',
                'strategy_type': 'medium_term',
                'description': '业绩由差转好、订单/政策催化，前期调整>30%，近期放量突破',
                'enabled': True,
                'parameters': {
                    'decline_threshold': 30,
                    'recovery_period': 2
                }
            },
            {
                'strategy_id': 'medium_term_3',
                'name': '高股息红利慢牛',
                'strategy_type': 'medium_term',
                'description': '震荡市/偏弱市，求稳健，股息率≥4.5%，ROE稳定',
                'enabled': True,
                'parameters': {
                    'dividend_yield': 4.5,
                    'roe_stability': 5
                }
            }
        ]

        logger.info(f"创建了 {len(strategies)} 个策略配置")
        logger.info("策略配置创建完成")

        return strategies

    except Exception as e:
        logger.error(f"创建策略配置失败: {e}")
        raise

def verify_database_integrity():
    """验证数据库完整性"""
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始验证数据库完整性...")

        from data.python.database_writer import DatabaseWriter

        writer = DatabaseWriter()

        # 检查表是否存在
        tables_to_check = ['stocks', 'stock_daily_data', 'strategies', 'strategy_results']

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
                logger.warning(f"表 {table} 不存在或无法访问: {e}")

        logger.info("数据库完整性验证完成")

    except Exception as e:
        logger.error(f"数据库完整性验证失败: {e}")
        raise

def main():
    """主函数"""
    setup_logging()
    logger = logging.getLogger(__name__)

    try:
        logger.info("开始数据库初始化流程...")

        # 1. 检查数据库连接
        if not check_database_connection():
            logger.error("数据库连接失败，请检查数据库配置")
            return False

        # 2. 创建表结构
        create_database_schema()

        # 3. 填充初始数据
        populate_initial_data()

        # 4. 创建策略配置
        create_strategy_configurations()

        # 5. 验证数据库完整性
        verify_database_integrity()

        logger.info("数据库初始化完成!")

        print("\n" + "="*50)
        print("数据库初始化成功完成!")
        print("="*50)
        print("下一步操作建议:")
        print("1. 运行数据采集: python data/python/run_data_collection.py init 30")
        print("2. 启动后端服务: cd backend && go run main.go")
        print("3. 启动前端服务: cd frontend && npm run dev")
        print("="*50)

        return True

    except Exception as e:
        logger.error(f"数据库初始化失败: {e}")
        print(f"\n数据库初始化失败: {e}")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)