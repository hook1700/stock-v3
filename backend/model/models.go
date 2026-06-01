package model

import (
	"time"
)

// Stock 股票基本信息
type Stock struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	Code        string    `gorm:"type:varchar(20);unique_index" json:"code"`
	Name        string    `gorm:"type:varchar(100)" json:"name"`
	Industry    string    `gorm:"type:varchar(100)" json:"industry"`
	Market      string    `gorm:"type:varchar(10)" json:"market"`
	ListingDate time.Time `json:"listing_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StockDailyData 股票日线数据
type StockDailyData struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	StockCode   string    `gorm:"type:varchar(20);index" json:"stock_code"`
	TradeDate   time.Time `gorm:"index" json:"trade_date"`
	OpenPrice   float64   `gorm:"type:decimal(10,3)" json:"open_price"`
	HighPrice   float64   `gorm:"type:decimal(10,3)" json:"high_price"`
	LowPrice    float64   `gorm:"type:decimal(10,3)" json:"low_price"`
	ClosePrice  float64   `gorm:"type:decimal(10,3)" json:"close_price"`
	Volume      int64     `json:"volume"`
	Amount      float64   `gorm:"type:decimal(15,2)" json:"amount"`
	TurnoverRate float64  `gorm:"type:decimal(8,4)" json:"turnover_rate"`
	PERatio     float64   `gorm:"type:decimal(10,4)" json:"pe_ratio"`
	PBRatio     float64   `gorm:"type:decimal(10,4)" json:"pb_ratio"`
	CreatedAt   time.Time `json:"created_at"`
}

// TechnicalIndicator 技术指标
type TechnicalIndicator struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	StockCode  string    `gorm:"type:varchar(20);index" json:"stock_code"`
	TradeDate  time.Time `gorm:"index" json:"trade_date"`
	MA5        float64   `gorm:"type:decimal(10,3)" json:"ma5"`
	MA10       float64   `gorm:"type:decimal(10,3)" json:"ma10"`
	MA20       float64   `gorm:"type:decimal(10,3)" json:"ma20"`
	MA60       float64   `gorm:"type:decimal(10,3)" json:"ma60"`
	VolumeMA5  int64     `json:"volume_ma5"`
	VolumeMA10 int64     `json:"volume_ma10"`
	RSI        float64   `gorm:"type:decimal(8,4)" json:"rsi"`
	MACD       float64   `gorm:"type:decimal(10,4)" json:"macd"`
	KDJK       float64   `gorm:"type:decimal(8,4)" json:"kdj_k"`
	KDJD       float64   `gorm:"type:decimal(8,4)" json:"kdj_d"`
	KDJJ       float64   `gorm:"type:decimal(8,4)" json:"kdj_j"`
	CreatedAt  time.Time `json:"created_at"`
}

// Strategy 策略配置
type Strategy struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	StrategyID  string    `gorm:"type:varchar(50);unique_index" json:"strategy_id"`
	Name        string    `gorm:"type:varchar(100)" json:"name"`
	StrategyType string   `gorm:"type:varchar(20)" json:"strategy_type"`
	Description string    `gorm:"type:text" json:"description"`
	Enabled     bool      `json:"enabled"`
	Parameters  string    `gorm:"type:jsonb" json:"parameters"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StrategyResult 策略执行结果
type StrategyResult struct {
	ID               uint      `gorm:"primary_key" json:"id"`
	StrategyID       string    `gorm:"index" json:"strategy_id"`
	StrategyType     string    `gorm:"index" json:"strategy_type"`     // 策略类型：short_term, medium_term, long_term
	TradeDate        time.Time `gorm:"index" json:"trade_date"`
	StockCode        string    `gorm:"index" json:"stock_code"`
	Score            float64   `gorm:"type:decimal(8,4)" json:"score"`
	BuyPrice         float64   `gorm:"type:decimal(10,3)" json:"buy_price"`
	StopLossPrice    float64   `gorm:"type:decimal(10,3)" json:"stop_loss_price"`
	TakeProfitPrice  float64   `gorm:"type:decimal(10,3)" json:"take_profit_price"`
	LogicDescription string    `gorm:"type:text" json:"logic_description"`
	Indicators       string    `gorm:"type:jsonb" json:"indicators"`
	CreatedAt        time.Time `json:"created_at"`
}

// SectorFundFlow 板块资金流
type SectorFundFlow struct {
	ID              uint      `gorm:"primary_key" json:"id"`
	SectorName      string    `gorm:"index" json:"sector_name"`
	TradeDate       time.Time `gorm:"index" json:"trade_date"`
	NetInflow       float64   `gorm:"type:decimal(15,2)" json:"net_inflow"`
	MainNetInflow   float64   `gorm:"type:decimal(15,2)" json:"main_net_inflow"`
	RetailNetInflow float64   `gorm:"type:decimal(15,2)" json:"retail_net_inflow"`
	TurnoverRate    float64   `gorm:"type:decimal(8,4)" json:"turnover_rate"`
	CreatedAt       time.Time `json:"created_at"`
}

