package strategy

import (
	"fmt"
	"math"
	"sort"
	"time"

	"stock-strategy-backend/model"
)

// ShortTermStrategyExecutor 短线策略执行器
type ShortTermStrategyExecutor struct {
	dailyDataRepo *model.StockDailyDataRepository
}

// NewShortTermStrategyExecutor 创建短线策略执行器
func NewShortTermStrategyExecutor() *ShortTermStrategyExecutor {
	return &ShortTermStrategyExecutor{
		dailyDataRepo: &model.StockDailyDataRepository{},
	}
}

// ExecuteShortTermStrategy 执行短线策略
func (e *ShortTermStrategyExecutor) ExecuteShortTermStrategy(strategyID string, stock *model.Stock, tradeDate time.Time) (*model.StrategyResult, error) {
	// 获取最近90个交易日的日线数据
	startDate := tradeDate.AddDate(0, 0, -120)
	dailyData, err := e.dailyDataRepo.GetDailyData(stock.Code, startDate, tradeDate)
	if err != nil || len(dailyData) < 30 {
		return nil, fmt.Errorf("获取日线数据失败或数据不足")
	}

	var score float64
	var logic string
	var buyPrice, stopLoss, takeProfit float64

	switch strategyID {
	case "short_term_1":
		// 均线回踩低吸策略
		score, logic, buyPrice, stopLoss, takeProfit = e.executeMAPullbackStrategy(dailyData)
	case "short_term_2":
		// 突破缩量回踩策略
		score, logic, buyPrice, stopLoss, takeProfit = e.executeBreakoutPullbackStrategy(dailyData)
	case "short_term_3":
		// 强势股10日线反抽策略
		score, logic, buyPrice, stopLoss, takeProfit = e.executeStrongStockReboundStrategy(dailyData)
	default:
		return nil, fmt.Errorf("未知短线策略: %s", strategyID)
	}

	if score <= 0.5 {
		return nil, nil // 评分不足，不生成结果
	}

	result := &model.StrategyResult{
		StrategyID:       strategyID,
		TradeDate:        tradeDate,
		StockCode:        stock.Code,
		Score:            score,
		BuyPrice:         buyPrice,
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		LogicDescription: logic,
		CreatedAt:        time.Now(),
	}

	return result, nil
}

