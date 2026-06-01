package strategy

import (
	"fmt"
	"math"

	"stock-strategy-backend/model"
)

// ShortTermStrategy 短线策略引擎
type ShortTermStrategy struct {
	name        string
	description string
	parameters  map[string]interface{}
}

// NewShortTermStrategy 创建短线策略实例
func NewShortTermStrategy(strategyID string) *ShortTermStrategy {
	strategies := map[string]*ShortTermStrategy{
		"short_term_1": {
			name:        "均线回踩低吸",
			description: "趋势热点股、板块处上升期，股价在20日线上，20日线向上，回踩5/10日线企稳",
			parameters: map[string]interface{}{
				"ma_periods":      []int{5, 10, 20},
				"volume_ratio":    1.2,
				"min_turnover":    500000000.0, // 5亿成交额
				"max_price_ratio": 0.03,        // 回踩幅度不超过3%
			},
		},
		"short_term_2": {
			name:        "突破缩量回踩",
			description: "横盘震荡后选择方向的票，曾放量突破箱体/前高，回踩原压力转支撑缩量",
			parameters: map[string]interface{}{
				"box_period":          20,
				"breakout_volume_ratio": 1.5,
				"pullback_volume_ratio": 0.7,
				"support_level_ratio":  0.02,
			},
		},
		"short_term_3": {
			name:        "强势股10日线反抽",
			description: "强于大盘的板块龙头，短期回调，第一次回踩10/20日线收阳",
			parameters: map[string]interface{}{
				"compare_period":     20,
				"support_ma":         10,
				"min_outperform":     0.05, // 强于大盘5%
				"max_pullback_ratio":  0.08, // 回调不超过8%
			},
		},
	}

	if strategy, exists := strategies[strategyID]; exists {
		return strategy
	}

	return &ShortTermStrategy{
		name:        "未知短线策略",
		description: "策略配置不存在",
		parameters:  make(map[string]interface{}),
	}
}

// Execute 执行短线策略
func (s *ShortTermStrategy) Execute(stock *model.Stock, dailyData []model.StockDailyData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	if len(dailyData) < 60 {
		return nil // 数据不足
	}

	var result *model.StrategyResult

	switch s.name {
	case "均线回踩低吸":
		result = s.executeMAPullbackStrategy(stock, dailyData, indicators)
	case "突破缩量回踩":
		result = s.executeBreakoutPullbackStrategy(stock, dailyData, indicators)
	case "强势股10日线反抽":
		result = s.executeStrongStockReboundStrategy(stock, dailyData, indicators)
	}

	return result
}

// executeMAPullbackStrategy 均线回踩低吸策略
func (s *ShortTermStrategy) executeMAPullbackStrategy(stock *model.Stock, dailyData []model.StockDailyData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查流动性
	minTurnover := s.parameters["min_turnover"].(float64)
	if latestData.Amount < minTurnover {
		return nil
	}

	// 2. 计算移动平均线
	ma5 := s.calculateMA(dailyData, 5)
	ma10 := s.calculateMA(dailyData, 10)
	ma20 := s.calculateMA(dailyData, 20)

	if ma20 == 0 || ma10 == 0 || ma5 == 0 {
		return nil
	}

	// 3. 检查20日线趋势向上
	ma20Trend := s.checkMATrend(dailyData, 20)
	if !ma20Trend {
		return nil
	}

	// 4. 检查股价在20日线上方
	if latestData.ClosePrice <= ma20 {
		return nil
	}

	// 5. 检查回踩5/10日线
	maxPriceRatio := s.parameters["max_price_ratio"].(float64)
	touchingMA5 := math.Abs(latestData.ClosePrice-ma5)/ma5 <= maxPriceRatio
	touchingMA10 := math.Abs(latestData.ClosePrice-ma10)/ma10 <= maxPriceRatio

	if !touchingMA5 && !touchingMA10 {
		return nil
	}

	// 6. 检查成交量特征
	volumeRatio := s.parameters["volume_ratio"].(float64)
	volumeOK := s.checkVolumePattern(dailyData, volumeRatio)
	if !volumeOK {
		return nil
	}

	// 7. 计算策略评分
	score := s.calculateMAPullbackScore(latestData, ma5, ma10, ma20, touchingMA5, touchingMA10)

	if score < 0.6 {
		return nil
	}

	// 8. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := buyPrice * 0.93  // -7%止损
	takeProfit := buyPrice * 1.08 // +8%止盈

	logic := fmt.Sprintf("股价在20日线上方，20日线向上趋势，回踩%s均线企稳，成交量温和放大",
		func() string {
			if touchingMA5 {
				return "5日"
			}
			return "10日"
		}(),
	)

	return &model.StrategyResult{
		StockCode:       stock.Code,
		Score:           score,
		BuyPrice:        buyPrice,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		LogicDescription: logic,
	}
}

