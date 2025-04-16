package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	logger "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/horacedh/cronjob-executor/webresult"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Created in 2025-03-19 08:51.
// @author Horace

// SignalNotify /*进程正常结束无法捕获, shutdown 进程结束是执行的方法，object 方法的参数，Ctrl + C 或者 kill pid 回调*/
func SignalNotify(shutdown func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		s := <-c // terminated or interrupt
		logger.Info("received signal is ", s)
		// 执行关闭方法
		shutdown()
		logger.Info("cron-job executor shutdown.")
		logger.Flush()
		// 确保最终退出
		os.Exit(0)
	}()
}

// UniDecode 解码成中文
func UniDecode(data string) string {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(data), `\\u`, `\u`, -1))
	if err != nil {
		return data
	}
	return str
}

// GetLocalIP 获取本地IP地址
func GetLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		// 排除回环接口和无效接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			if ip := ipNet.IP.To4(); ip != nil {
				return ip.String()
			}
		}
	}

	// 默认返回本地地址
	return "127.0.0.1"
}

// ToJsonString 转换成json字符串
func ToJsonString(data interface{}) string {
	jsonString, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("marshal json failed, data: %v, err: %v", data, err)
		return ""
	}
	return string(jsonString)
}

// FormatTime 格式化时间
func FormatTime(times int64) string {
	t := time.Unix(0, times*int64(time.Millisecond))
	return t.Format("2006-01-02 15:04:05")
}

// RenderMsgObject 渲染消息对象
func RenderMsgObject(context *gin.Context, msgObject webresult.MsgObject) {
	context.Header("Content-Type", "application/json; charset=utf-8")
	context.Render(200, render.JSON{
		Data: msgObject,
	})
}

// Sign 参数签名
func Sign(signKey string, token string, times string, body string, params map[string]interface{}) string {
	// 创建有序参数集合
	paramsMap := make(map[string]interface{})
	paramsMap["_signKey_"] = signKey
	paramsMap["times"] = times
	paramsMap["token"] = token
	paramsMap["rb"] = body

	// 合并额外参数
	for k, v := range params {
		paramsMap[k] = v
	}

	// 获取排序后的键
	keys := make([]string, 0, len(paramsMap))
	for k := range paramsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建参数字符串
	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", paramsMap[key]))
		sb.WriteString("&")
	}

	// 生成MD5签名
	hash := md5.Sum([]byte(sb.String()))
	return fmt.Sprintf("%x", hash)
}