// UserOperation 用户操作记录
type UserOperation struct {
	ID            uint      `gorm:"primary_key" json:"id"`
	UserID        string    `gorm:"type:varchar(50);index" json:"user_id"`
	OperationType string    `gorm:"type:varchar(50)" json:"operation_type"`
	Parameters    string    `gorm:"type:jsonb" json:"parameters"`
	ResultCount   int       `json:"result_count"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// FinancialData 财务数据
type FinancialData struct {
	ID                uint      `gorm:"primary_key" json:"id"`
	StockCode         string    `gorm:"type:varchar(20);index" json:"stock_code"`
	Quarter           string    `gorm:"type:varchar(10);index" json:"quarter"` // 例如: 2024Q1
	ReportDate        time.Time `json:"report_date"`
	
	// 盈利能力
	ROE               float64   `gorm:"type:decimal(8,4)" json:"roe"`                // 净资产收益率
	ROA               float64   `gorm:"type:decimal(8,4)" json:"roa"`                // 总资产收益率
	GrossMargin       float64   `gorm:"type:decimal(8,4)" json:"gross_margin"`       // 毛利率
	NetMargin         float64   `gorm:"type:decimal(8,4)" json:"net_margin"`        // 净利率
	
	// 成长性
	RevenueGrowth     float64   `gorm:"type:decimal(8,4)" json:"revenue_growth"`     // 营收增长率
	ProfitGrowth     float64   `gorm:"type:decimal(8,4)" json:"profit_growth"`     // 利润增长率
	LastQuarterGrowth float64   `gorm:"type:decimal(8,4)" json:"last_quarter_growth"` // 最近季度增长
	
	// 财务健康
	DebtRatio         float64   `gorm:"type:decimal(8,4)" json:"debt_ratio"`         // 资产负债率
	CurrentRatio      float64   `gorm:"type:decimal(8,4)" json:"current_ratio"`      // 流动比率
	CashFlow          float64   `gorm:"type:decimal(15,2)" json:"cash_flow"`        // 经营现金流
	
	// 估值与分红
	DividendYield    float64   `gorm:"type:decimal(8,4)" json:"dividend_yield"`    // 股息率
	PE               float64   `gorm:"type:decimal(10,4)" json:"pe"`               // 市盈率
	PB               float64   `gorm:"type:decimal(10,4)" json:"pb"`               // 市净率
	
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StockFilterRequest 股票筛选请求
type StockFilterRequest struct {
	StockCodes   []string `json:"stock_codes"`
	NameKeyword  string   `json:"name_keyword"`
	Industry     string   `json:"industry"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	SortBy       string   `json:"sort_by"`
	SortOrder    string   `json:"sort_order"`
	Page         int      `json:"page"`
	PageSize     int      `json:"page_size"`
}

// StockListResponse 股票列表响应
type StockListResponse struct {
	Stocks     []StockListItem `json:"stocks"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

// StockListItem 股票列表项
type StockListItem struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Industry    string  `json:"industry"`
	ClosePrice  float64 `json:"close_price"`
	ChangeRate  float64 `json:"change_rate"`
	Amount      float64 `json:"amount"`
	TurnoverRate float64 `json:"turnover_rate"`
}

// StrategyResultResponse 策略结果响应
type StrategyResultResponse struct {
	StrategyID      string                `json:"strategy_id"`
	StrategyName    string                `json:"strategy_name"`
	TradeDate       string                `json:"trade_date"`
	Results         []StrategyResultItem  `json:"results"`
	TotalCount      int                   `json:"total_count"`
}

// StrategyResultItem 策略结果项
type StrategyResultItem struct {
	StockCode       string  `json:"stock_code"`
	StockName       string  `json:"stock_name"`
	Score           float64 `json:"score"`
	BuyPrice        float64 `json:"buy_price"`
	StopLossPrice   float64 `json:"stop_loss_price"`
	TakeProfitPrice float64 `json:"take_profit_price"`
	Logic           string  `json:"logic"`
}

// FundFlowResponse 资金流响应
type FundFlowResponse struct {
	SectorName      string  `json:"sector_name"`
	TodayInflow     float64 `json:"today_inflow"`
	FiveDaysInflow  float64 `json:"five_days_inflow"`
	TenDaysInflow   float64 `json:"ten_days_inflow"`
	FifteenDaysInflow float64 `json:"fifteen_days_inflow"`
	TwentyDaysInflow float64 `json:"twenty_days_inflow"`
}