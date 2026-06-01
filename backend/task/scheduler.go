package task

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"stock-strategy-backend/config"
	"stock-strategy-backend/model"
	"stock-strategy-backend/strategy"
)

// TaskScheduler 任务调度器
type TaskScheduler struct {
	cron         *cron.Cron
	config       *config.Config
	engine       *strategy.StrategyEngine
	strategyRepo *model.StrategyRepository
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(cfg *config.Config) *TaskScheduler {
	return &TaskScheduler{
		cron:         cron.New(),
		config:       cfg,
		engine:       strategy.NewStrategyEngine(),
		strategyRepo: &model.StrategyRepository{},
	}
}

// parseTimeToCronSpec 将 "HH:MM" 格式时间转换为 cron 分钟和小时字段
func parseTimeToCronSpec(timeStr string) (string, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("无效的时间格式: %s，期望格式为 HH:MM", timeStr)
	}
	hour := parts[0]
	minute := parts[1]
	// 基本校验
	if len(hour) == 0 || len(minute) == 0 {
		return "", fmt.Errorf("无效的时间格式: %s", timeStr)
	}
	return minute + " " + hour, nil
}

// Start 启动任务调度器
func (s *TaskScheduler) Start() error {
	// 先初始化数据库中的默认策略（如果不存在）
	if err := s.initDefaultStrategies(); err != nil {
		log.Printf("初始化默认策略失败: %v", err)
	}

	// 添加每日数据更新任务
	dailyUpdateTime := s.config.Strategy.DailyUpdateTime
	if dailyUpdateTime == "" {
		dailyUpdateTime = "17:45"
	}

	dailyCronSpec, err := parseTimeToCronSpec(dailyUpdateTime)
	if err != nil {
		return err
	}

	_, err = s.cron.AddFunc(dailyCronSpec+" * * 1-5", s.dailyDataUpdateTask)
	if err != nil {
		return err
	}

	// 添加策略执行任务
	strategyRunTime := s.config.Strategy.StrategyRunTime
	if strategyRunTime == "" {
		strategyRunTime = "18:00"
	}

	strategyCronSpec, err := parseTimeToCronSpec(strategyRunTime)
	if err != nil {
		return err
	}

	_, err = s.cron.AddFunc(strategyCronSpec+" * * 1-5", s.strategyExecutionTask)
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

// initDefaultStrategies 初始化默认策略到数据库
func (s *TaskScheduler) initDefaultStrategies() error {
	defaultStrategies := []model.Strategy{
		{StrategyID: "short_term_1", Name: "短线策略1-均线回踩", StrategyType: "short_term", Description: "短期均线回踩买入策略", Enabled: true},
		{StrategyID: "short_term_2", Name: "短线策略2-突破回踩", StrategyType: "short_term", Description: "短期突破回踩策略", Enabled: true},
		{StrategyID: "short_term_3", Name: "短线策略3-强势股反弹", StrategyType: "short_term", Description: "短期强势股反弹策略", Enabled: true},
		{StrategyID: "medium_term_1", Name: "中线策略1-均线回踩", StrategyType: "medium_term", Description: "中期均线回踩买入策略", Enabled: true},
		{StrategyID: "medium_term_2", Name: "中线策略2-突破回踩", StrategyType: "medium_term", Description: "中期突破回踩策略", Enabled: true},
		{StrategyID: "medium_term_3", Name: "中线策略3-强势股反弹", StrategyType: "medium_term", Description: "中期强势股反弹策略", Enabled: true},
		{StrategyID: "long_term_1", Name: "长线策略1-均线回踩", StrategyType: "long_term", Description: "长期均线回踩买入策略", Enabled: true},
		{StrategyID: "long_term_2", Name: "长线策略2-突破回踩", StrategyType: "long_term", Description: "长期突破回踩策略", Enabled: true},
		{StrategyID: "long_term_3", Name: "长线策略3-强势股反弹", StrategyType: "long_term", Description: "长期强势股反弹策略", Enabled: true},
	}

	for _, st := range defaultStrategies {
		_, err := s.strategyRepo.GetStrategyByID(st.StrategyID)
		if err != nil {
			// 策略不存在，创建默认策略
			if err := s.strategyRepo.CreateStrategy(&st); err != nil {
				log.Printf("创建默认策略 %s 失败: %v", st.StrategyID, err)
			} else {
				log.Printf("创建默认策略: %s", st.StrategyID)
			}
		}
	}

	return nil
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

	// 从数据库获取所有启用的策略
	strategies, err := s.strategyRepo.GetEnabledStrategies()
	if err != nil {
		log.Printf("获取策略列表失败: %v", err)
		return
	}

	if len(strategies) == 0 {
		log.Println("没有启用的策略，跳过执行")
		return
	}

	successCount := 0
	errorCount := 0

	for _, strategy := range strategies {
		err := s.engine.ExecuteStrategy(strategy.StrategyID, tradeDate)
		if err != nil {
			log.Printf("策略 %s 执行失败: %v", strategy.StrategyID, err)
			errorCount++
		} else {
			log.Printf("策略 %s 执行成功", strategy.StrategyID)
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

	// 从数据库获取所有启用的策略
	strategies, err := s.strategyRepo.GetEnabledStrategies()
	if err != nil {
		log.Printf("获取策略列表失败: %v", err)
		return err
	}

	if len(strategies) == 0 {
		log.Println("没有启用的策略")
		return fmt.Errorf("没有启用的策略")
	}

	successCount := 0
	errorCount := 0

	for _, strategy := range strategies {
		err := s.engine.ExecuteStrategy(strategy.StrategyID, time.Now())
		if err != nil {
			log.Printf("策略 %s 执行失败: %v", strategy.StrategyID, err)
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