package strategy

import (
	"fmt"
	"math"

	"stock-strategy-backend/model"
)

// LongTermStrategy 长线策略引擎
type LongTermStrategy struct {
	name        string
	description string
	parameters  map[string]interface{}
}

// NewLongTermStrategy 创建长线策略实例
func NewLongTermStrategy(strategyID string) *LongTermStrategy {
	strategies := map[string]*LongTermStrategy{
		"long_term_1": {
			name:        "价值投资低估值",
			description: "寻找被低估的优质股票，PE/PB低于行业平均，ROE稳定，股息率高",
			parameters: map[string]interface{}{
				"pe_threshold":        15.0,  // PE低于15
				"pb_threshold":        2.0,   // PB低于2
				"roe_threshold":       10.0,  // ROE大于10%
				"dividend_yield":     3.0,   // 股息率大于3%
				"debt_ratio_max":      0.6,   // 负债率小于60%
				"hold_period":         252,   // 持有期约1年(交易日)
			},
		},
		"long_term_2": {
			name:        "成长股长期持有",
			description: "选择高成长股，营收/利润持续增长，行业前景好，适合3-5年持有",
			parameters: map[string]interface{}{
				"revenue_growth_min":   0.2,   // 营收增长>20%
				"profit_growth_min":    0.15,  // 利润增长>15%
				"roe_min":              15.0,  // ROE>15%
				"market_share":         0.1,   // 市场份额>10%
				"industry_growth":      0.1,   // 行业增长>10%
				"hold_period":          504,   // 持有期约2年(交易日)
			},
		},
		"long_term_3": {
			name:        "龙头股长期配置",
			description: "行业龙头股，竞争优势明显，长期趋势向上，适合长期配置",
			parameters: map[string]interface{}{
				"market_cap_min":      10000000000.0, // 市值>100亿
				"industry_rank":       3,              // 行业排名前3
				"roe_stable":          8.0,           // ROE稳定>8%
				"profit_stable":       5,              // 连续5年盈利
				"ma200_trend":         true,          // 200日均线向上
				"hold_period":         756,           // 持有期约3年(交易日)
			},
		},
	}

	if strategy, exists := strategies[strategyID]; exists {
		return strategy
	}

	return &LongTermStrategy{
		name:        "未知长线策略",
		description: "策略配置不存在",
		parameters:  make(map[string]interface{}),
	}
}

// Execute 执行长线策略
func (s *LongTermStrategy) Execute(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	if len(dailyData) < 250 { // 至少需要1年数据
		return nil
	}

	var result *model.StrategyResult

	switch s.name {
	case "价值投资低估值":
		result = s.executeValueStrategy(stock, dailyData, financialData, indicators)
	case "成长股长期持有":
		result = s.executeGrowthStrategy(stock, dailyData, financialData, indicators)
	case "龙头股长期配置":
		result = s.executeLeaderStrategy(stock, dailyData, financialData, indicators)
	}

	return result
}

// executeValueStrategy 价值投资策略
func (s *LongTermStrategy) executeValueStrategy(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查估值条件
	if !s.checkValuationConditions(financialData) {
		return nil
	}

	// 2. 检查财务健康
	if !s.checkFinancialHealth(financialData) {
		return nil
	}

	// 3. 检查股息率
	if !s.checkDividendCondition(financialData) {
		return nil
	}

	// 4. 检查长期趋势
	if !s.checkLongTermTrend(dailyData) {
		return nil
	}

	// 5. 计算评分
	score := s.calculateValueScore(latestData, financialData, indicators)

	if score < 0.6 {
		return nil
	}

	// 6. 生成买卖点建议（长线策略更注重安全边际）
	buyPrice := latestData.ClosePrice
	stopLoss := buyPrice * 0.80  // 长线止损较宽：-20%
	takeProfit := buyPrice * 1.50 // 长线止盈较高：+50%

	logic := fmt.Sprintf("低估值价值股，PE=%.1f,PB=%.1f,ROE=%.1f%%,股息率=%.2f%%",
		financialData.PE, financialData.PB, financialData.ROE, financialData.DividendYield)

	return &model.StrategyResult{
		StockCode:        stock.Code,
		Score:            score,
		BuyPrice:         buyPrice,
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		LogicDescription: logic,
	}
}

