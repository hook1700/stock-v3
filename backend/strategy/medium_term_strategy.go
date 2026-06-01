package strategy

import (
	"fmt"
	"math"

	"stock-strategy-backend/model"
)

// MediumTermStrategy 中线策略引擎
type MediumTermStrategy struct {
	name        string
	description string
	parameters  map[string]interface{}
}

// NewMediumTermStrategy 创建中线策略实例
func NewMediumTermStrategy(strategyID string) *MediumTermStrategy {
	strategies := map[string]*MediumTermStrategy{
		"medium_term_1": {
			name:        "行业成长均线多头",
			description: "行业景气股，偏中长线主做，近两季营收/净利同比↑，ROE≥10%，股价站上60日线，5>10>20>60",
			parameters: map[string]interface{}{
				"roe_threshold":           10.0,
				"growth_threshold":        0.1,   // 10%增长
				"ma_sequence":             []int{5, 10, 20, 60},
				"platform_breakout_ratio": 0.05,  // 平台突破5%
				"pullback_ratio":          0.03,  // 回踩3%
			},
		},
		"medium_term_2": {
			name:        "困境反转业绩拐点",
			description: "业绩由差转好、订单/政策催化，前期调整>30%，近期季报营收/净利拐点向上，行业有利好",
			parameters: map[string]interface{}{
				"decline_threshold":       0.3,   // 前期调整30%
				"recovery_period":         2,     // 恢复期2个季度
				"turnaround_growth":       0.15,  // 拐点增长15%
				"volume_increase_ratio":   1.5,   // 成交量放大1.5倍
			},
		},
		"medium_term_3": {
			name:        "高股息红利慢牛",
			description: "震荡市/偏弱市，求稳健，银行/电力/高速/运营商/部分资源股，股息率≥4%~5%，ROE稳定",
			parameters: map[string]interface{}{
				"dividend_yield":          4.5,
				"roe_stability":           5,     // 连续5年ROE稳定
				"debt_ratio":              0.6,   // 负债率不超过60%
				"volatility_threshold":    0.3,   // 年波动率不超过30%
			},
		},
	}

	if strategy, exists := strategies[strategyID]; exists {
		return strategy
	}

	return &MediumTermStrategy{
		name:        "未知中线策略",
		description: "策略配置不存在",
		parameters:  make(map[string]interface{}),
	}
}

// Execute 执行中线策略
func (s *MediumTermStrategy) Execute(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	if len(dailyData) < 120 { // 至少需要4个月数据
		return nil
	}

	var result *model.StrategyResult

	switch s.name {
	case "行业成长均线多头":
		result = s.executeIndustryGrowthStrategy(stock, dailyData, financialData, indicators)
	case "困境反转业绩拐点":
		result = s.executeTurnaroundStrategy(stock, dailyData, financialData, indicators)
	case "高股息红利慢牛":
		result = s.executeDividendStrategy(stock, dailyData, financialData, indicators)
	}

	return result
}

// executeIndustryGrowthStrategy 行业成长均线多头策略
func (s *MediumTermStrategy) executeIndustryGrowthStrategy(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查基本面条件
	if !s.checkFinancialConditions(financialData) {
		return nil
	}

	// 2. 检查均线多头排列
	maSequence := s.parameters["ma_sequence"].([]int)
	if !s.checkMASequence(dailyData, maSequence) {
		return nil
	}

	// 3. 检查平台突破
	platformBreakoutRatio := s.parameters["platform_breakout_ratio"].(float64)
	breakoutOK := s.checkPlatformBreakout(dailyData, platformBreakoutRatio)
	if !breakoutOK {
		return nil
	}

	// 4. 检查回踩支撑
	pullbackRatio := s.parameters["pullback_ratio"].(float64)
	supportOK := s.checkSupportLevel(dailyData, pullbackRatio)
	if !supportOK {
		return nil
	}

	// 5. 计算评分
	score := s.calculateIndustryGrowthScore(latestData, dailyData, financialData, indicators)

	if score < 0.6 {
		return nil
	}

	// 6. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := s.calculateStopLossPrice(dailyData, maSequence)
	takeProfit := buyPrice * 1.15 // +15%止盈

	logic := fmt.Sprintf("行业景气股，基本面良好，均线多头排列，平台突破后回踩支撑企稳")

	return &model.StrategyResult{
		StockCode:       stock.Code,
		Score:           score,
		BuyPrice:        buyPrice,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		LogicDescription: logic,
	}
}

