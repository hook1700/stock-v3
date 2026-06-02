"""
简化版数据采集器 - 专注于获取当日数据，不依赖历史数据
使用AKShare获取实时行情，简化流动性筛选
"""

import logging
import sys
import os
import pandas as pd
import akshare as ak
from datetime import datetime, timedelta
from typing import List, Dict, Optional
import re

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('simplified_data_collection.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


class SimplifiedDataCollector:
    """简化版数据采集器 - 使用AKShare获取当日数据"""

    def __init__(self):
        self.stock_pool_exclude = ["ST", "*ST", "退市", "退"]

    def get_all_a_stocks(self) -> List[Dict[str, str]]:
        """获取所有A股股票基本信息 - 使用AKShare"""
        try:
            logger.info("获取A股股票列表...")

            # 使用AKShare获取股票列表
            stock_info_df = ak.stock_info_a_code_name()

            stocks = []
            for _, row in stock_info_df.iterrows():
                stock = {
                    'code': row['code'],
                    'name': row['name'],
                    'market': self._get_market_type(row['code'])
                }
                stocks.append(stock)

            logger.info(f"获取到 {len(stocks)} 只A股股票")
            return stocks

        except Exception as e:
            logger.error(f"获取股票列表失败: {e}")
            return []

    def get_stock_spot_data(self, stock_codes: List[str] = None) -> pd.DataFrame:
        """获取股票实时行情数据"""
        try:
            logger.info("获取股票实时行情数据...")

            # 获取全部A股实时行情
            spot_df = ak.stock_zh_a_spot_em()

            if stock_codes:
                # 筛选指定股票代码
                spot_df = spot_df[spot_df['代码'].isin(stock_codes)]

            logger.info(f"获取到 {len(spot_df)} 只股票的实时行情")
            return spot_df

        except Exception as e:
            logger.error(f"获取实时行情数据失败: {e}")
            return pd.DataFrame()

    def filter_stocks_by_liquidity(self, stocks: List[Dict], min_turnover: int = 200000000) -> List[Dict]:
        """根据流动性筛选股票 - 使用当日成交额"""
        try:
            logger.info("开始流动性筛选...")

            # 获取股票代码列表
            stock_codes = [stock['code'] for stock in stocks]

            # 获取实时行情数据
            spot_data = self.get_stock_spot_data(stock_codes)

            if spot_data.empty:
                logger.warning("无法获取实时行情数据，跳过流动性筛选")
                return stocks

            # 筛选流动性股票
            filtered_stocks = []
            for stock in stocks:
                stock_code = stock['code']

                # 查找该股票的实时数据
                stock_spot = spot_data[spot_data['代码'] == stock_code]

                if not stock_spot.empty:
                    # 获取成交额（单位：万元）
                    turnover = stock_spot.iloc[0]['成交额']

                    # 转换为元（AKShare返回的是万元）
                    turnover_yuan = turnover * 10000

                    if turnover_yuan >= min_turnover:
                        stock['current_turnover'] = turnover_yuan
                        filtered_stocks.append(stock)

            logger.info(f"流动性筛选完成: 从{len(stocks)}只股票中筛选出{len(filtered_stocks)}只")
            return filtered_stocks

        except Exception as e:
            logger.error(f"流动性筛选失败: {e}")
            return stocks

    def get_industry_classification(self) -> pd.DataFrame:
        """获取行业分类数据"""
        try:
            logger.info("获取行业分类数据...")

            # 使用AKShare获取申万行业分类
            industry_df = ak.stock_industry_sw()

            logger.info(f"获取到 {len(industry_df)} 条行业分类数据")
            return industry_df

        except Exception as e:
            logger.error(f"获取行业分类数据失败: {e}")
            return pd.DataFrame()

    def get_sector_fund_flow(self) -> pd.DataFrame:
        """获取板块资金流数据"""
        try:
            logger.info("获取板块资金流数据...")

            # 获取行业资金流排名
            fund_flow_df = ak.stock_sector_fund_flow_rank(
                indicator="今日",
                sector_type="行业资金流"
            )

            logger.info(f"获取到 {len(fund_flow_df)} 条资金流数据")
            return fund_flow_df

        except Exception as e:
            logger.error(f"获取资金流数据失败: {e}")
            return pd.DataFrame()

    def get_stock_fundamentals(self, stock_code: str) -> Optional[Dict]:
        """获取股票基本面数据"""
        try:
            # 使用AKShare获取股票基本信息
            stock_info = ak.stock_individual_info_em(symbol=stock_code)

            if stock_info.empty:
                return None

            # 转换为字典格式
            fundamentals = {}
            for _, row in stock_info.iterrows():
                fundamentals[row['item']] = row['value']

            return fundamentals

        except Exception as e:
            logger.debug(f"获取股票 {stock_code} 基本面数据失败: {e}")
            return None

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


def collect_daily_data_simplified() -> Dict:
    """简化的每日数据采集任务"""
    collector = SimplifiedDataCollector()

    try:
        logger.info("开始执行简化的每日数据采集任务")

        # 1. 获取所有A股股票
        all_stocks = collector.get_all_a_stocks()
        if not all_stocks:
            logger.error("无法获取股票列表，任务终止")
            return {}

        # 2. 过滤ST股票
        pattern = "|".join(re.escape(x) for x in collector.stock_pool_exclude)
        filtered_stocks = [s for s in all_stocks if not re.search(pattern, s.get('name', ''))]
        logger.info(f"过滤ST股票后剩余: {len(filtered_stocks)} 只")

        # 3. 流动性筛选（使用当日成交额）
        liquid_stocks = collector.filter_stocks_by_liquidity(
            filtered_stocks,
            min_turnover=200000000  # 2亿成交额
        )

        # 4. 获取行业分类数据
        industry_data = collector.get_industry_classification()

        # 5. 获取资金流数据
        fund_flow_data = collector.get_sector_fund_flow()

        # 6. 获取实时行情数据
        spot_data = collector.get_stock_spot_data([s['code'] for s in liquid_stocks])

        result = {
            'all_stocks': all_stocks,
            'liquid_stocks': liquid_stocks,
            'industry_data': industry_data,
            'fund_flow_data': fund_flow_data,
            'spot_data': spot_data,
            'collection_time': datetime.now().isoformat()
        }

        logger.info("简化的每日数据采集任务完成")
        return result

    except Exception as e:
        logger.error(f"简化的每日数据采集任务失败: {e}")
        return {}


def test_simplified_collector():
    """测试简化版数据采集器"""
    logger.info("开始测试简化版数据采集器...")

    collector = SimplifiedDataCollector()

    try:
        # 测试获取股票列表
        stocks = collector.get_all_a_stocks()
        logger.info(f"测试结果: 获取到 {len(stocks)} 只股票")

        if stocks:
            # 测试流动性筛选
            liquid_stocks = collector.filter_stocks_by_liquidity(stocks[:100])
            logger.info(f"流动性筛选测试: 从100只股票中筛选出{len(liquid_stocks)}只")

            # 测试行业数据
            industry_data = collector.get_industry_classification()
            logger.info(f"行业数据测试: 获取到{len(industry_data)}条数据")

            # 测试资金流数据
            fund_flow = collector.get_sector_fund_flow()
            logger.info(f"资金流数据测试: 获取到{len(fund_flow)}条数据")

        logger.info("简化版数据采集器测试完成")

    except Exception as e:
        logger.error(f"简化版数据采集器测试失败: {e}")


if __name__ == "__main__":
    # 测试简化版数据采集器
    test_simplified_collector()

    # 执行简化的数据采集
    result = collect_daily_data_simplified()
    print(f"采集结果: {len(result.get('liquid_stocks', []))} 只流动性股票")