"""
临时数据采集器 - 解决当前BaoStock查询问题
先确保能获取到基本数据，PostgreSQL连接问题后续解决
"""

import logging
import sys
import os
import pandas as pd
import baostock as bs
import akshare as ak
from datetime import datetime, timedelta, date
import time
import re
from typing import List, Dict, Optional

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('temp_data_collection.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


class TempDataCollector:
    """临时数据采集器 - 专注于解决当前BaoStock查询问题"""

    def __init__(self):
        self.bs_connected = False
        self._init_baostock()

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

    def get_all_a_stocks_robust(self) -> List[Dict[str, str]]:
        """稳健地获取所有A股股票基本信息"""
        try:
            if not self.bs_connected:
                logger.error("BaoStock未连接")
                return []

            # 使用AKShare作为备用方案
            try:
                logger.info("尝试使用AKShare获取股票列表...")
                stock_info_df = ak.stock_info_a_code_name()

                stocks = []
                for _, row in stock_info_df.iterrows():
                    stock = {
                        'code': row['code'],
                        'name': row['name'],
                        'market': self._get_market_type(row['code'])
                    }
                    stocks.append(stock)

                logger.info(f"AKShare获取到 {len(stocks)} 只A股股票")
                return stocks

            except Exception as e:
                logger.warning(f"AKShare获取股票列表失败: {e}，尝试使用BaoStock")

            # 使用BaoStock作为主方案
            rs = bs.query_stock_basic(code_name="", code="")

            # 检查查询结果
            if rs is None:
                logger.error("BaoStock查询返回None")
                return []

            if rs.error_code != '0':
                logger.error(f"BaoStock查询失败: {rs.error_msg}")
                return []

            stock_list = []
            while rs.next():
                stock_list.append(rs.get_row_data())

            df = pd.DataFrame(stock_list, columns=rs.fields)
            # 只保留A股上市状态
            df = df[(df["type"] == "1") & (df["status"] == "1")]

            stocks = []
            for _, row in df.iterrows():
                stock = {
                    'code': self._code_from_bs(row['code']),
                    'name': row['code_name'],
                    'market': self._get_market_type(self._code_from_bs(row['code']))
                }
                stocks.append(stock)

            logger.info(f"BaoStock获取到 {len(stocks)} 只A股股票")
            return stocks

        except Exception as e:
            logger.error(f"获取股票列表失败: {e}")
            return []

    def get_stock_daily_data_robust(self, stock_code: str, start_date: str, end_date: str) -> pd.DataFrame:
        """稳健地获取股票日线数据"""
        try:
            if not self.bs_connected:
                logger.error("BaoStock未连接")
                return pd.DataFrame()

            bs_code = self._code_to_bs(stock_code)

            # 尝试使用AKShare作为备用方案
            try:
                logger.info(f"尝试使用AKShare获取股票 {stock_code} 日线数据...")
                stock_zh_a_hist_df = ak.stock_zh_a_hist(
                    symbol=stock_code,
                    period="daily",
                    start_date=start_date,
                    end_date=end_date,
                    adjust="hfq"
                )

                if not stock_zh_a_hist_df.empty:
                    # 重命名列以兼容系统
                    stock_zh_a_hist_df = stock_zh_a_hist_df.rename(columns={
                        "日期": "date",
                        "开盘": "open",
                        "最高": "high",
                        "最低": "low",
                        "收盘": "close",
                        "成交量": "volume",
                        "成交额": "amount",
                        "涨跌幅": "change_pct",
                        "换手率": "turnover"
                    })

                    stock_zh_a_hist_df["date"] = pd.to_datetime(stock_zh_a_hist_df["date"])
                    logger.info(f"AKShare获取到股票 {stock_code} 的 {len(stock_zh_a_hist_df)} 条日线数据")
                    return stock_zh_a_hist_df

            except Exception as e:
                logger.debug(f"AKShare获取股票 {stock_code} 日线数据失败: {e}")

            # 使用BaoStock作为主方案
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
                logger.debug(f"股票 {stock_code} 在 {start_date} 到 {end_date} 期间无数据")
                return pd.DataFrame()

            df = pd.DataFrame(data_list, columns=rs.fields)
            df["date"] = pd.to_datetime(df["date"])

            # 数据类型转换
            for col in ["open", "high", "low", "close", "volume", "amount", "turn", "pctChg"]:
                df[col] = pd.to_numeric(df[col], errors="coerce")

            # 重命名列以兼容系统
            df = df.rename(columns={
                "turn": "turnover",
                "pctChg": "change_pct",
            })

            # 过滤停牌日
            df = df[df["volume"] > 0]
            logger.info(f"BaoStock获取到股票 {stock_code} 的 {len(df)} 条日线数据")
            return df

        except Exception as e:
            logger.error(f"获取股票 {stock_code} 日线数据失败: {e}")
            return pd.DataFrame()

    def filter_stocks_by_liquidity_simple(self, stocks: List[Dict], min_turnover: int = 200000000) -> List[Dict]:
        """简化版流动性筛选 - 使用当日成交额"""
        try:
            logger.info("开始简化版流动性筛选...")

            # 获取当日实时行情
            try:
                spot_df = ak.stock_zh_a_spot_em()

                filtered_stocks = []
                for stock in stocks:
                    stock_code = stock['code']

                    # 查找该股票的实时数据
                    stock_spot = spot_df[spot_df['代码'] == stock_code]

                    if not stock_spot.empty:
                        # 获取成交额（单位：万元）
                        turnover = stock_spot.iloc[0]['成交额']

                        # 转换为元（AKShare返回的是万元）
                        turnover_yuan = turnover * 10000

                        if turnover_yuan >= min_turnover:
                            stock['current_turnover'] = turnover_yuan
                            filtered_stocks.append(stock)

                logger.info(f"简化版流动性筛选完成: 从{len(stocks)}只股票中筛选出{len(filtered_stocks)}只")
                return filtered_stocks

            except Exception as e:
                logger.warning(f"AKShare实时行情获取失败: {e}，跳过流动性筛选")
                return stocks

        except Exception as e:
            logger.error(f"流动性筛选失败: {e}")
            return stocks

    def collect_basic_data_to_csv(self, output_dir: str = "data_output"):
        """采集基础数据并保存到CSV文件（临时方案）"""
        try:
            logger.info("开始采集基础数据到CSV文件...")

            # 创建输出目录
            os.makedirs(output_dir, exist_ok=True)

            # 1. 获取所有A股股票
            all_stocks = self.get_all_a_stocks_robust()
            if not all_stocks:
                logger.error("无法获取股票列表")
                return False

            # 保存股票基本信息
            stocks_df = pd.DataFrame(all_stocks)
            stocks_file = os.path.join(output_dir, "stocks_basic.csv")
            stocks_df.to_csv(stocks_file, index=False, encoding='utf-8-sig')
            logger.info(f"股票基本信息保存到: {stocks_file}")

            # 2. 过滤ST股票
            pattern = "|".join(["ST", "\\*ST", "退市", "退"])
            filtered_stocks = [s for s in all_stocks if not re.search(pattern, s.get('name', ''))]
            logger.info(f"过滤ST股票后剩余: {len(filtered_stocks)} 只")

            # 3. 流动性筛选
            liquid_stocks = self.filter_stocks_by_liquidity_simple(filtered_stocks)
            logger.info(f"流动性筛选后剩余: {len(liquid_stocks)} 只")

            # 4. 采集流动性股票的日线数据
            today = datetime.now().strftime('%Y%m%d')
            yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y%m%d')

            daily_data_list = []

            # 限制采集数量，避免超时
            max_collection = min(len(liquid_stocks), 100)

            for i, stock in enumerate(liquid_stocks[:max_collection]):
                try:
                    stock_code = stock['code']

                    # 获取最近2天的数据
                    daily_data = self.get_stock_daily_data_robust(stock_code, yesterday, today)

                    if not daily_data.empty:
                        # 添加股票代码和名称
                        daily_data['stock_code'] = stock_code
                        daily_data['stock_name'] = stock['name']
                        daily_data_list.append(daily_data)

                    if (i + 1) % 20 == 0:
                        logger.info(f"日线数据采集进度: {i+1}/{max_collection}")

                    # 避免请求过快
                    time.sleep(0.1)

                except Exception as e:
                    logger.error(f"采集股票 {stock_code} 数据失败: {e}")

            if daily_data_list:
                # 合并所有日线数据
                all_daily_data = pd.concat(daily_data_list, ignore_index=True)
                daily_file = os.path.join(output_dir, "stock_daily_data.csv")
                all_daily_data.to_csv(daily_file, index=False, encoding='utf-8-sig')
                logger.info(f"日线数据保存到: {daily_file}，共 {len(all_daily_data)} 条记录")

            # 5. 获取行业分类数据
            try:
                industry_df = ak.stock_industry_sw()
                industry_file = os.path.join(output_dir, "industry_classification.csv")
                industry_df.to_csv(industry_file, index=False, encoding='utf-8-sig')
                logger.info(f"行业分类数据保存到: {industry_file}")
            except Exception as e:
                logger.warning(f"获取行业分类数据失败: {e}")

            # 6. 获取资金流数据
            try:
                fund_flow_df = ak.stock_sector_fund_flow_rank(
                    indicator="今日",
                    sector_type="行业资金流"
                )
                fund_flow_file = os.path.join(output_dir, "sector_fund_flow.csv")
                fund_flow_df.to_csv(fund_flow_file, index=False, encoding='utf-8-sig')
                logger.info(f"资金流数据保存到: {fund_flow_file}")
            except Exception as e:
                logger.warning(f"获取资金流数据失败: {e}")

            logger.info("基础数据采集完成")
            return True

        except Exception as e:
            logger.error(f"基础数据采集失败: {e}")
            return False

    def _get_market_type(self, stock_code: str) -> str:
        """根据股票代码判断市场类型"""
        if stock_code.startswith('6'):
            return 'SH'  # 上海主板
        elif stock_code.startswith('0') or stock_code.startswith('3'):
            return 'SZ'  # 深圳主板/创业板
        elif stock_code.startswith('8'):
            return 'BJ'  # 北京交易所
        else:
            return 'OTHER'

    def close(self):
        """关闭连接"""
        if self.bs_connected:
            bs.logout()
            self.bs_connected = False
            logger.info("BaoStock连接已关闭")

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


def collect_data_to_csv():
    """采集数据到CSV文件（临时方案）"""
    collector = TempDataCollector()

    try:
        logger.info("开始执行临时数据采集任务...")

        # 采集数据到CSV文件
        success = collector.collect_basic_data_to_csv("temp_data_output")

        if success:
            logger.info("临时数据采集任务完成")
        else:
            logger.error("临时数据采集任务失败")

    except Exception as e:
        logger.error(f"临时数据采集任务失败: {e}")
    finally:
        collector.close()


def test_temp_collector():
    """测试临时数据采集器"""
    logger.info("开始测试临时数据采集器...")

    collector = TempDataCollector()

    try:
        # 测试获取股票列表
        stocks = collector.get_all_a_stocks_robust()
        logger.info(f"测试结果: 获取到 {len(stocks)} 只股票")

        if stocks:
            # 测试获取日线数据
            if len(stocks) > 0:
                test_stock = stocks[0]
                yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y%m%d')
                daily_data = collector.get_stock_daily_data_robust(test_stock['code'], yesterday, yesterday)
                logger.info(f"日线数据测试: 获取到 {len(daily_data)} 条数据")

        logger.info("临时数据采集器测试完成")

    except Exception as e:
        logger.error(f"临时数据采集器测试失败: {e}")
    finally:
        collector.close()


if __name__ == "__main__":
    # 测试临时数据采集器
    test_temp_collector()

    # 执行数据采集任务
    # collect_data_to_csv()