// executeTurnaroundStrategy 困境反转业绩拐点策略
func (s *MediumTermStrategy) executeTurnaroundStrategy(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查前期调整幅度
	declineThreshold := s.parameters["decline_threshold"].(float64)
	if !s.checkPreviousDecline(dailyData, declineThreshold) {
		return nil
	}

	// 2. 检查业绩拐点
	recoveryPeriod := s.parameters["recovery_period"].(int)
	turnaroundGrowth := s.parameters["turnaround_growth"].(float64)
	if !s.checkPerformanceTurnaround(financialData, recoveryPeriod, turnaroundGrowth) {
		return nil
	}

	// 3. 检查技术面突破
	volumeIncreaseRatio := s.parameters["volume_increase_ratio"].(float64)
	if !s.checkTechnicalBreakout(dailyData, volumeIncreaseRatio) {
		return nil
	}

	// 4. 检查行业利好
	if !s.checkIndustryCatalyst(stock.Industry) {
		return nil
	}

	// 5. 计算评分
	score := s.calculateTurnaroundScore(latestData, dailyData, financialData, indicators)

	if score < 0.6 {
		return nil
	}

	// 6. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := latestData.ClosePrice * 0.88 // -12%止损
	takeProfit := buyPrice * 1.25 // +25%止盈（困境反转空间较大）

	logic := fmt.Sprintf("业绩拐点确认，前期充分调整，技术面突破，行业有利好催化")

	return &model.StrategyResult{
		StockCode:       stock.Code,
		Score:           score,
		BuyPrice:        buyPrice,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		LogicDescription: logic,
	}
}

// executeDividendStrategy 高股息红利慢牛策略
func (s *MediumTermStrategy) executeDividendStrategy(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查股息率
	dividendYield := s.parameters["dividend_yield"].(float64)
	if !s.checkDividendYield(financialData, dividendYield) {
		return nil
	}

	// 2. 检查ROE稳定性
	roeStability := s.parameters["roe_stability"].(int)
	if !s.checkROEStability(financialData, roeStability) {
		return nil
	}

	// 3. 检查负债率
	debtRatio := s.parameters["debt_ratio"].(float64)
	if !s.checkDebtRatio(financialData, debtRatio) {
		return nil
	}

	// 4. 检查波动率
	volatilityThreshold := s.parameters["volatility_threshold"].(float64)
	if !s.checkVolatility(dailyData, volatilityThreshold) {
		return nil
	}

	// 5. 检查技术面趋势
	if !s.checkSlowBullTrend(dailyData) {
		return nil
	}

	// 6. 计算评分
	score := s.calculateDividendScore(latestData, dailyData, financialData, indicators)

	if score < 0.6 {
		return nil
	}

	// 7. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := latestData.ClosePrice * 0.85 // -15%止损（慢牛策略止损较宽）
	takeProfit := buyPrice * 1.20 // +20%止盈

	logic := fmt.Sprintf("高股息率，ROE稳定，负债合理，慢牛趋势确立")

	return &model.StrategyResult{
		StockCode:       stock.Code,
		Score:           score,
		BuyPrice:        buyPrice,
		StopLossPrice:   stopLoss,
		TakeProfitPrice: takeProfit,
		LogicDescription: logic,
	}
}

// checkFinancialConditions 检查基本面条件
func (s *MediumTermStrategy) checkFinancialConditions(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	// 检查ROE
	roeThreshold := s.parameters["roe_threshold"].(float64)
	if financialData.ROE < roeThreshold {
		return false
	}

	// 检查增长性
	growthThreshold := s.parameters["growth_threshold"].(float64)
	if financialData.RevenueGrowth < growthThreshold || financialData.ProfitGrowth < growthThreshold {
		return false
	}

	return true
}

// checkMASequence 检查均线多头排列
func (s *MediumTermStrategy) checkMASequence(dailyData []model.StockDailyData, maSequence []int) bool {
	if len(dailyData) < maSequence[len(maSequence)-1] {
		return false
	}

	// 计算各周期均线
	maValues := make([]float64, len(maSequence))
	for i, period := range maSequence {
		maValues[i] = s.calculateMA(dailyData, period)
		if maValues[i] == 0 {
			return false
		}
	}

	// 检查是否多头排列（短周期>长周期）
	for i := 1; i < len(maValues); i++ {
		if maValues[i-1] <= maValues[i] {
			return false
		}
	}

	return true
}

