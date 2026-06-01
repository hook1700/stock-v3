package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stock-strategy-backend/model"
	"stock-strategy-backend/strategy"
)

// StrategyHandler 策略相关API处理器
type StrategyHandler struct {
	strategyRepo *model.StrategyRepository
	resultRepo   *model.StrategyResultRepository
	engine       *strategy.StrategyEngine
}

// NewStrategyHandler 创建策略处理器
func NewStrategyHandler() *StrategyHandler {
	return &StrategyHandler{
		strategyRepo: &model.StrategyRepository{},
		resultRepo:   &model.StrategyResultRepository{},
		engine:       strategy.NewStrategyEngine(),
	}
}

// GetStrategies 获取所有策略
func (h *StrategyHandler) GetStrategies(c *gin.Context) {
	strategies, err := h.strategyRepo.GetEnabledStrategies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": strategies,
		"count":      len(strategies),
	})
}

// GetStrategyResults 获取策略执行结果
func (h *StrategyHandler) GetStrategyResults(c *gin.Context) {
	strategyID := c.Param("strategy_id")
	dateStr := c.Query("date")

	var tradeDate time.Time
	var err error

	if dateStr == "" {
		tradeDate = time.Now()
	} else {
		tradeDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
			return
		}
	}

	// 获取策略结果
	results, err := h.resultRepo.GetResultsByDateAndStrategy(tradeDate, strategyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略结果失败"})
		return
	}

	// 获取策略信息
	strategyInfo, err := h.strategyRepo.GetStrategyByID(strategyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略信息失败"})
		return
	}

	// 构建响应
	response := model.StrategyResultResponse{
		StrategyID:   strategyID,
		StrategyName: strategyInfo.Name,
		TradeDate:    tradeDate.Format("2006-01-02"),
		Results:      h.buildResultItems(results),
		TotalCount:   len(results),
	}

	c.JSON(http.StatusOK, response)
}

// RunStrategy 执行策略
func (h *StrategyHandler) RunStrategy(c *gin.Context) {
	strategyID := c.Param("strategy_id")
	dateStr := c.Query("date")

	var tradeDate time.Time
	var err error

	if dateStr == "" {
		tradeDate = time.Now()
	} else {
		tradeDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误"})
			return
		}
	}

	// 执行策略
	err = h.engine.ExecuteStrategy(strategyID, tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "策略执行失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "策略执行成功",
		"strategy_id": strategyID,
		"trade_date":  tradeDate.Format("2006-01-02"),
	})
}

// GetStrategyHistory 获取策略历史结果
func (h *StrategyHandler) GetStrategyHistory(c *gin.Context) {
	strategyID := c.Param("strategy_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 解析日期
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
		return
	}

	// 获取历史结果
	historyResults, err := h.resultRepo.GetResultsByStrategyAndPeriod(strategyID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略历史失败"})
		return
	}

	// 按日期分组结果
	resultsByDate := make(map[string][]model.StrategyResult)
	for _, result := range historyResults {
		dateStr := result.TradeDate.Format("2006-01-02")
		resultsByDate[dateStr] = append(resultsByDate[dateStr], result)
	}

	c.JSON(http.StatusOK, gin.H{
		"strategy_id": strategyID,
		"start_date":  startDateStr,
		"end_date":    endDateStr,
		"results":     resultsByDate,
		"total_days":  len(resultsByDate),
	})
}

// buildResultItems 构建策略结果项
func (h *StrategyHandler) buildResultItems(results []model.StrategyResult) []model.StrategyResultItem {
	var items []model.StrategyResultItem
	stockRepo := &model.StockRepository{}

	for _, result := range results {
		// 获取股票名称
		stock, err := stockRepo.GetStockByCode(result.StockCode)
		stockName := ""
		if err == nil {
			stockName = stock.Name
		}

		item := model.StrategyResultItem{
			StockCode:       result.StockCode,
			StockName:       stockName,
			Score:           result.Score,
			BuyPrice:        result.BuyPrice,
			StopLossPrice:   result.StopLossPrice,
			TakeProfitPrice: result.TakeProfitPrice,
			Logic:           result.LogicDescription,
		}

		items = append(items, item)
	}

	return items
}

// GetStrategiesByType 按类型获取策略
func (h *StrategyHandler) GetStrategiesByType(c *gin.Context) {
	strategyType := c.Param("type")

	strategies, err := h.strategyRepo.GetEnabledStrategies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略列表失败"})
		return
	}

	var filtered []model.Strategy
	for _, s := range strategies {
		if s.StrategyType == strategyType {
			filtered = append(filtered, s)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": filtered,
		"type":       strategyType,
		"count":      len(filtered),
	})
}

// UpdateStrategyStatus 更新策略状态
func (h *StrategyHandler) UpdateStrategyStatus(c *gin.Context) {
	strategyID := c.Param("strategy_id")

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	strategy, err := h.strategyRepo.GetStrategyByID(strategyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "策略不存在"})
		return
	}

	strategy.Enabled = req.Enabled
	if err := model.DB.Save(strategy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新策略状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "策略状态更新成功",
		"strategy_id": strategyID,
		"enabled":     req.Enabled,
	})
}