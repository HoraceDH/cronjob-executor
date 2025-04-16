package services

// Created in 2025-03-23 22:06.
// @author Horace

import (
	"fmt"
	logger "github.com/cihub/seelog"
	"github.com/emirpasic/gods/queues/priorityqueue"
	godsutils "github.com/emirpasic/gods/utils"
	"github.com/horacedh/cronjob-executor/bean"
	"github.com/horacedh/cronjob-executor/context"
	"github.com/horacedh/cronjob-executor/task"
	"github.com/horacedh/cronjob-executor/utils"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

// 单例模式
var (
	dispatcherService     DispatcherService
	dispatcherServiceOnce sync.Once
)

// DispatcherService 接口
type DispatcherService interface {
	// Start 开始调度，
	Start(address string)
	// AddTask 添加任务
	AddTask(params *task.TaskParams) int
	// getTask 线程安全的获取任务
	getTask() *task.TaskParams
	// invokeTask 执行任务
	invokeTask(address string, params *task.TaskParams)
}

// dispatcherServiceImpl 实现类
type dispatcherServiceImpl struct {
	mu sync.Mutex
	// taskQueue 任务队列，按照执行时间升序排序
	taskQueue *priorityqueue.Queue
	// handlers 任务处理方法集合，key为包路径+方法名，value为方法的反射值
	handlers map[string]*reflect.Value
	// taskOptions 任务配置集合，key为包路径+方法名，value为任务配置
	taskOptions map[string]*bean.TaskOptions
}

// getTask 线程安全的获取任务
func (dispatcherService *dispatcherServiceImpl) getTask() *task.TaskParams {
	dispatcherService.mu.Lock()
	defer dispatcherService.mu.Unlock()
	taskParams, _ := dispatcherService.taskQueue.Dequeue()
	if taskParams == nil {
		return nil
	}
	return taskParams.(*task.TaskParams)
}

// Start 开始调度
func (dispatcherService *dispatcherServiceImpl) Start(address string) {
	// 启动任务结果发送
	go GetResultSendService().Start()

	// 队列中还有元素，或者还没停机
	for dispatcherService.taskQueue.Size() > 0 || !context.Shutdown.Load() {
		taskParams := dispatcherService.getTask()

		// 如果队列为空，则休眠一段时间，等待任务到来
		if taskParams == nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		// 如果还没达到执行时间，重新添加到队列并休眠指定时间
		var remaining = taskParams.ExecutionTime - time.Now().UnixMilli()
		if remaining > 0 {
			dispatcherService.AddTask(taskParams)
			time.Sleep(time.Duration(remaining)*time.Millisecond - 1)
			continue
		}

		// 执行任务
		go dispatcherService.invokeTask(address, taskParams)
	}

	time.Sleep(time.Second * 3)
	logger.Infof("dispatcher service stopped.")
	context.DispatcherStopped.Store(true)
}

// AddTask 添加任务
func (dispatcherService *dispatcherServiceImpl) AddTask(params *task.TaskParams) int {
	dispatcherService.mu.Lock()
	defer dispatcherService.mu.Unlock()

	dispatcherService.taskQueue.Enqueue(params)
	return dispatcherService.taskQueue.Size()
}

// invokeTask 执行任务
func (dispatcherService *dispatcherServiceImpl) invokeTask(address string, params *task.TaskParams) {
	reflectValue := dispatcherService.handlers[params.Method]
	if reflectValue == nil {
		logger.Warnf("dispatch task error, target method is null, task:%s,", utils.ToJsonString(params))
		GetResultSendService().AddResult(&task.TaskResult{
			TaskLogId: params.TaskLogId,
			TaskId:    params.TaskId,
			State:     task.EXECUTION_FAILED_NOT_FOUND,
		})
		return
	}

	startTime := time.Now().UnixMilli()
	var state = task.EXECUTION_SUCCESS
	var failureReason string
	var handlerResult *task.HandlerResult

	// 如果任务执行发生异常
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("cron job task handler exception, realExecutionTime:%s, executionTime:%s, task:%s, msg:%v",
				utils.FormatTime(startTime), utils.FormatTime(params.ExecutionTime), utils.ToJsonString(params), r)
			state = task.EXECUTION_FAILED
			failureReason = fmt.Sprintf("%v\r\n\r\n%s", r, string(debug.Stack()))
		}

		endTime := time.Now().UnixMilli()
		GetResultSendService().AddResult(&task.TaskResult{
			TaskLogId:         params.TaskLogId,
			TaskId:            params.TaskId,
			State:             state,
			FailedReason:      failureReason,
			RealExecutionTime: startTime,
			ElapsedTime:       int(endTime - startTime),
			Address:           address,
		})
	}()

	// 检查执行延迟
	//delayTime := startTime - params.ExecutionTime
	//if delayTime > 0 {
	//	if params.ReceivedDispatcherTime > params.ExecutionTime {
	//		logger.Warnf("cron job task invoke delay, maybe the network fluctuates or new scheduler instance start, id:%d, delayTime:%dms, executionTime:%s, receivedDispatcherTime:%s, realExecutionTime:%s, taskParams:%s",
	//			params.TaskLogId, delayTime,
	//			utils2.FormatTime(params.ExecutionTime),
	//			utils2.FormatTime(params.ReceivedDispatcherTime),
	//			utils2.FormatTime(startTime),
	//			utils2.ToJsonString(params))
	//	} else {
	//		logger.Warnf("cron job task invoke delay, maybe the GC or CPU of the executor is busy, id:%d, delayTime:%dms, executionTime:%s, receivedDispatcherTime:%s, realExecutionTime:%s, taskParams:%s",
	//			params.TaskLogId, delayTime,
	//			utils2.FormatTime(params.ExecutionTime),
	//			utils2.FormatTime(params.ReceivedDispatcherTime),
	//			utils2.FormatTime(startTime),
	//			utils2.ToJsonString(params))
	//	}
	//}

	// 执行目标方法
	results := reflectValue.Call([]reflect.Value{reflect.ValueOf(params)})
	if len(results) == 0 || results[0].IsNil() {
		state = task.EXECUTION_FAILED
		failureReason = fmt.Sprintf("result is null, please check the return value of the method: %s", params.Method)
		logger.Errorf("cron job task handler failed, result is nil, realExecutionTime:%s, executionTime:%s, params:%s",
			utils.FormatTime(startTime), utils.FormatTime(params.ExecutionTime), utils.ToJsonString(params))
		return
	}

	handlerResult = results[0].Interface().(*task.HandlerResult)
	if !handlerResult.IsSuccess() {
		state = task.EXECUTION_FAILED
		failureReason = fmt.Sprintf("cron job task handler failed, code:%d, msg:%s", handlerResult.Code, handlerResult.Msg)
		logger.Errorf("cron job task handler failed, code:%d, msg:%s, realExecutionTime:%s, executionTime:%s, params:%s",
			handlerResult.Code, handlerResult.Msg,
			utils.FormatTime(startTime), utils.FormatTime(params.ExecutionTime), utils.ToJsonString(params))
	}

}

// InitDispatcherService 初始化
func InitDispatcherService(handlers map[string]*reflect.Value, taskOptions map[string]*bean.TaskOptions) DispatcherService {
	dispatcherServiceOnce.Do(func() {
		dispatcherService = &dispatcherServiceImpl{
			mu:          sync.Mutex{},
			taskQueue:   priorityqueue.NewWith(taskComparator),
			handlers:    handlers,
			taskOptions: taskOptions,
		}
	})
	return dispatcherService
}

// GetDispatcherService 获取实例对象
func GetDispatcherService() DispatcherService {
	return dispatcherService
}

// taskComparator 比较器
func taskComparator(a, b interface{}) int {
	priorityA := a.(*task.TaskParams).ExecutionTime
	priorityB := b.(*task.TaskParams).ExecutionTime
	return godsutils.Int64Comparator(priorityA, priorityB) // "-" descending order
}