// checkPlatformBreakout 检查平台突破
func (s *MediumTermStrategy) checkPlatformBreakout(dailyData []model.StockDailyData, breakoutRatio float64) bool {
	if len(dailyData) < 60 {
		return false
	}

	// 寻找平台区间（最近20-40个交易日）
	platformStart := len(dailyData) - 40
	platformEnd := len(dailyData) - 20
	platformData := dailyData[platformStart:platformEnd]

	// 计算平台高点
	platformHigh, _ := s.findPlatformRange(platformData)

	// 检查是否突破平台
	latestData := dailyData[len(dailyData)-1]
	breakoutLevel := platformHigh * (1 + breakoutRatio)

	return latestData.ClosePrice >= breakoutLevel
}

// checkSupportLevel 检查回踩支撑
func (s *MediumTermStrategy) checkSupportLevel(dailyData []model.StockDailyData, pullbackRatio float64) bool {
	latestData := dailyData[len(dailyData)-1]

	// 计算关键均线支撑
	ma20 := s.calculateMA(dailyData, 20)
	ma60 := s.calculateMA(dailyData, 60)

	if ma20 == 0 || ma60 == 0 {
		return false
	}

	// 检查是否回踩均线
	supportLevel := math.Max(ma20, ma60)
	pullbackDistance := math.Abs(latestData.ClosePrice-supportLevel) / supportLevel

	return pullbackDistance <= pullbackRatio
}

// checkPreviousDecline 检查前期调整幅度
func (s *MediumTermStrategy) checkPreviousDecline(dailyData []model.StockDailyData, declineThreshold float64) bool {
	if len(dailyData) < 120 {
		return false
	}

	// 寻找前期高点（6个月前）
	highPoint := len(dailyData) - 120
	highPrice := dailyData[highPoint].HighPrice

	// 寻找近期低点（1个月前）
	lowPoint := len(dailyData) - 20
	lowPrice := dailyData[lowPoint].LowPrice

	if highPrice == 0 {
		return false
	}

	declineRatio := (highPrice - lowPrice) / highPrice
	return declineRatio >= declineThreshold
}

// checkPerformanceTurnaround 检查业绩拐点
func (s *MediumTermStrategy) checkPerformanceTurnaround(financialData *model.FinancialData, recoveryPeriod int, turnaroundGrowth float64) bool {
	if financialData == nil {
		return false
	}

	// 检查最近几个季度的业绩变化
	// 这里需要具体的财务数据实现
	// 简化实现：检查最近季度是否增长
	return financialData.LastQuarterGrowth >= turnaroundGrowth
}

// checkTechnicalBreakout 检查技术面突破
func (s *MediumTermStrategy) checkTechnicalBreakout(dailyData []model.StockDailyData, volumeIncreaseRatio float64) bool {
	if len(dailyData) < 30 {
		return false
	}

	// 检查成交量放大
	recentVolume := dailyData[len(dailyData)-1].Volume
	averageVolume := s.calculateAverageVolume(dailyData, 30)

	if averageVolume == 0 {
		return false
	}

	return float64(recentVolume) >= float64(averageVolume)*volumeIncreaseRatio
}

// checkIndustryCatalyst 检查行业利好
func (s *MediumTermStrategy) checkIndustryCatalyst(industry string) bool {
	// 这里需要集成行业利好数据
	// 简化实现：返回true
	return true
}

// checkDividendYield 检查股息率
func (s *MediumTermStrategy) checkDividendYield(financialData *model.FinancialData, dividendYield float64) bool {
	if financialData == nil {
		return false
	}

	return financialData.DividendYield >= dividendYield
}

// checkROEStability 检查ROE稳定性
func (s *MediumTermStrategy) checkROEStability(financialData *model.FinancialData, stabilityYears int) bool {
	if financialData == nil {
		return false
	}

	// 检查连续几年的ROE稳定性
	// 简化实现：检查最近一年的ROE
	return financialData.ROE >= 8.0 // 简化标准
}

// checkDebtRatio 检查负债率
func (s *MediumTermStrategy) checkDebtRatio(financialData *model.FinancialData, debtRatio float64) bool {
	if financialData == nil {
		return false
	}

	return financialData.DebtRatio <= debtRatio
}

// checkVolatility 检查波动率
func (s *MediumTermStrategy) checkVolatility(dailyData []model.StockDailyData, volatilityThreshold float64) bool {
	if len(dailyData) < 250 {
		return false
	}

	// 计算年化波动率
	returns := make([]float64, len(dailyData)-1)
	for i := 1; i < len(dailyData); i++ {
		if dailyData[i-1].ClosePrice > 0 {
			returns[i-1] = (dailyData[i].ClosePrice - dailyData[i-1].ClosePrice) / dailyData[i-1].ClosePrice
		}
	}

	volatility := s.calculateVolatility(returns)
	return volatility <= volatilityThreshold
}