// executeGrowthStrategy 成长股策略
func (s *LongTermStrategy) executeGrowthStrategy(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查成长性
	if !s.checkGrowthConditions(financialData) {
		return nil
	}

	// 2. 检查盈利能力
	if !s.checkProfitability(financialData) {
		return nil
	}

	// 3. 检查行业前景
	if !s.checkIndustryProspect(stock.Industry) {
		return nil
	}

	// 4. 检查技术面趋势
	if !s.checkLongTermTrend(dailyData) {
		return nil
	}

	// 5. 计算评分
	score := s.calculateGrowthScore(latestData, financialData, indicators)

	if score < 0.6 {
		return nil
	}

	// 6. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := buyPrice * 0.75  // 成长股波动大，止损更宽：-25%
	takeProfit := buyPrice * 2.00 // 成长股目标收益高：+100%

	logic := fmt.Sprintf("高成长股，营收增长%.1f%%,利润增长%.1f%%,ROE=%.1f%%",
		financialData.RevenueGrowth*100, financialData.ProfitGrowth*100, financialData.ROE)

	return &model.StrategyResult{
		StockCode:        stock.Code,
		Score:            score,
		BuyPrice:         buyPrice,
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		LogicDescription: logic,
	}
}

// executeLeaderStrategy 龙头股策略
func (s *LongTermStrategy) executeLeaderStrategy(stock *model.Stock, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) *model.StrategyResult {
	latestData := dailyData[len(dailyData)-1]

	// 1. 检查市值规模
	if !s.checkMarketCap(stock, latestData) {
		return nil
	}

	// 2. 检查盈利能力稳定性
	if !s.checkStability(financialData) {
		return nil
	}

	// 3. 检查竞争优势（简化：通过ROE和利润率判断）
	if !s.checkCompetitiveAdvantage(financialData) {
		return nil
	}

	// 4. 检查长期趋势
	if !s.checkLongTermTrend(dailyData) {
		return nil
	}

	// 5. 计算评分
	score := s.calculateLeaderScore(latestData, dailyData, financialData, indicators)

	if score < 0.6 {
		return nil
	}

	// 6. 生成买卖点建议
	buyPrice := latestData.ClosePrice
	stopLoss := buyPrice * 0.85  // 龙头股相对稳健：-15%
	takeProfit := buyPrice * 1.80 // 龙头股长期收益：+80%

	logic := fmt.Sprintf("行业龙头股，市值%.0f亿,ROE=%.1f%%,长期趋势向上",
		latestData.ClosePrice*100000000/100000000, financialData.ROE) // 简化市值计算

	return &model.StrategyResult{
		StockCode:        stock.Code,
		Score:            score,
		BuyPrice:         buyPrice,
		StopLossPrice:    stopLoss,
		TakeProfitPrice:  takeProfit,
		LogicDescription: logic,
	}
}

// checkValuationConditions 检查估值条件
func (s *LongTermStrategy) checkValuationConditions(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	peThreshold := s.parameters["pe_threshold"].(float64)
	pbThreshold := s.parameters["pb_threshold"].(float64)

	return financialData.PE > 0 && financialData.PE < peThreshold &&
		financialData.PB > 0 && financialData.PB < pbThreshold
}

// checkFinancialHealth 检查财务健康
func (s *LongTermStrategy) checkFinancialHealth(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	roeThreshold := s.parameters["roe_threshold"].(float64)
	debtRatioMax := s.parameters["debt_ratio_max"].(float64)

	return financialData.ROE >= roeThreshold && financialData.DebtRatio <= debtRatioMax
}

// checkDividendCondition 检查股息率条件
func (s *LongTermStrategy) checkDividendCondition(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	dividendYield := s.parameters["dividend_yield"].(float64)
	return financialData.DividendYield >= dividendYield
}

// checkLongTermTrend 检查长期趋势
func (s *LongTermStrategy) checkLongTermTrend(dailyData []model.StockDailyData) bool {
	if len(dailyData) < 200 {
		return false
	}

	// 计算200日均线趋势
	ma200 := s.calculateMA(dailyData, 200)
	ma100 := s.calculateMA(dailyData, 100)

	if ma200 == 0 || ma100 == 0 {
		return false
	}

	// 长期趋势向上：100日均线 > 200日均线
	return ma100 > ma200
}

// checkGrowthConditions 检查成长性条件
func (s *LongTermStrategy) checkGrowthConditions(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	revenueGrowthMin := s.parameters["revenue_growth_min"].(float64)
	profitGrowthMin := s.parameters["profit_growth_min"].(float64)

	return financialData.RevenueGrowth >= revenueGrowthMin &&
		financialData.ProfitGrowth >= profitGrowthMin
}

