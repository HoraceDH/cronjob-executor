package bean

// Created in 2025-03-23 16:11.
// @author Horace
// Options 配置选项

type ExecutorOptions struct {
	// address 调度平台地址，例如：http://127.0.0.1:9527
	Address string
	// tenant 租户编码
	Tenant string
	// appName 应用名，一般英文代号
	AppName string
	// appDesc 应用描述，应用中文名称或者对应用的描述
	AppDesc string
	// tag 执行器标签
	Tag string
	// signKey 签名Key
	SignKey string
}

// RouterStrategy 路由策略枚举定义
type RouterStrategy int32

const (
	// RANDOM 随机策略，适合绝大多数场景
	RANDOM RouterStrategy = 1
	// SHARDING 分片策略，适合大任务分片拆分执行场景
	SHARDING RouterStrategy = 2
)

// ExpiredStrategy 过期策略枚举定义
type ExpiredStrategy int32

const (
	// ExpiredDiscard 过期的任务直接丢弃，适合密集型场景
	ExpiredDiscard ExpiredStrategy = 1
	// ExpiredExecute 过期的任务依然调度，适合松散型场景
	ExpiredExecute ExpiredStrategy = 2
)

// FailureStrategy 失败策略枚举定义
type FailureStrategy int

const (
	// FailureRetry 失败重试
	FailureRetry FailureStrategy = 1
	// FailureDiscard 失败丢弃
	FailureDiscard FailureStrategy = 2
)

// TaskOptions 任务配置
type TaskOptions struct {
	// Name 任务名称
	Name string
	// Cron CRON表达式
	Cron string
	// RouterStrategy 路由策略，默认随机策略
	RouterStrategy RouterStrategy
	// ExpiredStrategy 过期策略，默认过期依然调度
	ExpiredStrategy ExpiredStrategy
	// ExpireTime 过期时间，毫秒，默认3分钟过期，任务超过过期时间调度时，则走过期策略，最大过期时间5分钟，3 * 60 * 1000
	ExpireTime int
	// FailureStrategy 失败策略，默认失败重试
	FailureStrategy FailureStrategy
	// MaxRetryCount 失败最大重试次数
	MaxRetryCount int
	// FailureRetryInterval 失败重试间隔时间，毫秒
	FailureRetryInterval int
	// Timeout  任务超时时间，超过此时间没有反馈执行结果给调度器，则认为执行器执行失败，调度器按照策略进行重试，单位毫秒，最大10秒钟，如果是消耗大量时间的任务，建立使用独立线程池运行
	Timeout int
	// Remark 任务备注，主要是用来描述任务详情，用来做什么样的任务？方便后期维护和管理
	Remark string
}

// ExecutorRegisterParams 执行器注册参数
type ExecutorRegisterParams struct {
	// tenant 租户编码
	Tenant string `json:"tenant"`
	// appName 应用名，一般英文代号
	AppName string `json:"appName"`
	// AppDesc 应用描述，应用中文名称或者对应用的描述
	AppDesc string `json:"appDesc"`
	// HostName 主机名
	HostName string `json:"hostName"`
	// Tag 执行器标签
	Tag string `json:"tag"`
	// Version 执行器SDK版本
	Version string `json:"version"`
	// Address 执行器地址，host:port
	Address string `json:"address"`
}

// TaskRegisterParams 任务注册参数
type TaskRegisterParams struct {
	AppDesc              string `json:"appDesc"`
	AppName              string `json:"appName"`
	Cron                 string `json:"cron"`
	ExpiredStrategy      int    `json:"expiredStrategy"`
	ExpiredTime          int    `json:"expiredTime"`
	FailureRetryInterval int    `json:"failureRetryInterval"`
	FailureStrategy      int    `json:"failureStrategy"`
	MaxRetryCount        int    `json:"maxRetryCount"`
	Method               string `json:"method"`
	Name                 string `json:"name"`
	Remark               string `json:"remark"`
	RouterStrategy       int    `json:"routerStrategy"`
	Tag                  string `json:"tag"`
	Tenant               string `json:"tenant"`
	Timeout              int    `json:"timeout"`
}