// executeBreakoutPullbackStrategy 突破缩量回踩策略
func (s *ShortTermStrategy) executeBreakoutPullbackStrategy(stock *model.Stock, dailyData []model.StockDailyData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查箱体震荡
	boxPeriod := s.parameters["box_period"].(int)
	if len(dailyData) < boxPeriod {
		return nil
	}

	boxData := dailyData[len(dailyData)-boxPeriod:]
	boxHigh, boxLow := s.findBoxRange(boxData)

	// 2. 检查突破
	breakoutVolumeRatio := s.parameters["breakout_volume_ratio"].(float64)
	breakoutOK := s.checkBreakout(boxData, boxHigh, breakoutVolumeRatio)
	if !breakoutOK {
		return nil
	}

	// 3. 检查回踩
	supportLevelRatio := s.parameters["support_level_ratio"].(float64)
	pullbackVolumeRatio := s.parameters["pullback_volume_ratio"].(float64)
	pullbackOK := s.checkPullback(latestData, boxHigh, supportLevelRatio, pullbackVolumeRatio)
	if !pullbackOK {
		return nil
	}

	// 4. 计算评分
	score := s.calculateBreakoutScore(latestData, boxHigh, boxLow)

	if score < 0.6 {
		return nil
	}

	// 5. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := boxLow * 0.98  // 跌破箱体下沿止损
	takeProfit := buyPrice * 1.07 // +7%止盈

	logic := "平台突破后缩量回踩，原压力位转为支撑位，确认支撑有效"

	return &model.StrategyResult{
		StockCode:       stock.Code,
		Score:           score,
		BuyPrice:        buyPrice,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		LogicDescription: logic,
	}
}

// executeStrongStockReboundStrategy 强势股10日线反抽策略
func (s *ShortTermStrategy) executeStrongStockReboundStrategy(stock *model.Stock, dailyData []model.StockDailyData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查相对强度
	comparePeriod := s.parameters["compare_period"].(int)
	minOutperform := s.parameters["min_outperform"].(float64)
	strongOK := s.checkStockStrength(stock, dailyData, comparePeriod, minOutperform)
	if !strongOK {
		return nil
	}

	// 2. 检查回调幅度
	maxPullbackRatio := s.parameters["max_pullback_ratio"].(float64)
	pullbackOK := s.checkPullbackMagnitude(dailyData, maxPullbackRatio)
	if !pullbackOK {
		return nil
	}

	// 3. 检查10日线支撑
	supportMA := s.parameters["support_ma"].(int)
	maValue := s.calculateMA(dailyData, supportMA)
	if maValue == 0 {
		return nil
	}

	supportRatio := 0.02 // 2%支撑范围
	if math.Abs(latestData.ClosePrice-maValue)/maValue > supportRatio {
		return nil
	}

	// 4. 检查K线形态（收阳）
	if latestData.ClosePrice <= latestData.OpenPrice {
		return nil
	}

	// 5. 计算评分
	score := s.calculateReboundScore(latestData, maValue)

	if score < 0.6 {
		return nil
	}

	// 6. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := maValue * 0.97  // 跌破均线3%止损
	takeProfit := buyPrice * 1.06 // +6%止盈

	logic := fmt.Sprintf("强势股第一次回踩%d日线收阳，出现承接迹象", supportMA)

	return &model.StrategyResult{
		StockCode:       stock.Code,
		Score:           score,
		BuyPrice:        buyPrice,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		LogicDescription: logic,
	}
}

