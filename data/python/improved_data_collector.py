"""
改进的数据采集器 - 基于stock-v2-min的成功方案
使用BaoStock + SQLite本地存储，避免网络问题
"""

import logging
import sys
import os
import pandas as pd
import baostock as bs
import akshare as ak
from datetime import datetime, timedelta, date
from pathlib import Path
from typing import List, Dict, Optional
import re
import math
import time

# 添加项目路径
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('improved_data_collection.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

# 导入数据库写入器
from database_writer import DatabaseWriter

# 配置
CONFIG = {
    'kline_days': 250,  # 历史K线天数
    'top_n_kline': 0,   # 0=全部股票，可设置限制加快速度
    'min_turnover': 200000000,  # 2亿成交额筛选
    'stock_pool_exclude': ["ST", "*ST", "退市", "退"]
}


def _code_to_bs(code: str) -> str:
    """将纯数字股票代码转换为 BaoStock 格式"""
    if code.startswith(("sh.", "sz.")):
        return code
    if code.startswith(("6", "9")):
        return f"sh.{code}"
    else:
        return f"sz.{code}"


def _code_from_bs(bs_code: str) -> str:
    """将 BaoStock 格式代码转换为纯数字代码"""
    if "." in bs_code:
        return bs_code.split(".")[1]
    return bs_code


class ImprovedDataCollector:
    """改进的数据采集器 - 基于BaoStock的稳定方案"""

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

    def get_all_a_stocks(self) -> List[Dict[str, str]]:
        """获取所有A股股票基本信息 - 使用BaoStock更稳定"""
        try:
            if not self.bs_connected:
                logger.error("BaoStock未连接")
                return []

            # 使用BaoStock获取股票基本信息
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
                    'code': _code_from_bs(row['code']),
                    'name': row['code_name'],
                    'market': self._get_market_type(_code_from_bs(row['code']))
                }
                stocks.append(stock)

            logger.info(f"获取到 {len(stocks)} 只A股股票")
            return stocks

        except Exception as e:
            logger.error(f"获取股票列表失败: {e}")
            return []

    def get_stock_daily_data(self, stock_code: str, start_date: str, end_date: str) -> pd.DataFrame:
        """获取股票日线数据 - 使用BaoStock"""
        try:
            if not self.bs_connected:
                logger.error("BaoStock未连接")
                return pd.DataFrame()

            bs_code = _code_to_bs(stock_code)
            rs = bs.query_history_k_data_plus(
                bs_code,
                "date,code,open,high,low,close,volume,amount,turn,pctChg",
                start_date=start_date,
                end_date=end_date,
                frequency="d",
                adjustflag="2"  # 前复权
            )

            # 检查查询结果是否有效
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

            # 重命名列以兼容现有系统
            df = df.rename(columns={
                "turn": "turnover",
                "pctChg": "change_pct",
            })

            # 过滤停牌日
            df = df[df["volume"] > 0]
            return df

        except Exception as e:
            logger.error(f"获取股票 {stock_code} 日线数据失败: {e}")
            return pd.DataFrame()

    def filter_stocks_by_liquidity(self, stocks: List[Dict], min_turnover: int = 200000000, use_realtime: bool = True) -> List[Dict]:
        """
        根据流动性筛选股票
        
        Args:
            stocks: 股票列表
            min_turnover: 最小成交额阈值（元）
            use_realtime: True=使用AKShare实时行情（快速），False=使用BaoStock历史数据（精确）
        """
        try:
            if use_realtime:
                return self._filter_by_realtime_akshare(stocks, min_turnover)
            else:
                return self._filter_by_history_baostock(stocks, min_turnover)
        except Exception as e:
            logger.error(f"流动性筛选失败: {e}")
            return stocks

    def _filter_by_realtime_akshare(self, stocks: List[Dict], min_turnover: int) -> List[Dict]:
        """
        使用AKShare实时行情进行流动性筛选（快速，一次性获取全市场数据）
        
        注意：AKShare返回的成交额单位为"元"，直接可比
        """
        try:
            logger.info("使用AKShare获取实时行情进行流动性筛选...")
            
            # 一次性获取全市场A股实时行情（包含成交额）
            logger.info("正在调用AKShare获取全市场实时行情...")
            spot_df = ak.stock_zh_a_spot_em()
            
            if spot_df.empty:
                logger.warning("AKShare返回空数据，降级使用BaoStock历史数据")
                return self._filter_by_history_baostock(stocks, min_turnover)
            
            logger.info(f"AKShare返回 {len(spot_df)} 只股票的实时行情")
            
            # 构建代码到成交额的映射（AKShare代码列名为'代码'，成交额列名为'成交额'）
            # 注意：AKShare 返回的代码可能不带前缀，需要统一格式
            turnover_map = {}
            for _, row in spot_df.iterrows():
                code = str(row['代码']).zfill(6)  # 确保6位代码
                # AKShare 成交额单位为元
                turnover_map[code] = float(row['成交额'])
            
            # 筛选流动性股票
            filtered_stocks = []
            total_count = len(stocks)
            
            for i, stock in enumerate(stocks):
                stock_code = stock['code']
                # 统一代码格式（确保6位）
                normalized_code = stock_code.zfill(6)
                
                # 从映射中获取成交额
                turnover = turnover_map.get(normalized_code, 0)
                
                if turnover >= min_turnover:
                    stock['avg_turnover'] = turnover
                    stock['turnover_source'] = 'realtime'
                    filtered_stocks.append(stock)
            
            logger.info(f"流动性筛选完成(AKShare实时): 从{total_count}只股票中筛选出{len(filtered_stocks)}只")
            return filtered_stocks
            
        except Exception as e:
            logger.error(f"AKShare实时流动性筛选失败: {e}，降级使用BaoStock历史数据")
            return self._filter_by_history_baostock(stocks, min_turnover)

    def _filter_by_history_baostock(self, stocks: List[Dict], min_turnover: int) -> List[Dict]:
        """
        使用BaoStock历史数据进行流动性筛选（精确，但速度慢）
        作为AKShare方案的降级备选
        """
        try:
            filtered_stocks = []
            total_count = len(stocks)

            for i, stock in enumerate(stocks):
                if i % 100 == 0:
                    logger.info(f"流动性筛选进度(BaoStock): {i}/{total_count}")

                stock_code = stock['code']

                # 获取最近20个交易日的成交额数据
                end_date = datetime.now().strftime('%Y%m%d')
                start_date = (datetime.now() - timedelta(days=60)).strftime('%Y%m%d')

                daily_data = self.get_stock_daily_data(stock_code, start_date, end_date)

                if not daily_data.empty and len(daily_data) >= 20:
                    # 计算日均成交额
                    avg_turnover = daily_data['amount'].tail(20).mean()

                    if avg_turnover >= min_turnover:
                        stock['avg_turnover'] = avg_turnover
                        stock['turnover_source'] = 'history'
                        filtered_stocks.append(stock)

            logger.info(f"流动性筛选完成(BaoStock历史): 从{total_count}只股票中筛选出{len(filtered_stocks)}只")
            return filtered_stocks

        except Exception as e:
            logger.error(f"BaoStock历史流动性筛选失败: {e}")
            return stocks

    def get_industry_classification(self) -> pd.DataFrame:
        """获取行业分类数据 - 使用BaoStock"""
        try:
            if not self.bs_connected:
                return pd.DataFrame()

            # 获取股票列表
            stocks = self.get_all_a_stocks()
            if not stocks:
                return pd.DataFrame()

            industry_data = []

            # 抽样获取行业信息（避免请求过多）
            sample_codes = [stock['code'] for stock in stocks[:300]]

            for i, code in enumerate(sample_codes):
                try:
                    bs_code = _code_to_bs(code)
                    rs_ind = bs.query_stock_industry(code=bs_code)

                    # 检查查询结果
                    if rs_ind is None:
                        continue
                    if rs_ind.error_code != '0':
                        continue

                    while rs_ind.next():
                        row = rs_ind.get_row_data()
                        if len(row) > 3:
                            industry_data.append({
                                'code': code,
                                'industry_name': row[3] if row[3] else ""
                            })

                    if (i + 1) % 50 == 0:
                        logger.info(f"行业数据进度: {i+1}/{len(sample_codes)}")

                except Exception as e:
                    logger.debug(f"获取股票 {code} 行业信息失败: {e}")

            if industry_data:
                df = pd.DataFrame(industry_data)
                logger.info(f"获取到 {len(df)} 只股票的行业信息")
                return df
            else:
                return pd.DataFrame()

        except Exception as e:
            logger.error(f"获取行业数据失败: {e}")
            return pd.DataFrame()

    def get_sector_fund_flow(self, date: str = None) -> pd.DataFrame:
        """获取板块资金流数据 - 使用AKShare"""
        try:
            if not date:
                date = datetime.now().strftime('%Y%m%d')

            # 获取当日资金流数据
            fund_flow_data = ak.stock_sector_fund_flow_rank(
                indicator="今日",
                sector_type="行业资金流"
            )
            return fund_flow_data
        except Exception as e:
            logger.error(f"获取资金流数据失败: {e}")
            return pd.DataFrame()

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


