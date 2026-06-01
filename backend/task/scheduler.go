package task

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"stock-strategy-backend/config"
	"stock-strategy-backend/strategy"
)

// TaskScheduler 任务调度器
type TaskScheduler struct {
	cron      *cron.Cron
	config    *config.Config
	engine    *strategy.StrategyEngine
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(cfg *config.Config) *TaskScheduler {
	return &TaskScheduler{
		cron:   cron.New(),
		config: cfg,
		engine: strategy.NewStrategyEngine(),
	}
}

// Start 启动任务调度器
func (s *TaskScheduler) Start() error {
	// 添加每日数据更新任务
	dailyUpdateTime := s.config.Strategy.DailyUpdateTime
	if dailyUpdateTime == "" {
		dailyUpdateTime = "17:45"
	}

	_, err := s.cron.AddFunc("0 "+dailyUpdateTime+" * * 1-5", s.dailyDataUpdateTask)
	if err != nil {
		return err
	}

	// 添加策略执行任务
	strategyRunTime := s.config.Strategy.StrategyRunTime
	if strategyRunTime == "" {
		strategyRunTime = "18:00"
	}

	_, err = s.cron.AddFunc("0 "+strategyRunTime+" * * 1-5", s.strategyExecutionTask)
	if err != nil {
		return err
	}

	// 启动调度器
	s.cron.Start()

	log.Printf("任务调度器已启动，数据更新时间: %s，策略执行时间: %s",
		dailyUpdateTime, strategyRunTime)

	return nil
}

// Stop 停止任务调度器
func (s *TaskScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		log.Println("任务调度器已停止")
	}
}

// dailyDataUpdateTask 每日数据更新任务
func (s *TaskScheduler) dailyDataUpdateTask() {
	if s.config.Strategy.WeekendSkip && isWeekend(time.Now()) {
		log.Println("周末跳过数据更新")
		return
	}

	log.Println("开始执行每日数据更新任务...")

	// 这里调用数据采集模块的每日数据采集功能
	// 实际项目中需要集成数据采集模块
	log.Println("数据更新任务执行完成")
}

// strategyExecutionTask 策略执行任务
func (s *TaskScheduler) strategyExecutionTask() {
	if s.config.Strategy.WeekendSkip && isWeekend(time.Now()) {
		log.Println("周末跳过策略执行")
		return
	}

	log.Println("开始执行策略执行任务...")

	tradeDate := time.Now()

	// 获取所有启用的策略并执行
	strategies := []string{
		"short_term_1", "short_term_2", "short_term_3",
		"medium_term_1", "medium_term_2", "medium_term_3",
		"long_term_1", "long_term_2", "long_term_3",
	}

	successCount := 0
	errorCount := 0

	for _, strategyID := range strategies {
		err := s.engine.ExecuteStrategy(strategyID, tradeDate)
		if err != nil {
			log.Printf("策略 %s 执行失败: %v", strategyID, err)
			errorCount++
		} else {
			log.Printf("策略 %s 执行成功", strategyID)
			successCount++
		}
	}

	log.Printf("策略执行任务完成，成功: %d，失败: %d", successCount, errorCount)
}

// isWeekend 判断是否为周末
func isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// RunStrategyImmediately 立即执行指定策略
func (s *TaskScheduler) RunStrategyImmediately(strategyID string) error {
	log.Printf("立即执行策略: %s", strategyID)

	err := s.engine.ExecuteStrategy(strategyID, time.Now())
	if err != nil {
		log.Printf("策略 %s 立即执行失败: %v", strategyID, err)
		return err
	}

	log.Printf("策略 %s 立即执行成功", strategyID)
	return nil
}

// RunAllStrategiesImmediately 立即执行所有策略
func (s *TaskScheduler) RunAllStrategiesImmediately() error {
	log.Println("立即执行所有策略")

	strategies := []string{
		"short_term_1", "short_term_2", "short_term_3",
		"medium_term_1", "medium_term_2", "medium_term_3",
		"long_term_1", "long_term_2", "long_term_3",
	}

	successCount := 0
	errorCount := 0

	for _, strategyID := range strategies {
		err := s.engine.ExecuteStrategy(strategyID, time.Now())
		if err != nil {
			log.Printf("策略 %s 执行失败: %v", strategyID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("所有策略立即执行完成，成功: %d，失败: %d", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d 个策略执行失败", errorCount)
	}

	return nil
}