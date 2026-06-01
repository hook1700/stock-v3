"""
数据库写入模块
将采集的股票数据写入PostgreSQL数据库
"""

import logging
import psycopg2
from psycopg2.extras import execute_values
import pandas as pd
from datetime import datetime
from typing import List, Dict, Any

logger = logging.getLogger(__name__)

class DatabaseWriter:
    """数据库写入器"""

    def __init__(self, host='localhost', port=5432, user='postgres',
                 password='password', dbname='stock_strategy'):
        self.connection_params = {
            'host': host,
            'port': port,
            'user': user,
            'password': password,
            'dbname': dbname
        }
        self._test_connection()

    def _test_connection(self):
        """测试数据库连接"""
        try:
            conn = psycopg2.connect(**self.connection_params)
            conn.close()
            logger.info("数据库连接测试成功")
        except Exception as e:
            logger.error(f"数据库连接失败: {e}")
            raise

    def _get_connection(self):
        """获取数据库连接"""
        return psycopg2.connect(**self.connection_params)

    def save_stock_basic_info(self, stocks: List[Dict[str, Any]]):
        """保存股票基本信息"""
        if not stocks:
            logger.warning("股票基本信息为空，跳过保存")
            return

        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            # 批量插入或更新股票基本信息
            sql = """
            INSERT INTO stocks (code, name, industry, market, listing_date, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (code) DO UPDATE SET
                name = EXCLUDED.name,
                industry = EXCLUDED.industry,
                market = EXCLUDED.market,
                listing_date = EXCLUDED.listing_date,
                updated_at = EXCLUDED.updated_at
            """

            data = []
            current_time = datetime.now()
            for stock in stocks:
                data.append((
                    stock.get('code'),
                    stock.get('name'),
                    stock.get('industry'),
                    stock.get('market'),
                    stock.get('listing_date'),
                    current_time,
                    current_time
                ))

            execute_values(cursor, sql, data)
            conn.commit()
            logger.info(f"成功保存 {len(stocks)} 条股票基本信息")

        except Exception as e:
            logger.error(f"保存股票基本信息失败: {e}")
            if conn:
                conn.rollback()
        finally:
            if cursor:
                cursor.close()
            if conn:
                conn.close()

    def save_stock_daily_data(self, stock_code: str, daily_data: pd.DataFrame):
        """保存股票日线数据"""
        if daily_data.empty:
            logger.warning(f"股票 {stock_code} 日线数据为空，跳过保存")
            return

        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            sql = """
            INSERT INTO stock_daily_data
            (stock_code, trade_date, open_price, high_price, low_price, close_price,
             volume, amount, turnover_rate, pe_ratio, pb_ratio, created_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (stock_code, trade_date) DO UPDATE SET
                open_price = EXCLUDED.open_price,
                high_price = EXCLUDED.high_price,
                low_price = EXCLUDED.low_price,
                close_price = EXCLUDED.close_price,
                volume = EXCLUDED.volume,
                amount = EXCLUDED.amount,
                turnover_rate = EXCLUDED.turnover_rate,
                pe_ratio = EXCLUDED.pe_ratio,
                pb_ratio = EXCLUDED.pb_ratio
            """

            data = []
            current_time = datetime.now()

            for _, row in daily_data.iterrows():
                data.append((
                    stock_code,
                    row.get('date'),
                    row.get('open', 0),
                    row.get('high', 0),
                    row.get('low', 0),
                    row.get('close', 0),
                    row.get('volume', 0),
                    row.get('amount', 0),
                    row.get('turnover_rate', 0),
                    row.get('pe_ratio', 0),
                    row.get('pb_ratio', 0),
                    current_time
                ))

            execute_values(cursor, sql, data)
            conn.commit()
            logger.info(f"成功保存股票 {stock_code} 的 {len(daily_data)} 条日线数据")

        except Exception as e:
            logger.error(f"保存股票 {stock_code} 日线数据失败: {e}")
            if conn:
                conn.rollback()
        finally:
            if cursor:
                cursor.close()
            if conn:
                conn.close()

    def save_liquid_stocks(self, liquid_stocks: List[str]):
        """保存流动性股票列表（标记为高流动性）"""
        if not liquid_stocks:
            logger.warning("流动性股票列表为空，跳过保存")
            return

        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            # 创建流动性标记表（如果不存在）
            cursor.execute("""
                CREATE TABLE IF NOT EXISTS liquid_stocks (
                    id SERIAL PRIMARY KEY,
                    stock_code VARCHAR(20) UNIQUE NOT NULL,
                    is_liquid BOOLEAN DEFAULT TRUE,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """)

            # 批量插入流动性股票
            sql = """
            INSERT INTO liquid_stocks (stock_code, is_liquid)
            VALUES (%s, TRUE)
            ON CONFLICT (stock_code) DO UPDATE SET
                is_liquid = EXCLUDED.is_liquid
            """

            data = [(code,) for code in liquid_stocks]
            execute_values(cursor, sql, data)
            conn.commit()
            logger.info(f"成功标记 {len(liquid_stocks)} 只流动性股票")

        except Exception as e:
            logger.error(f"保存流动性股票列表失败: {e}")
            if conn:
                conn.rollback()
        finally:
            if cursor:
                cursor.close()
            if conn:
                conn.close()

    def save_industry_data(self, industry_data: pd.DataFrame):
        """保存行业分类数据"""
        if industry_data.empty:
            logger.warning("行业数据为空，跳过保存")
            return

        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            # 创建行业分类表（如果不存在）
            cursor.execute("""
                CREATE TABLE IF NOT EXISTS industry_classification (
                    id SERIAL PRIMARY KEY,
                    stock_code VARCHAR(20) NOT NULL,
                    industry_name VARCHAR(100),
                    industry_level INTEGER,
                    update_date DATE,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    UNIQUE(stock_code, industry_level)
                )
            """)

            sql = """
            INSERT INTO industry_classification (stock_code, industry_name, industry_level, update_date)
            VALUES (%s, %s, %s, %s)
            ON CONFLICT (stock_code, industry_level) DO UPDATE SET
                industry_name = EXCLUDED.industry_name,
                update_date = EXCLUDED.update_date
            """

            data = []
            current_date = datetime.now().date()

            for _, row in industry_data.iterrows():
                data.append((
                    row.get('code'),
                    row.get('industry_name'),
                    1,  # 一级行业分类
                    current_date
                ))

            execute_values(cursor, sql, data)
            conn.commit()
            logger.info(f"成功保存 {len(industry_data)} 条行业分类数据")

        except Exception as e:
            logger.error(f"保存行业分类数据失败: {e}")
            if conn:
                conn.rollback()
        finally:
            if cursor:
                cursor.close()
            if conn:
                conn.close()

    def save_fund_flow_data(self, fund_flow_data: pd.DataFrame):
        """保存资金流数据"""
        if fund_flow_data.empty:
            logger.warning("资金流数据为空，跳过保存")
            return

        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            sql = """
            INSERT INTO sector_fund_flow
            (sector_name, trade_date, net_inflow, main_net_inflow, retail_net_inflow, turnover_rate, created_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (sector_name, trade_date) DO UPDATE SET
                net_inflow = EXCLUDED.net_inflow,
                main_net_inflow = EXCLUDED.main_net_inflow,
                retail_net_inflow = EXCLUDED.retail_net_inflow,
                turnover_rate = EXCLUDED.turnover_rate
            """

            data = []
            current_time = datetime.now()

            for _, row in fund_flow_data.iterrows():
                data.append((
                    row.get('sector_name'),
                    row.get('trade_date'),
                    row.get('net_inflow', 0),
                    row.get('main_net_inflow', 0),
                    row.get('retail_net_inflow', 0),
                    row.get('turnover_rate', 0),
                    current_time
                ))

            execute_values(cursor, sql, data)
            conn.commit()
            logger.info(f"成功保存 {len(fund_flow_data)} 条资金流数据")

        except Exception as e:
            logger.error(f"保存资金流数据失败: {e}")
            if conn:
                conn.rollback()
        finally:
            if cursor:
                cursor.close()
            if conn:
                conn.close()

    def save_strategy_results(self, results: List[Dict[str, Any]]):
        """保存策略执行结果"""
        if not results:
            logger.warning("策略结果为空，跳过保存")
            return

        try:
            conn = self._get_connection()
            cursor = conn.cursor()

            sql = """
            INSERT INTO strategy_results
            (strategy_id, trade_date, stock_code, score, buy_price, stop_loss_price,
             take_profit_price, logic_description, indicators, created_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (strategy_id, trade_date, stock_code) DO UPDATE SET
                score = EXCLUDED.score,
                buy_price = EXCLUDED.buy_price,
                stop_loss_price = EXCLUDED.stop_loss_price,
                take_profit_price = EXCLUDED.take_profit_price,
                logic_description = EXCLUDED.logic_description,
                indicators = EXCLUDED.indicators
            """

            data = []
            current_time = datetime.now()

            for result in results:
                data.append((
                    result.get('strategy_id'),
                    result.get('trade_date'),
                    result.get('stock_code'),
                    result.get('score', 0),
                    result.get('buy_price', 0),
                    result.get('stop_loss_price', 0),
                    result.get('take_profit_price', 0),
                    result.get('logic_description', ''),
                    result.get('indicators', '{}'),
                    current_time
                ))

            execute_values(cursor, sql, data)
            conn.commit()
            logger.info(f"成功保存 {len(results)} 条策略结果")

        except Exception as e:
            logger.error(f"保存策略结果失败: {e}")
            if conn:
                conn.rollback()
        finally:
            if cursor:
                cursor.close()
            if conn:
                conn.close()