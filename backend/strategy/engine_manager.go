package strategy

import (
	"fmt"
	"log"
	"time"

	"stock-strategy-backend/model"
)

// StrategyEngineManager 策略引擎管理器
type StrategyEngineManager struct {
	shortTermStrategies map[string]*ShortTermStrategy
	mediumTermStrategies map[string]*MediumTermStrategy
	longTermStrategies   map[string]*LongTermStrategy
	stockRepo           *model.StockRepository
	dailyDataRepo       *model.StockDailyDataRepository
	resultRepo          *model.StrategyResultRepository
}

// NewStrategyEngineManager 创建策略引擎管理器
func NewStrategyEngineManager() *StrategyEngineManager {
	return &StrategyEngineManager{
		shortTermStrategies: make(map[string]*ShortTermStrategy),
		mediumTermStrategies: make(map[string]*MediumTermStrategy),
		longTermStrategies:   make(map[string]*LongTermStrategy),
		stockRepo:           &model.StockRepository{},
		dailyDataRepo:       &model.StockDailyDataRepository{},
		resultRepo:          &model.StrategyResultRepository{},
	}
}

// InitializeStrategies 初始化所有策略
func (m *StrategyEngineManager) InitializeStrategies() error {
	log.Println("初始化策略引擎...")

	// 初始化短线策略
	shortTermIDs := []string{"short_term_1", "short_term_2", "short_term_3"}
	for _, id := range shortTermIDs {
		m.shortTermStrategies[id] = NewShortTermStrategy(id)
	}

	// 初始化中线策略
	mediumTermIDs := []string{"medium_term_1", "medium_term_2", "medium_term_3"}
	for _, id := range mediumTermIDs {
		m.mediumTermStrategies[id] = NewMediumTermStrategy(id)
	}

	// 初始化长线策略
	longTermIDs := []string{"long_term_1", "long_term_2", "long_term_3"}
	for _, id := range longTermIDs {
		m.longTermStrategies[id] = NewLongTermStrategy(id)
	}

	log.Printf("策略引擎初始化完成: %d个短线策略, %d个中线策略, %d个长线策略",
		len(m.shortTermStrategies), len(m.mediumTermStrategies), len(m.longTermStrategies))

	return nil
}

// ExecuteAllStrategies 执行所有策略
func (m *StrategyEngineManager) ExecuteAllStrategies(tradeDate time.Time) (map[string][]*model.StrategyResult, error) {
	log.Printf("开始执行所有策略，交易日期: %s", tradeDate.Format("2006-01-02"))

	results := make(map[string][]*model.StrategyResult)

	// 获取所有股票数据
	stocks, err := m.stockRepo.GetAllStocks()
	if err != nil {
		return nil, err
	}

	// 获取流动性股票（简化实现：取前100只）
	liquidStocks := m.filterLiquidStocks(stocks)

	log.Printf("处理 %d 只流动性股票", len(liquidStocks))

	// 执行短线策略
	shortTermResults, err := m.executeShortTermStrategies(liquidStocks, tradeDate)
	if err != nil {
		log.Printf("短线策略执行失败: %v", err)
	} else {
		results["short_term"] = shortTermResults
		log.Printf("短线策略执行完成: %d 个信号", len(shortTermResults))
	}

	// 执行中线策略
	mediumTermResults, err := m.executeMediumTermStrategies(liquidStocks, tradeDate)
	if err != nil {
		log.Printf("中线策略执行失败: %v", err)
	} else {
		results["medium_term"] = mediumTermResults
		log.Printf("中线策略执行完成: %d 个信号", len(mediumTermResults))
	}

	// 执行长线策略
	longTermResults, err := m.executeLongTermStrategies(liquidStocks, tradeDate)
	if err != nil {
		log.Printf("长线策略执行失败: %v", err)
	} else {
		results["long_term"] = longTermResults
		log.Printf("长线策略执行完成: %d 个信号", len(longTermResults))
	}

	// 保存结果到数据库
	totalSignals := len(shortTermResults) + len(mediumTermResults) + len(longTermResults)
	if err := m.saveStrategyResults(results, tradeDate); err != nil {
		log.Printf("保存策略结果失败: %v", err)
	} else {
		log.Printf("策略结果保存完成，共 %d 个信号", totalSignals)
	}

	return results, nil
}

// ExecuteSingleStrategy 执行单个策略
func (m *StrategyEngineManager) ExecuteSingleStrategy(strategyID string, tradeDate time.Time) ([]*model.StrategyResult, error) {
	log.Printf("执行单个策略: %s, 日期: %s", strategyID, tradeDate.Format("2006-01-02"))

	// 获取股票数据
	stocks, err := m.stockRepo.GetAllStocks()
	if err != nil {
		return nil, err
	}

	liquidStocks := m.filterLiquidStocks(stocks)

	var results []*model.StrategyResult

	// 根据策略类型执行
	if strategy, exists := m.shortTermStrategies[strategyID]; exists {
		results, err = m.executeShortTermStrategy(strategy, liquidStocks, tradeDate)
	} else if strategy, exists := m.mediumTermStrategies[strategyID]; exists {
		results, err = m.executeMediumTermStrategy(strategy, liquidStocks, tradeDate)
	} else if strategy, exists := m.longTermStrategies[strategyID]; exists {
		results, err = m.executeLongTermStrategy(strategy, liquidStocks, tradeDate)
	} else {
		return nil, fmt.Errorf("策略不存在: %s", strategyID)
	}

	if err != nil {
		return nil, err
	}

	// 保存结果
	if err := m.saveSingleStrategyResults(strategyID, results, tradeDate); err != nil {
		log.Printf("保存策略结果失败: %v", err)
	}

	log.Printf("策略 %s 执行完成: %d 个信号", strategyID, len(results))
	return results, nil
}

