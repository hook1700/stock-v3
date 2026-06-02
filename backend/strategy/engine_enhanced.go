package strategy

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"stock-strategy-backend/model"
)

// StrategyEngineEnhanced 增强版策略引擎
type StrategyEngineEnhanced struct {
	stockRepo     *model.StockRepository
	dailyDataRepo *model.StockDailyDataRepository
	strategyRepo  *model.StrategyRepository
	resultRepo    *model.StrategyResultRepository
}

// NewStrategyEngineEnhanced 创建增强版策略引擎
func NewStrategyEngineEnhanced() *StrategyEngineEnhanced {
	return &StrategyEngineEnhanced{
		stockRepo:     &model.StockRepository{},
		dailyDataRepo: &model.StockDailyDataRepository{},
		strategyRepo:  &model.StrategyRepository{},
		resultRepo:    &model.StrategyResultRepository{},
	}
}

// ExecuteStrategy 执行策略
func (e *StrategyEngineEnhanced) ExecuteStrategy(strategyID string, tradeDate time.Time) error {
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

	// 按评分排序，只保留前20只股票
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > 20 {
		results = results[:20]
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
func (e *StrategyEngineEnhanced) executeShortTermStrategy(strategy *model.Strategy, stocks []model.Stock, tradeDate time.Time) ([]model.StrategyResult, error) {
	var results []model.StrategyResult

	for _, stock := range stocks {
		// 流动性筛选（日均成交额≥2亿）
		if !e.checkLiquidity(stock.Code, tradeDate, 200000000) {
			continue
		}

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

		if score > 0.5 { // 只保留评分>0.5的股票
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

// maPullbackStrategy 均线回踩低吸策略（完整实现）
func (e *StrategyEngineEnhanced) maPullbackStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近60个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -90)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 30 {
		return 0, "", 0, 0, 0
	}

	// 计算技术指标
	latestData := dailyData[len(dailyData)-1]
	ma20 := e.calculateMA(dailyData, 20)
	ma10 := e.calculateMA(dailyData, 10)
	ma5 := e.calculateMA(dailyData, 5)

	// 计算前10日的MA20用于趋势判断
	if len(dailyData) < 30 {
		return 0, "", 0, 0, 0
	}
	prevMA20 := e.calculateMA(dailyData[:len(dailyData)-10], 20)

	// 策略逻辑判断
	// 1. 股价在20日线上方
	// 2. 20日线趋势向上
	// 3. 回踩5/10日线（价格接近均线2-3%）
	if latestData.ClosePrice > ma20 && ma20 > prevMA20 {
		// 检查是否回踩5日线或10日线
		distanceToMA5 := math.Abs(latestData.ClosePrice-ma5) / ma5
		distanceToMA10 := math.Abs(latestData.ClosePrice-ma10) / ma10

		if distanceToMA5 < 0.02 || distanceToMA10 < 0.03 {
			// 成交量验证：近期放量上涨→缩量回踩
			_ = e.calculateVolumeMA(dailyData, 5) // volumeMA5
			_ = e.calculateVolumeMA(dailyData, 10) // volumeMA10

			// 检查成交量模式
			volumeScore := e.analyzeVolumePattern(dailyData)

			score := 0.7 + volumeScore*0.3 // 基础分0.7 + 成交量得分
			buyPrice := latestData.ClosePrice
			stopLoss := buyPrice * 0.93 // -7%止损
			takeProfit := buyPrice * 1.08 // +8%止盈
			logic := fmt.Sprintf("股价在20日线上方，20日线向上(%.2f→%.2f)，回踩%s线企稳，成交量模式良好",
				prevMA20, ma20, e.getClosestMA(latestData.ClosePrice, ma5, ma10))

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// breakoutPullbackStrategy 突破缩量回踩策略（完整实现）
func (e *StrategyEngineEnhanced) breakoutPullbackStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近90个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -120)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 60 {
		return 0, "", 0, 0, 0
	}

	// 识别箱体震荡区间（15-30日）
	boxData := dailyData[len(dailyData)-45:] // 最近45个交易日
	if len(boxData) < 30 {
		return 0, "", 0, 0, 0
	}

	// 计算箱体上下沿
	boxHigh, boxLow := e.calculateBoxRange(boxData[:30]) // 前30个交易日为箱体区间

	// 检查突破信号（最近15个交易日）
	breakoutData := boxData[30:]
	breakoutDetected := false
	var breakoutPrice, breakoutVolume float64

	for i, data := range breakoutData {
		if data.ClosePrice > boxHigh && data.Volume > int64(float64(e.calculateVolumeMA(boxData[:30], 20))*1.5) {
			breakoutDetected = true
			breakoutPrice = data.ClosePrice
			breakoutVolume = float64(data.Volume)
			break
		}

		// 如果已经检查了5个交易日还没突破，认为突破失败
		if i >= 5 {
			break
		}
	}

	if !breakoutDetected {
		return 0, "", 0, 0, 0
	}

	// 检查回踩（最近几个交易日）
	latestData := dailyData[len(dailyData)-1]
	pullbackRatio := math.Abs(latestData.ClosePrice-boxHigh) / boxHigh

	// 回踩确认：价格接近原压力位，成交量萎缩
	if pullbackRatio < 0.03 && float64(latestData.Volume) < breakoutVolume*0.7 {
		score := 0.65
		buyPrice := latestData.ClosePrice
		stopLoss := math.Min(buyPrice*0.93, boxLow) // 止损价或箱体下沿
		takeProfit := buyPrice * 1.12 // 突破后目标涨幅
		logic := fmt.Sprintf("平台突破(%.2f)后缩量回踩原压力位，支撑有效，目标涨幅12%%", breakoutPrice)

		return score, logic, buyPrice, stopLoss, takeProfit
	}

	return 0, "", 0, 0, 0
}

// strongStockReboundStrategy 强势股10日线反抽策略（完整实现）
func (e *StrategyEngineEnhanced) strongStockReboundStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近60个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -90)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 40 {
		return 0, "", 0, 0, 0
	}

	// 强势股筛选：近20日涨幅 > 0（简化版，实际应该对比大盘）
	recent20Data := dailyData[len(dailyData)-20:]
	if len(recent20Data) < 20 {
		return 0, "", 0, 0, 0
	}

	priceChange := (recent20Data[19].ClosePrice - recent20Data[0].ClosePrice) / recent20Data[0].ClosePrice
	if priceChange <= 0 {
		return 0, "", 0, 0, 0
	}

	// 检查是否第一次回踩10日线
	ma10 := e.calculateMA(dailyData, 10)
	latestData := dailyData[len(dailyData)-1]

	// 回踩确认：价格接近10日线，出现阳线
	distanceToMA10 := math.Abs(latestData.ClosePrice-ma10) / ma10
	isPositive := latestData.ClosePrice > latestData.OpenPrice

	if distanceToMA10 < 0.03 && isPositive {
		// 检查是否是第一次回踩（最近20个交易日内没有其他回踩）
		if e.isFirstPullback(dailyData, 10, 20) {
			score := 0.68 + math.Min(priceChange*2, 0.2) // 涨幅越大得分越高
			buyPrice := latestData.ClosePrice
			stopLoss := buyPrice * 0.93
			takeProfit := buyPrice * 1.10
			logic := fmt.Sprintf("强势股(近20日涨幅%.1f%%)第一次回踩10日线收阳，反弹概率较高", priceChange*100)

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// executeMediumTermStrategy 执行中线策略（基于技术指标）
func (e *StrategyEngineEnhanced) executeMediumTermStrategy(strategy *model.Strategy, stocks []model.Stock, tradeDate time.Time) ([]model.StrategyResult, error) {
	var results []model.StrategyResult

	for _, stock := range stocks {
		// 流动性筛选（日均成交额≥2亿）
		if !e.checkLiquidity(stock.Code, tradeDate, 200000000) {
			continue
		}

		var score float64
		var logic string
		var buyPrice, stopLoss, takeProfit float64

		switch strategy.StrategyID {
		case "medium_term_1":
			// 行业成长均线多头策略（技术面版）
			score, logic, buyPrice, stopLoss, takeProfit = e.industryGrowthStrategy(&stock, tradeDate)
		case "medium_term_2":
			// 困境反转策略（技术面版）
			score, logic, buyPrice, stopLoss, takeProfit = e.turnaroundStrategy(&stock, tradeDate)
		case "medium_term_3":
			// 高股息策略（技术面版）
			score, logic, buyPrice, stopLoss, takeProfit = e.dividendStrategy(&stock, tradeDate)
		}

		if score > 0.5 {
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

// industryGrowthStrategy 行业成长均线多头策略（技术面实现）
func (e *StrategyEngineEnhanced) industryGrowthStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近120个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -150)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 60 {
		return 0, "", 0, 0, 0
	}

	// 检查均线多头排列：5>10>20>60
	ma5 := e.calculateMA(dailyData, 5)
	ma10 := e.calculateMA(dailyData, 10)
	ma20 := e.calculateMA(dailyData, 20)
	ma60 := e.calculateMA(dailyData, 60)

	if ma5 > ma10 && ma10 > ma20 && ma20 > ma60 {
		latestData := dailyData[len(dailyData)-1]

		// 检查是否突破后回踩
		distanceToMA20 := math.Abs(latestData.ClosePrice-ma20) / ma20
		distanceToMA60 := math.Abs(latestData.ClosePrice-ma60) / ma60

		if distanceToMA20 < 0.03 || distanceToMA60 < 0.05 {
			score := 0.75
			buyPrice := latestData.ClosePrice
			stopLoss := buyPrice * 0.88 // -12%止损
			takeProfit := buyPrice * 1.25 // +25%止盈
			logic := "均线多头排列(5>10>20>60)，突破后回踩关键均线企稳，中线趋势良好"

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// executeLongTermStrategy 执行长线策略（占位实现）
func (e *StrategyEngineEnhanced) executeLongTermStrategy(strategy *model.Strategy, stocks []model.Stock, tradeDate time.Time) ([]model.StrategyResult, error) {
	// TODO: 长线策略需要基本面数据，当前数据源无法获取，暂不实现
	log.Printf("长线策略 %s 需要基本面数据，当前数据源无法支持，跳过执行", strategy.Name)
	return []model.StrategyResult{}, nil
}

// ========== 工具函数 ==========

// checkLiquidity 检查股票流动性
func (e *StrategyEngineEnhanced) checkLiquidity(stockCode string, tradeDate time.Time, minTurnover int64) bool {
	startDate := tradeDate.AddDate(0, 0, -30)
	dailyData, err := e.dailyDataRepo.GetDailyData(stockCode, startDate, tradeDate)
	if err != nil || len(dailyData) < 20 {
		return false
	}

	// 计算最近20个交易日的日均成交额
	var totalAmount int64
	count := 0
	for i := len(dailyData) - 1; i >= 0 && count < 20; i-- {
		totalAmount += int64(dailyData[i].Amount)
		count++
	}

	avgTurnover := totalAmount / int64(count)
	return avgTurnover >= minTurnover
}

// calculateMA 计算移动平均线
func (e *StrategyEngineEnhanced) calculateMA(dailyData []model.StockDailyData, period int) float64 {
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
func (e *StrategyEngineEnhanced) calculateVolumeMA(dailyData []model.StockDailyData, period int) int64 {
	if len(dailyData) < period {
		return 0
	}

	var sum int64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].Volume
	}

	return sum / int64(period)
}

// analyzeVolumePattern 分析成交量模式
func (e *StrategyEngineEnhanced) analyzeVolumePattern(dailyData []model.StockDailyData) float64 {
	if len(dailyData) < 15 {
		return 0
	}

	// 检查放量上涨→缩量回踩模式
	recent10Data := dailyData[len(dailyData)-10:]

	// 计算成交量变化
	volumeIncrease := false
	volumeDecrease := false

	// 检查前5个交易日是否放量
	if len(recent10Data) >= 5 {
		volumeMA5 := e.calculateVolumeMA(recent10Data[:5], 5)
		volumeMA10 := e.calculateVolumeMA(dailyData[len(dailyData)-15:len(dailyData)-5], 10)
		if volumeMA5 > int64(float64(volumeMA10)*1.2) {
			volumeIncrease = true
		}
	}

	// 检查最近3个交易日是否缩量
	if len(recent10Data) >= 3 {
		recentVolume := e.calculateVolumeMA(recent10Data[len(recent10Data)-3:], 3)
		prevVolume := e.calculateVolumeMA(recent10Data[:3], 3)
		if recentVolume < int64(float64(prevVolume)*0.8) {
			volumeDecrease = true
		}
	}

	if volumeIncrease && volumeDecrease {
		return 0.8 // 良好的成交量模式
	}

	return 0.3 // 一般的成交量模式
}

// getClosestMA 获取最接近的均线
func (e *StrategyEngineEnhanced) getClosestMA(price, ma5, ma10 float64) string {
	distance5 := math.Abs(price - ma5)
	distance10 := math.Abs(price - ma10)

	if distance5 < distance10 {
		return "5日"
	}
	return "10日"
}

// calculateBoxRange 计算箱体区间
func (e *StrategyEngineEnhanced) calculateBoxRange(dailyData []model.StockDailyData) (float64, float64) {
	if len(dailyData) == 0 {
		return 0, 0
	}

	var highs, lows []float64
	for _, data := range dailyData {
		highs = append(highs, data.HighPrice)
		lows = append(lows, data.LowPrice)
	}

	sort.Float64s(highs)
	sort.Float64s(lows)

	// 取70%分位数作为箱体上下沿
	highIdx := int(float64(len(highs)) * 0.7)
	lowIdx := int(float64(len(lows)) * 0.3)

	if highIdx >= len(highs) {
		highIdx = len(highs) - 1
	}
	if lowIdx < 0 {
		lowIdx = 0
	}

	return highs[highIdx], lows[lowIdx]
}

// isFirstPullback 检查是否是第一次回踩
func (e *StrategyEngineEnhanced) isFirstPullback(dailyData []model.StockDailyData, maPeriod, lookback int) bool {
	if len(dailyData) < lookback+maPeriod {
		return true
	}

	// 检查过去lookback个交易日内是否有其他回踩
	recentData := dailyData[len(dailyData)-lookback:]

	for i := 0; i < len(recentData)-5; i++ {
		windowData := dailyData[len(dailyData)-lookback+i:len(dailyData)-lookback+i+maPeriod]
		if len(windowData) < maPeriod {
			continue
		}

		ma := e.calculateMA(windowData, maPeriod)
		distance := math.Abs(recentData[i].ClosePrice-ma) / ma

		if distance < 0.03 {
			return false // 发现之前有回踩
		}
	}

	return true
}

// turnaroundStrategy 困境反转策略（技术面实现）
func (e *StrategyEngineEnhanced) turnaroundStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 获取最近180个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -240)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 120 {
		return 0, "", 0, 0, 0
	}

	// 检查前期调整幅度（从最高点到最低点）
	var maxPrice, minPrice float64
	for _, data := range dailyData[:90] { // 前90个交易日
		if data.HighPrice > maxPrice {
			maxPrice = data.HighPrice
		}
		if data.LowPrice < minPrice || minPrice == 0 {
			minPrice = data.LowPrice
		}
	}

	declineRatio := (maxPrice - minPrice) / maxPrice
	if declineRatio < 0.3 { // 调整幅度不足30%
		return 0, "", 0, 0, 0
	}

	// 检查近期是否放量突破60日线
	recentData := dailyData[len(dailyData)-30:]
	ma60 := e.calculateMA(dailyData, 60)

	breakoutDetected := false
	for _, data := range recentData {
		if data.ClosePrice > ma60 && data.Volume > int64(float64(e.calculateVolumeMA(dailyData[len(dailyData)-60:], 20))*1.3) {
			breakoutDetected = true
			break
		}
	}

	if breakoutDetected {
		latestData := dailyData[len(dailyData)-1]
		score := 0.7
		buyPrice := latestData.ClosePrice
		stopLoss := buyPrice * 0.90 // -10%止损
		takeProfit := buyPrice * 1.20 // +20%止盈
		logic := fmt.Sprintf("前期调整%.1f%%，近期放量突破60日线，困境反转信号", declineRatio*100)

		return score, logic, buyPrice, stopLoss, takeProfit
	}

	return 0, "", 0, 0, 0
}

// dividendStrategy 高股息策略（技术面实现）
func (e *StrategyEngineEnhanced) dividendStrategy(stock *model.Stock, tradeDate time.Time) (float64, string, float64, float64, float64) {
	// 高股息策略需要基本面数据，当前无法实现
	// 这里提供一个基于技术面的简化版本

	// 获取最近120个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -150)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 60 {
		return 0, "", 0, 0, 0
	}

	// 检查年线趋势（使用240日线近似年线）
	ma240 := e.calculateMA(dailyData, 240)
	if ma240 == 0 {
		return 0, "", 0, 0, 0
	}

	// 检查是否在年线或60日线附近
	latestData := dailyData[len(dailyData)-1]
	distanceToMA60 := math.Abs(latestData.ClosePrice-e.calculateMA(dailyData, 60)) / e.calculateMA(dailyData, 60)
	distanceToMA240 := math.Abs(latestData.ClosePrice-ma240) / ma240

	if distanceToMA60 < 0.05 || distanceToMA240 < 0.08 {
		// 检查波动率（高股息股通常波动较小）
		volatility := e.calculateVolatility(dailyData[len(dailyData)-20:])

		if volatility < 0.02 { // 日波动率小于2%
			score := 0.65
			buyPrice := latestData.ClosePrice
			stopLoss := buyPrice * 0.85 // -15%止损
			takeProfit := buyPrice * 1.15 // +15%止盈
			logic := "在年线/60日线附近企稳，波动率低，具备防御性特征"

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// calculateVolatility 计算波动率
func (e *StrategyEngineEnhanced) calculateVolatility(dailyData []model.StockDailyData) float64 {
	if len(dailyData) < 2 {
		return 0
	}

	var returns []float64
	for i := 1; i < len(dailyData); i++ {
		dailyReturn := (dailyData[i].ClosePrice - dailyData[i-1].ClosePrice) / dailyData[i-1].ClosePrice
		returns = append(returns, math.Abs(dailyReturn))
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}

	return sum / float64(len(returns))
}