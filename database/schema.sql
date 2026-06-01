-- 股票策略分析系统数据库表结构

-- 股票基本信息表
CREATE TABLE IF NOT EXISTS stocks (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,        -- 股票代码
    name VARCHAR(100) NOT NULL,              -- 股票名称
    industry VARCHAR(100),                   -- 所属行业
    market VARCHAR(10),                      -- 市场类型（SH/SZ/BJ）
    listing_date DATE,                       -- 上市日期
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 股票日线数据表
CREATE TABLE IF NOT EXISTS stock_daily_data (
    id SERIAL PRIMARY KEY,
    stock_code VARCHAR(20) NOT NULL,         -- 股票代码
    trade_date DATE NOT NULL,                -- 交易日期
    open_price DECIMAL(10,3),                -- 开盘价
    high_price DECIMAL(10,3),                -- 最高价
    low_price DECIMAL(10,3),                 -- 最低价
    close_price DECIMAL(10,3),               -- 收盘价
    volume BIGINT,                           -- 成交量
    amount DECIMAL(15,2),                    -- 成交额
    turnover_rate DECIMAL(8,4),              -- 换手率
    pe_ratio DECIMAL(10,4),                  -- 市盈率
    pb_ratio DECIMAL(10,4),                  -- 市净率
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stock_code, trade_date),
    FOREIGN KEY (stock_code) REFERENCES stocks(code) ON DELETE CASCADE
);

-- 技术指标表
CREATE TABLE IF NOT EXISTS technical_indicators (
    id SERIAL PRIMARY KEY,
    stock_code VARCHAR(20) NOT NULL,         -- 股票代码
    trade_date DATE NOT NULL,                -- 交易日期
    ma5 DECIMAL(10,3),                       -- 5日均价
    ma10 DECIMAL(10,3),                      -- 10日均价
    ma20 DECIMAL(10,3),                      -- 20日均价
    ma60 DECIMAL(10,3),                      -- 60日均价
    volume_ma5 BIGINT,                       -- 5日平均成交量
    volume_ma10 BIGINT,                      -- 10日平均成交量
    rsi DECIMAL(8,4),                        -- RSI指标
    macd DECIMAL(10,4),                      -- MACD指标
    kdj_k DECIMAL(8,4),                      -- KDJ-K值
    kdj_d DECIMAL(8,4),                      -- KDJ-D值
    kdj_j DECIMAL(8,4),                      -- KDJ-J值
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stock_code, trade_date),
    FOREIGN KEY (stock_code) REFERENCES stocks(code) ON DELETE CASCADE
);

-- 策略配置表
CREATE TABLE IF NOT EXISTS strategies (
    id SERIAL PRIMARY KEY,
    strategy_id VARCHAR(50) NOT NULL UNIQUE, -- 策略标识
    name VARCHAR(100) NOT NULL,              -- 策略名称
    strategy_type VARCHAR(20) NOT NULL,      -- 策略类型（short_term/medium_term/long_term）
    description TEXT,                        -- 策略描述
    enabled BOOLEAN DEFAULT TRUE,            -- 是否启用
    parameters JSONB,                        -- 策略参数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 策略执行结果表
CREATE TABLE IF NOT EXISTS strategy_results (
    id SERIAL PRIMARY KEY,
    strategy_id VARCHAR(50) NOT NULL,        -- 策略标识
    trade_date DATE NOT NULL,                -- 执行日期
    stock_code VARCHAR(20) NOT NULL,         -- 股票代码
    score DECIMAL(8,4),                      -- 策略评分
    buy_price DECIMAL(10,3),                 -- 建议买入价
    stop_loss_price DECIMAL(10,3),           -- 止损价
    take_profit_price DECIMAL(10,3),         -- 止盈价
    logic_description TEXT,                  -- 逻辑说明
    indicators JSONB,                        -- 技术指标数据
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (strategy_id) REFERENCES strategies(strategy_id) ON DELETE CASCADE,
    FOREIGN KEY (stock_code) REFERENCES stocks(code) ON DELETE CASCADE,
    UNIQUE(strategy_id, trade_date, stock_code)
);

-- 板块资金流数据表
CREATE TABLE IF NOT EXISTS sector_fund_flow (
    id SERIAL PRIMARY KEY,
    sector_name VARCHAR(100) NOT NULL,       -- 板块名称
    trade_date DATE NOT NULL,                -- 交易日期
    net_inflow DECIMAL(15,2),                -- 净流入金额
    main_net_inflow DECIMAL(15,2),           -- 主力净流入
    retail_net_inflow DECIMAL(15,2),         -- 散户净流入
    turnover_rate DECIMAL(8,4),              -- 换手率
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(sector_name, trade_date)
);

-- 用户操作记录表
CREATE TABLE IF NOT EXISTS user_operations (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) DEFAULT 'system',    -- 用户ID
    operation_type VARCHAR(50) NOT NULL,     -- 操作类型（search/filter/view）
    parameters JSONB,                        -- 操作参数
    result_count INTEGER,                    -- 结果数量
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_stock_daily_data_code_date ON stock_daily_data(stock_code, trade_date);
CREATE INDEX IF NOT EXISTS idx_strategy_results_date_strategy ON strategy_results(trade_date, strategy_id);
CREATE INDEX IF NOT EXISTS idx_strategy_results_stock ON strategy_results(stock_code);
CREATE INDEX IF NOT EXISTS idx_sector_fund_flow_date ON sector_fund_flow(trade_date);
CREATE INDEX IF NOT EXISTS idx_user_operations_created_at ON user_operations(created_at);

-- 插入默认策略配置
INSERT INTO strategies (strategy_id, name, strategy_type, description, parameters) VALUES
('short_term_1', '均线回踩低吸', 'short_term', '趋势热点股、板块处上升期', '{"ma_periods": [5, 10, 20], "volume_ratio": 1.2}'),
('short_term_2', '突破缩量回踩', 'short_term', '横盘震荡后选择方向的票', '{"box_period": 20, "breakout_volume_ratio": 1.5}'),
('short_term_3', '强势股10日线反抽', 'short_term', '强于大盘的板块龙头，短期回调', '{"compare_period": 20, "support_ma": 10}'),
('medium_term_1', '行业成长均线多头', 'medium_term', '行业景气股，偏中长线主做', '{"ma_sequence": [5, 10, 20, 60], "roe_threshold": 10}'),
('medium_term_2', '困境反转业绩拐点', 'medium_term', '业绩由差转好、订单/政策催化', '{"decline_threshold": 30, "recovery_period": 2}'),
('medium_term_3', '高股息红利慢牛', 'medium_term', '震荡市/偏弱市，求稳健', '{"dividend_yield": 4.5, "roe_stability": 5}'),
('long_term_1', '优质白马龙头', 'long_term', '连续3年ROE≥15%，净利↑，现金流好', '{"roe_years": 3, "roe_threshold": 15}'),
('long_term_2', '红利再投收息策略', 'long_term', '不想频繁操作，重视现金流', '{"dividend_years": 5, "yield_threshold": 4.5}'),
('long_term_3', '真成长PEG低吸', 'long_term', '能接受波动、愿研究行业和增速', '{"peg_threshold": 1, "growth_threshold": 20}');

-- 创建更新时间的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要更新时间的表创建触发器
CREATE TRIGGER update_stocks_updated_at BEFORE UPDATE ON stocks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_strategies_updated_at BEFORE UPDATE ON strategies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();