def collect_daily_data_improved(save_to_db: bool = True) -> Dict:
    """改进的每日数据采集任务
    
    Args:
        save_to_db: 是否保存数据到数据库，默认True
    """
    collector = ImprovedDataCollector()
    writer = DatabaseWriter() if save_to_db else None

    try:
        logger.info("开始执行改进的每日数据采集任务")

        # 1. 获取所有A股股票
        all_stocks = collector.get_all_a_stocks()
        if not all_stocks:
            logger.error("无法获取股票列表，任务终止")
            return {}

        # 2. 保存到数据库（股票基本信息）
        if writer:
            writer.save_stock_basic_info(all_stocks)
            logger.info(f"已保存 {len(all_stocks)} 只股票基本信息到数据库")

        # 3. 过滤ST股票
        pattern = "|".join(re.escape(x) for x in CONFIG['stock_pool_exclude'])
        filtered_stocks = [s for s in all_stocks if not re.search(pattern, s.get('name', ''))]
        logger.info(f"过滤ST股票后剩余: {len(filtered_stocks)} 只")

        # 4. 流动性筛选
        liquid_stocks = collector.filter_stocks_by_liquidity(
            filtered_stocks,
            min_turnover=CONFIG['min_turnover']
        )

        # 5. 获取行业分类数据
        industry_data = collector.get_industry_classification()
        
        # 6. 保存行业数据到数据库
        if writer and not industry_data.empty:
            writer.save_industry_data(industry_data)
            logger.info(f"已保存行业分类数据到数据库")

        # 7. 获取资金流数据
        fund_flow_data = collector.get_sector_fund_flow()
        
        # 8. 保存资金流数据到数据库
        if writer and not fund_flow_data.empty:
            writer.save_fund_flow_data(fund_flow_data)
            logger.info(f"已保存资金流数据到数据库")

        # 9. 采集今日股票数据并保存到数据库
        today = datetime.now().strftime('%Y%m%d')
        
        # 限制采集数量，避免超时
        max_collection = min(len(liquid_stocks), 100)
        
        saved_count = 0
        for i, stock in enumerate(liquid_stocks[:max_collection]):
            try:
                daily_data = collector.get_stock_daily_data(
                    stock['code'],
                    today,
                    today
                )

                if not daily_data.empty and writer:
                    # 保存日线数据到数据库
                    writer.save_stock_daily_data(stock['code'], daily_data)
                    saved_count += 1

                if (i + 1) % 20 == 0:
                    logger.info(f"今日数据采集进度: {i+1}/{max_collection}")

            except Exception as e:
                logger.error(f"采集股票 {stock['code']} 数据失败: {e}")

        logger.info(f"今日数据已保存 {saved_count} 只股票的日线数据到数据库")

        result = {
            'all_stocks': all_stocks,
            'liquid_stocks': liquid_stocks,
            'industry_data': industry_data,
            'fund_flow_data': fund_flow_data,
            'saved_count': saved_count,
            'collection_time': datetime.now().isoformat()
        }

        logger.info("改进的每日数据采集任务完成")
        return result

    except Exception as e:
        logger.error(f"改进的每日数据采集任务失败: {e}")
        return {}
    finally:
        collector.close()


