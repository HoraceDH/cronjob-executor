package services

// Created in 2025-03-26 16:06.
// @author Horace

import (
	logger "github.com/cihub/seelog"
	"github.com/emirpasic/gods/queues/priorityqueue"
	"github.com/emirpasic/gods/utils"
	"github.com/horacedh/cronjob-executor/context"
	"github.com/horacedh/cronjob-executor/task"
	"sync"
	"time"
)

// 单例模式
var (
	resultSendService     ResultSendService
	resultSendServiceOnce sync.Once
)

// ResultSendService 接口
type ResultSendService interface {
	// Start 开始发送任务结果
	Start()
	// AddResult 添加任务结果
	AddResult(result *task.TaskResult) int
}

// resultSendServiceImpl 实现类
type resultSendServiceImpl struct {
	mu sync.Mutex
	// resultQueue 任务结果队列，按照执行时间升序排序
	resultQueue *priorityqueue.Queue
}

// AddResult 添加任务结果
func (resultSendService *resultSendServiceImpl) AddResult(result *task.TaskResult) int {
	resultSendService.mu.Lock()
	defer resultSendService.mu.Unlock()

	resultSendService.resultQueue.Enqueue(result)
	return resultSendService.resultQueue.Size()
}

// Start 开始发送任务结果
func (resultSendService *resultSendServiceImpl) Start() {
	context.WaitGroup.Add(1)

	// 没有停机或者队列还有元素
	for resultSendService.resultQueue.Size() > 0 || !context.Shutdown.Load() || !context.DispatcherStopped.Load() {
		taskResult := resultSendService.getTaskResult()

		// 如果队列为空，则休眠一段时间，等待任务到来
		if taskResult == nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		// 发送http请求
		success := GetOpenApiService().SendTaskResult(taskResult)
		if !success {
			resultSendService.AddResult(taskResult)
		}
	}

	logger.Infof("result send service is stopped.")
	context.WaitGroup.Done()
}

// getTaskResult 线程安全的获取任务结果
func (resultSendService *resultSendServiceImpl) getTaskResult() *task.TaskResult {
	resultSendService.mu.Lock()
	defer resultSendService.mu.Unlock()
	taskResult, _ := resultSendService.resultQueue.Dequeue()
	if taskResult == nil {
		return nil
	}
	return taskResult.(*task.TaskResult)
}

// GetResultSendService 获取实例对象
func GetResultSendService() ResultSendService {
	resultSendServiceOnce.Do(func() {
		resultSendService = &resultSendServiceImpl{
			mu:          sync.Mutex{},
			resultQueue: priorityqueue.NewWith(taskResultComparator),
		}
	})
	return resultSendService
}

// taskResultComparator 比较器
func taskResultComparator(a, b interface{}) int {
	priorityA := a.(*task.TaskResult).RealExecutionTime
	priorityB := b.(*task.TaskResult).RealExecutionTime
	return utils.Int64Comparator(priorityA, priorityB) // "-" descending order
}
