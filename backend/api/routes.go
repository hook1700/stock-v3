package api

import (
	"github.com/gin-gonic/gin"
)

// RegisterStockRoutes 注册股票相关路由
func RegisterStockRoutes(router *gin.Engine) {
	handler := NewStockHandler()

	stockGroup := router.Group("/api/stocks")
	{
		stockGroup.POST("/list", handler.GetStockList)
		stockGroup.GET("/:code", handler.GetStockDetail)
		stockGroup.GET("/:code/history", handler.GetStockHistory)
		stockGroup.GET("/industries", handler.GetIndustries)
		stockGroup.GET("/search", handler.SearchStocks)
	}
}

// RegisterStrategyRoutes 注册策略相关路由
func RegisterStrategyRoutes(router *gin.Engine) {
	handler := NewStrategyHandler()

	strategyGroup := router.Group("/api/strategies")
	{
		strategyGroup.GET("", handler.GetStrategies)
		strategyGroup.GET("/:strategy_id/results", handler.GetStrategyResults)
		strategyGroup.POST("/:strategy_id/run", handler.RunStrategy)
		strategyGroup.GET("/:strategy_id/history", handler.GetStrategyHistory)
		strategyGroup.GET("/types/:type", handler.GetStrategiesByType)
		strategyGroup.PUT("/:strategy_id/status", handler.UpdateStrategyStatus)
	}
}

// RegisterFundFlowRoutes 注册资金流相关路由
func RegisterFundFlowRoutes(router *gin.Engine) {
	handler := NewFundFlowHandler()

	fundFlowGroup := router.Group("/api/fund-flow")
	{
		fundFlowGroup.GET("/sectors", handler.GetSectorFundFlow)
		fundFlowGroup.GET("/sectors/:sector/history", handler.GetSectorHistory)
		fundFlowGroup.GET("/summary", handler.GetFundFlowSummary)
		fundFlowGroup.GET("/trend", handler.GetFundFlowTrend)
	}
}

// RegisterUserRoutes 注册用户相关路由
func RegisterUserRoutes(router *gin.Engine) {
	handler := NewUserHandler()

	userGroup := router.Group("/api/user")
	{
		userGroup.POST("/operations", handler.SaveUserOperation)
		userGroup.GET("/operations", handler.GetUserOperations)
		userGroup.GET("/preferences", handler.GetUserPreferences)
		userGroup.PUT("/preferences", handler.UpdateUserPreferences)
	}
}

// RegisterAllRoutes 注册所有路由
func RegisterAllRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": "2026-06-01T00:00:00Z",
			"version":   "1.0.0",
		})
	})

	// API路由组
	apiGroup := router.Group("/api")
	{
		// 股票相关
		stockGroup := apiGroup.Group("/stocks")
		{
			stockGroup.POST("/list", GetStockList)
			stockGroup.GET("/:code", GetStockDetail)
			stockGroup.GET("/:code/history", GetStockHistory)
			stockGroup.GET("/industries", GetIndustries)
			stockGroup.GET("/search", SearchStocks)
		}

		// 策略相关
		strategyGroup := apiGroup.Group("/strategies")
		{
			strategyGroup.GET("", GetStrategies)
			strategyGroup.GET("/:strategy_id/results", GetStrategyResults)
			strategyGroup.POST("/:strategy_id/run", RunStrategy)
			strategyGroup.GET("/:strategy_id/history", GetStrategyHistory)
			strategyGroup.GET("/types/:type", GetStrategiesByType)
			strategyGroup.PUT("/:strategy_id/status", UpdateStrategyStatus)
		}

		// 资金流相关
		fundFlowGroup := apiGroup.Group("/fund-flow")
		{
			fundFlowGroup.GET("/sectors", GetSectorFundFlow)
			fundFlowGroup.GET("/sectors/:sector/history", GetSectorHistory)
			fundFlowGroup.GET("/summary", GetFundFlowSummary)
			fundFlowGroup.GET("/trend", GetFundFlowTrend)
		}

		// 用户相关
		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/operations", SaveUserOperation)
			userGroup.GET("/operations", GetUserOperations)
			userGroup.GET("/preferences", GetUserPreferences)
			userGroup.PUT("/preferences", UpdateUserPreferences)
		}

		// 系统管理
		systemGroup := apiGroup.Group("/system")
		{
			systemGroup.GET("/status", GetSystemStatus)
			systemGroup.POST("/tasks/run", RunTaskImmediately)
			systemGroup.GET("/tasks/status", GetTaskStatus)
		}
	}
}

