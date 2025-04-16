package services

// Created in 2025-03-23 16:22.
// @author Horace

import (
	"github.com/horacedh/cronjob-executor/context"
	"github.com/horacedh/cronjob-executor/utils"
	"sync"
	"time"
)

// 单例模式
var (
	heartbeatService     HeartbeatService
	heartbeatServiceOnce sync.Once
)

// HeartbeatService 接口
type HeartbeatService interface {
	// Start 启动心跳
	Start(address string)
}

// heartbeatServiceImpl 实现类
type heartbeatServiceImpl struct {
}

// Start 启动心跳
func (heartbeatService *heartbeatServiceImpl) Start(address string) {
	// 每3秒一次注册
	utils.GetScheduler().ScheduleAtFixedRate(time.Second*3, true, func() {
		// 如果已经停止，则不再发起心跳
		if context.Shutdown.Load() {
			//logger.Warnf("cronjob executor is shutdown, don't heartbeat, address:%s", address)
			return
		}
		if !GetRegisterService().IsSuccess() {
			time.Sleep(time.Second)
			return
		}
		GetOpenApiService().Heartbeat(address)
	})
}

// GetHeartbeatService 获取实例对象
func GetHeartbeatService() HeartbeatService {
	heartbeatServiceOnce.Do(func() {
		heartbeatService = &heartbeatServiceImpl{}
	})
	return heartbeatService
}
