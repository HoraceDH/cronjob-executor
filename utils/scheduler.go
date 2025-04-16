package utils

// Created in 2025-03-23 14:54.
// @author Horace

import (
	logger "github.com/cihub/seelog"
	"sync"
	"time"
)

// 单例模式
var (
	scheduler     Scheduler
	schedulerOnce sync.Once
)

// Scheduler 接口
type Scheduler interface {
	// 开始调度
	start(period time.Duration)
	// ScheduleAtFixedRate 按照固定频率运行
	ScheduleAtFixedRate(period time.Duration, initRun bool, task func())
}

// SchedulerImpl 实现类
type schedulerImpl struct {
	// mu 互斥锁
	mu sync.Mutex
	// tickers ticker map集合，key是period值，value是ticker
	tickers map[time.Duration]*time.Ticker
	// tasksMap 周期性任务集合，key是period值，value是函数数组
	tasksMap map[time.Duration][]func()
}

// init 开始调度
func (scheduler *schedulerImpl) start(period time.Duration) {
	// 在携程中扫描并启动Ticker
	go func() {
		ticker := scheduler.tickers[period]
		if ticker == nil {
			logger.Errorf("start scheduler fail, ticker is nil, period: %s", period)
			return
		}

		// 达到调度周期后，会有回调
		for range ticker.C {
			// 获取所有的任务并发起调用
			tasks := scheduler.tasksMap[period]

			// 在携程中调用调度的任务
			for _, task := range tasks {
				go task()
			}
		}
	}()
}

// ScheduleAtFixedRate 按照固定频率运行
func (scheduler *schedulerImpl) ScheduleAtFixedRate(period time.Duration, initRun bool, task func()) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	// 如果启动时运行一次
	if initRun {
		go task()
	}

	// 初始化ticker
	ticker := scheduler.tickers[period]
	if ticker == nil {
		ticker = time.NewTicker(period)
		scheduler.tickers[period] = ticker

		// 开始调度
		scheduler.start(period)
	}

	// 添加任务到集合中
	tasks := scheduler.tasksMap[period]
	if tasks == nil {
		tasks = make([]func(), 0)
	}
	tasks = append(tasks, task)
	scheduler.tasksMap[period] = tasks
}

// GetScheduler 获取实例对象
func GetScheduler() Scheduler {
	schedulerOnce.Do(func() {
		scheduler = &schedulerImpl{
			mu:       sync.Mutex{},
			tickers:  make(map[time.Duration]*time.Ticker),
			tasksMap: make(map[time.Duration][]func()),
		}
	})
	return scheduler
}
