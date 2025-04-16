package task

// Created in 2025-03-18 20:15.
// @author Horace

// TaskParams 任务参数
type TaskParams struct {
	// Page 页码
	Page int32 `json:"page"`
	// Total 总页数
	Total int32 `json:"total"`
	// TaskLogId 任务日志ID
	TaskLogId int64 `json:"taskLogId"`
	// TaskId 任务ID
	TaskId int64 `json:"taskId"`
	// Method 任务方法，类全限定名
	Method string `json:"method"`
	// ExeType 执行类型，0：表示常规任务调度，1：管理后台立即执行，2：过期执行
	ExeType int32 `json:"exeType"`
	// Cron 表达式
	Cron string `json:"cron"`
	// Tag 任务标签
	Tag string `json:"tag"`
	// ExecutionTime 执行时间
	ExecutionTime int64 `json:"executionTime"`
	// ReceivedDispatcherTime 调度时间，这里表示接到调度任务的时间，并不是服务端记录的调度时间，查问题时也可以拿来与服务端做对比，服务端是先更新调度时间，再发起请求给执行器
	ReceivedDispatcherTime int64 `json:"receivedDispatcherTime"`
	// Params 任务自定义参数
	Params string `json:"params"`
}

// HandlerResult 任务处理结果
type HandlerResult struct {
	// Code 处理结果编码
	Code int32
	// Msg 处理结果描述信息
	Msg string
}

// IsSuccess 是否成功
func (result *HandlerResult) IsSuccess() bool {
	return result.Code == 0
}

// Success 成功
func Success() *HandlerResult {
	return &HandlerResult{Code: 0, Msg: "success"}
}

// Failed 失败
func Failed(msg string) *HandlerResult {
	return &HandlerResult{Code: 1, Msg: msg}
}

// TaskHandler 任务处理器接口
type TaskHandler interface {
	// Handle 任务处理方法
	Handle(params *TaskParams) *HandlerResult
}

// TaskLogState 任务日志状态
type TaskLogState int32

const (
	// INITIALIZE 初始化
	INITIALIZE TaskLogState = 1
	// QUEUEING 队列中
	QUEUEING TaskLogState = 2
	// EXECUTION 调度中
	EXECUTION TaskLogState = 3
	// EXECUTION_SUCCESS 执行成功
	EXECUTION_SUCCESS TaskLogState = 4
	// EXECUTION_FAILED 执行失败
	EXECUTION_FAILED TaskLogState = 5
	// EXECUTION_CANCEL 取消执行（预生成日志后，任务取消等情况）
	EXECUTION_CANCEL TaskLogState = 6
	// EXECUTION_EXPIRED 任务过期（超过执行时间而未被调度）
	EXECUTION_EXPIRED TaskLogState = 7
	// EXECUTION_FAILED_DISCARD 执行失败，已丢弃
	EXECUTION_FAILED_DISCARD TaskLogState = 8
	// EXECUTION_FAILED_RETRYING 执行失败，重试中
	EXECUTION_FAILED_RETRYING TaskLogState = 9
	// EXECUTION_FAILED_NOT_FOUND 执行失败，未找到执行方法
	EXECUTION_FAILED_NOT_FOUND TaskLogState = 10
)

// TaskResult 任务执行结果
type TaskResult struct {
	// TaskLogId 任务日志ID
	TaskLogId int64 `json:"taskLogId"`
	// TaskId 任务ID
	TaskId int64 `json:"taskId"`
	// State 任务状态
	State TaskLogState `json:"state"`
	// FailedReason 失败原因
	FailedReason string `json:"failedReason"`
	// RealExecutionTime 实际执行时间
	RealExecutionTime int64 `json:"realExecutionTime"`
	// ElapsedTime 耗时
	ElapsedTime int `json:"elapsedTime"`
	// Address 执行器地址
	Address string `json:"address"`
}
