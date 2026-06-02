"""
策略执行器 - 连接Python数据采集和Golang策略引擎
"""

import logging
import json
import sys
import os
from datetime import datetime, timedelta
from typing import Dict, List, Optional

# 添加项目路径
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from stock_data_collector import StockDataCollector
from database_writer import DatabaseWriter

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class StrategyExecutor:
    """策略执行器"""

    def __init__(self):
        self.collector = StockDataCollector()
        self.writer = DatabaseWriter()

    def execute_strategy(self, strategy_id: str, trade_date: str = None) -> Dict:
        """执行单个策略"""
        if not trade_date:
            trade_date = datetime.now().strftime('%Y-%m-%d')

        logger.info(f"开始执行策略 {strategy_id}，交易日期: {trade_date}")

        try:
            # 1. 获取所有A股股票
            all_stocks = self.collector.get_all_a_stocks()
            logger.info(f"获取到 {len(all_stocks)} 只股票")

            # 2. 流动性筛选（日均成交额≥2亿）
            liquid_stocks = self.collector.filter_stocks_by_liquidity(all_stocks, min_turnover=200000000)
            logger.info(f"流动性筛选后剩余 {len(liquid_stocks)} 只股票")

            # 3. 根据策略类型执行不同的策略逻辑
            strategy_results = []

            if strategy_id.startswith('short_term'):
                strategy_results = self.execute_short_term_strategy(strategy_id, liquid_stocks, trade_date)
            elif strategy_id.startswith('medium_term'):
                strategy_results = self.execute_medium_term_strategy(strategy_id, liquid_stocks, trade_date)
            else:
                logger.warning(f"暂不支持策略类型: {strategy_id}")
                return {"error": f"暂不支持策略类型: {strategy_id}"}

            # 4. 保存策略结果到数据库
            if strategy_results:
                self.writer.save_strategy_results(strategy_id, trade_date, strategy_results)
                logger.info(f"策略 {strategy_id} 执行完成，生成 {len(strategy_results)} 条结果")

            return {
                "strategy_id": strategy_id,
                "trade_date": trade_date,
                "total_stocks": len(liquid_stocks),
                "results_count": len(strategy_results),
                "results": strategy_results[:10]  # 只返回前10条结果
            }

        except Exception as e:
            logger.error(f"执行策略 {strategy_id} 失败: {e}")
            return {"error": str(e)}

    def execute_short_term_strategy(self, strategy_id: str, stocks: List[Dict], trade_date: str) -> List[Dict]:
        """执行短线策略"""
        results = []

        # 计算交易日期范围（最近90个交易日）
        end_date = datetime.strptime(trade_date, '%Y-%m-%d')
        start_date = end_date - timedelta(days=120)

        for i, stock in enumerate(stocks):
            if i % 100 == 0:
                logger.info(f"处理短线策略进度: {i}/{len(stocks)}")

            try:
                stock_code = stock['code']

                # 获取股票日线数据
                daily_data = self.collector.get_stock_daily_data(
                    stock_code,
                    start_date.strftime('%Y%m%d'),
                    end_date.strftime('%Y%m%d')
                )

                if daily_data.empty or len(daily_data) < 30:
                    continue

                # 根据策略ID执行不同的短线策略
                if strategy_id == 'short_term_1':
                    result = self.ma_pullback_strategy(stock, daily_data, trade_date)
                elif strategy_id == 'short_term_2':
                    result = self.breakout_pullback_strategy(stock, daily_data, trade_date)
                elif strategy_id == 'short_term_3':
                    result = self.strong_stock_rebound_strategy(stock, daily_data, trade_date)
                else:
                    continue

                if result and result.get('score', 0) > 0.5:
                    results.append(result)

            except Exception as e:
                logger.error(f"处理股票 {stock['code']} 失败: {e}")
                continue

        # 按评分排序，取前20只股票
        results.sort(key=lambda x: x.get('score', 0), reverse=True)
        return results[:20]

    def ma_pullback_strategy(self, stock: Dict, daily_data, trade_date: str) -> Optional[Dict]:
        """均线回踩低吸策略"""
        try:
            # 获取最新数据
            latest_data = daily_data.iloc[-1]

            # 计算技术指标
            close_prices = daily_data['close'].astype(float)
            ma5 = close_prices.tail(5).mean()
            ma10 = close_prices.tail(10).mean()
            ma20 = close_prices.tail(20).mean()

            # 计算前10日的MA20用于趋势判断
            if len(daily_data) >= 30:
                prev_ma20 = close_prices.iloc[-30:-10].mean()
            else:
                prev_ma20 = ma20

            # 策略逻辑判断
            # 1. 股价在20日线上方
            # 2. 20日线趋势向上
            # 3. 回踩5/10日线（价格接近均线2-3%）
            if latest_data['close'] > ma20 and ma20 > prev_ma20:
                # 检查是否回踩5日线或10日线
                distance_to_ma5 = abs(latest_data['close'] - ma5) / ma5
                distance_to_ma10 = abs(latest_data['close'] - ma10) / ma10

                if distance_to_ma5 < 0.02 or distance_to_ma10 < 0.03:
                    # 成交量验证
                    volume_score = self.analyze_volume_pattern(daily_data)

                    # K线形态验证
                    candle_score = self.analyze_candle_pattern(latest_data)

                    score = 0.7 + volume_score * 0.2 + candle_score * 0.1
                    buy_price = latest_data['close']
                    stop_loss = buy_price * 0.93  # -7%止损
                    take_profit = buy_price * 1.08  # +8%止盈

                    # 确定最接近的均线
                    closest_ma = "5日" if distance_to_ma5 < distance_to_ma10 else "10日"
                    logic = f"股价在20日线上方，20日线向上({prev_ma20:.2f}→{ma20:.2f})，回踩{closest_ma}线企稳"

                    return {
                        'stock_code': stock['code'],
                        'stock_name': stock['name'],
                        'score': round(score, 4),
                        'buy_price': round(buy_price, 3),
                        'stop_loss': round(stop_loss, 3),
                        'take_profit': round(take_profit, 3),
                        'logic': logic,
                        'indicators': {
                            'ma5': round(ma5, 3),
                            'ma10': round(ma10, 3),
                            'ma20': round(ma20, 3),
                            'distance_to_ma5': round(distance_to_ma5, 4),
                            'distance_to_ma10': round(distance_to_ma10, 4)
                        }
                    }

        except Exception as e:
            logger.error(f"均线回踩策略执行失败: {e}")

        return None

    def breakout_pullback_strategy(self, stock: Dict, daily_data, trade_date: str) -> Optional[Dict]:
        """突破缩量回踩策略"""
        try:
            if len(daily_data) < 60:
                return None

            latest_data = daily_data.iloc[-1]

            # 识别箱体震荡区间（前30个交易日）
            box_data = daily_data.iloc[-45:-15] if len(daily_data) >= 45 else daily_data.iloc[-30:]

            # 计算箱体上下沿
            box_high = box_data['high'].max()
            box_low = box_data['low'].min()

            # 检查突破信号（最近15个交易日）
            breakout_data = daily_data.iloc[-15:]
            breakout_detected = False
            breakout_price = 0
            breakout_volume = 0

            for _, data in breakout_data.iterrows():
                if data['close'] > box_high and data['volume'] > box_data['volume'].mean() * 1.5:
                    breakout_detected = True
                    breakout_price = data['close']
                    breakout_volume = data['volume']
                    break

            if not breakout_detected:
                return None

            # 检查回踩
            pullback_ratio = abs(latest_data['close'] - box_high) / box_high

            # 回踩确认：价格接近原压力位，成交量萎缩
            if pullback_ratio < 0.03 and latest_data['volume'] < breakout_volume * 0.7:
                score = 0.65
                buy_price = latest_data['close']
                stop_loss = min(buy_price * 0.93, box_low)
                take_profit = buy_price * 1.12
                logic = f"平台突破({breakout_price:.2f})后缩量回踩原压力位，支撑有效"

                return {
                    'stock_code': stock['code'],
                    'stock_name': stock['name'],
                    'score': round(score, 4),
                    'buy_price': round(buy_price, 3),
                    'stop_loss': round(stop_loss, 3),
                    'take_profit': round(take_profit, 3),
                    'logic': logic,
                    'indicators': {
                        'box_high': round(box_high, 3),
                        'box_low': round(box_low, 3),
                        'breakout_price': round(breakout_price, 3),
                        'pullback_ratio': round(pullback_ratio, 4)
                    }
                }

        except Exception as e:
            logger.error(f"突破缩量回踩策略执行失败: {e}")

        return None

    def strong_stock_rebound_strategy(self, stock: Dict, daily_data, trade_date: str) -> Optional[Dict]:
        """强势股10日线反抽策略"""
        try:
            if len(daily_data) < 40:
                return None

            latest_data = daily_data.iloc[-1]

            # 强势股筛选：近20日涨幅 > 0
            recent_20_data = daily_data.iloc[-20:]
            if len(recent_20_data) < 20:
                return None

            price_change = (recent_20_data.iloc[-1]['close'] - recent_20_data.iloc[0]['close']) / recent_20_data.iloc[0]['close']
            if price_change <= 0:
                return None

            # 计算10日线
            close_prices = daily_data['close'].astype(float)
            ma10 = close_prices.tail(10).mean()

            # 回踩确认：价格接近10日线，出现阳线
            distance_to_ma10 = abs(latest_data['close'] - ma10) / ma10
            is_positive = latest_data['close'] > latest_data['open']

            if distance_to_ma10 < 0.03 and is_positive:
                # 检查是否是第一次回踩
                if self.is_first_pullback(daily_data, 10, 20):
                    score = 0.68 + min(price_change * 2, 0.2)
                    buy_price = latest_data['close']
                    stop_loss = buy_price * 0.93
                    take_profit = buy_price * 1.10
                    logic = f"强势股(近20日涨幅{price_change*100:.1f}%)第一次回踩10日线收阳，反弹概率较高"

                    return {
                        'stock_code': stock['code'],
                        'stock_name': stock['name'],
                        'score': round(score, 4),
                        'buy_price': round(buy_price, 3),
                        'stop_loss': round(stop_loss, 3),
                        'take_profit': round(take_profit, 3),
                        'logic': logic,
                        'indicators': {
                            'ma10': round(ma10, 3),
                            'price_change_20d': round(price_change, 4),
                            'distance_to_ma10': round(distance_to_ma10, 4)
                        }
                    }

        except Exception as e:
            logger.error(f"强势股反抽策略执行失败: {e}")

        return None

    def execute_medium_term_strategy(self, strategy_id: str, stocks: List[Dict], trade_date: str) -> List[Dict]:
        """执行中线策略（技术面实现）"""
        results = []

        # 计算交易日期范围（最近180个交易日）
        end_date = datetime.strptime(trade_date, '%Y-%m-%d')
        start_date = end_date - timedelta(days=240)

        for i, stock in enumerate(stocks):
            if i % 100 == 0:
                logger.info(f"处理中线策略进度: {i}/{len(stocks)}")

            try:
                stock_code = stock['code']

                # 获取股票日线数据
                daily_data = self.collector.get_stock_daily_data(
                    stock_code,
                    start_date.strftime('%Y%m%d'),
                    end_date.strftime('%Y%m%d')
                )

                if daily_data.empty or len(daily_data) < 60:
                    continue

                if strategy_id == 'medium_term_1':
                    result = self.industry_growth_strategy(stock, daily_data, trade_date)
                elif strategy_id == 'medium_term_2':
                    result = self.turnaround_strategy(stock, daily_data, trade_date)
                else:
                    continue

                if result and result.get('score', 0) > 0.5:
                    results.append(result)

            except Exception as e:
                logger.error(f"处理股票 {stock['code']} 失败: {e}")
                continue

        # 按评分排序，取前15只股票
        results.sort(key=lambda x: x.get('score', 0), reverse=True)
        return results[:15]

    def industry_growth_strategy(self, stock: Dict, daily_data, trade_date: str) -> Optional[Dict]:
        """行业成长均线多头策略（技术面实现）"""
        try:
            if len(daily_data) < 60:
                return None

            latest_data = daily_data.iloc[-1]

            # 检查均线多头排列：5>10>20>60
            close_prices = daily_data['close'].astype(float)
            ma5 = close_prices.tail(5).mean()
            ma10 = close_prices.tail(10).mean()
            ma20 = close_prices.tail(20).mean()
            ma60 = close_prices.tail(60).mean() if len(daily_data) >= 60 else ma20

            if ma5 > ma10 and ma10 > ma20 and ma20 > ma60:
                # 检查是否突破后回踩
                distance_to_ma20 = abs(latest_data['close'] - ma20) / ma20
                distance_to_ma60 = abs(latest_data['close'] - ma60) / ma60

                if distance_to_ma20 < 0.03 or distance_to_ma60 < 0.05:
                    score = 0.75
                    buy_price = latest_data['close']
                    stop_loss = buy_price * 0.88  # -12%止损
                    take_profit = buy_price * 1.25  # +25%止盈
                    logic = "均线多头排列(5>10>20>60)，突破后回踩关键均线企稳，中线趋势良好"

                    return {
                        'stock_code': stock['code'],
                        'stock_name': stock['name'],
                        'score': round(score, 4),
                        'buy_price': round(buy_price, 3),
                        'stop_loss': round(stop_loss, 3),
                        'take_profit': round(take_profit, 3),
                        'logic': logic,
                        'indicators': {
                            'ma5': round(ma5, 3),
                            'ma10': round(ma10, 3),
                            'ma20': round(ma20, 3),
                            'ma60': round(ma60, 3)
                        }
                    }

        except Exception as e:
            logger.error(f"行业成长策略执行失败: {e}")

        return None

    def turnaround_strategy(self, stock: Dict, daily_data, trade_date: str) -> Optional[Dict]:
        """困境反转策略（技术面实现）"""
        try:
            if len(daily_data) < 120:
                return None

            latest_data = daily_data.iloc[-1]

            # 检查前期调整幅度（从最高点到最低点）
            early_data = daily_data.iloc[:90] if len(daily_data) >= 90 else daily_data
            max_price = early_data['high'].max()
            min_price = early_data['low'].min()

            decline_ratio = (max_price - min_price) / max_price
            if decline_ratio < 0.3:  # 调整幅度不足30%
                return None

            # 检查近期是否放量突破60日线
            recent_data = daily_data.iloc[-30:]
            ma60 = daily_data['close'].tail(60).mean()

            breakout_detected = False
            for _, data in recent_data.iterrows():
                if data['close'] > ma60 and data['volume'] > daily_data['volume'].tail(60).mean() * 1.3:
                    breakout_detected = True
                    break

            if breakout_detected:
                score = 0.7
                buy_price = latest_data['close']
                stop_loss = buy_price * 0.90  # -10%止损
                take_profit = buy_price * 1.20  # +20%止盈
                logic = f"前期调整{decline_ratio*100:.1f}%，近期放量突破60日线，困境反转信号"

                return {
                    'stock_code': stock['code'],
                    'stock_name': stock['name'],
                    'score': round(score, 4),
                    'buy_price': round(buy_price, 3),
                    'stop_loss': round(stop_loss, 3),
                    'take_profit': round(take_profit, 3),
                    'logic': logic,
                    'indicators': {
                        'decline_ratio': round(decline_ratio, 4),
                        'ma60': round(ma60, 3)
                    }
                }

        except Exception as e:
            logger.error(f"困境反转策略执行失败: {e}")

        return None

    # ========== 工具函数 ==========

    def analyze_volume_pattern(self, daily_data):
        """分析成交量模式"""
        if len(daily_data) < 15:
            return 0

        # 检查放量上涨→缩量回踩模式
        recent_data = daily_data.iloc[-10:]

        # 检查前5个交易日是否放量
        if len(recent_data) >= 5:
            volume_ma5 = recent_data.iloc[:5]['volume'].mean()
            volume_ma10 = daily_data.iloc[-15:-5]['volume'].mean()
            if volume_ma5 > volume_ma10 * 1.2:
                volume_increase = True
            else:
                volume_increase = False
        else:
            volume_increase = False

        # 检查最近3个交易日是否缩量
        if len(recent_data) >= 3:
            recent_volume = recent_data.iloc[-3:]['volume'].mean()
            prev_volume = recent_data.iloc[:3]['volume'].mean()
            if recent_volume < prev_volume * 0.8:
                volume_decrease = True
            else:
                volume_decrease = False
        else:
            volume_decrease = False

        if volume_increase and volume_decrease:
            return 0.8  # 良好的成交量模式

        return 0.3  # 一般的成交量模式

    def analyze_candle_pattern(self, data):
        """分析K线形态"""
        body_size = abs(data['close'] - data['open'])
        total_range = data['high'] - data['low']

        if total_range == 0:
            return 0

        body_ratio = body_size / total_range

        # 判断K线形态
        if body_ratio < 0.3:
            return 0.8  # 小实体（十字星、纺锤线）
        elif body_ratio > 0.7:
            if data['close'] > data['open']:
                return 0.9  # 大阳线
            return 0.2  # 大阴线
        else:
            if data['close'] > data['open']:
                return 0.7  # 中阳线
            return 0.4  # 中阴线

    def is_first_pullback(self, daily_data, ma_period, lookback):
        """检查是否是第一次回踩"""
        if len(daily_data) < lookback + ma_period:
            return True

        # 检查过去lookback个交易日内是否有其他回踩
        recent_data = daily_data.iloc[-lookback:]

        for i in range(len(recent_data) - 5):
            window_data = daily_data.iloc[-lookback+i:-lookback+i+ma_period]
            if len(window_data) < ma_period:
                continue

            ma = window_data['close'].mean()
            distance = abs(recent_data.iloc[i]['close'] - ma) / ma

            if distance < 0.03:
                return False  # 发现之前有回踩

        return True

    def execute_all_strategies(self, trade_date: str = None) -> Dict:
        """执行所有策略"""
        if not trade_date:
            trade_date = datetime.now().strftime('%Y-%m-%d')

        logger.info(f"开始执行所有策略，交易日期: {trade_date}")

        strategies = [
            'short_term_1',  # 均线回踩低吸
            'short_term_2',  # 突破缩量回踩
            'short_term_3',  # 强势股10日线反抽
            'medium_term_1', # 行业成长均线多头
            'medium_term_2', # 困境反转
        ]

        all_results = {}

        for strategy_id in strategies:
            try:
                result = self.execute_strategy(strategy_id, trade_date)
                all_results[strategy_id] = result
                logger.info(f"策略 {strategy_id} 执行完成: {result.get('results_count', 0)} 条结果")
            except Exception as e:
                logger.error(f"执行策略 {strategy_id} 失败: {e}")
                all_results[strategy_id] = {"error": str(e)}

        logger.info("所有策略执行完成")
        return all_results


def main():
    """主函数"""
    executor = StrategyExecutor()

    # 解析命令行参数
    if len(sys.argv) > 1:
        if sys.argv[1] == "--strategy":
            strategy_id = sys.argv[2] if len(sys.argv) > 2 else "short_term_1"
            trade_date = sys.argv[3] if len(sys.argv) > 3 else None

            result = executor.execute_strategy(strategy_id, trade_date)
            print(json.dumps(result, indent=2, ensure_ascii=False))

        elif sys.argv[1] == "--all":
            trade_date = sys.argv[2] if len(sys.argv) > 2 else None

            result = executor.execute_all_strategies(trade_date)
            print(json.dumps(result, indent=2, ensure_ascii=False))

        else:
            print("用法: python strategy_executor.py [--strategy <strategy_id> [trade_date]] | [--all [trade_date]]")
    else:
        # 默认执行短线策略1
        result = executor.execute_strategy("short_term_1")
        print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()