// executeShortTermStrategies 执行所有短线策略
func (m *StrategyEngineManager) executeShortTermStrategies(stocks []model.Stock, tradeDate time.Time) ([]*model.StrategyResult, error) {
	var allResults []*model.StrategyResult

	for strategyID, strategy := range m.shortTermStrategies {
		results, err := m.executeShortTermStrategy(strategy, stocks, tradeDate)
		if err != nil {
			log.Printf("短线策略 %s 执行失败: %v", strategyID, err)
			continue
		}

		allResults = append(allResults, results...)
		log.Printf("短线策略 %s 完成: %d 个信号", strategyID, len(results))
	}

	return allResults, nil
}

// executeMediumTermStrategies 执行所有中线策略
func (m *StrategyEngineManager) executeMediumTermStrategies(stocks []model.Stock, tradeDate time.Time) ([]*model.StrategyResult, error) {
	var allResults []*model.StrategyResult

	for strategyID, strategy := range m.mediumTermStrategies {
		results, err := m.executeMediumTermStrategy(strategy, stocks, tradeDate)
		if err != nil {
			log.Printf("中线策略 %s 执行失败: %v", strategyID, err)
			continue
		}

		allResults = append(allResults, results...)
		log.Printf("中线策略 %s 完成: %d 个信号", strategyID, len(results))
	}

	return allResults, nil
}

// executeLongTermStrategies 执行所有长线策略
func (m *StrategyEngineManager) executeLongTermStrategies(stocks []model.Stock, tradeDate time.Time) ([]*model.StrategyResult, error) {
	var allResults []*model.StrategyResult

	for strategyID, strategy := range m.longTermStrategies {
		results, err := m.executeLongTermStrategy(strategy, stocks, tradeDate)
		if err != nil {
			log.Printf("长线策略 %s 执行失败: %v", strategyID, err)
			continue
		}

		allResults = append(allResults, results...)
		log.Printf("长线策略 %s 完成: %d 个信号", strategyID, len(results))
	}

	return allResults, nil
}