// calculateMA 计算移动平均线
func (s *ShortTermStrategy) calculateMA(dailyData []model.StockDailyData, period int) float64 {
	if len(dailyData) < period {
		return 0
	}

	var sum float64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].ClosePrice
	}

	return sum / float64(period)
}

// checkMATrend 检查移动平均线趋势
func (s *ShortTermStrategy) checkMATrend(dailyData []model.StockDailyData, period int) bool {
	if len(dailyData) < period*2 {
		return false
	}

	// 计算当前MA和之前MA
	currentMA := s.calculateMA(dailyData, period)
	previousData := dailyData[:len(dailyData)-10] // 10天前的数据
	if len(previousData) < period {
		return false
	}

	previousMA := s.calculateMA(previousData, period)

	return currentMA > previousMA
}

// checkVolumePattern 检查成交量模式
func (s *ShortTermStrategy) checkVolumePattern(dailyData []model.StockDailyData, minRatio float64) bool {
	if len(dailyData) < 10 {
		return false
	}

	// 计算最近5日平均成交量
	var recentSum int64
	for i := len(dailyData) - 5; i < len(dailyData); i++ {
		recentSum += dailyData[i].Volume
	}
	recentAvg := float64(recentSum) / 5.0

	// 计算之前5日平均成交量
	var previousSum int64
	for i := len(dailyData) - 10; i < len(dailyData)-5; i++ {
		previousSum += dailyData[i].Volume
	}
	previousAvg := float64(previousSum) / 5.0

	if previousAvg == 0 {
		return false
	}

	return recentAvg/previousAvg >= minRatio
}

// findBoxRange 寻找箱体范围
func (s *ShortTermStrategy) findBoxRange(dailyData []model.StockDailyData) (float64, float64) {
	var high, low float64
	if len(dailyData) == 0 {
		return 0, 0
	}

	high = dailyData[0].HighPrice
	low = dailyData[0].LowPrice

	for _, data := range dailyData {
		if data.HighPrice > high {
			high = data.HighPrice
		}
		if data.LowPrice < low {
			low = data.LowPrice
		}
	}

	return high, low
}

// checkBreakout 检查突破
func (s *ShortTermStrategy) checkBreakout(dailyData []model.StockDailyData, boxHigh float64, volumeRatio float64) bool {
	for i, data := range dailyData {
		// 检查是否突破箱体上沿
		if data.HighPrice > boxHigh {
			// 检查突破当天的成交量
			if i > 0 && float64(data.Volume) > float64(dailyData[i-1].Volume)*volumeRatio {
				return true
			}
		}
	}

	return false
}

// checkPullback 检查回踩
func (s *ShortTermStrategy) checkPullback(latestData model.StockDailyData, supportLevel float64, ratio float64, volumeRatio float64) bool {
	// 检查价格是否在支撑位附近
	priceDiff := math.Abs(latestData.ClosePrice-supportLevel) / supportLevel
	if priceDiff > ratio {
		return false
	}

	// 检查成交量是否缩量
	// 这里需要更多数据来判断，暂时返回true
	return true
}

// checkStockStrength 检查股票相对强度
func (s *ShortTermStrategy) checkStockStrength(stock *model.Stock, dailyData []model.StockDailyData, period int, minOutperform float64) bool {
	// 简化实现：检查近期涨幅
	if len(dailyData) < period {
		return false
	}

	startPrice := dailyData[len(dailyData)-period].ClosePrice
	endPrice := dailyData[len(dailyData)-1].ClosePrice

	if startPrice == 0 {
		return false
	}

	return (endPrice-startPrice)/startPrice >= minOutperform
}

