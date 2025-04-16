package webserver

// Created in 2025-03-19 20:17.
// @author Horace

import (
	"encoding/json"
	logger "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	cronjobContext "github.com/horacedh/cronjob-executor/context"
	"github.com/horacedh/cronjob-executor/services"
	"github.com/horacedh/cronjob-executor/task"
	"github.com/horacedh/cronjob-executor/utils"
	"github.com/horacedh/cronjob-executor/webresult"
	"io"
	"sync"
	"time"
)

// 单例模式
var (
	executorController     ExecutorController
	executorControllerOnce sync.Once
)

// ExecutorController 接口
type ExecutorController interface {
	// Dispatcher 任务分发接口
	Dispatcher() gin.HandlerFunc
}

// ExecutorControllerImpl 实现类
type ExecutorControllerImpl struct {
}

// Dispatcher 任务分发接口
func (controller ExecutorControllerImpl) Dispatcher() gin.HandlerFunc {
	return func(context *gin.Context) {
		bytes, err := io.ReadAll(context.Request.Body)
		if err != nil {
			logger.Errorf("received execute request, read request body failed, err: %v", err)
			utils.RenderMsgObject(context, webresult.ERROR)
			return
		}

		// 校验签名
		var body = string(bytes)
		if !controller.verifySign(context, body) {
			utils.RenderMsgObject(context, webresult.ERROR_SIGN)
			return
		}

		// 如果已经停机
		if cronjobContext.Shutdown.Load() {
			logger.Warnf("received execute request, executor is not running, ignore the task, params:%s", body)
			utils.RenderMsgObject(context, webresult.ERROR_EXECUTE_SHUTDOWN)
			return
		}

		// 加入队列
		var taskParams = task.TaskParams{}
		err = json.Unmarshal(bytes, &taskParams)
		if err != nil {
			logger.Errorf("received execute request, json unmarshal failed, params:%s, err: %v", body, err)
			utils.RenderMsgObject(context, webresult.ERROR)
			return
		}

		taskParams.ReceivedDispatcherTime = time.Now().UnixMilli()
		var queueSize = services.GetDispatcherService().AddTask(&taskParams)
		logger.Debugf("received execute request, queueSize:%d, params:%v", queueSize, utils.ToJsonString(taskParams))
		utils.RenderMsgObject(context, webresult.SUCCESS)
	}
}

// verifySign 校验签名
func (controller ExecutorControllerImpl) verifySign(context *gin.Context, body string) bool {
	var signKey = cronjobContext.SignKey.Load()
	var sign = context.GetHeader("sign")
	var token = context.GetHeader("token")
	var times = context.GetHeader("times")
	var params = make(map[string]interface{})
	serverSign := utils.Sign(signKey.(string), token, times, body, params)
	var success = serverSign == sign
	if !success {
		logger.Errorf("received execute request, sign verify failed, signKey:%s, serverSign:%s, sign:%s, token:%s, times:%s, params:%s", signKey, serverSign, sign, token, times, body)
	}
	return success
}

// GetExecutorController 获取实例对象
func GetExecutorController() ExecutorController {
	executorControllerOnce.Do(func() {
		executorController = &ExecutorControllerImpl{}
	})
	return executorController
}