// executeShortTermStrategy 执行单个短线策略
func (m *StrategyEngineManager) executeShortTermStrategy(strategy *ShortTermStrategy, stocks []model.Stock, tradeDate time.Time) ([]*model.StrategyResult, error) {
	var results []*model.StrategyResult

	for _, stock := range stocks {
		// 获取股票日线数据
		dailyData, err := m.getStockDailyData(stock.Code, tradeDate, 60) // 获取60天数据
		if err != nil {
			continue
		}

		// 获取技术指标
		indicators, err := m.getTechnicalIndicators(stock.Code, tradeDate)
		if err != nil {
			continue
		}

		// 执行策略
		result := strategy.Execute(&stock, dailyData, indicators)
		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// executeMediumTermStrategy 执行单个中线策略
func (m *StrategyEngineManager) executeMediumTermStrategy(strategy *MediumTermStrategy, stocks []model.Stock, tradeDate time.Time) ([]*model.StrategyResult, error) {
	var results []*model.StrategyResult

	for _, stock := range stocks {
		// 获取股票日线数据
		dailyData, err := m.getStockDailyData(stock.Code, tradeDate, 120) // 获取120天数据
		if err != nil {
			continue
		}

		// 获取财务数据
		financialData, err := m.getFinancialData(stock.Code)
		if err != nil {
			continue
		}

		// 获取技术指标
		indicators, err := m.getTechnicalIndicators(stock.Code, tradeDate)
		if err != nil {
			continue
		}

		// 执行策略
		result := strategy.Execute(&stock, dailyData, financialData, indicators)
		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// executeLongTermStrategy 执行单个长线策略
func (m *StrategyEngineManager) executeLongTermStrategy(strategy *LongTermStrategy, stocks []model.Stock, tradeDate time.Time) ([]*model.StrategyResult, error) {
	var results []*model.StrategyResult

	for _, stock := range stocks {
		// 获取股票日线数据
		dailyData, err := m.getStockDailyData(stock.Code, tradeDate, 250) // 获取250天数据
		if err != nil {
			continue
		}

		// 获取财务数据
		financialData, err := m.getFinancialData(stock.Code)
		if err != nil {
			continue
		}

		// 获取技术指标
		indicators, err := m.getTechnicalIndicators(stock.Code, tradeDate)
		if err != nil {
			continue
		}

		// 执行策略
		result := strategy.Execute(&stock, dailyData, financialData, indicators)
		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// filterLiquidStocks 筛选流动性股票
func (m *StrategyEngineManager) filterLiquidStocks(stocks []model.Stock) []model.Stock {
	var liquidStocks []model.Stock

	// 简化实现：取前100只股票
	// 实际实现应该根据成交额、市值等指标筛选
	maxStocks := 100
	if len(stocks) > maxStocks {
		liquidStocks = stocks[:maxStocks]
	} else {
		liquidStocks = stocks
	}

	return liquidStocks
}

// getStockDailyData 获取股票日线数据
func (m *StrategyEngineManager) getStockDailyData(stockCode string, endDate time.Time, days int) ([]model.StockDailyData, error) {
	startDate := endDate.AddDate(0, 0, -days)
	return m.dailyDataRepo.GetDailyData(stockCode, startDate, endDate)
}

// getFinancialData 获取财务数据
func (m *StrategyEngineManager) getFinancialData(stockCode string) (*model.FinancialData, error) {
	// 这里需要实现财务数据获取逻辑
	// 简化实现：返回空数据
	return &model.FinancialData{}, nil
}

// getTechnicalIndicators 获取技术指标
func (m *StrategyEngineManager) getTechnicalIndicators(stockCode string, date time.Time) ([]model.TechnicalIndicator, error) {
	// 这里需要实现技术指标获取逻辑
	// 简化实现：返回空数据
	return []model.TechnicalIndicator{}, nil
}

// saveStrategyResults 保存策略结果
func (m *StrategyEngineManager) saveStrategyResults(results map[string][]*model.StrategyResult, tradeDate time.Time) error {
	for strategyType, strategyResults := range results {
		for _, result := range strategyResults {
			// 设置策略类型和交易日期
			result.StrategyType = strategyType
			result.TradeDate = tradeDate

			// 保存到数据库
			if err := m.resultRepo.CreateResult(result); err != nil {
				log.Printf("保存策略结果失败: %v", err)
			}
		}
	}

	return nil
}

// saveSingleStrategyResults 保存单个策略结果
func (m *StrategyEngineManager) saveSingleStrategyResults(strategyID string, results []*model.StrategyResult, tradeDate time.Time) error {
	for _, result := range results {
		result.StrategyID = strategyID
		result.TradeDate = tradeDate

		if err := m.resultRepo.CreateResult(result); err != nil {
			log.Printf("保存策略结果失败: %v", err)
		}
	}

	return nil
}

// GetStrategyList 获取策略列表
func (m *StrategyEngineManager) GetStrategyList() []map[string]interface{} {
	var strategies []map[string]interface{}

	// 添加短线策略
	for id, strategy := range m.shortTermStrategies {
		strategies = append(strategies, map[string]interface{}{
			"id":          id,
			"name":        strategy.GetName(),
			"type":        "short_term",
			"description": strategy.GetDescription(),
			"parameters":  strategy.GetParameters(),
		})
	}

	// 添加中线策略
	for id, strategy := range m.mediumTermStrategies {
		strategies = append(strategies, map[string]interface{}{
			"id":          id,
			"name":        strategy.GetName(),
			"type":        "medium_term",
			"description": strategy.GetDescription(),
			"parameters":  strategy.GetParameters(),
		})
	}

	// 添加长线策略
	for id, strategy := range m.longTermStrategies {
		strategies = append(strategies, map[string]interface{}{
			"id":          id,
			"name":        strategy.GetName(),
			"type":        "long_term",
			"description": strategy.GetDescription(),
			"parameters":  strategy.GetParameters(),
		})
	}

	return strategies
}

// GetStrategyPerformance 获取策略历史表现
func (m *StrategyEngineManager) GetStrategyPerformance(strategyID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	// 获取策略结果
	results, err := m.resultRepo.GetResultsByStrategyAndPeriod(strategyID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 计算表现指标
	performance := m.calculateStrategyPerformance(results)

	return performance, nil
}

// calculateStrategyPerformance 计算策略表现
func (m *StrategyEngineManager) calculateStrategyPerformance(results []model.StrategyResult) map[string]interface{} {
	if len(results) == 0 {
		return map[string]interface{}{
			"total_signals": 0,
			"avg_score":     0,
			"success_rate":  0,
			"avg_profit":    0,
		}
	}

	var totalScore float64
	var successCount int
	var totalProfit float64

	for _, result := range results {
		totalScore += result.Score

		// 简化实现：假设评分>0.7为成功
		if result.Score > 0.7 {
			successCount++
		}

		// 计算理论收益（简化实现）
		if result.TakeProfitPrice > result.BuyPrice {
			totalProfit += (result.TakeProfitPrice - result.BuyPrice) / result.BuyPrice
		}
	}

	return map[string]interface{}{
		"total_signals": len(results),
		"avg_score":     totalScore / float64(len(results)),
		"success_rate":  float64(successCount) / float64(len(results)),
		"avg_profit":    totalProfit / float64(len(results)),
	}
}

// Close 关闭策略引擎
func (m *StrategyEngineManager) Close() {
	log.Println("关闭策略引擎管理器")
	// 清理资源
}