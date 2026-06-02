"""
PostgreSQL数据采集器 - 专为stock-v3项目设计
将采集的数据直接写入PostgreSQL数据库，供Golang后端使用
"""

import logging
import sys
import os
import pandas as pd
import baostock as bs
import akshare as ak
from datetime import datetime, timedelta, date
import psycopg2
from psycopg2.extras import execute_values
import time
import re
from typing import List, Dict, Optional

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('postgresql_data_collection.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# PostgreSQL配置（应与Golang后端配置一致）
POSTGRES_CONFIG = {
    'host': 'localhost',
    'port': '5432',
    'user': 'postgres',
    'password': 'password',
    'database': 'stock_strategy',
    'sslmode': 'disable'
}


class PostgreSQLDataCollector:
    """PostgreSQL数据采集器 - 专为stock-v3项目设计"""

    def __init__(self):
        self.bs_connected = False
        self.pg_conn = None
        self._init_baostock()
        self._init_postgresql()

    def _init_baostock(self):
        """初始化BaoStock连接"""
        try:
            lg = bs.login()
            if lg.error_code == '0':
                self.bs_connected = True
                logger.info("BaoStock连接成功")
            else:
                logger.error(f"BaoStock连接失败: {lg.error_msg}")
        except Exception as e:
            logger.error(f"BaoStock初始化失败: {e}")

    def _init_postgresql(self):
        """初始化PostgreSQL连接"""
        try:
            self.pg_conn = psycopg2.connect(**POSTGRES_CONFIG)
            logger.info("PostgreSQL连接成功")
        except Exception as e:
            logger.error(f"PostgreSQL连接失败: {e}")
            self.pg_conn = None

    def _code_to_bs(self, code: str) -> str:
        """将纯数字股票代码转换为BaoStock格式"""
        if code.startswith('6'):
            return f'sh.{code}'
        else:
            return f'sz.{code}'

    def _code_from_bs(self, bs_code: str) -> str:
        """将BaoStock格式代码转换为纯数字代码"""
        if '.' in bs_code:
            return bs_code.split('.')[1]
        return bs_code

    def get_all_a_stocks(self) -> List[Dict[str, str]]:
        """获取所有A股股票基本信息"""
        try:
            if not self.bs_connected:
                logger.error("BaoStock未连接")
                return []

            rs = bs.query_stock_basic(code_name="", code="")
            stock_list = []
            while (rs.error_code == '0') and rs.next():
                stock_list.append(rs.get_row_data())

            df = pd.DataFrame(stock_list, columns=rs.fields)
            # 只保留A股上市状态
            df = df[(df["type"] == "1") & (df["status"] == "1")]

            stocks = []
            for _, row in df.iterrows():
                stock = {
                    'code': self._code_from_bs(row['code']),
                    'name': row['code_name'],
                    'market': self._get_market_type(self._code_from_bs(row['code'])),
                    'industry': '',
                    'listing_date': None
                }
                stocks.append(stock)

            logger.info(f"获取到 {len(stocks)} 只A股股票")
            return stocks

        except Exception as e:
            logger.error(f"获取股票列表失败: {e}")
            return []

    def save_stocks_to_postgresql(self, stocks: List[Dict]):
        """将股票基本信息保存到PostgreSQL"""
        if not self.pg_conn:
            logger.error("PostgreSQL未连接")
            return False

        try:
            cursor = self.pg_conn.cursor()

            # 清空现有数据
            cursor.execute("DELETE FROM stocks")

            # 批量插入数据
            insert_data = []
            for stock in stocks:
                insert_data.append((
                    stock['code'],
                    stock['name'],
                    stock['industry'],
                    stock['market'],
                    stock['listing_date'] or datetime.now(),
                    datetime.now(),
                    datetime.now()
                ))

            execute_values(
                cursor,
                """INSERT INTO stocks
                (code, name, industry, market, listing_date, created_at, updated_at)
                VALUES %s""",
                insert_data
            )

            self.pg_conn.commit()
            logger.info(f"成功保存 {len(stocks)} 只股票信息到PostgreSQL")
            return True

        except Exception as e:
            logger.error(f"保存股票信息到PostgreSQL失败: {e}")
            self.pg_conn.rollback()
            return False

    def get_stock_daily_data(self, stock_code: str, start_date: str, end_date: str) -> pd.DataFrame:
        """获取股票日线数据"""
        try:
            if not self.bs_connected:
                return pd.DataFrame()

            bs_code = self._code_to_bs(stock_code)
            rs = bs.query_history_k_data_plus(
                bs_code,
                "date,code,open,high,low,close,volume,amount,turn,pctChg",
                start_date=start_date,
                end_date=end_date,
                frequency="d",
                adjustflag="2"  # 前复权
            )

            # 检查查询结果
            if rs is None:
                logger.warning(f"查询股票 {stock_code} 日线数据返回None")
                return pd.DataFrame()

            if rs.error_code != '0':
                logger.warning(f"查询股票 {stock_code} 日线数据失败: {rs.error_msg}")
                return pd.DataFrame()

            data_list = []
            while rs.next():
                data_list.append(rs.get_row_data())

            if not data_list:
                return pd.DataFrame()

            df = pd.DataFrame(data_list, columns=rs.fields)
            df["date"] = pd.to_datetime(df["date"])

            # 数据类型转换
            for col in ["open", "high", "low", "close", "volume", "amount", "turn", "pctChg"]:
                df[col] = pd.to_numeric(df[col], errors="coerce")

            # 重命名列以兼容PostgreSQL表结构
            df = df.rename(columns={
                "turn": "turnover_rate",
                "pctChg": "change_pct",
                "date": "trade_date"
            })

            # 过滤停牌日
            df = df[df["volume"] > 0]
            return df

        except Exception as e:
            logger.error(f"获取股票 {stock_code} 日线数据失败: {e}")
            return pd.DataFrame()

    def save_daily_data_to_postgresql(self, daily_data: pd.DataFrame, stock_code: str):
        """将日线数据保存到PostgreSQL"""
        if not self.pg_conn or daily_data.empty:
            return False

        try:
            cursor = self.pg_conn.cursor()

            # 批量插入数据
            insert_data = []
            for _, row in daily_data.iterrows():
                insert_data.append((
                    stock_code,
                    row['trade_date'],
                    row['open'],
                    row['high'],
                    row['low'],
                    row['close'],
                    row['volume'],
                    row['amount'],
                    row['turnover_rate'],
                    row['change_pct'],
                    datetime.now()
                ))

            execute_values(
                cursor,
                """INSERT INTO stock_daily_data
                (stock_code, trade_date, open_price, high_price, low_price, close_price,
                 volume, amount, turnover_rate, change_pct, created_at)
                VALUES %s""",
                insert_data
            )

            self.pg_conn.commit()
            logger.info(f"成功保存股票 {stock_code} 的 {len(daily_data)} 条日线数据到PostgreSQL")
            return True

        except Exception as e:
            logger.error(f"保存日线数据到PostgreSQL失败: {e}")
            self.pg_conn.rollback()
            return False

    def collect_and_save_historical_data(self, days: int = 30):
        """采集并保存历史数据"""
        try:
            logger.info(f"开始采集最近{days}天的历史数据")

            # 获取所有股票
            stocks = self.get_all_a_stocks()
            if not stocks:
                logger.error("无法获取股票列表")
                return

            # 保存股票基本信息
            self.save_stocks_to_postgresql(stocks)

            # 计算日期范围
            end_date = datetime.now()
            start_date = end_date - timedelta(days=days)

            # 分批采集历史数据
            batch_size = 50
            total_count = len(stocks)

            for i, stock in enumerate(stocks):
                try:
                    stock_code = stock['code']

                    # 获取日线数据
                    daily_data = self.get_stock_daily_data(
                        stock_code,
                        start_date.strftime('%Y%m%d'),
                        end_date.strftime('%Y%m%d')
                    )

                    if not daily_data.empty:
                        # 保存到PostgreSQL
                        self.save_daily_data_to_postgresql(daily_data, stock_code)

                    if (i + 1) % batch_size == 0:
                        logger.info(f"历史数据采集进度: {i+1}/{total_count}")

                    # 避免请求过快
                    time.sleep(0.1)

                except Exception as e:
                    logger.error(f"采集股票 {stock_code} 数据失败: {e}")

            logger.info("历史数据采集完成")

        except Exception as e:
            logger.error(f"历史数据采集任务失败: {e}")

    def _get_market_type(self, stock_code: str) -> str:
        """根据股票代码判断市场类型"""
        if stock_code.startswith('6'):
            return 'SH'
        elif stock_code.startswith('0') or stock_code.startswith('3'):
            return 'SZ'
        elif stock_code.startswith('8'):
            return 'BJ'
        else:
            return 'OTHER'

    def close(self):
        """关闭连接"""
        if self.bs_connected:
            bs.logout()
            self.bs_connected = False
            logger.info("BaoStock连接已关闭")

        if self.pg_conn:
            self.pg_conn.close()
            logger.info("PostgreSQL连接已关闭")

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


def collect_daily_data_to_postgresql():
    """每日数据采集任务 - 将数据保存到PostgreSQL"""
    collector = PostgreSQLDataCollector()

    try:
        logger.info("开始执行PostgreSQL数据采集任务")

        # 采集最近30天的历史数据
        collector.collect_and_save_historical_data(days=30)

        logger.info("PostgreSQL数据采集任务完成")

    except Exception as e:
        logger.error(f"PostgreSQL数据采集任务失败: {e}")
    finally:
        collector.close()


def test_postgresql_collector():
    """测试PostgreSQL数据采集器"""
    logger.info("开始测试PostgreSQL数据采集器...")

    collector = PostgreSQLDataCollector()

    try:
        # 测试获取股票列表
        stocks = collector.get_all_a_stocks()
        logger.info(f"测试结果: 获取到 {len(stocks)} 只股票")

        if stocks:
            # 测试保存到PostgreSQL
            success = collector.save_stocks_to_postgresql(stocks[:10])  # 只测试前10只
            logger.info(f"PostgreSQL保存测试: {'成功' if success else '失败'}")

            # 测试获取日线数据
            if len(stocks) > 0:
                test_stock = stocks[0]
                yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y%m%d')
                daily_data = collector.get_stock_daily_data(test_stock['code'], yesterday, yesterday)
                logger.info(f"日线数据测试: 获取到 {len(daily_data)} 条数据")

        logger.info("PostgreSQL数据采集器测试完成")

    except Exception as e:
        logger.error(f"PostgreSQL数据采集器测试失败: {e}")
    finally:
        collector.close()


if __name__ == "__main__":
    # 测试PostgreSQL数据采集器
    test_postgresql_collector()

    # 执行数据采集任务
    # collect_daily_data_to_postgresql()