// checkSlowBullTrend 检查慢牛趋势
func (s *MediumTermStrategy) checkSlowBullTrend(dailyData []model.StockDailyData) bool {
	if len(dailyData) < 120 {
		return false
	}

	// 检查年线趋势向上
	ma250 := s.calculateMA(dailyData, 250)
	ma120 := s.calculateMA(dailyData, 120)

	if ma250 == 0 || ma120 == 0 {
		return false
	}

	return ma120 > ma250
}

// calculateMA 计算移动平均线
func (s *MediumTermStrategy) calculateMA(dailyData []model.StockDailyData, period int) float64 {
	if len(dailyData) < period {
		return 0
	}

	var sum float64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].ClosePrice
	}

	return sum / float64(period)
}

// findPlatformRange 寻找平台范围
func (s *MediumTermStrategy) findPlatformRange(dailyData []model.StockDailyData) (float64, float64) {
	if len(dailyData) == 0 {
		return 0, 0
	}

	high := dailyData[0].HighPrice
	low := dailyData[0].LowPrice

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

// calculateAverageVolume 计算平均成交量
func (s *MediumTermStrategy) calculateAverageVolume(dailyData []model.StockDailyData, period int) int64 {
	if len(dailyData) < period {
		return 0
	}

	var sum int64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].Volume
	}

	return sum / int64(period)
}

// calculateVolatility 计算波动率
func (s *MediumTermStrategy) calculateVolatility(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	var variance float64
	for _, r := range returns {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance * 252) // 年化波动率
}

// calculateStopLossPrice 计算止损价格
func (s *MediumTermStrategy) calculateStopLossPrice(dailyData []model.StockDailyData, maSequence []int) float64 {
	// 使用最长的均线作为止损位
	longestMA := s.calculateMA(dailyData, maSequence[len(maSequence)-1])
	return longestMA * 0.95 // 跌破5%止损
}

// calculateIndustryGrowthScore 计算行业成长策略评分
func (s *MediumTermStrategy) calculateIndustryGrowthScore(data model.StockDailyData, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) float64 {
	score := 0.5

	// 基本面评分
	if financialData != nil {
		if financialData.ROE >= 15 {
			score += 0.15
		} else if financialData.ROE >= 10 {
			score += 0.1
		}

		if financialData.RevenueGrowth >= 0.2 {
			score += 0.1
		} else if financialData.RevenueGrowth >= 0.1 {
			score += 0.05
		}
	}

	// 技术面评分
	if s.checkMASequence(dailyData, []int{5, 10, 20, 60}) {
		score += 0.15
	}

	// 成交量评分
	if data.Volume > s.calculateAverageVolume(dailyData, 20) {
		score += 0.05
	}

	return math.Min(score, 1.0)
}

// calculateTurnaroundScore 计算困境反转策略评分
func (s *MediumTermStrategy) calculateTurnaroundScore(data model.StockDailyData, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) float64 {
	score := 0.5

	// 调整幅度评分
	if s.checkPreviousDecline(dailyData, 0.3) {
		score += 0.15
	}

	// 业绩拐点评分
	if financialData != nil && financialData.LastQuarterGrowth >= 0.15 {
		score += 0.2
	}

	// 技术突破评分
	volumeIncreaseRatio := s.parameters["volume_increase_ratio"].(float64)
	if s.checkTechnicalBreakout(dailyData, volumeIncreaseRatio) {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// calculateDividendScore 计算高股息策略评分
func (s *MediumTermStrategy) calculateDividendScore(data model.StockDailyData, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) float64 {
	score := 0.5

	// 股息率评分
	if financialData != nil {
		if financialData.DividendYield >= 6 {
			score += 0.2
		} else if financialData.DividendYield >= 4.5 {
			score += 0.15
		}

		// ROE稳定性评分
		if financialData.ROE >= 12 {
			score += 0.1
		} else if financialData.ROE >= 8 {
			score += 0.05
		}

		// 负债率评分
		if financialData.DebtRatio <= 0.5 {
			score += 0.1
		}
	}

	// 波动率评分
	if s.checkVolatility(dailyData, 0.3) {
		score += 0.05
	}

	return math.Min(score, 1.0)
}

// GetName 获取策略名称
func (s *MediumTermStrategy) GetName() string {
	return s.name
}

// GetDescription 获取策略描述
func (s *MediumTermStrategy) GetDescription() string {
	return s.description
}

// GetParameters 获取策略参数
func (s *MediumTermStrategy) GetParameters() map[string]interface{} {
	return s.parameters
}