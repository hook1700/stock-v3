package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stock-strategy-backend/model"
)

// StockHandler 股票相关API处理器
type StockHandler struct {
	stockRepo     *model.StockRepository
	dailyDataRepo *model.StockDailyDataRepository
}

// NewStockHandler 创建股票处理器
func NewStockHandler() *StockHandler {
	return &StockHandler{
		stockRepo:     &model.StockRepository{},
		dailyDataRepo: &model.StockDailyDataRepository{},
	}
}

// GetStockList 获取股票列表
func (h *StockHandler) GetStockList(c *gin.Context) {
	var req model.StockFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	if req.SortBy == "" {
		req.SortBy = "code"
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc"
	}

	// 获取所有股票
	stocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取股票列表失败"})
		return
	}

	// 应用筛选条件
	filteredStocks := h.filterStocks(stocks, req)

	// 分页
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > len(filteredStocks) {
		end = len(filteredStocks)
	}

	pagedStocks := filteredStocks[start:end]

	// 构建响应
	response := model.StockListResponse{
		Stocks:     pagedStocks,
		TotalCount: len(filteredStocks),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetStockDetail 获取股票详情
func (h *StockHandler) GetStockDetail(c *gin.Context) {
	stockCode := c.Param("code")

	// 获取股票基本信息
	stock, err := h.stockRepo.GetStockByCode(stockCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "股票不存在"})
		return
	}

	// 获取最新日线数据
	dailyData, err := h.dailyDataRepo.GetLatestData(stockCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取股票数据失败"})
		return
	}

	response := gin.H{
		"stock":      stock,
		"daily_data": dailyData,
	}

	c.JSON(http.StatusOK, response)
}

// GetStockHistory 获取股票历史数据
func (h *StockHandler) GetStockHistory(c *gin.Context) {
	stockCode := c.Param("code")
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

	// 获取历史数据
	historyData, err := h.dailyDataRepo.GetDailyData(stockCode, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stock_code": stockCode,
		"start_date": startDateStr,
		"end_date":   endDateStr,
		"data":       historyData,
		"count":      len(historyData),
	})
}

// filterStocks 筛选股票
func (h *StockHandler) filterStocks(stocks []model.Stock, req model.StockFilterRequest) []model.StockListItem {
	var result []model.StockListItem

	for _, stock := range stocks {
		// 名称关键词筛选
		if req.NameKeyword != "" {
			if !containsString(stock.Name, req.NameKeyword) && !containsString(stock.Code, req.NameKeyword) {
				continue
			}
		}

		// 行业筛选
		if req.Industry != "" && stock.Industry != req.Industry {
			continue
		}

		// 获取最新价格数据
		dailyData, err := h.dailyDataRepo.GetLatestData(stock.Code)
		if err != nil {
			continue
		}

		// 构建列表项
		item := model.StockListItem{
			Code:         stock.Code,
			Name:         stock.Name,
			Industry:     stock.Industry,
			ClosePrice:   dailyData.ClosePrice,
			ChangeRate:   0, // 需要计算涨跌幅
			Amount:       dailyData.Amount,
			TurnoverRate: dailyData.TurnoverRate,
		}

		result = append(result, item)
	}

	// 排序
	result = h.sortStocks(result, req.SortBy, req.SortOrder)

	return result
}

// sortStocks 排序股票列表
func (h *StockHandler) sortStocks(stocks []model.StockListItem, sortBy, sortOrder string) []model.StockListItem {
	// 简化的排序实现
	// 实际项目中需要实现完整的排序逻辑
	return stocks
}

// containsString 检查字符串包含
func containsString(s, substr string) bool {
	// 简化的包含检查
	// 实际项目中可能需要更复杂的匹配逻辑
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetIndustries 获取行业列表
func (h *StockHandler) GetIndustries(c *gin.Context) {
	stocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取行业列表失败"})
		return
	}

	industryMap := make(map[string]bool)
	for _, stock := range stocks {
		if stock.Industry != "" {
			industryMap[stock.Industry] = true
		}
	}

	var industries []string
	for industry := range industryMap {
		industries = append(industries, industry)
	}

	c.JSON(http.StatusOK, gin.H{"industries": industries, "count": len(industries)})
}

// SearchStocks 搜索股票
func (h *StockHandler) SearchStocks(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	stocks, err := h.stockRepo.GetAllStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索股票失败"})
		return
	}

	var results []model.StockListItem
	for _, stock := range stocks {
		if containsString(stock.Name, keyword) || containsString(stock.Code, keyword) {
			item := model.StockListItem{
				Code:     stock.Code,
				Name:     stock.Name,
				Industry: stock.Industry,
			}
			results = append(results, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stocks": results,
		"count":  len(results),
	})
}