// checkProfitability 检查盈利能力
func (s *LongTermStrategy) checkProfitability(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	roeMin := s.parameters["roe_min"].(float64)
	return financialData.ROE >= roeMin && financialData.NetMargin > 0
}

// checkIndustryProspect 检查行业前景
func (s *LongTermStrategy) checkIndustryProspect(industry string) bool {
	// 简化实现：假设所有行业都有前景
	// 实际应该根据行业数据判断
	return len(industry) > 0
}

// checkMarketCap 检查市值规模
func (s *LongTermStrategy) checkMarketCap(stock *model.Stock, latestData model.StockDailyData) bool {
	// 简化实现：通过价格判断（实际应该从财务数据获取市值）
	// 这里暂时返回true
	return true
}

// checkStability 检查稳定性
func (s *LongTermStrategy) checkStability(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	roeStable := s.parameters["roe_stable"].(float64)
	// profitStable := s.parameters["profit_stable"].(int) // 暂时未使用，保留以备将来扩展

	// 简化实现：检查ROE是否稳定
	return financialData.ROE >= roeStable
}

// checkCompetitiveAdvantage 检查竞争优势
func (s *LongTermStrategy) checkCompetitiveAdvantage(financialData *model.FinancialData) bool {
	if financialData == nil {
		return false
	}

	// 通过高ROE和净利率判断竞争优势
	return financialData.ROE >= 12.0 && financialData.NetMargin >= 5.0
}

// calculateMA 计算移动平均线
func (s *LongTermStrategy) calculateMA(dailyData []model.StockDailyData, period int) float64 {
	if len(dailyData) < period {
		return 0
	}

	var sum float64
	for i := len(dailyData) - period; i < len(dailyData); i++ {
		sum += dailyData[i].ClosePrice
	}

	return sum / float64(period)
}

// calculateValueScore 计算价值投资策略评分
func (s *LongTermStrategy) calculateValueScore(data model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) float64 {
	score := 0.5

	// 估值评分
	if financialData != nil {
		if financialData.PE < 10 {
			score += 0.15
		} else if financialData.PE < 15 {
			score += 0.1
		}

		if financialData.PB < 1.5 {
			score += 0.1
		} else if financialData.PB < 2.0 {
			score += 0.05
		}

		// 股息率评分
		if financialData.DividendYield >= 5.0 {
			score += 0.15
		} else if financialData.DividendYield >= 3.0 {
			score += 0.1
		}
	}

	// 财务健康评分
	if financialData != nil && financialData.DebtRatio < 0.5 {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// calculateGrowthScore 计算成长股策略评分
func (s *LongTermStrategy) calculateGrowthScore(data model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) float64 {
	score := 0.5

	// 成长性评分
	if financialData != nil {
		if financialData.RevenueGrowth >= 0.3 {
			score += 0.2
		} else if financialData.RevenueGrowth >= 0.2 {
			score += 0.15
		}

		if financialData.ProfitGrowth >= 0.25 {
			score += 0.15
		} else if financialData.ProfitGrowth >= 0.15 {
			score += 0.1
		}
	}

	// 盈利能力评分
	if financialData != nil && financialData.ROE >= 20 {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// calculateLeaderScore 计算龙头股策略评分
func (s *LongTermStrategy) calculateLeaderScore(data model.StockDailyData, dailyData []model.StockDailyData, financialData *model.FinancialData, indicators []model.TechnicalIndicator) float64 {
	score := 0.5

	// 稳定性评分
	if financialData != nil {
		if financialData.ROE >= 15 {
			score += 0.15
		} else if financialData.ROE >= 10 {
			score += 0.1
		}

		if financialData.NetMargin >= 10 {
			score += 0.1
		}
	}

	// 趋势评分
	ma200 := s.calculateMA(dailyData, 200)
	if ma200 > 0 && data.ClosePrice > ma200*1.2 { // 股价高于200日均线20%
		score += 0.15
	}

	return math.Min(score, 1.0)
}

// GetName 获取策略名称
func (s *LongTermStrategy) GetName() string {
	return s.name
}

// GetDescription 获取策略描述
func (s *LongTermStrategy) GetDescription() string {
	return s.description
}

// GetParameters 获取策略参数
func (s *LongTermStrategy) GetParameters() map[string]interface{} {
	return s.parameters
}