// GetStockList 获取股票列表
func GetStockList(c *gin.Context) {
	handler := NewStockHandler()
	handler.GetStockList(c)
}

// GetStockDetail 获取股票详情
func GetStockDetail(c *gin.Context) {
	handler := NewStockHandler()
	handler.GetStockDetail(c)
}

// GetStockHistory 获取股票历史数据
func GetStockHistory(c *gin.Context) {
	handler := NewStockHandler()
	handler.GetStockHistory(c)
}

// GetIndustries 获取行业列表
func GetIndustries(c *gin.Context) {
	handler := NewStockHandler()
	handler.GetIndustries(c)
}

// SearchStocks 搜索股票
func SearchStocks(c *gin.Context) {
	handler := NewStockHandler()
	handler.SearchStocks(c)
}

// GetStrategies 获取策略列表
func GetStrategies(c *gin.Context) {
	handler := NewStrategyHandler()
	handler.GetStrategies(c)
}

// GetStrategyResults 获取策略结果
func GetStrategyResults(c *gin.Context) {
	handler := NewStrategyHandler()
	handler.GetStrategyResults(c)
}

// RunStrategy 执行策略
func RunStrategy(c *gin.Context) {
	handler := NewStrategyHandler()
	handler.RunStrategy(c)
}

// GetStrategyHistory 获取策略历史
func GetStrategyHistory(c *gin.Context) {
	handler := NewStrategyHandler()
	handler.GetStrategyHistory(c)
}

// GetStrategiesByType 按类型获取策略
func GetStrategiesByType(c *gin.Context) {
	handler := NewStrategyHandler()
	handler.GetStrategiesByType(c)
}

// UpdateStrategyStatus 更新策略状态
func UpdateStrategyStatus(c *gin.Context) {
	handler := NewStrategyHandler()
	handler.UpdateStrategyStatus(c)
}

// GetSectorFundFlow 获取板块资金流
func GetSectorFundFlow(c *gin.Context) {
	handler := NewFundFlowHandler()
	handler.GetSectorFundFlow(c)
}

// GetSectorHistory 获取板块历史资金流
func GetSectorHistory(c *gin.Context) {
	handler := NewFundFlowHandler()
	handler.GetSectorHistory(c)
}

// GetFundFlowSummary 获取资金流摘要
func GetFundFlowSummary(c *gin.Context) {
	handler := NewFundFlowHandler()
	handler.GetFundFlowSummary(c)
}

// GetFundFlowTrend 获取资金流趋势
func GetFundFlowTrend(c *gin.Context) {
	handler := NewFundFlowHandler()
	handler.GetFundFlowTrend(c)
}

// SaveUserOperation 保存用户操作
func SaveUserOperation(c *gin.Context) {
	handler := NewUserHandler()
	handler.SaveUserOperation(c)
}

// GetUserOperations 获取用户操作记录
func GetUserOperations(c *gin.Context) {
	handler := NewUserHandler()
	handler.GetUserOperations(c)
}

// GetUserPreferences 获取用户偏好
func GetUserPreferences(c *gin.Context) {
	handler := NewUserHandler()
	handler.GetUserPreferences(c)
}

// UpdateUserPreferences 更新用户偏好
func UpdateUserPreferences(c *gin.Context) {
	handler := NewUserHandler()
	handler.UpdateUserPreferences(c)
}

// GetSystemStatus 获取系统状态
func GetSystemStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":     "running",
		"version":    "1.0.0",
		"start_time": "2026-06-01T00:00:00Z",
		"uptime":     "24h",
	})
}

// RunTaskImmediately 立即执行任务
func RunTaskImmediately(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "任务已触发",
		"task_id": "task_001",
	})
}

// GetTaskStatus 获取任务状态
func GetTaskStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"tasks": []gin.H{
			{
				"id":     "daily_update",
				"status": "completed",
				"last_run": "2026-06-01T17:45:00Z",
			},
			{
				"id":     "strategy_run",
				"status": "completed",
				"last_run": "2026-06-01T18:00:00Z",
			},
		},
	})
}