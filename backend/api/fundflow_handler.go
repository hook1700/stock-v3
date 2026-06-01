package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stock-strategy-backend/model"
)

// FundFlowHandler 资金流处理器
type FundFlowHandler struct {
	flowRepo *model.SectorFundFlowRepository
}

// NewFundFlowHandler 创建资金流处理器
func NewFundFlowHandler() *FundFlowHandler {
	return &FundFlowHandler{
		flowRepo: &model.SectorFundFlowRepository{},
	}
}

// GetSectorFundFlow 获取板块资金流
func (h *FundFlowHandler) GetSectorFundFlow(c *gin.Context) {
	flows, err := h.flowRepo.GetFlowsByDate(time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取板块资金流失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"flows": flows, "count": len(flows)})
}

// GetSectorHistory 获取板块历史资金流
func (h *FundFlowHandler) GetSectorHistory(c *gin.Context) {
	sector := c.Param("sector")
	days := 30 // 默认30天
	flows, err := h.flowRepo.GetFlowsBySector(sector, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史资金流失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sector": sector, "flows": flows, "count": len(flows)})
}

// GetFundFlowSummary 获取资金流摘要
func (h *FundFlowHandler) GetFundFlowSummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "资金流摘要", "date": time.Now().Format("2006-01-02")})
}

// GetFundFlowTrend 获取资金流趋势
func (h *FundFlowHandler) GetFundFlowTrend(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "资金流趋势", "date": time.Now().Format("2006-01-02")})
}
