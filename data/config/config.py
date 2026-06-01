"""
数据源配置模块
配置股票数据获取的相关参数
"""

import os
from typing import Dict, List

class DataSourceConfig:
    """数据源配置类"""

    # BaoStock 配置
    BAOSTOCK_CONFIG = {
        'host': '127.0.0.1',
        'port': 3306,
        'user': 'baostock',
        'password': 'password',
        'database': 'baostock'
    }

    # AKShare 配置
    AKSHARE_CONFIG = {
        'base_url': 'http://api.akshare.akfamily.xyz',
        'timeout': 30,
        'retry_times': 3
    }

    # 备用数据源配置
    BACKUP_SOURCES = [
        {
            'name': '新浪财经',
            'url': 'http://hq.sinajs.cn',
            'enabled': True
        },
        {
            'name': '东方财富',
            'url': 'http://push2.eastmoney.com',
            'enabled': True
        }
    ]

class StockFilterConfig:
    """股票筛选配置类"""

    # 通用筛选条件
    COMMON_FILTERS = {
        'min_daily_turnover': 200000000,  # 最低日均成交额2亿
        'max_daily_turnover': None,       # 最高日均成交额无限制
        'exclude_st': True,               # 排除ST股票
        'exclude_suspended': True,        # 排除停牌股票
    }

    # 短线策略筛选条件
    SHORT_TERM_FILTERS = {
        'min_daily_turnover': 500000000,  # 短线最低5亿成交额
        'min_price': 5.0,                 # 最低股价
        'max_price': 200.0,               # 最高股价
    }

    # 申万行业分类配置
    SW_INDUSTRY_CONFIG = {
        'level': 1,                       # 一级行业分类
        'update_frequency': 'daily'       # 每日更新
    }

class StrategyConfig:
    """策略执行配置类"""

    # 策略执行时间
    EXECUTION_TIME = {
        'daily_update': '17:45',          # 每日数据更新时间
        'strategy_run': '18:00',          # 策略执行时间
        'weekend_skip': True              # 周末跳过
    }

    # 策略配置
    STRATEGIES = {
        'short_term_1': {  # 短线策略①: 均线回踩低吸
            'enabled': True,
            'name': '均线回踩低吸',
            'type': 'short_term',
            'description': '趋势热点股、板块处上升期'
        },
        'short_term_2': {  # 短线策略②: 突破缩量回踩
            'enabled': True,
            'name': '突破缩量回踩',
            'type': 'short_term',
            'description': '横盘震荡后选择方向的票'
        },
        'short_term_3': {  # 短线策略③: 强势股10日线反抽
            'enabled': True,
            'name': '强势股10日线反抽',
            'type': 'short_term',
            'description': '强于大盘的板块龙头，短期回调'
        },
        'medium_term_1': {  # 中线策略①: 行业成长 + 均线多头
            'enabled': True,
            'name': '行业成长均线多头',
            'type': 'medium_term',
            'description': '行业景气股，偏中长线主做'
        },
        'medium_term_2': {  # 中线策略②: 困境反转 / 业绩拐点
            'enabled': True,
            'name': '困境反转业绩拐点',
            'type': 'medium_term',
            'description': '业绩由差转好、订单/政策催化'
        },
        'medium_term_3': {  # 中线策略③: 高股息红利慢牛
            'enabled': True,
            'name': '高股息红利慢牛',
            'type': 'medium_term',
            'description': '震荡市/偏弱市，求稳健'
        },
        'long_term_1': {  # 长线策略①: 优质白马龙头
            'enabled': True,
            'name': '优质白马龙头',
            'type': 'long_term',
            'description': '连续3年ROE≥15%，净利↑，现金流好'
        },
        'long_term_2': {  # 长线策略②: 红利再投收息策略
            'enabled': True,
            'name': '红利再投收息策略',
            'type': 'long_term',
            'description': '不想频繁操作，重视现金流'
        },
        'long_term_3': {  # 长线策略③: 真成长PEG低吸
            'enabled': True,
            'name': '真成长PEG低吸',
            'type': 'long_term',
            'description': '能接受波动、愿研究行业和增速'
        }
    }

def get_config() -> Dict:
    """获取完整配置"""
    return {
        'data_source': DataSourceConfig,
        'stock_filter': StockFilterConfig,
        'strategy': StrategyConfig
    }