package services

// Created in 2025-03-23 17:02.
// @author Horace

import (
	"bytes"
	"encoding/json"
	logger "github.com/cihub/seelog"
	"github.com/horacedh/cronjob-executor/bean"
	"github.com/horacedh/cronjob-executor/context"
	"github.com/horacedh/cronjob-executor/httpclients"
	"github.com/horacedh/cronjob-executor/task"
	"github.com/horacedh/cronjob-executor/utils"
	"sync"
)

var apiExecutorRegister = "/openapi/executor/register"
var apiExecutorUnregister = "/openapi/executor/unregister"
var apiExecutorHeartbeat = "/openapi/executor/heartbeat"
var apiTaskRegister = "/openapi/task/register"
var apiTaskExecuteComplete = "/openapi/task/complete"

// 单例模式
var (
	openApiService     OpenApiService
	openApiServiceOnce sync.Once
)

// OpenApiService 接口
type OpenApiService interface {
	// RegisterExecutor 注册执行器
	RegisterExecutor(params bean.ExecutorRegisterParams) bool
	// SetHost 设置主机地址
	SetHost(address string)
	// RegisterTask 注册任务
	RegisterTask(params []bean.TaskRegisterParams) bool
	// Heartbeat 心跳
	Heartbeat(address string) bool
	// UnregisterExecutor 注销执行器
	UnregisterExecutor(address string) bool
	// SendTaskResult 发送任务结果
	SendTaskResult(result *task.TaskResult) bool
}

// openApiServiceImpl 实现类
type openApiServiceImpl struct {
	host string
}

// SendTaskResult 发送任务结果
func (openApiService *openApiServiceImpl) SendTaskResult(params *task.TaskResult) bool {
	var commonHeaders = openApiService.getCommonHeaders()

	var url = openApiService.host + apiTaskExecuteComplete
	jsonParams, _ := json.Marshal(params)
	result := httpclients.GetHttpClient().PostRequest(url, commonHeaders, nil, bytes.NewReader(jsonParams))
	success := result.IsSuccess()
	if success {
		logger.Debugf("cron job send task result success, serverAddress:%s, params:%v", openApiService.host, params)
	} else {
		logger.Errorf("cron job send task result failed, serverAddress:%s, result:%v, params:%v", openApiService.host, result.MsgObject, params)
	}
	return success
}

// UnregisterExecutor 注销执行器
func (openApiService *openApiServiceImpl) UnregisterExecutor(address string) bool {
	var commonHeaders = openApiService.getCommonHeaders()
	var params = make(map[string]string)
	params["address"] = address

	var url = openApiService.host + apiExecutorUnregister
	jsonParams, _ := json.Marshal(params)
	result := httpclients.GetHttpClient().PostRequest(url, commonHeaders, nil, bytes.NewReader(jsonParams))
	success := result.IsSuccess()
	if success {
		logger.Infof("cron job unregister success, serverAddress:%s, params:%v", openApiService.host, params)
	} else {
		logger.Errorf("cron job unregister failed, serverAddress:%s, result:%v, params:%v", openApiService.host, result.MsgObject, params)
	}
	return success
}

// Heartbeat 心跳
func (openApiService *openApiServiceImpl) Heartbeat(address string) bool {
	var commonHeaders = openApiService.getCommonHeaders()
	var params = make(map[string]string)
	params["address"] = address

	var url = openApiService.host + apiExecutorHeartbeat
	jsonParams, _ := json.Marshal(params)
	result := httpclients.GetHttpClient().PostRequest(url, commonHeaders, nil, bytes.NewReader(jsonParams))
	success := result.IsSuccess()
	if !success {
		logger.Errorf("cron job heartbeat failed, serverAddress:%s, result:%v, params:%v", openApiService.host, result.MsgObject, params)
	}
	return success
}

// RegisterTask 注册任务
func (openApiService *openApiServiceImpl) RegisterTask(params []bean.TaskRegisterParams) bool {
	var commonHeaders = openApiService.getCommonHeaders()

	var url = openApiService.host + apiTaskRegister
	jsonParams, _ := json.Marshal(params)
	result := httpclients.GetHttpClient().PostRequest(url, commonHeaders, nil, bytes.NewReader(jsonParams))
	success := result.IsSuccess()
	if success {
		logger.Debugf("cron job task register success, serverAddress:%s, params:%v", openApiService.host, utils.ToJsonString(params))
	} else {
		logger.Errorf("cron job task register failed, serverAddress:%s, result:%v, params:%v", openApiService.host, result.MsgObject, utils.ToJsonString(params))
	}
	return success
}

// SetHost 设置主机地址
func (openApiService *openApiServiceImpl) SetHost(address string) {
	openApiService.host = address
}

// RegisterExecutor 注册执行器
func (openApiService *openApiServiceImpl) RegisterExecutor(params bean.ExecutorRegisterParams) bool {
	var commonHeaders = openApiService.getCommonHeaders()

	var url = openApiService.host + apiExecutorRegister
	jsonParams, _ := json.Marshal(params)
	result := httpclients.GetHttpClient().PostRequest(url, commonHeaders, nil, bytes.NewReader(jsonParams))
	success := result.IsSuccess()
	if success {
		logger.Debugf("cron job executor register success, serverAddress:%s, params:%v", openApiService.host, utils.ToJsonString(params))
	} else {
		logger.Errorf("cron job executor register failed, serverAddress:%s, result:%v, params:%v", openApiService.host, result.MsgObject, utils.ToJsonString(params))
	}
	return success
}

// getCommonHeaders 获取公共请求头
func (openApiService *openApiServiceImpl) getCommonHeaders() map[string]interface{} {
	headers := make(map[string]interface{})
	headers["User-Agent"] = "CronJob-Go-SDK"
	headers["Content-Type"] = "application/json; charset=utf-8"
	headers["SDK-Version"] = context.Version
	return headers
}

// GetOpenApiService 获取实例对象
func GetOpenApiService() OpenApiService {
	openApiServiceOnce.Do(func() {
		openApiService = &openApiServiceImpl{}
	})
	return openApiService
}