// executeMAPullbackStrategy 均线回踩低吸策略（完整实现）
func (e *ShortTermStrategyExecutor) executeMAPullbackStrategy(dailyData []model.StockDailyData) (float64, string, float64, float64, float64) {
	if len(dailyData) < 30 {
		return 0, "", 0, 0, 0
	}

	latestData := dailyData[len(dailyData)-1]

	// 计算技术指标
	ma5 := e.calculateMA(dailyData, 5)
	ma10 := e.calculateMA(dailyData, 10)
	ma20 := e.calculateMA(dailyData, 20)

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
			volumeScore := e.analyzeVolumePattern(dailyData)

			// 形态确认：收小阳线或十字星
			candleScore := e.analyzeCandlePattern(latestData)

			score := 0.7 + volumeScore*0.2 + candleScore*0.1
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

// executeBreakoutPullbackStrategy 突破缩量回踩策略（完整实现）
func (e *ShortTermStrategyExecutor) executeBreakoutPullbackStrategy(dailyData []model.StockDailyData) (float64, string, float64, float64, float64) {
	if len(dailyData) < 60 {
		return 0, "", 0, 0, 0
	}

	// 识别箱体震荡区间（前30个交易日）
	boxData := dailyData[len(dailyData)-45:] // 最近45个交易日
	if len(boxData) < 30 {
		return 0, "", 0, 0, 0
	}

	// 计算箱体上下沿
	boxHigh, boxLow := e.calculateBoxRange(boxData[:30])

	// 检查突破信号（最近15个交易日）
	breakoutData := boxData[30:]
	breakoutDetected := false
	var breakoutPrice float64
	var breakoutVolume int64

	for i, data := range breakoutData {
		if i >= 5 { // 只检查前5个交易日
			break
		}

		if data.ClosePrice > boxHigh && data.Volume > int64(float64(e.calculateVolumeMA(boxData[:30], 20))*1.5) {
			breakoutDetected = true
			breakoutPrice = data.ClosePrice
			breakoutVolume = data.Volume
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
	if pullbackRatio < 0.03 && float64(latestData.Volume) < float64(breakoutVolume)*0.7 {
		// 形态确认：收盘站稳支撑线
		if latestData.ClosePrice > boxHigh*0.98 {
			score := 0.65
			buyPrice := latestData.ClosePrice
			stopLoss := math.Min(buyPrice*0.93, boxLow) // 止损价或箱体下沿
			takeProfit := buyPrice * 1.12 // 突破后目标涨幅
			logic := fmt.Sprintf("平台突破(%.2f)后缩量回踩原压力位，支撑有效，目标涨幅12%%", breakoutPrice)

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// executeStrongStockReboundStrategy 强势股10日线反抽策略（完整实现）
func (e *ShortTermStrategyExecutor) executeStrongStockReboundStrategy(dailyData []model.StockDailyData) (float64, string, float64, float64, float64) {
	if len(dailyData) < 40 {
		return 0, "", 0, 0, 0
	}

	// 强势股筛选：近20日涨幅 > 0
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
			// 成交量验证：回调缩量，反弹放量
			volumeScore := e.analyzeReboundVolumePattern(dailyData)

			score := 0.68 + math.Min(priceChange*2, 0.2) + volumeScore*0.1
			buyPrice := latestData.ClosePrice
			stopLoss := buyPrice * 0.93
			takeProfit := buyPrice * 1.10
			logic := fmt.Sprintf("强势股(近20日涨幅%.1f%%)第一次回踩10日线收阳，反弹概率较高", priceChange*100)

			return score, logic, buyPrice, stopLoss, takeProfit
		}
	}

	return 0, "", 0, 0, 0
}

// ========== 技术指标计算工具函数 ==========

// calculateMA 计算移动平均线
func (e *ShortTermStrategyExecutor) calculateMA(dailyData []model.StockDailyData, period int) float64 {
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
func (e *ShortTermStrategyExecutor) calculateVolumeMA(dailyData []model.StockDailyData, period int) int64 {
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
func (e *ShortTermStrategyExecutor) analyzeVolumePattern(dailyData []model.StockDailyData) float64 {
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

// analyzeReboundVolumePattern 分析反弹成交量模式
func (e *ShortTermStrategyExecutor) analyzeReboundVolumePattern(dailyData []model.StockDailyData) float64 {
	if len(dailyData) < 10 {
		return 0
	}

	// 检查回调缩量→反弹放量模式
	recent5Data := dailyData[len(dailyData)-5:]
	recent10Data := dailyData[len(dailyData)-10:]

	// 回调期间成交量（前5个交易日）
	declineVolume := e.calculateVolumeMA(recent10Data[:5], 5)

	// 反弹期间成交量（最近5个交易日）
	reboundVolume := e.calculateVolumeMA(recent5Data, 5)

	if reboundVolume > int64(float64(declineVolume)*1.1) {
		return 0.7 // 反弹放量
	}

	return 0.3 // 成交量一般
}

// analyzeCandlePattern 分析K线形态
func (e *ShortTermStrategyExecutor) analyzeCandlePattern(data model.StockDailyData) float64 {
	// 计算实体大小
	bodySize := math.Abs(data.ClosePrice - data.OpenPrice)
	totalRange := data.HighPrice - data.LowPrice

	if totalRange == 0 {
		return 0
	}

	// 实体比例
	bodyRatio := bodySize / totalRange

	// 判断K线形态
	if bodyRatio < 0.3 {
		// 小实体（十字星、纺锤线）
		return 0.8
	} else if bodyRatio > 0.7 {
		// 大实体（大阳线/大阴线）
		if data.ClosePrice > data.OpenPrice {
			return 0.9 // 大阳线
		}
		return 0.2 // 大阴线
	}

	// 中等实体
	if data.ClosePrice > data.OpenPrice {
		return 0.7 // 中阳线
	}
	return 0.4 // 中阴线
}

// getClosestMA 获取最接近的均线
func (e *ShortTermStrategyExecutor) getClosestMA(price, ma5, ma10 float64) string {
	distance5 := math.Abs(price - ma5)
	distance10 := math.Abs(price - ma10)

	if distance5 < distance10 {
		return "5日"
	}
	return "10日"
}

// calculateBoxRange 计算箱体区间
func (e *ShortTermStrategyExecutor) calculateBoxRange(dailyData []model.StockDailyData) (float64, float64) {
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
func (e *ShortTermStrategyExecutor) isFirstPullback(dailyData []model.StockDailyData, maPeriod, lookback int) bool {
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

// FilterByLiquidity 根据流动性筛选股票
func (e *ShortTermStrategyExecutor) FilterByLiquidity(stocks []model.Stock, tradeDate time.Time, minTurnover int64) []model.Stock {
	var liquidStocks []model.Stock

	for _, stock := range stocks {
		if e.checkLiquidity(stock.Code, tradeDate, minTurnover) {
			liquidStocks = append(liquidStocks, stock)
		}
	}

	return liquidStocks
}

// checkLiquidity 检查股票流动性
func (e *ShortTermStrategyExecutor) checkLiquidity(stockCode string, tradeDate time.Time, minTurnover int64) bool {
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

// CalculateVolatility 计算波动率
func (e *ShortTermStrategyExecutor) CalculateVolatility(dailyData []model.StockDailyData) float64 {
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

// GetStrategyDescription 获取策略描述
func (e *ShortTermStrategyExecutor) GetStrategyDescription(strategyID string) string {
	descriptions := map[string]string{
		"short_term_1": "均线回踩低吸策略：趋势热点股、板块处上升期，股价在20日线上方，20日线向上，回踩5/10日线企稳",
		"short_term_2": "突破缩量回踩策略：横盘震荡后选择方向的票，平台突破后缩量回踩原压力位转支撑",
		"short_term_3": "强势股10日线反抽策略：强于大盘的板块龙头，短期回调至10日线出现承接",
	}

	return descriptions[strategyID]
}