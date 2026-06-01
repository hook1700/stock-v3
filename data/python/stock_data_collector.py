"""
股票数据采集器
使用BaoStock和AKShare获取股票数据
"""

import logging
import pandas as pd
import baostock as bs
import akshare as ak
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple
import time
import math

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class StockDataCollector:
    """股票数据采集器"""

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
        """获取所有A股股票基本信息"""
        try:
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

    def get_stock_daily_data(self, stock_code: str, start_date: str, end_date: str) -> pd.DataFrame:
        """获取股票日线数据"""
        try:
            if self.bs_connected:
                # 使用BaoStock获取日线数据
                rs = bs.query_history_k_data_plus(
                    stock_code,
                    "date,code,open,high,low,close,volume,amount,turn",
                    start_date=start_date,
                    end_date=end_date,
                    frequency="d",
                    adjustflag="3"  # 后复权
                )

                data_list = []
                while (rs.error_code == '0') & rs.next():
                    data_list.append(rs.get_row_data())

                if data_list:
                    df = pd.DataFrame(data_list, columns=rs.fields)
                    # 数据类型转换
                    df['open'] = df['open'].astype(float)
                    df['high'] = df['high'].astype(float)
                    df['low'] = df['low'].astype(float)
                    df['close'] = df['close'].astype(float)
                    df['volume'] = df['volume'].astype(int)
                    df['amount'] = df['amount'].astype(float)
                    df['turn'] = df['turn'].astype(float)
                    df['date'] = pd.to_datetime(df['date'])

                    return df
                else:
                    return pd.DataFrame()

            else:
                # 备用方案：使用AKShare
                stock_zh_a_hist_df = ak.stock_zh_a_hist(
                    symbol=stock_code,
                    period="daily",
                    start_date=start_date,
                    end_date=end_date,
                    adjust="hfq"
                )
                return stock_zh_a_hist_df

        except Exception as e:
            logger.error(f"获取股票 {stock_code} 日线数据失败: {e}")
            return pd.DataFrame()

    def get_industry_classification(self) -> pd.DataFrame:
        """获取申万行业分类数据"""
        try:
            # 使用AKShare获取申万行业分类
            stock_industry_sw_df = ak.stock_industry_sw()
            return stock_industry_sw_df
        except Exception as e:
            logger.error(f"获取申万行业数据失败: {e}")
            return pd.DataFrame()

    def get_sector_fund_flow(self, date: str = None) -> pd.DataFrame:
        """获取板块资金流数据"""
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

    def get_stock_basic_info(self, stock_code: str) -> Dict:
        """获取股票基本面信息"""
        try:
            # 使用AKShare获取股票基本信息
            stock_info = ak.stock_individual_info_em(symbol=stock_code)

            basic_info = {}
            for _, row in stock_info.iterrows():
                basic_info[row['item']] = row['value']

            return basic_info

        except Exception as e:
            logger.error(f"获取股票 {stock_code} 基本信息失败: {e}")
            return {}

    def filter_stocks_by_liquidity(self, stocks: List[Dict], min_turnover: int = 200000000) -> List[Dict]:
        """根据流动性筛选股票"""
        try:
            filtered_stocks = []
            total_count = len(stocks)

            for i, stock in enumerate(stocks):
                if i % 100 == 0:
                    logger.info(f"流动性筛选进度: {i}/{total_count}")

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
                        filtered_stocks.append(stock)

            logger.info(f"流动性筛选完成: 从{total_count}只股票中筛选出{len(filtered_stocks)}只")
            return filtered_stocks

        except Exception as e:
            logger.error(f"流动性筛选失败: {e}")
            return stocks

    def calculate_technical_indicators(self, daily_data: pd.DataFrame) -> Dict:
        """计算技术指标"""
        if daily_data.empty:
            return {}

        try:
            # 计算移动平均线
            close_prices = daily_data['close'].astype(float)

            indicators = {
                'ma5': close_prices.tail(5).mean(),
                'ma10': close_prices.tail(10).mean(),
                'ma20': close_prices.tail(20).mean(),
                'ma60': close_prices.tail(60).mean() if len(daily_data) >= 60 else 0,
            }

            # 计算成交量均线
            volumes = daily_data['volume'].astype(int)
            indicators['volume_ma5'] = volumes.tail(5).mean()
            indicators['volume_ma10'] = volumes.tail(10).mean()

            # 计算涨跌幅
            if len(daily_data) >= 2:
                current_close = close_prices.iloc[-1]
                prev_close = close_prices.iloc[-2]
                indicators['change_rate'] = (current_close - prev_close) / prev_close

            # 计算RSI（简化版）
            indicators['rsi'] = self._calculate_rsi(close_prices.tail(14))

            return indicators

        except Exception as e:
            logger.error(f"计算技术指标失败: {e}")
            return {}

    def _calculate_rsi(self, prices: pd.Series, period: int = 14) -> float:
        """计算RSI指标"""
        if len(prices) < period + 1:
            return 0

        try:
            delta = prices.diff()
            gain = (delta.where(delta > 0, 0)).rolling(window=period).mean()
            loss = (-delta.where(delta < 0, 0)).rolling(window=period).mean()

            rs = gain.iloc[-1] / loss.iloc[-1] if loss.iloc[-1] != 0 else 0
            rsi = 100 - (100 / (1 + rs))

            return rsi
        except:
            return 0

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


