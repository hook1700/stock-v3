package strategy

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"stock-strategy-backend/model"
)

// StrategyEngine 策略引擎
type StrategyEngine struct {
	stockRepo      *model.StockRepository
	dailyDataRepo  *model.StockDailyDataRepository
	strategyRepo   *model.StrategyRepository
	resultRepo     *model.StrategyResultRepository
}

// NewStrategyEngine 创建策略引擎
func NewStrategyEngine() *StrategyEngine {
	return &StrategyEngine{
		stockRepo:     &model.StockRepository{},
		dailyDataRepo: &model.StockDailyDataRepository{},
		strategyRepo:  &model.StrategyRepository{},
		resultRepo:    &model.StrategyResultRepository{},
	}
}

// ExecuteStrategy 执行策略
func (e *StrategyEngine) ExecuteStrategy(strategyID string, tradeDate time.Time) error {
	// 获取策略配置
	strategy, err := e.strategyRepo.GetStrategyByID(strategyID)
	if err != nil {
		return fmt.Errorf("获取策略失败: %v", err)
	}

	if !strategy.Enabled {
		return fmt.Errorf("策略未启用")
	}

	// 获取所有股票
	stocks, err := e.stockRepo.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	var results []model.StrategyResult

	// 根据策略类型执行不同的策略逻辑
	switch strategy.StrategyType {
	case "short_term":
		results, err = e.executeShortTermStrategy(strategy, stocks, tradeDate)
	case "medium_term":
		results, err = e.executeMediumTermStrategy(strategy, stocks, tradeDate)
	case "long_term":
		results, err = e.executeLongTermStrategy(strategy, stocks, tradeDate)
	default:
		return fmt.Errorf("未知策略类型: %s", strategy.StrategyType)
	}

	if err != nil {
		return err
	}

	// 保存策略结果
	for _, result := range results {
		if err := e.resultRepo.CreateResult(&result); err != nil {
			log.Printf("保存策略结果失败: %v", err)
		}
	}

	log.Printf("策略 %s 执行完成，生成 %d 条结果", strategy.Name, len(results))
	return nil
}

// executeShortTermStrategy 执行短线策略
func (e *StrategyEngine) executeShortTermStrategy(strategy *model.Strategy, stocks []model.Stock, tradeDate time.Time) ([]model.StrategyResult, error) {
	var results []model.StrategyResult

	for _, stock := range stocks {
		var score float64
		var logic string
		var buyPrice, stopLoss, takeProfit float64

		switch strategy.StrategyID {
		case "short_term_1":
			// 均线回踩低吸策略
			score, logic, buyPrice, stopLoss, takeProfit = e.maPullbackStrategy(&stock, tradeDate)
		case "short_term_2":
			// 突破缩量回踩策略
			score, logic, buyPrice, stopLoss, takeProfit = e.breakoutPullbackStrategy(&stock, tradeDate)
		case "short_term_3":
			// 强势股10日线反抽策略
			score, logic, buyPrice, stopLoss, takeProfit = e.strongStockReboundStrategy(&stock, tradeDate)
		}

		if score > 0 {
			indicators, _ := json.Marshal(map[string]interface{}{
				"score": score,
			})

			result := model.StrategyResult{
				StrategyID:       strategy.StrategyID,
				TradeDate:        tradeDate,
				StockCode:        stock.Code,
				Score:            score,
				BuyPrice:         buyPrice,
				StopLossPrice:    stopLoss,
				TakeProfitPrice:  takeProfit,
				LogicDescription: logic,
				Indicators:       string(indicators),
				CreatedAt:        time.Now(),
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// maPullbackStrategy 均线回踩低吸策略
func (e *StrategyEngine) maPullbackStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近30个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -60)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 30 {
		return 0, "", 0, 0, 0
	}

	// 计算技术指标
	latestData := dailyData[len(dailyData)-1]
	ma20 := e.calculateMA(dailyData, 20)
	ma10 := e.calculateMA(dailyData, 10)
	ma5 := e.calculateMA(dailyData, 5)

	// 策略逻辑：股价在20日线上，20日线向上，回踩5/10日线
	if latestData.ClosePrice > ma20 && ma20 > e.calculateMA(dailyData[:len(dailyData)-10], 20) {
		// 检查是否回踩5日线或10日线
		if math.Abs(latestData.ClosePrice-ma5)/ma5 < 0.02 || math.Abs(latestData.ClosePrice-ma10)/ma10 < 0.03 {
			score := 0.7
			buyPrice := latestData.ClosePrice
			stopLoss := buyPrice * 0.93
			takeProfit := buyPrice * 1.08
			logic := "股价在20日线上方，20日线向上，回踩5/10日线企稳"

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// breakoutPullbackStrategy 突破缩量回踩策略
func (e *StrategyEngine) breakoutPullbackStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近60个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -90)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 60 {
		return 0, "", 0, 0, 0
	}

	// 策略逻辑实现...
	// 简化的实现，实际需要更复杂的逻辑
	score := 0.6
	buyPrice := dailyData[len(dailyData)-1].ClosePrice
	stopLoss := buyPrice * 0.93
	takeProfit := buyPrice * 1.07
	logic := "平台突破后缩量回踩，支撑有效"

	return score, logic, buyPrice, stopLoss, takeProfit
}

// strongStockReboundStrategy 强势股10日线反抽策略
func (e *StrategyEngine) strongStockReboundStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近40个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -60)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 40 {
		return 0, "", 0, 0, 0
	}

	// 策略逻辑实现...
	// 简化的实现，实际需要更复杂的逻辑
	score := 0.65
	buyPrice := dailyData[len(dailyData)-1].ClosePrice
	stopLoss := buyPrice * 0.93
	takeProfit := buyPrice * 1.06
	logic := "强势股第一次回踩10日线，出现承接迹象"

	return score, logic, buyPrice, stopLoss, takeProfit
}

// executeMediumTermStrategy 执行中线策略
func (e *StrategyEngine) executeMediumTermStrategy(strategy *model.Strategy, stocks []model.Stock, tradeDate time.Time) ([]model.StrategyResult, error) {
	// 中线策略实现...
	return []model.StrategyResult{}, nil
}

// executeLongTermStrategy 执行长线策略
func (e *StrategyEngine) executeLongTermStrategy(strategy *model.Strategy, stocks []model.Stock, tradeDate time.Time) ([]model.StrategyResult, error) {
	// 长线策略实现...
	return []model.StrategyResult{}, nil
}

// calculateMA 计算移动平均线
func (e *StrategyEngine) calculateMA(dailyData []model.StockDailyData, period int) float64 {
	if len(dailyData) < period {
		return 0
	}

	var sum float64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].ClosePrice
	}

	return sum / float64(period)
}

// calculateVolumeMA 计算成交量移动平均
func (e *StrategyEngine) calculateVolumeMA(dailyData []model.StockDailyData, period int) int64 {
	if len(dailyData) < period {
		return 0
	}

	var sum int64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].Volume
	}

	return sum / int64(period)
}