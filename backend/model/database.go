package model

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"stock-strategy-backend/config"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	dbConfig := cfg.Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)

	var err error
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// 设置连接池
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	DB.DB().SetConnMaxLifetime(time.Hour)

	// 启用Logger
	DB.LogMode(true)

	// 自动迁移表结构
	DB.AutoMigrate(
		&Stock{},
		&StockDailyData{},
		&TechnicalIndicator{},
		&Strategy{},
		&StrategyResult{},
		&SectorFundFlow{},
		&UserOperation{},
	)

	log.Println("Database connected successfully")
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// StockRepository 股票数据仓库
type StockRepository struct{}

func (r *StockRepository) CreateStock(stock *Stock) error {
	return DB.Create(stock).Error
}

func (r *StockRepository) GetStockByCode(code string) (*Stock, error) {
	var stock Stock
	err := DB.Where("code = ?", code).First(&stock).Error
	return &stock, err
}

func (r *StockRepository) GetStocksByIndustry(industry string) ([]Stock, error) {
	var stocks []Stock
	err := DB.Where("industry = ?", industry).Find(&stocks).Error
	return stocks, err
}

func (r *StockRepository) GetAllStocks() ([]Stock, error) {
	var stocks []Stock
	err := DB.Find(&stocks).Error
	return stocks, err
}

// StockDailyDataRepository 股票日线数据仓库
type StockDailyDataRepository struct{}

func (r *StockDailyDataRepository) CreateDailyData(data *StockDailyData) error {
	return DB.Create(data).Error
}

func (r *StockDailyDataRepository) GetDailyData(stockCode string, startDate, endDate time.Time) ([]StockDailyData, error) {
	var data []StockDailyData
	err := DB.Where("stock_code = ? AND trade_date BETWEEN ? AND ?", stockCode, startDate, endDate).
		Order("trade_date").Find(&data).Error
	return data, err
}

func (r *StockDailyDataRepository) GetLatestData(stockCode string) (*StockDailyData, error) {
	var data StockDailyData
	err := DB.Where("stock_code = ?", stockCode).Order("trade_date DESC").First(&data).Error
	return &data, err
}

// StrategyRepository 策略仓库
type StrategyRepository struct{}

func (r *StrategyRepository) GetEnabledStrategies() ([]Strategy, error) {
	var strategies []Strategy
	err := DB.Where("enabled = ?", true).Find(&strategies).Error
	return strategies, err
}

func (r *StrategyRepository) GetStrategyByID(strategyID string) (*Strategy, error) {
	var strategy Strategy
	err := DB.Where("strategy_id = ?", strategyID).First(&strategy).Error
	return &strategy, err
}

// StrategyResultRepository 策略结果仓库
type StrategyResultRepository struct{}

func (r *StrategyResultRepository) CreateResult(result *StrategyResult) error {
	return DB.Create(result).Error
}

func (r *StrategyResultRepository) GetResultsByDateAndStrategy(tradeDate time.Time, strategyID string) ([]StrategyResult, error) {
	var results []StrategyResult
	err := DB.Where("trade_date = ? AND strategy_id = ?", tradeDate, strategyID).
		Order("score DESC").Find(&results).Error
	return results, err
}

func (r *StrategyResultRepository) GetResultsByStrategyAndPeriod(strategyID string, startDate, endDate time.Time) ([]StrategyResult, error) {
	var results []StrategyResult
	err := DB.Where("strategy_id = ? AND trade_date BETWEEN ? AND ?", strategyID, startDate, endDate).
		Order("trade_date DESC, score DESC").Find(&results).Error
	return results, err
}

// SectorFundFlowRepository 板块资金流仓库
type SectorFundFlowRepository struct{}

func (r *SectorFundFlowRepository) CreateFlow(flow *SectorFundFlow) error {
	return DB.Create(flow).Error
}

func (r *SectorFundFlowRepository) GetFlowsByDate(tradeDate time.Time) ([]SectorFundFlow, error) {
	var flows []SectorFundFlow
	err := DB.Where("trade_date = ?", tradeDate).Order("net_inflow DESC").Find(&flows).Error
	return flows, err
}

func (r *SectorFundFlowRepository) GetFlowsBySector(sectorName string, days int) ([]SectorFundFlow, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	var flows []SectorFundFlow
	err := DB.Where("sector_name = ? AND trade_date >= ?", sectorName, startDate).
		Order("trade_date").Find(&flows).Error
	return flows, err
}