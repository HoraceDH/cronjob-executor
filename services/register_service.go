package services

// Created in 2025-03-18 20:06.
// @author Horace

import (
	logger "github.com/cihub/seelog"
	"github.com/horacedh/cronjob-executor/bean"
	"github.com/horacedh/cronjob-executor/context"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// 单例模式
var (
	registerService     RegisterService
	registerServiceOnce sync.Once
)

// RegisterService 接口
type RegisterService interface {
	// RegisterExecutor 注册执行器
	RegisterExecutor(option bean.ExecutorOptions, address string)
	// RegisterTask 注册任务
	RegisterTask(options bean.ExecutorOptions, option map[string]*bean.TaskOptions)
	// IsSuccess 是否注册成功
	IsSuccess() bool
	// Unregister 注销执行器
	Unregister(address string) bool
}

// registerServiceImpl 实现类
type registerServiceImpl struct {
	// success 标志位，用于判断是否已经成功注册
	success atomic.Bool
}

// Unregister 注销执行器
func (registerService *registerServiceImpl) Unregister(address string) bool {
	registerService.success.Store(false)
	return GetOpenApiService().UnregisterExecutor(address)
}

// IsSuccess 是否注册成功
func (registerService *registerServiceImpl) IsSuccess() bool {
	return registerService.success.Load()
}

// RegisterExecutor 注册执行器
func (registerService *registerServiceImpl) RegisterExecutor(options bean.ExecutorOptions, address string) {
	registerParams := registerService.buildExecutorRegisterParams(options, address)
	success := GetOpenApiService().RegisterExecutor(registerParams)
	registerService.success.Store(success)

	// 如果已经停止，则不再重试
	if context.Shutdown.Load() {
		logger.Warnf("cronjob executor is shutdown, don't retry register executor, params:%v", registerParams)
		return
	}

	if !success {
		// 如果注册不成功，则重试一次
		time.Sleep(time.Second)
		registerService.RegisterExecutor(options, address)
	}
}

// RegisterTask 注册任务
func (registerService *registerServiceImpl) RegisterTask(executorOptions bean.ExecutorOptions, taskOptions map[string]*bean.TaskOptions) {
	var registerParams = registerService.buildTaskRegisterParams(executorOptions, taskOptions)
	success := GetOpenApiService().RegisterTask(registerParams)

	// 如果已经停止，则不再重试
	if context.Shutdown.Load() {
		logger.Warnf("cronjob executor is shutdown, don't retry register task, params:%v", registerParams)
		return
	}

	if !success {
		// 如果注册不成功，则重试一次
		time.Sleep(time.Second)
		registerService.RegisterTask(executorOptions, taskOptions)
	}
}

// buildExecutorRegisterParams 构建注册执行器的参数
func (registerService *registerServiceImpl) buildExecutorRegisterParams(options bean.ExecutorOptions, address string) bean.ExecutorRegisterParams {
	hostName, _ := os.Hostname()
	return bean.ExecutorRegisterParams{
		Tenant:   options.Tenant,
		AppName:  options.AppName,
		AppDesc:  options.AppDesc,
		HostName: hostName,
		Tag:      options.Tag,
		Address:  address,
		Version:  context.Version,
	}
}

// buildTaskRegisterParams 构建注册任务的参数
func (registerService *registerServiceImpl) buildTaskRegisterParams(options bean.ExecutorOptions, taskOptions map[string]*bean.TaskOptions) []bean.TaskRegisterParams {
	params := make([]bean.TaskRegisterParams, 0)
	for key, taskOption := range taskOptions {
		params = append(params, bean.TaskRegisterParams{
			AppDesc:              options.AppDesc,
			AppName:              options.AppName,
			Cron:                 taskOption.Cron,
			ExpiredStrategy:      int(taskOption.ExpiredStrategy),
			ExpiredTime:          taskOption.ExpireTime,
			FailureRetryInterval: taskOption.FailureRetryInterval,
			FailureStrategy:      int(taskOption.FailureStrategy),
			MaxRetryCount:        taskOption.MaxRetryCount,
			Method:               key,
			Name:                 taskOption.Name,
			Remark:               taskOption.Remark,
			RouterStrategy:       int(taskOption.RouterStrategy),
			Tag:                  options.Tag,
			Tenant:               options.Tenant,
			Timeout:              taskOption.Timeout,
		})
	}
	return params
}

// GetRegisterService 获取实例对象
func GetRegisterService() RegisterService {
	registerServiceOnce.Do(func() {
		registerService = &registerServiceImpl{}
	})
	return registerService
}