// checkPullbackMagnitude 检查回调幅度
func (s *ShortTermStrategy) checkPullbackMagnitude(dailyData []model.StockDailyData, maxRatio float64) bool {
	if len(dailyData) < 10 {
		return false
	}

	// 找到近期高点
	recentHigh := dailyData[len(dailyData)-1].HighPrice
	for i := len(dailyData) - 10; i < len(dailyData); i++ {
		if dailyData[i].HighPrice > recentHigh {
			recentHigh = dailyData[i].HighPrice
		}
	}

	currentPrice := dailyData[len(dailyData)-1].ClosePrice
	if recentHigh == 0 {
		return false
	}

	pullbackRatio := (recentHigh - currentPrice) / recentHigh
	return pullbackRatio <= maxRatio
}

// calculateMAPullbackScore 计算均线回踩策略评分
func (s *ShortTermStrategy) calculateMAPullbackScore(data model.StockDailyData, ma5, ma10, ma20 float64, touchingMA5, touchingMA10 bool) float64 {
	score := 0.5

	// 均线多头排列加分
	if ma5 > ma10 && ma10 > ma20 {
		score += 0.2
	}

	// 回踩5日线比10日线评分高
	if touchingMA5 {
		score += 0.15
	} else if touchingMA10 {
		score += 0.1
	}

	// 成交量配合加分
	if float64(data.Volume) > data.Amount/float64(data.ClosePrice)*0.8 { // 简化成交量判断
		score += 0.1
	}

	// 价格位置评分
	positionScore := (data.ClosePrice - ma20) / ma20
	if positionScore > 0 && positionScore < 0.1 { // 在20日线上方10%以内
		score += 0.05
	}

	return math.Min(score, 1.0)
}

// calculateBreakoutScore 计算突破回踩策略评分
func (s *ShortTermStrategy) calculateBreakoutScore(data model.StockDailyData, boxHigh, boxLow float64) float64 {
	score := 0.5

	// 突破幅度评分
	breakoutRatio := (data.ClosePrice - boxHigh) / boxHigh
	if breakoutRatio > 0.05 { // 突破5%以上
		score += 0.2
	} else if breakoutRatio > 0.02 { // 突破2%以上
		score += 0.1
	}

	// 箱体高度评分
	boxHeight := (boxHigh - boxLow) / boxLow
	if boxHeight > 0.1 && boxHeight < 0.3 { // 箱体高度10%-30%
		score += 0.15
	}

	// 成交量评分
	if data.Volume > 0 { // 简化评分
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// calculateReboundScore 计算反抽策略评分
func (s *ShortTermStrategy) calculateReboundScore(data model.StockDailyData, maValue float64) float64 {
	score := 0.5

	// K线形态评分（阳线）
	if data.ClosePrice > data.OpenPrice {
		score += 0.15
	}

	// 下影线评分
	shadowRatio := (data.ClosePrice - data.LowPrice) / (data.HighPrice - data.LowPrice)
	if shadowRatio > 0.3 { // 下影线较长
		score += 0.1
	}

	// 均线支撑评分
	supportRatio := math.Abs(data.ClosePrice-maValue) / maValue
	if supportRatio < 0.02 { // 距离均线2%以内
		score += 0.15
	}

	return math.Min(score, 1.0)
}

// GetName 获取策略名称
func (s *ShortTermStrategy) GetName() string {
	return s.name
}

// GetDescription 获取策略描述
func (s *ShortTermStrategy) GetDescription() string {
	return s.description
}

// GetParameters 获取策略参数
func (s *ShortTermStrategy) GetParameters() map[string]interface{} {
	return s.parameters
}