def collect_historical_data_improved(days: int = 250, save_to_db: bool = True):
    """改进的历史数据采集
    
    Args:
        days: 采集最近N天的历史数据
        save_to_db: 是否保存数据到数据库，默认True
    """
    collector = ImprovedDataCollector()
    writer = DatabaseWriter() if save_to_db else None

    try:
        logger.info(f"开始采集最近{days}天的历史数据")

        # 获取所有股票代码
        all_stocks = collector.get_all_a_stocks()
        if not all_stocks:
            logger.error("无法获取股票列表，任务终止")
            return

        # 先保存股票基本信息到数据库
        if writer:
            writer.save_stock_basic_info(all_stocks)
            logger.info(f"已保存 {len(all_stocks)} 只股票基本信息到数据库")

        # 计算日期范围
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days)

        start_str = start_date.strftime('%Y%m%d')
        end_str = end_date.strftime('%Y%m%d')

        logger.info(f"采集日期范围: {start_str} ~ {end_str}")

        # 分批采集历史数据
        batch_size = 50
        total_batches = math.ceil(len(all_stocks) / batch_size)

        total_saved = 0
        for batch_num in range(total_batches):
            start_idx = batch_num * batch_size
            end_idx = min((batch_num + 1) * batch_size, len(all_stocks))
            batch = all_stocks[start_idx:end_idx]

            logger.info(f"处理批次 {batch_num + 1}/{total_batches} (股票 {start_idx+1}-{end_idx})")

            for stock in batch:
                try:
                    daily_data = collector.get_stock_daily_data(
                        stock['code'],
                        start_str,
                        end_str
                    )

                    # 保存到数据库
                    if not daily_data.empty and writer:
                        writer.save_stock_daily_data(stock['code'], daily_data)
                        total_saved += len(daily_data)
                        logger.debug(f"保存股票 {stock['code']} 的 {len(daily_data)} 条历史数据")
                    elif not daily_data.empty:
                        # 不保存数据库时只记录日志
                        logger.debug(f"采集到股票 {stock['code']} 的 {len(daily_data)} 条历史数据（未保存到数据库）")

                except Exception as e:
                    logger.error(f"采集股票 {stock['code']} 历史数据失败: {e}")

            # 批次间延迟，避免请求过快
            if batch_num < total_batches - 1:
                time.sleep(1)

        if writer:
            logger.info(f"历史数据采集完成，共保存 {total_saved} 条数据到数据库")
        else:
            logger.info("历史数据采集完成（未保存到数据库）")

    except Exception as e:
        logger.error(f"历史数据采集失败: {e}")
    finally:
        collector.close()


if __name__ == "__main__":
    # 测试改进的数据采集
    print("测试改进的数据采集...")

    # 测试单次采集（保存到数据库）
    result = collect_daily_data_improved(save_to_db=True)
    print(f"采集结果: {len(result.get('liquid_stocks', []))} 只流动性股票")
    print(f"保存到数据库: {result.get('saved_count', 0)} 只股票")

    # 测试历史数据采集（小批量，保存到数据库）
    # collect_historical_data_improved(days=30, save_to_db=True)