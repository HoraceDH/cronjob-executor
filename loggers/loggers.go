package loggers

import (
	"github.com/cihub/seelog"
	"path/filepath"
)

// 当前项目或者应用运行目录
var logFileName = "logger.xml"

func getCurrentPath() string {
	// 初始化当前项目路径
	path, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		seelog.Error("init current path error", err)
		return "/"
	}
	return path + "/"
}

// init 初始化函数
func init() {
	var path = getCurrentPath()
	initLogger(path)
}

// initLogger 初始化日志框架
func initLogger(path string) {
	var logFile = path + "etc/" + logFileName
	log, err := seelog.LoggerFromConfigAsFile(logFile)
	if err != nil {
		seelog.Infof("init logger use default string, load config file error: %v, logFile:%s", err, logFile)
		log, err = seelog.LoggerFromConfigAsString(getConfigString())
	} else {
		seelog.Infof("init logger use config file, logFile:%s", logFile)
	}

	_ = seelog.ReplaceLogger(log)
	defer seelog.Flush()
}

// getConfigString 获取默认配置
func getConfigString() string {
	var config = `<!-- type：sync, asynctimer, asyncloop，asyncinterval：单位纳秒，这里指定10毫秒-->
<seelog type="adaptive" mininterval="10000000" maxinterval="100000000" critmsgcount="5">
    <outputs formatid="configId">
        <!--<console/>-->

        <!--info以上日志-->
        <filter levels="info,error,warn,error,critical">
            <rollingfile filename="logs/app.log" type="date" namemode="postfix" datepattern="2006-01-02_15" maxrolls="168"/>
        </filter>

        <!--错误日志-->
        <filter levels="error,critical">
            <rollingfile filename="logs/error.log" type="date" namemode="postfix" datepattern="2006-01-02_15" maxrolls="168"/>
        </filter>
    </outputs>
    <formats>
		<format id="configId" format="%Date(2006-01-02 15:04:05.000) [%LEVEL] %File:%Line - %Msg%n"/>
    </formats>
</seelog>
`
	return config
}
