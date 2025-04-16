package cronjob

// Created in 2025-03-18 20:00.
// @author Horace

import (
	logger "github.com/cihub/seelog"
	"github.com/horacedh/cronjob-executor/bean"
	"github.com/horacedh/cronjob-executor/context"
	"github.com/horacedh/cronjob-executor/httpclients"
	_ "github.com/horacedh/cronjob-executor/loggers"
	"github.com/horacedh/cronjob-executor/services"
	"github.com/horacedh/cronjob-executor/task"
	"github.com/horacedh/cronjob-executor/utils"
	"github.com/horacedh/cronjob-executor/webserver"
	"reflect"
	"sync"
	"time"
)

// 单例模式
var (
	executorClient     ExecutorClient
	executorClientOnce sync.Once
)

// ExecutorClient 接口
type ExecutorClient interface {
	// AddTask 添加任务
	AddTask(handler task.TaskHandler, options bean.TaskOptions)
	// Start 启动执行器客户端
	Start()
	// Stop 停止执行器客户端
	stop()
}

// ExecutorClientImpl 实现类
type executorClientImpl struct {
	// options 配置参数
	options bean.ExecutorOptions
	// handlers 任务处理方法集合，key为包路径+方法名，value为方法的反射值
	handlers map[string]*reflect.Value
	// taskOptions 任务配置集合，key为包路径+方法名，value为任务配置
	taskOptions map[string]*bean.TaskOptions
}

// AddTask 添加任务处理器
func (client *executorClientImpl) AddTask(handler task.TaskHandler, options bean.TaskOptions) {
	// 检查参数
	if options.Cron == "" {
		panic("invalid cronjob options, please set cron.")
	}
	if options.Name == "" {
		panic("invalid task options, please set name.")
	}

	if options.RouterStrategy == 0 {
		options.RouterStrategy = bean.RANDOM
	}
	if options.ExpiredStrategy == 0 {
		options.ExpiredStrategy = bean.ExpiredExecute
	}
	if options.ExpireTime == 0 {
		options.ExpireTime = 3 * 60 * 1000
	}
	if options.FailureStrategy == 0 {
		options.FailureStrategy = bean.FailureRetry
	}
	if options.MaxRetryCount == 0 {
		options.MaxRetryCount = 5
	}
	if options.FailureRetryInterval == 0 {
		options.FailureRetryInterval = 5000
	}
	if options.Timeout == 0 {
		options.Timeout = 10000
	}

	// 生成唯一key（包路径+方法名）
	key := client.options.AppName + "/" + reflect.TypeOf(handler).String() + ".Handle"

	// 缓存Handle方法的反射值
	handleMethod := reflect.ValueOf(handler).MethodByName("Handle")
	if !handleMethod.IsValid() {
		panic("Handle method can not be found")
	}

	client.handlers[key] = &handleMethod
	client.taskOptions[key] = &options
}

// Stop 停止
func (client *executorClientImpl) stop() {
	context.Shutdown.Store(true)

	// 向调度器发送下线的请求
	services.GetRegisterService().Unregister(webserver.GetHttpServer().GetAddress())

	// 等待调度器处理完成，等待结果发送完成
	context.WaitGroup.Wait()
}

// Start 启动
func (client *executorClientImpl) Start() {
	// 注册关闭钩子
	utils.SignalNotify(client.stop)

	// 初始化HttpClient
	httpclients.Init(httpclients.Options{
		Timeout: time.Second * 5,
		SignKey: client.options.SignKey,
	})

	// 启动Http服务
	httpServer := webserver.GetHttpServer()
	go httpServer.Start()

	// 等待HttpServer启动成功
	for !httpServer.IsStarted() {
		time.Sleep(time.Millisecond * 200)
	}

	// 每30秒注册执行器和任务
	utils.GetScheduler().ScheduleAtFixedRate(time.Second*30, true, func() {
		if context.Shutdown.Load() {
			return
		}

		// 注册执行器和任务
		services.GetRegisterService().RegisterExecutor(client.options, httpServer.GetAddress())

		time.Sleep(time.Second)

		// 注册任务
		services.GetRegisterService().RegisterTask(client.options, client.taskOptions)
	})

	// 开始心跳，如果执行器未注册成功，则不会开始心跳
	services.GetHeartbeatService().Start(httpServer.GetAddress())

	// 开始调度
	dispatcherService := services.InitDispatcherService(client.handlers, client.taskOptions)
	logger.Infof("start cron-job executor success, options: %s", utils.ToJsonString(client.options))
	dispatcherService.Start(httpServer.GetAddress())

	// 保持主goroutine存活
	select {}
}

// checkOptions 检查参数
func checkOptions(options *bean.ExecutorOptions) {
	if options.Address == "" || options.SignKey == "" || options.Tenant == "" || options.AppName == "" || options.AppDesc == "" {
		panic("invalid cronjob bean, please set parameters.")
	}
}

// GetExecutorClient 获取实例对象，需要先调用init方法
func GetExecutorClient(option *bean.ExecutorOptions) ExecutorClient {
	executorClientOnce.Do(func() {
		// 检查参数
		checkOptions(option)
		if option.Tag == "" {
			option.Tag = "common"
		}
		services.GetOpenApiService().SetHost(option.Address)
		executorClient = &executorClientImpl{
			options:     *option,
			handlers:    make(map[string]*reflect.Value),
			taskOptions: make(map[string]*bean.TaskOptions),
		}
		context.SignKey.Store(option.SignKey)
	})
	return executorClient
}