def collect_daily_data() -> Dict:
    """每日数据采集任务"""
    collector = StockDataCollector()

    try:
        logger.info("开始执行每日数据采集任务")

        # 1. 获取所有A股股票
        all_stocks = collector.get_all_a_stocks()

        # 2. 流动性筛选
        liquid_stocks = collector.filter_stocks_by_liquidity(
            all_stocks,
            min_turnover=200000000  # 2亿成交额
        )

        # 3. 获取行业分类数据
        industry_data = collector.get_industry_classification()

        # 4. 获取资金流数据
        fund_flow_data = collector.get_sector_fund_flow()

        # 5. 采集今日股票数据
        today = datetime.now().strftime('%Y%m%d')
        stock_daily_data = {}

        for stock in liquid_stocks[:100]:  # 限制采集数量，避免超时
            try:
                daily_data = collector.get_stock_daily_data(
                    stock['code'],
                    today,
                    today
                )

                if not daily_data.empty:
                    indicators = collector.calculate_technical_indicators(daily_data)
                    stock_daily_data[stock['code']] = {
                        'daily_data': daily_data,
                        'indicators': indicators
                    }

            except Exception as e:
                logger.error(f"采集股票 {stock['code']} 数据失败: {e}")

        result = {
            'all_stocks': all_stocks,
            'liquid_stocks': liquid_stocks,
            'industry_data': industry_data,
            'fund_flow_data': fund_flow_data,
            'stock_daily_data': stock_daily_data,
            'collection_time': datetime.now().isoformat()
        }

        logger.info("每日数据采集任务完成")
        return result

    except Exception as e:
        logger.error(f"每日数据采集任务失败: {e}")
        return {}
    finally:
        collector.close()


def collect_historical_data(days: int = 365):
    """采集历史数据（用于初始化）"""
    collector = StockDataCollector()

    try:
        logger.info(f"开始采集最近{days}天的历史数据")

        # 获取所有股票代码
        all_stocks = collector.get_all_a_stocks()

        # 计算日期范围
        end_date = datetime.now()
        start_date = end_date - timedelta(days=days)

        # 分批采集历史数据
        batch_size = 50
        total_batches = math.ceil(len(all_stocks) / batch_size)

        for batch_num in range(total_batches):
            start_idx = batch_num * batch_size
            end_idx = min((batch_num + 1) * batch_size, len(all_stocks))
            batch = all_stocks[start_idx:end_idx]

            logger.info(f"处理批次 {batch_num + 1}/{total_batches}")

            for stock in batch:
                try:
                    daily_data = collector.get_stock_daily_data(
                        stock['code'],
                        start_date.strftime('%Y%m%d'),
                        end_date.strftime('%Y%m%d')
                    )

                    # 这里可以保存到数据库
                    if not daily_data.empty:
                        logger.debug(f"采集到股票 {stock['code']} 的 {len(daily_data)} 条历史数据")

                except Exception as e:
                    logger.error(f"采集股票 {stock['code']} 历史数据失败: {e}")

            # 批次间延迟，避免请求过快
            time.sleep(1)

        logger.info("历史数据采集完成")

    except Exception as e:
        logger.error(f"历史数据采集失败: {e}")
    finally:
        collector.close()


if __name__ == "__main__":
    # 测试数据采集
    print("测试股票数据采集...")

    # 测试单次采集
    result = collect_daily_data()
    print(f"采集结果: {len(result.get('liquid_stocks', []))} 只流动性股票")

    # 测试历史数据采集（小批量）
    # collect_historical_data(days=30)