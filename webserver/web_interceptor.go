package webserver

import (
	logger "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/horacedh/cronjob-executor/utils"
	"github.com/horacedh/cronjob-executor/webresult"
	"net/http"
	"runtime/debug"
	"time"
)

// loggerInterceptor 日志打印中间件
func loggerInterceptor() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()
		request := context.Request

		// 计算执行时间，确保发生异常时也有日志打印
		defer func() {
			useTime := time.Since(startTime).Milliseconds()
			if useTime > 200 {
				logger.Warnf("request url, use: %dms, method: %s, url: %s, clientIp: %s, headers: %s", useTime, request.Method, request.RequestURI, request.RemoteAddr, request.Header)
			} else {
				logger.Debugf("request url, use: %dms, method: %s, url: %s, clientIp: %s, headers: %s", useTime, request.Method, request.RequestURI, request.RemoteAddr, request.Header)
			}
		}()

		// 调用下一个中间件
		context.Next()
	}
}

// globalErrorHandler 全局错误处理器
func globalErrorHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			status := context.Writer.Status()
			if err := recover(); err != nil {
				request := context.Request
				logger.Errorf("global error handler, err: %v, method: %s, url: %s, clientIp: %s, headers: %s, stack: %v", err, request.Method, request.RequestURI, request.RemoteAddr, request.Header, string(debug.Stack()))
				utils.RenderMsgObject(context, webresult.ERROR)
				return
			}

			if status != http.StatusOK {
				context.Writer.WriteHeader(http.StatusOK)
				utils.RenderMsgObject(context, webresult.MsgObject{
					Code: status,
					Msg:  "请求失败",
				})
			}
		}()

		// 调用下一个中间件
		context.Next()
	}
}
