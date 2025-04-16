package webserver

// Created in 2025-03-19 20:01.
// @author Horace

import (
	"fmt"
	logger "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/horacedh/cronjob-executor/utils"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

// 单例模式
var (
	webServer      WebServer
	httpServerOnce sync.Once
)

// WebServer 接口
type WebServer interface {
	// Start 启动服务器
	Start()
	// IsStarted 是否已经启动
	IsStarted() bool
	// GetAddress 获取地址，ip:host 格式
	GetAddress() string
}

// webServerImpl 实现类
type webServerImpl struct {
	// 初始端口，失败后递增
	Port int32
	// 启动状态
	started atomic.Bool
}

// GetAddress 获取地址，ip:host 格式
func (webServer *webServerImpl) GetAddress() string {
	return fmt.Sprintf("%s:%d", utils.GetLocalIP(), webServer.Port)
}

// IsStarted 是否已经启动
func (webServer *webServerImpl) IsStarted() bool {
	return webServer.started.Load()
}

func (webServer *webServerImpl) Start() {
	gin.SetMode(gin.ReleaseMode)

	// 获取一个默认的WEB引擎
	engine := gin.New()
	engine.With(func(engine *gin.Engine) {
		engine.Use(loggerInterceptor(), globalErrorHandler())

		// 初始化路由
		initRouter(engine)
	})

	// 一直尝试绑定端口，直到成功
	var listen net.Listener
	for {
		listenTemp, err := net.Listen("tcp4", fmt.Sprintf(":%d", webServer.Port))
		if err != nil {
			logger.Warnf("start cron-job web server failed, incremental Port retry, Port: %d, error: %s", webServer.Port, err.Error())
			logger.Flush()

			// 如果端口已经绑定，则递增端口并重新启动
			if strings.Contains(err.Error(), "address already in use") {
				webServer.Port++
			}
			continue
		}
		listen = listenTemp
		break
	}

	logger.Infof("start cron-job web server, at Port: %d", webServer.Port)
	webServer.started.Store(true)
	err := engine.RunListener(listen)
	if err != nil {
		logger.Errorf("start cron-job web server error, Port: %d, error: %s", webServer.Port, err.Error())
	}
	logger.Flush()
}

// initRouter 初始化路由
func initRouter(engine *gin.Engine) {
	// 任务分发接口
	executorController := GetExecutorController()
	engine.POST("/dispatch", executorController.Dispatcher())
}

// GetHttpServer 获取实例对象
func GetHttpServer() WebServer {
	httpServerOnce.Do(func() {
		webServer = &webServerImpl{
			Port:    8527,
			started: atomic.Bool{},
		}
	})
	return webServer
}
